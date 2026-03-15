// Package reporting owns saved dashboard definitions and the backend API
// contract used by the frontend reporting experience. The store is intentionally
// local-first: it persists dashboards under the platform data root and seeds
// itself from repo-managed dashboard manifests so product behavior remains
// versioned and inspectable.
package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Store describes the reporting persistence behavior needed by the API and
// admin surfaces.
type Store interface {
	ListDashboards() ([]Dashboard, error)
	SaveDashboard(Dashboard) error
	DeleteDashboard(string) error
}

// MultiStore lets the platform prefer one reporting store while keeping a
// secondary local-first mirror. This is useful when PostgreSQL is available
// but we still want filesystem-backed resilience and an easy seed source.
type MultiStore struct {
	primary   Store
	secondary Store
}

// NewMultiStore creates a reporting store that prefers the primary backend but
// falls back to the secondary backend when needed.
func NewMultiStore(primary, secondary Store) Store {
	return &MultiStore{primary: primary, secondary: secondary}
}

// Dashboard summarizes a saved internal reporting view.
type Dashboard struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Widgets     []DashboardWidget `json:"widgets" yaml:"widgets"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
}

// DashboardWidget defines one reporting element powered by a constrained
// analytics query rather than arbitrary SQL.
type DashboardWidget struct {
	ID          string      `json:"id" yaml:"id"`
	Name        string      `json:"name" yaml:"name"`
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description,omitempty" yaml:"description"`
	DatasetRef  string      `json:"dataset_ref,omitempty" yaml:"dataset_ref"`
	MetricRef   string      `json:"metric_ref,omitempty" yaml:"metric_ref"`
	ValueField  string      `json:"value_field,omitempty" yaml:"value_field"`
	XAxis       string      `json:"x_axis,omitempty" yaml:"x_axis"`
	YAxis       string      `json:"y_axis,omitempty" yaml:"y_axis"`
	Limit       int         `json:"limit,omitempty" yaml:"limit"`
	Filters     WidgetQuery `json:"filters,omitempty" yaml:"filters"`
}

// WidgetQuery captures the constrained filter contract shared with the
// analytics service.
type WidgetQuery struct {
	FromMonth string `json:"from_month,omitempty" yaml:"from_month"`
	ToMonth   string `json:"to_month,omitempty" yaml:"to_month"`
	Category  string `json:"category,omitempty" yaml:"category"`
}

// MemoryStore keeps dashboards in memory when the filesystem path is
// unavailable. It remains useful as a fallback in tests or unusual startup
// failures.
type MemoryStore struct {
	mu         sync.RWMutex
	dashboards []Dashboard
}

// NewMemoryStore creates a store with a default finance dashboard.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{dashboards: defaultDashboards()}
}

// ListDashboards returns a copy for callers.
func (s *MemoryStore) ListDashboards() ([]Dashboard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Dashboard, len(s.dashboards))
	copy(out, s.dashboards)
	return out, nil
}

// SaveDashboard upserts a dashboard in memory.
func (s *MemoryStore) SaveDashboard(dashboard Dashboard) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dashboard.UpdatedAt = time.Now().UTC()
	for index := range s.dashboards {
		if s.dashboards[index].ID == dashboard.ID {
			s.dashboards[index] = dashboard
			return nil
		}
	}
	s.dashboards = append(s.dashboards, dashboard)
	return nil
}

// DeleteDashboard removes a saved dashboard from memory.
func (s *MemoryStore) DeleteDashboard(dashboardID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := range s.dashboards {
		if s.dashboards[index].ID == dashboardID {
			s.dashboards = append(s.dashboards[:index], s.dashboards[index+1:]...)
			return nil
		}
	}
	return nil
}

// FileStore persists dashboards under the platform data root and seeds missing
// state from repo-managed dashboard manifests.
type FileStore struct {
	mu            sync.RWMutex
	path          string
	dashboardRoot string
}

// NewFileStore constructs a filesystem-backed dashboard store.
func NewFileStore(dataRoot, dashboardRoot string) (*FileStore, error) {
	store := &FileStore{
		path:          filepath.Join(dataRoot, "control_plane", "dashboards.json"),
		dashboardRoot: dashboardRoot,
	}
	if err := store.ensureSeeded(); err != nil {
		return nil, err
	}
	return store, nil
}

// ListDashboards returns saved dashboards, seeded from manifests on first run.
func (s *FileStore) ListDashboards() ([]Dashboard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readDashboardsLocked()
}

// SaveDashboard upserts a dashboard and persists the full dashboard list.
func (s *FileStore) SaveDashboard(dashboard Dashboard) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateDashboard(dashboard); err != nil {
		return err
	}
	dashboard.UpdatedAt = time.Now().UTC()

	dashboards, err := s.readDashboardsLocked()
	if err != nil {
		return err
	}
	for index := range dashboards {
		if dashboards[index].ID == dashboard.ID {
			dashboards[index] = dashboard
			return s.writeDashboardsLocked(dashboards)
		}
	}
	dashboards = append(dashboards, dashboard)
	slices.SortFunc(dashboards, func(left, right Dashboard) int {
		return compareStrings(left.Name, right.Name)
	})
	return s.writeDashboardsLocked(dashboards)
}

// DeleteDashboard removes one dashboard from persisted local-first state.
func (s *FileStore) DeleteDashboard(dashboardID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dashboards, err := s.readDashboardsLocked()
	if err != nil {
		return err
	}
	filtered := dashboards[:0]
	for _, dashboard := range dashboards {
		if dashboard.ID != dashboardID {
			filtered = append(filtered, dashboard)
		}
	}
	return s.writeDashboardsLocked(filtered)
}

