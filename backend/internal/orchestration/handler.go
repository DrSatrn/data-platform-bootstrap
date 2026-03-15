// This file exposes the HTTP surface for pipeline listing and sample run
// inspection. The handler currently reads manifests directly so the first slice
// is immediately useful even before persistent orchestration storage is added.
package orchestration

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/authz"
	"github.com/streanor/data-platform/backend/internal/shared"
)

// PipelineLoader defines the manifest-loading behavior the orchestration API
// depends on. Declaring the interface locally prevents an import cycle between
// the orchestration domain and the manifest adapter.
type PipelineLoader interface {
	LoadPipelines() ([]Pipeline, error)
}

// PipelineHandler serves the orchestration-focused API routes.
type PipelineHandler struct {
	loader  PipelineLoader
	store   Store
	control *ControlService
	logger  *slog.Logger
	authz   *authz.Service
	audit   audit.Store
}

// NewPipelineHandler constructs the pipeline API surface.
func NewPipelineHandler(loader PipelineLoader, store Store, control *ControlService, logger *slog.Logger, authService *authz.Service, auditStore audit.Store) http.Handler {
	return &PipelineHandler{
		loader:  loader,
		store:   store,
		control: control,
		logger:  logger,
		authz:   authService,
		audit:   auditStore,
	}
}

func (h *PipelineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.handleTrigger(w, r)
		return
	}

	pipelines, err := h.loader.LoadPipelines()
	if err != nil {
		h.logger.Error("failed to load pipelines", slog.String("error", err.Error()))
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load pipelines",
		})
		return
	}

	runs, err := h.store.ListPipelineRuns()
	if err != nil {
		h.logger.Error("failed to load pipeline runs", slog.String("error", err.Error()))
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load pipeline runs",
		})
		return
	}

	validationErrors := make(map[string]string)
	for _, pipeline := range pipelines {
		if err := ValidatePipeline(pipeline); err != nil {
			validationErrors[pipeline.ID] = err.Error()
		}
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"pipelines":         pipelines,
		"runs":              runs,
		"validation_errors": validationErrors,
	})
}

func (h *PipelineHandler) handleTrigger(w http.ResponseWriter, r *http.Request) {
	principal := h.authz.ResolveRequest(r)
	if !authz.Allowed(principal, authz.RoleEditor) {
		_ = h.audit.Append(audit.Event{
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "trigger_pipeline",
			Resource:     "unknown",
			Outcome:      "forbidden",
		})
		shared.WriteJSON(w, http.StatusForbidden, map[string]any{
			"error": "editor role required to trigger pipelines",
		})
		return
	}

	var payload struct {
		PipelineID string `json:"pipeline_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid trigger payload",
		})
		return
	}

	run, err := h.control.TriggerPipeline(context.Background(), payload.PipelineID, "manual_api")
	if err != nil {
		_ = h.audit.Append(audit.Event{
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "trigger_pipeline",
			Resource:     payload.PipelineID,
			Outcome:      "failure",
			Details: map[string]any{
				"error": err.Error(),
			},
		})
		h.logger.Error("failed to trigger pipeline", slog.String("error", err.Error()))
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}
	_ = h.audit.Append(audit.Event{
		ActorSubject: principal.Subject,
		ActorRole:    string(principal.Role),
		Action:       "trigger_pipeline",
		Resource:     payload.PipelineID,
		Outcome:      "success",
		Details: map[string]any{
			"run_id": run.ID,
		},
	})

	shared.WriteJSON(w, http.StatusAccepted, map[string]any{
		"run": run,
	})
}
