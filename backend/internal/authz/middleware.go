// This file provides small HTTP helpers for enforcing the current role model
// consistently across handlers. Centralizing the check keeps the access story
// aligned with the docs as more product surfaces become token-gated.
package authz

import (
	"net/http"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// RequireRole wraps an HTTP handler with a minimum-role check.
func RequireRole(service *Service, minimum Role, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := service.ResolveRequest(r)
		if Allowed(principal, minimum) {
			next.ServeHTTP(w, r)
			return
		}
		shared.WriteJSON(w, http.StatusForbidden, map[string]any{
			"error":         string(minimum) + " role required",
			"required_role": minimum,
			"current_role":  principal.Role,
		})
	})
}
