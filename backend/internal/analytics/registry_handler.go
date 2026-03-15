// This file exposes the semantic metric registry used by the frontend metric
// browser. It keeps the serving layer curated by returning repo-managed metric
// definitions plus lightweight preview series.
package analytics

import (
	"net/http"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// MetricLoader describes the manifest access required by the metrics registry.
type MetricLoader interface {
	LoadMetrics() ([]MetricDefinition, error)
}

// MetricCatalogHandler serves the semantic metric registry.
type MetricCatalogHandler struct {
	loader  MetricLoader
	service *Service
}

// NewMetricCatalogHandler constructs the metric registry handler.
func NewMetricCatalogHandler(loader MetricLoader, service *Service) http.Handler {
	return &MetricCatalogHandler{loader: loader, service: service}
}

func (h *MetricCatalogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.loader.LoadMetrics()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	type metricPreview struct {
		Definition MetricDefinition `json:"definition"`
		Preview    []map[string]any `json:"preview"`
	}

	out := make([]metricPreview, 0, len(metrics))
	for _, metric := range metrics {
		preview := []map[string]any{}
		if result, err := h.service.QueryMetric(metric.ID, QueryOptions{Limit: 6}); err == nil {
			preview = result.Series
		}
		out = append(out, metricPreview{
			Definition: metric,
			Preview:    preview,
		})
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{"metrics": out})
}
