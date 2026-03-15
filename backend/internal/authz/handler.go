// This file exposes a small session endpoint so browser clients can understand
// what the current bearer token is allowed to do.
package authz

import (
	"net/http"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// SessionHandler serves the current resolved session.
type SessionHandler struct {
	service *Service
}

// NewSessionHandler constructs a session endpoint.
func NewSessionHandler(service *Service) http.Handler {
	return &SessionHandler{service: service}
}

func (h *SessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
		return
	}
	shared.WriteJSON(w, http.StatusOK, h.service.SessionForRequest(r))
}
