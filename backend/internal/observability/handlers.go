// This file exposes first-party diagnostics APIs for system overview and log
// inspection. These endpoints back the built-in admin portal rather than
// relying on external metrics dashboards.
package observability

import (
	"net/http"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/shared"
)

// PipelineLoader describes the pipeline loading dependency needed for system
// overview responses.
type PipelineLoader interface {
	LoadPipelines() ([]orchestration.Pipeline, error)
}

// AssetLoader describes the asset loading dependency needed for system
// overview responses.
type AssetLoader interface {
	LoadAssets() ([]metadata.DataAsset, error)
}

// OverviewHandler returns a system summary payload.
type OverviewHandler struct {
	cfg        config.Settings
	service    *Service
	pipelines  PipelineLoader
	assets     AssetLoader
	runHistory orchestration.Store
}

// NewOverviewHandler constructs the system overview handler.
func NewOverviewHandler(
	cfg config.Settings,
	service *Service,
	pipelines PipelineLoader,
	assets AssetLoader,
	runHistory orchestration.Store,
) http.Handler {
	return &OverviewHandler{
		cfg:        cfg,
		service:    service,
		pipelines:  pipelines,
		assets:     assets,
		runHistory: runHistory,
	}
}

func (h *OverviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pipelines, err := h.pipelines.LoadPipelines()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load pipelines for system overview",
		})
		return
	}

	assets, err := h.assets.LoadAssets()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load assets for system overview",
		})
		return
	}

	runHistory, err := h.runHistory.ListPipelineRuns()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load run history for system overview",
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"environment":     h.cfg.Environment,
		"http_addr":       h.cfg.HTTPAddr,
		"web_addr":        h.cfg.WebAddr,
		"known_pipelines": len(pipelines),
		"known_assets":    len(assets),
		"run_history":     len(runHistory),
		"telemetry": h.service.Snapshot(map[string]string{
			"environment": h.cfg.Environment,
			"api_base":    h.cfg.APIBaseURL,
		}),
	})
}

// RecentLogsHandler returns the in-memory log buffer.
type RecentLogsHandler struct {
	service *Service
}

// NewRecentLogsHandler constructs the recent log API.
func NewRecentLogsHandler(service *Service) http.Handler {
	return &RecentLogsHandler{service: service}
}

func (h *RecentLogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"logs": h.service.RecentLogs(),
	})
}
