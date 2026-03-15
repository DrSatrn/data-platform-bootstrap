// This file exposes the admin terminal command endpoint used by both the web
// portal and the local CLI. Requests are token-gated when an admin token is
// configured.
package admin

import (
	"encoding/json"
	"net/http"

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
}

// NewHandler constructs an admin-terminal handler.
func NewHandler(cfg config.Settings, authService *authz.Service, service *Service) http.Handler {
	return &Handler{
		cfg:     cfg,
		authz:   authService,
		service: service,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
		return
	}

	if !authz.Allowed(h.authz.ResolveRequest(r), authz.RoleAdmin) {
		shared.WriteJSON(w, http.StatusForbidden, map[string]any{
			"error": "admin role required for terminal access",
		})
		return
	}

	var payload executeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid admin command payload",
		})
		return
	}

	result := h.service.Execute(payload.Command)
	status := http.StatusOK
	if !result.Success {
		status = http.StatusBadRequest
	}

	shared.WriteJSON(w, status, result)
}
