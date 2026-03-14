// This file exposes the constrained analytics API used by the reporting UI.
// The handler intentionally returns curated data only and does not accept
// arbitrary SQL or uncontrolled asset references.
package analytics

import (
	"net/http"
	"strconv"

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
	dataset := r.URL.Query().Get("dataset")
	metric := r.URL.Query().Get("metric")
	options := QueryOptions{
		FromMonth: r.URL.Query().Get("from_month"),
		ToMonth:   r.URL.Query().Get("to_month"),
		Category:  r.URL.Query().Get("category"),
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		parsed, err := strconv.Atoi(limit)
		if err != nil || parsed < 0 {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": "limit must be a non-negative integer",
			})
			return
		}
		options.Limit = parsed
	}

	var (
		result QueryResult
		err    error
	)
	switch {
	case metric != "":
		result, err = h.service.QueryMetric(metric, options)
	case dataset != "":
		result, err = h.service.QueryDataset(dataset, options)
	default:
		result, err = h.service.SampleDashboard()
	}
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"dashboard": result,
		"query":     result,
	})
}
