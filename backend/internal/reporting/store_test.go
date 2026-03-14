// These tests cover the local-first reporting store that seeds itself from
// dashboard manifests and persists dashboard edits under the data root.
package reporting

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStoreSeedsFromDashboardManifest(t *testing.T) {
	root := t.TempDir()
	dashboardRoot := filepath.Join(root, "dashboards")
	if err := os.MkdirAll(dashboardRoot, 0o755); err != nil {
		t.Fatalf("mkdir dashboard root: %v", err)
	}

	manifest := `id: finance_overview
name: Finance Overview
description: Test dashboard
widgets:
  - id: savings_rate
    name: Savings Rate
    type: kpi
    metric_ref: metrics_savings_rate
`
	if err := os.WriteFile(filepath.Join(dashboardRoot, "finance_overview.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write dashboard manifest: %v", err)
	}

	store, err := NewFileStore(root, dashboardRoot)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}

	dashboards, err := store.ListDashboards()
	if err != nil {
		t.Fatalf("list dashboards: %v", err)
	}
	if len(dashboards) != 1 {
		t.Fatalf("expected 1 dashboard, got %d", len(dashboards))
	}
	if dashboards[0].ID != "finance_overview" {
		t.Fatalf("unexpected dashboard id: %s", dashboards[0].ID)
	}
}

func TestFileStoreSaveDashboardPersists(t *testing.T) {
	root := t.TempDir()
	dashboardRoot := filepath.Join(root, "dashboards")
	if err := os.MkdirAll(dashboardRoot, 0o755); err != nil {
		t.Fatalf("mkdir dashboard root: %v", err)
	}

	store, err := NewFileStore(root, dashboardRoot)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	if err := store.SaveDashboard(Dashboard{
		ID:          "team_ops",
		Name:        "Team Ops",
		Description: "Tracks operator-facing metrics.",
		Widgets: []DashboardWidget{
			{ID: "cashflow", Name: "Cashflow", Type: "table", DatasetRef: "mart_monthly_cashflow"},
		},
	}); err != nil {
		t.Fatalf("save dashboard: %v", err)
	}

	again, err := NewFileStore(root, dashboardRoot)
	if err != nil {
		t.Fatalf("new file store second pass: %v", err)
	}
	dashboards, err := again.ListDashboards()
	if err != nil {
		t.Fatalf("list dashboards second pass: %v", err)
	}
	found := false
	for _, dashboard := range dashboards {
		if dashboard.ID == "team_ops" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected saved dashboard to persist")
	}
}
