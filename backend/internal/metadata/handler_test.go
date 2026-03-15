// These tests cover the runtime freshness enrichment so the catalog UI can
// trust that local materializations are being classified consistently.
package metadata

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFreshnessStatusMarksMissingAssets(t *testing.T) {
	handler := &CatalogHandler{dataRoot: t.TempDir()}
	asset := DataAsset{
		ID: "mart_monthly_cashflow",
		Freshness: Freshness{
			ExpectedWithin: "24h",
			WarnAfter:      "48h",
		},
	}

	status := handler.freshnessStatus(asset)
	if status.State != "missing" {
		t.Fatalf("expected missing status, got %s", status.State)
	}
}

func TestFreshnessStatusMarksLateAndStaleAssets(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "mart", "mart_monthly_cashflow.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir mart dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`[]`), 0o644); err != nil {
		t.Fatalf("write mart file: %v", err)
	}

	handler := &CatalogHandler{dataRoot: root}
	asset := DataAsset{
		ID: "mart_monthly_cashflow",
		Freshness: Freshness{
			ExpectedWithin: "2h",
			WarnAfter:      "4h",
		},
	}

	lateAt := time.Now().Add(-3 * time.Hour)
	if err := os.Chtimes(path, lateAt, lateAt); err != nil {
		t.Fatalf("set late modtime: %v", err)
	}
	if status := handler.freshnessStatus(asset); status.State != "late" {
		t.Fatalf("expected late status, got %s", status.State)
	}

	staleAt := time.Now().Add(-5 * time.Hour)
	if err := os.Chtimes(path, staleAt, staleAt); err != nil {
		t.Fatalf("set stale modtime: %v", err)
	}
	if status := handler.freshnessStatus(asset); status.State != "stale" {
		t.Fatalf("expected stale status, got %s", status.State)
	}
}

func TestFreshnessStatusMarksFreshAssets(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "metrics", "metrics_savings_rate.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir metrics dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`[]`), 0o644); err != nil {
		t.Fatalf("write metrics file: %v", err)
	}
	recentAt := time.Now().Add(-30 * time.Minute)
	if err := os.Chtimes(path, recentAt, recentAt); err != nil {
		t.Fatalf("set fresh modtime: %v", err)
	}

	handler := &CatalogHandler{dataRoot: root}
	asset := DataAsset{
		ID: "metrics_savings_rate",
		Freshness: Freshness{
			ExpectedWithin: "2h",
			WarnAfter:      "4h",
		},
	}

	status := handler.freshnessStatus(asset)
	if status.State != "fresh" {
		t.Fatalf("expected fresh status, got %s", status.State)
	}
}
