// This file exposes run artifact listing and content retrieval. The endpoints
// are intentionally narrow and only serve files produced under the configured
// artifact root.
package storage

import (
	"net/http"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// Handler serves artifact endpoints.
type Handler struct {
	service *Service
}

// NewHandler constructs an artifact handler.
func NewHandler(service *Service) http.Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		runID := r.URL.Query().Get("run_id")
		relativePath := r.URL.Query().Get("path")
		if runID == "" {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "run_id is required"})
			return
		}
		if relativePath == "" {
			artifacts, err := h.service.ListRunArtifacts(runID)
			if err != nil {
				shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			shared.WriteJSON(w, http.StatusOK, map[string]any{"artifacts": artifacts})
			return
		}

		bytes, err := h.service.ReadRunArtifact(runID, relativePath)
		if err != nil {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", detectContentType(relativePath))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bytes)
	default:
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}
