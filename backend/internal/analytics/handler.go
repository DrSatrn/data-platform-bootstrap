// This file exposes the constrained analytics API used by the reporting UI.
// The handler intentionally returns curated data only and does not accept
// arbitrary SQL or uncontrolled asset references.
package analytics

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

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
	dataset, metric, options, err := parseAnalyticsRequest(r)
	if err != nil {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	var (
		result QueryResult
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
		status := http.StatusInternalServerError
		message := "failed to query analytics"
		if isClientAnalyticsError(err) {
			status = http.StatusBadRequest
			message = sanitizeAnalyticsError(err)
		}
		shared.WriteError(w, status, message, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"dashboard": result,
		"query":     result,
	})
}

type ExportHandler struct {
	service *Service
}

// NewExportHandler constructs the CSV export endpoint for curated analytics
// queries.
func NewExportHandler(service *Service) http.Handler {
	return &ExportHandler{service: service}
}

func (h *ExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dataset, metric, options, err := parseAnalyticsRequest(r)
	if err != nil {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}
	if dataset == "" && metric == "" {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "dataset or metric is required for export",
		})
		return
	}

	var (
		filename string
		payload  []byte
	)
	switch {
	case metric != "":
		filename = metric + ".csv"
		payload, err = h.service.ExportMetricCSV(metric, options)
	case dataset != "":
		filename = dataset + ".csv"
		payload, err = h.service.ExportDatasetCSV(dataset, options)
	}
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to export analytics"
		if isClientAnalyticsError(err) {
			status = http.StatusBadRequest
			message = sanitizeAnalyticsError(err)
		}
		shared.WriteError(w, status, message, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filename)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func isClientAnalyticsError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "unknown curated dataset") ||
		strings.Contains(message, "unknown metric") ||
		strings.Contains(message, "group_by") ||
		strings.Contains(message, "dataset or metric is required")
}

func sanitizeAnalyticsError(err error) string {
	message := err.Error()
	switch {
	case strings.Contains(message, "unknown curated dataset"):
		return "unknown curated dataset"
	case strings.Contains(message, "unknown metric"):
		return "unknown metric"
	case strings.Contains(message, "group_by"):
		return "unsupported group_by for this dataset"
	default:
		return "invalid analytics request"
	}
}

func parseAnalyticsRequest(r *http.Request) (string, string, QueryOptions, error) {
	dataset := r.URL.Query().Get("dataset")
	metric := r.URL.Query().Get("metric")
	options := QueryOptions{
		FromMonth:      r.URL.Query().Get("from_month"),
		ToMonth:        r.URL.Query().Get("to_month"),
		Category:       r.URL.Query().Get("category"),
		GroupBy:        r.URL.Query().Get("group_by"),
		DrillDimension: r.URL.Query().Get("drill_dimension"),
		DrillValue:     r.URL.Query().Get("drill_value"),
		SortBy:         r.URL.Query().Get("sort_by"),
		SortDirection:  r.URL.Query().Get("sort_direction"),
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		parsed, err := strconv.Atoi(limit)
		if err != nil || parsed < 0 {
			return "", "", QueryOptions{}, fmt.Errorf("limit must be a non-negative integer")
		}
		options.Limit = parsed
	}
	return dataset, metric, options, nil
}
