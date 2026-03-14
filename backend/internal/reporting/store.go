// Package reporting holds saved-report and dashboard models. The in-memory
// store keeps the vertical slice usable before persistent reporting state lands
// in PostgreSQL.
package reporting

import "sync"

// Dashboard summarizes a saved internal reporting view.
type Dashboard struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Widgets     []string `json:"widgets"`
}

// MemoryStore stores dashboards in memory.
type MemoryStore struct {
	mu         sync.RWMutex
	dashboards []Dashboard
}

// NewMemoryStore creates a store with a default finance dashboard.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		dashboards: []Dashboard{
			{
				ID:          "finance-overview",
				Name:        "Finance Overview",
				Description: "Tracks savings rate, spend by category, and budget variance.",
				Widgets:     []string{"kpi_savings_rate", "chart_cashflow", "chart_budget_variance"},
			},
		},
	}
}

// ListDashboards returns a copy for callers.
func (s *MemoryStore) ListDashboards() []Dashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Dashboard, len(s.dashboards))
	copy(out, s.dashboards)
	return out
}
