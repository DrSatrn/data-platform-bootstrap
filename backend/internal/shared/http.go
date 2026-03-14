// Package shared contains small reusable helpers with narrow scope. HTTP
// response helpers live here because several bounded contexts expose their own
// handlers but should still behave consistently.
package shared

import (
	"encoding/json"
	"net/http"
)

// WriteJSON returns a JSON payload with the provided status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
