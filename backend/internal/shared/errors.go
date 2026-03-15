// Package shared contains small reusable helpers with narrow scope. These
// error helpers keep HTTP handlers consistent and log full backend failures
// server-side without exposing raw internals to clients.
package shared

import (
	"log/slog"
	"net/http"
)

// WriteError logs the backend error and returns a stable JSON error payload.
func WriteError(w http.ResponseWriter, status int, clientMessage string, err error) {
	if err != nil {
		slog.Default().Error(
			"http handler returned error",
			slog.Int("status", status),
			slog.String("client_error", clientMessage),
			slog.String("error", err.Error()),
		)
	}
	WriteJSON(w, status, map[string]any{
		"error": clientMessage,
	})
}

// WriteRoleError standardizes authorization failures across handler-local role
// checks and the shared middleware.
func WriteRoleError(w http.ResponseWriter, requiredRole string, currentRole string) {
	WriteJSON(w, http.StatusForbidden, map[string]any{
		"error":         requiredRole + " role required",
		"required_role": requiredRole,
		"current_role":  currentRole,
	})
}