func (s *FileStore) ensureSeeded() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create dashboard store dir: %w", err)
	}
	bytes, err := os.ReadFile(s.path)
	if err == nil && len(bytes) > 0 {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read dashboard store: %w", err)
	}

	dashboards, err := s.loadSeedDashboards()
	if err != nil {
		return err
	}
	return s.writeDashboardsLocked(dashboards)
}

func (s *FileStore) loadSeedDashboards() ([]Dashboard, error) {
	pattern := filepath.Join(s.dashboardRoot, "*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob dashboard manifests: %w", err)
	}
	if len(matches) == 0 {
		return defaultDashboards(), nil
	}

	dashboards := make([]Dashboard, 0, len(matches))
	for _, match := range matches {
		bytes, err := os.ReadFile(match)
		if err != nil {
			return nil, fmt.Errorf("read dashboard manifest %s: %w", match, err)
		}
		var dashboard Dashboard
		if err := yaml.Unmarshal(bytes, &dashboard); err != nil {
			return nil, fmt.Errorf("decode dashboard manifest %s: %w", match, err)
		}
		if err := validateDashboard(dashboard); err != nil {
			return nil, fmt.Errorf("validate dashboard manifest %s: %w", match, err)
		}
		dashboards = append(dashboards, dashboard)
	}
	slices.SortFunc(dashboards, func(left, right Dashboard) int {
		return compareStrings(left.Name, right.Name)
	})
	return dashboards, nil
}

func (s *FileStore) readDashboardsLocked() ([]Dashboard, error) {
	bytes, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read dashboards file: %w", err)
	}
	var dashboards []Dashboard
	if err := json.Unmarshal(bytes, &dashboards); err != nil {
		return nil, fmt.Errorf("decode dashboards file: %w", err)
	}
	return dashboards, nil
}

func (s *FileStore) writeDashboardsLocked(dashboards []Dashboard) error {
	bytes, err := json.MarshalIndent(dashboards, "", "  ")
	if err != nil {
		return fmt.Errorf("encode dashboards file: %w", err)
	}
	if err := os.WriteFile(s.path, bytes, 0o644); err != nil {
		return fmt.Errorf("write dashboards file: %w", err)
	}
	return nil
}

func validateDashboard(dashboard Dashboard) error {
	if dashboard.ID == "" {
		return fmt.Errorf("dashboard id is required")
	}
	if dashboard.Name == "" {
		return fmt.Errorf("dashboard name is required")
	}
	if len(dashboard.Widgets) == 0 {
		return fmt.Errorf("dashboard %s must define at least one widget", dashboard.ID)
	}
	for _, widget := range dashboard.Widgets {
		if widget.ID == "" {
			return fmt.Errorf("dashboard %s contains a widget without an id", dashboard.ID)
		}
		if widget.Name == "" {
			return fmt.Errorf("dashboard %s widget %s is missing a name", dashboard.ID, widget.ID)
		}
		if widget.DatasetRef == "" && widget.MetricRef == "" {
			return fmt.Errorf("dashboard %s widget %s must reference a dataset or metric", dashboard.ID, widget.ID)
		}
	}
	return nil
}

// ValidateDashboardDefinition exposes dashboard validation to tooling that
// needs the same contract as the reporting store without duplicating logic.
func ValidateDashboardDefinition(dashboard Dashboard) error {
	return validateDashboard(dashboard)
}

func defaultDashboards() []Dashboard {
	return []Dashboard{
		{
			ID:          "finance_overview",
			Name:        "Finance Overview",
			Description: "Tracks savings rate, cashflow, category spend, and budget variance.",
			Widgets: []DashboardWidget{
				{ID: "savings_rate_kpi", Name: "Savings Rate", Type: "kpi", MetricRef: "metrics_savings_rate", ValueField: "savings_rate"},
				{ID: "cashflow_table", Name: "Monthly Cashflow", Type: "table", DatasetRef: "mart_monthly_cashflow"},
				{ID: "category_spend_table", Name: "Category Spend", Type: "table", DatasetRef: "mart_category_spend"},
				{ID: "budget_variance_table", Name: "Budget Variance", Type: "table", DatasetRef: "mart_budget_vs_actual"},
			},
		},
	}
}

func compareStrings(left, right string) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

// ListDashboards returns dashboards from the primary store when available and
// populated, otherwise it falls back to the secondary store.
func (s *MultiStore) ListDashboards() ([]Dashboard, error) {
	primaryDashboards, primaryErr := s.primary.ListDashboards()
	if primaryErr == nil && len(primaryDashboards) > 0 {
		return primaryDashboards, nil
	}
	return s.secondary.ListDashboards()
}

// SaveDashboard persists to both stores so the primary backend and local-first
// fallback stay aligned.
func (s *MultiStore) SaveDashboard(dashboard Dashboard) error {
	if err := s.primary.SaveDashboard(dashboard); err != nil {
		return err
	}
	return s.secondary.SaveDashboard(dashboard)
}

// DeleteDashboard removes a dashboard from both stores so the primary backend
// and the local-first mirror remain aligned.
func (s *MultiStore) DeleteDashboard(dashboardID string) error {
	if err := s.primary.DeleteDashboard(dashboardID); err != nil {
		return err
	}
	return s.secondary.DeleteDashboard(dashboardID)
}
