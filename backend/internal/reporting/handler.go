// This file exposes saved-dashboard APIs for the frontend. The endpoint is kept
// intentionally narrow because dashboards should be powered by curated
// analytics responses rather than ad hoc backend complexity.
package reporting

import (
	"encoding/json"
	"net/http"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/authz"
	"github.com/streanor/data-platform/backend/internal/shared"
)

// Handler serves reporting endpoints.
type Handler struct {
	store Store
	authz *authz.Service
	audit audit.Store
}

// NewHandler constructs the reporting handler.
func NewHandler(store Store, authService *authz.Service, auditStore audit.Store) http.Handler {
	return &Handler{store: store, authz: authService, audit: auditStore}
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
		principal := h.authz.ResolveRequest(r)
		if !authz.Allowed(principal, authz.RoleEditor) {
			_ = h.audit.Append(audit.Event{
				ActorUserID:  principal.UserID,
				ActorSubject: principal.Subject,
				ActorRole:    string(principal.Role),
				Action:       "save_dashboard",
				Resource:     "unknown",
				Outcome:      "forbidden",
			})
			shared.WriteJSON(w, http.StatusForbidden, map[string]any{
				"error": "editor role required to save dashboards",
			})
			return
		}
		var dashboard Dashboard
		if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": "invalid dashboard payload",
			})
			return
		}
		if err := h.store.SaveDashboard(dashboard); err != nil {
			_ = h.audit.Append(audit.Event{
				ActorUserID:  principal.UserID,
				ActorSubject: principal.Subject,
				ActorRole:    string(principal.Role),
				Action:       "save_dashboard",
				Resource:     dashboard.ID,
				Outcome:      "failure",
				Details: map[string]any{
					"error": err.Error(),
				},
			})
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
			return
		}
		_ = h.audit.Append(audit.Event{
			ActorUserID:  principal.UserID,
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "save_dashboard",
			Resource:     dashboard.ID,
			Outcome:      "success",
		})
		shared.WriteJSON(w, http.StatusCreated, map[string]any{
			"dashboard": dashboard,
		})
	case http.MethodDelete:
		principal := h.authz.ResolveRequest(r)
		if !authz.Allowed(principal, authz.RoleEditor) {
			_ = h.audit.Append(audit.Event{
				ActorUserID:  principal.UserID,
				ActorSubject: principal.Subject,
				ActorRole:    string(principal.Role),
				Action:       "delete_dashboard",
				Resource:     "unknown",
				Outcome:      "forbidden",
			})
			shared.WriteJSON(w, http.StatusForbidden, map[string]any{
				"error": "editor role required to delete dashboards",
			})
			return
		}
		dashboardID := r.URL.Query().Get("id")
		if dashboardID == "" {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": "dashboard id is required",
			})
			return
		}
		if err := h.store.DeleteDashboard(dashboardID); err != nil {
			_ = h.audit.Append(audit.Event{
				ActorUserID:  principal.UserID,
				ActorSubject: principal.Subject,
				ActorRole:    string(principal.Role),
				Action:       "delete_dashboard",
				Resource:     dashboardID,
				Outcome:      "failure",
				Details: map[string]any{
					"error": err.Error(),
				},
			})
			shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
			return
		}
		_ = h.audit.Append(audit.Event{
			ActorUserID:  principal.UserID,
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "delete_dashboard",
			Resource:     dashboardID,
			Outcome:      "success",
		})
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
