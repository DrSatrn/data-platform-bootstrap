// These tests cover the derived catalog trust model so the UI and validation
// tooling can rely on lineage and coverage calculations staying stable.
package metadata

import "testing"

func TestEnrichAssetsDerivesCoverageAndLineage(t *testing.T) {
	assets := []DataAsset{
		{
			ID:                "raw_transactions",
			Layer:             "raw",
			SourceRefs:        []string{"sample.transactions_csv"},
			DocumentationRefs: []string{"doc.raw_transactions"},
			QualityCheckRefs:  []string{"check_duplicate_transactions"},
			Columns: []Column{
				{Name: "transaction_id", Description: "Stable ID"},
				{Name: "amount", Description: "", IsPII: true},
			},
			FreshnessStatus: Status{State: "fresh"},
		},
		{
			ID:         "mart_monthly_cashflow",
			Layer:      "mart",
			SourceRefs: []string{"raw_transactions"},
			Columns: []Column{
				{Name: "month", Description: "Month"},
			},
			FreshnessStatus: Status{State: "late"},
		},
	}

	enriched := EnrichAssets(assets)
	if !enriched[0].Coverage.HasDocumentation {
		t.Fatalf("expected raw asset to be documented")
	}
	if !enriched[0].Coverage.HasQualityChecks {
		t.Fatalf("expected raw asset to have quality coverage")
	}
	if !enriched[0].Coverage.ContainsPII {
		t.Fatalf("expected raw asset pii detection")
	}
	if len(enriched[0].Lineage.Downstream) != 1 || enriched[0].Lineage.Downstream[0] != "mart_monthly_cashflow" {
		t.Fatalf("unexpected downstream lineage: %#v", enriched[0].Lineage.Downstream)
	}
	if len(enriched[1].Lineage.Upstream) != 1 || enriched[1].Lineage.Upstream[0] != "raw_transactions" {
		t.Fatalf("unexpected upstream lineage: %#v", enriched[1].Lineage.Upstream)
	}
}

func TestSummarizeAssetsAggregatesCoverageAndFreshness(t *testing.T) {
	assets := []DataAsset{
		{
			ID:              "raw_transactions",
			Layer:           "raw",
			FreshnessStatus: Status{State: "fresh"},
			Coverage: Coverage{
				HasDocumentation:  true,
				HasQualityChecks:  true,
				ContainsPII:       false,
				DocumentedColumns: 2,
				TotalColumns:      2,
			},
			Lineage: Lineage{Upstream: []string{"sample.transactions_csv"}},
		},
		{
			ID:              "mart_monthly_cashflow",
			Layer:           "mart",
			FreshnessStatus: Status{State: "stale"},
			Coverage: Coverage{
				HasDocumentation:  false,
				HasQualityChecks:  false,
				ContainsPII:       true,
				DocumentedColumns: 1,
				TotalColumns:      3,
			},
			Lineage: Lineage{Upstream: []string{"raw_transactions", "raw_account_balances"}},
		},
	}

	summary := SummarizeAssets(assets)
	if summary.TotalAssets != 2 {
		t.Fatalf("expected 2 assets, got %d", summary.TotalAssets)
	}
	if summary.ByLayer["raw"] != 1 || summary.ByLayer["mart"] != 1 {
		t.Fatalf("unexpected layer summary: %#v", summary.ByLayer)
	}
	if summary.ByFreshness["fresh"] != 1 || summary.ByFreshness["stale"] != 1 {
		t.Fatalf("unexpected freshness summary: %#v", summary.ByFreshness)
	}
	if summary.AssetsMissingDocs != 1 || summary.AssetsMissingQuality != 1 {
		t.Fatalf("unexpected governance summary: %#v", summary)
	}
	if summary.AssetsContainingPII != 1 {
		t.Fatalf("unexpected pii summary: %#v", summary)
	}
	if summary.DocumentedColumns != 3 || summary.TotalColumns != 5 {
		t.Fatalf("unexpected column coverage: %#v", summary)
	}
	if summary.LineageEdges != 3 {
		t.Fatalf("expected 3 lineage edges, got %d", summary.LineageEdges)
	}
}

func TestBuildEdgesDeduplicatesRelationships(t *testing.T) {
	edges := BuildEdges([]DataAsset{
		{ID: "mart_a", Lineage: Lineage{Upstream: []string{"raw_a", "raw_a", "raw_b"}}},
	})
	if len(edges) != 2 {
		t.Fatalf("expected 2 unique edges, got %d", len(edges))
	}
}
