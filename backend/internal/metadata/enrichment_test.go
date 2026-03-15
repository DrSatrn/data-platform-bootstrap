// These tests extend metadata coverage around freshness derivation, manifest
// lineage, and nil-safe enrichment behavior.
package metadata

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFreshnessStatusFallsBackToTimestampWhenSLAIsInvalid(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "mart", "mart_monthly_cashflow.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir mart dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`[]`), 0o644); err != nil {
		t.Fatalf("write mart file: %v", err)
	}
	updatedAt := time.Now().Add(-90 * time.Minute)
	if err := os.Chtimes(path, updatedAt, updatedAt); err != nil {
		t.Fatalf("set modtime: %v", err)
	}

	handler := &CatalogHandler{dataRoot: root}
	status := handler.freshnessStatus(DataAsset{
		ID: "mart_monthly_cashflow",
		Freshness: Freshness{
			ExpectedWithin: "not-a-duration",
			WarnAfter:      "also-bad",
		},
	})

	if status.State != "fresh" {
		t.Fatalf("expected fresh fallback status, got %+v", status)
	}
	if status.LastUpdated == "" || status.LagSeconds <= 0 {
		t.Fatalf("expected timestamp-derived freshness details, got %+v", status)
	}
}

func TestEnrichAssetsHandlesMissingOptionalFields(t *testing.T) {
	enriched := EnrichAssets([]DataAsset{{
		ID:    "mart_budget_vs_actual",
		Layer: "mart",
	}})

	if len(enriched) != 1 {
		t.Fatalf("expected one asset, got %+v", enriched)
	}
	if enriched[0].Coverage.TotalColumns != 0 || enriched[0].Coverage.HasDocumentation {
		t.Fatalf("expected zero-value coverage for sparse asset, got %+v", enriched[0].Coverage)
	}
	if len(enriched[0].Lineage.Upstream) != 0 || len(enriched[0].Lineage.Downstream) != 0 {
		t.Fatalf("expected empty lineage for sparse asset, got %+v", enriched[0].Lineage)
	}
}

func TestSummarizeAssetsHandlesEmptyInput(t *testing.T) {
	summary := SummarizeAssets(nil)

	if summary.TotalAssets != 0 || summary.TotalColumns != 0 || summary.DocumentedColumns != 0 {
		t.Fatalf("expected zero summary for empty assets, got %+v", summary)
	}
	if summary.ByLayer == nil || summary.ByFreshness == nil {
		t.Fatalf("expected initialized summary maps, got %+v", summary)
	}
}

func TestBuildEdgesExtractsManifestLineageFromSourceRefs(t *testing.T) {
	assets := EnrichAssets([]DataAsset{
		{ID: "raw_transactions", Layer: "raw", SourceRefs: []string{"sample.transactions_csv"}},
		{ID: "mart_monthly_cashflow", Layer: "mart", SourceRefs: []string{"raw_transactions", "raw_transactions"}},
	})

	edges := BuildEdges(assets)
	if len(edges) != 2 {
		t.Fatalf("expected two unique edges, got %+v", edges)
	}
}

func TestProjectStoreHandlesNilDependenciesAndLoaderErrors(t *testing.T) {
	if err := ProjectStore(nil, nil); err != nil {
		t.Fatalf("expected nil dependencies to be a no-op, got %v", err)
	}

	err := ProjectStore(enrichmentLoaderStub{err: errors.New("load failed")}, enrichmentStoreStub{})
	if err == nil || err.Error() != "load assets for projection: load failed" {
		t.Fatalf("expected wrapped loader error, got %v", err)
	}
}

type enrichmentLoaderStub struct {
	assets []DataAsset
	err    error
}

func (s enrichmentLoaderStub) LoadAssets() ([]DataAsset, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.assets, nil
}

type enrichmentStoreStub struct {
	seeded []DataAsset
}

func (s enrichmentStoreStub) SeedAssets([]DataAsset) error {
	return nil
}

func (s enrichmentStoreStub) ListAssets() ([]DataAsset, error) {
	return s.seeded, nil
}

func (s enrichmentStoreStub) UpdateAnnotations(AssetAnnotationsPatch) error {
	return nil
}
