// This file exposes the audit API used by the built-in operations UI.
package audit

import (
	"net/http"
	"strconv"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// Handler serves recent audit events.
type Handler struct {
	store Store
}

// NewHandler constructs an audit API handler.
func NewHandler(store Store) http.Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
		return
	}

	limit := 25
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	events, err := h.store.ListRecent(limit)
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"events": events,
	})
}
