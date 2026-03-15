// This file exposes the admin terminal command endpoint used by both the web
// portal and the local CLI. Requests are token-gated when an admin token is
// configured.
package admin

import (
	"encoding/json"
	"net/http"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/authz"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/shared"
)

type executeRequest struct {
	Command string `json:"command"`
}

// Handler serves the admin terminal API.
type Handler struct {
	cfg     config.Settings
	authz   *authz.Service
	service *Service
	audit   audit.Store
}

// NewHandler constructs an admin-terminal handler.
func NewHandler(cfg config.Settings, authService *authz.Service, service *Service, auditStore audit.Store) http.Handler {
	return &Handler{
		cfg:     cfg,
		authz:   authService,
		service: service,
		audit:   auditStore,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
		return
	}

	principal := h.authz.ResolveRequest(r)
	if !authz.Allowed(principal, authz.RoleAdmin) {
		_ = h.audit.Append(audit.Event{
			ActorUserID:  principal.UserID,
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "admin_command",
			Resource:     "terminal",
			Outcome:      "forbidden",
		})
		shared.WriteRoleError(w, string(authz.RoleAdmin), string(principal.Role))
		return
	}

	var payload executeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid admin command payload",
		})
		return
	}

	result := h.service.Execute(payload.Command, principal)
	status := http.StatusOK
	if !result.Success {
		status = http.StatusBadRequest
	}

	_ = h.audit.Append(audit.Event{
		ActorUserID:  principal.UserID,
		ActorSubject: principal.Subject,
		ActorRole:    string(principal.Role),
		Action:       "admin_command",
		Resource:     payload.Command,
		Outcome:      map[bool]string{true: "success", false: "failure"}[result.Success],
		Details: map[string]any{
			"output_preview": firstOutput(result.Output),
		},
	})

	shared.WriteJSON(w, status, result)
}

func firstOutput(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}
