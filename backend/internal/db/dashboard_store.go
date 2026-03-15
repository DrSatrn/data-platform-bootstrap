// This file provides PostgreSQL-backed persistence for saved dashboard
// definitions. The store uses the existing `dashboards` table and keeps the
// JSON definition explicit so reporting state remains inspectable and easy to
// migrate later.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streanor/data-platform/backend/internal/reporting"
)

// DashboardStore persists reporting dashboards in PostgreSQL.
type DashboardStore struct {
	conn *sql.DB
}

// NewDashboardStoreFromConn constructs a store from an existing connection
// pool.
func NewDashboardStoreFromConn(conn *sql.DB) *DashboardStore {
	return &DashboardStore{conn: conn}
}

// ListDashboards loads all dashboards ordered by name.
func (s *DashboardStore) ListDashboards() ([]reporting.Dashboard, error) {
	rows, err := s.conn.Query(`
		select id, name, description, definition
		from dashboards
		order by name
	`)
	if err != nil {
		return nil, fmt.Errorf("query dashboards: %w", err)
	}
	defer rows.Close()

	dashboards := []reporting.Dashboard{}
	for rows.Next() {
		var (
			dashboard  reporting.Dashboard
			definition []byte
		)
		if err := rows.Scan(&dashboard.ID, &dashboard.Name, &dashboard.Description, &definition); err != nil {
			return nil, fmt.Errorf("scan dashboard: %w", err)
		}
		if err := hydrateDashboardDefinition(&dashboard, definition); err != nil {
			return nil, err
		}
		dashboards = append(dashboards, dashboard)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dashboards: %w", err)
	}
	return dashboards, nil
}

// SaveDashboard upserts one dashboard definition.
func (s *DashboardStore) SaveDashboard(dashboard reporting.Dashboard) error {
	dashboard.UpdatedAt = time.Now().UTC()
	definition, err := json.Marshal(map[string]any{
		"owner":           dashboard.Owner,
		"tags":            dashboard.Tags,
		"shared_role":     dashboard.SharedRole,
		"default_filters": dashboard.DefaultFilters,
		"presets":         dashboard.Presets,
		"widgets":         dashboard.Widgets,
		"updated_at":      dashboard.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("encode dashboard definition: %w", err)
	}

	_, err = s.conn.Exec(`
		insert into dashboards (id, name, description, definition)
		values ($1, $2, $3, $4)
		on conflict (id) do update set
			name = excluded.name,
			description = excluded.description,
			definition = excluded.definition
	`, dashboard.ID, dashboard.Name, dashboard.Description, definition)
	if err != nil {
		return fmt.Errorf("save dashboard %s: %w", dashboard.ID, err)
	}
	return nil
}

// DeleteDashboard removes a saved dashboard definition from PostgreSQL.
func (s *DashboardStore) DeleteDashboard(dashboardID string) error {
	_, err := s.conn.Exec(`delete from dashboards where id = $1`, dashboardID)
	if err != nil {
		return fmt.Errorf("delete dashboard %s: %w", dashboardID, err)
	}
	return nil
}

func hydrateDashboardDefinition(dashboard *reporting.Dashboard, definition []byte) error {
	if len(definition) == 0 {
		return nil
	}
	var payload struct {
		Owner          string                      `json:"owner"`
		Tags           []string                    `json:"tags"`
		SharedRole     string                      `json:"shared_role"`
		DefaultFilters reporting.WidgetQuery       `json:"default_filters"`
		Presets        []reporting.DashboardPreset `json:"presets"`
		Widgets        []reporting.DashboardWidget `json:"widgets"`
		UpdatedAt      time.Time                   `json:"updated_at"`
	}
	if err := json.Unmarshal(definition, &payload); err != nil {
		return fmt.Errorf("decode dashboard definition: %w", err)
	}
	dashboard.Owner = payload.Owner
	dashboard.Tags = payload.Tags
	dashboard.SharedRole = payload.SharedRole
	dashboard.DefaultFilters = payload.DefaultFilters
	dashboard.Presets = payload.Presets
	dashboard.Widgets = payload.Widgets
	dashboard.UpdatedAt = payload.UpdatedAt
	return nil
}
