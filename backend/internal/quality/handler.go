// This file exposes quality status for the API and UI. A dedicated endpoint
// keeps operator concerns visible without forcing callers to infer quality from
// broader dataset metadata responses.
package quality

import (
	"net/http"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// Handler serves quality endpoints.
type Handler struct {
	service *Service
}

// NewHandler constructs a quality handler.
func NewHandler(service *Service) http.Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"checks": h.service.ListStatuses(),
	})
}
