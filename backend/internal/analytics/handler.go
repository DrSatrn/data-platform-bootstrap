// This file exposes the constrained analytics API used by the reporting UI.
// The handler intentionally returns curated data only and does not accept
// arbitrary SQL or uncontrolled asset references.
package analytics

import (
	"net/http"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// Handler serves analytics endpoints.
type Handler struct {
	service *Service
}

// NewHandler constructs the analytics HTTP handler.
func NewHandler(service *Service) http.Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.service.SampleDashboard()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"dashboard": dashboard,
	})
}
