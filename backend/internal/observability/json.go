// This file provides a tiny JSON helper used by small runtime-facing packages
// to avoid repeating the same marshaling boilerplate in handlers.
package observability

import "encoding/json"

func mustMarshal(payload any) string {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return `{"status":"error","message":"failed to encode payload"}`
	}

	return string(bytes)
}
