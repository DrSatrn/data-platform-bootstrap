// This file exposes saved-dashboard APIs for the frontend. The endpoint is kept
// intentionally narrow because dashboards should be powered by curated
// analytics responses rather than ad hoc backend complexity.
package reporting

import (
	"net/http"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// Handler serves reporting endpoints.
type Handler struct {
	store *MemoryStore
}

// NewHandler constructs the reporting handler.
func NewHandler(store *MemoryStore) http.Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"dashboards": h.store.ListDashboards(),
	})
}
