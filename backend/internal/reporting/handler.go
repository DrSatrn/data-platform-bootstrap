// This file exposes saved-dashboard APIs for the frontend. The endpoint is kept
// intentionally narrow because dashboards should be powered by curated
// analytics responses rather than ad hoc backend complexity.
package reporting

import (
	"encoding/json"
	"net/http"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// Handler serves reporting endpoints.
type Handler struct {
	store Store
}

// NewHandler constructs the reporting handler.
func NewHandler(store Store) http.Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dashboards, err := h.store.ListDashboards()
		if err != nil {
			shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
			return
		}
		shared.WriteJSON(w, http.StatusOK, map[string]any{
			"dashboards": dashboards,
		})
	case http.MethodPost:
		var dashboard Dashboard
		if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": "invalid dashboard payload",
			})
			return
		}
		if err := h.store.SaveDashboard(dashboard); err != nil {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
			return
		}
		shared.WriteJSON(w, http.StatusCreated, map[string]any{
			"dashboard": dashboard,
		})
	case http.MethodDelete:
		dashboardID := r.URL.Query().Get("id")
		if dashboardID == "" {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": "dashboard id is required",
			})
			return
		}
		if err := h.store.DeleteDashboard(dashboardID); err != nil {
			shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
			return
		}
		shared.WriteJSON(w, http.StatusOK, map[string]any{
			"deleted": dashboardID,
		})
	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
	}
}
