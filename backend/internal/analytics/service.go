// Package analytics provides a deliberately constrained analytics-serving
// surface. The current implementation returns manifest-backed sample responses
// so the frontend and API contracts can be built before the DuckDB adapter is
// fully wired in.
package analytics

// Service owns curated analytics responses.
type Service struct{}

// QueryResult is shaped for chart-friendly frontend consumption.
type QueryResult struct {
	Dataset string           `json:"dataset"`
	Series  []map[string]any `json:"series"`
}

// NewService creates an analytics service.
func NewService() *Service {
	return &Service{}
}

// SampleDashboard returns a starter dataset for the finance dashboard.
func (s *Service) SampleDashboard() QueryResult {
	return QueryResult{
		Dataset: "metrics_monthly_cashflow",
		Series: []map[string]any{
			{"month": "2026-01", "income": 7200, "expenses": 4100, "savings_rate": 0.43},
			{"month": "2026-02", "income": 7200, "expenses": 4350, "savings_rate": 0.40},
			{"month": "2026-03", "income": 7200, "expenses": 3980, "savings_rate": 0.45},
		},
	}
}
