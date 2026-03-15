// These tests deepen analytics coverage around constrained query validation
// and the staged fallback path from DuckDB to artifacts to sample data.
package analytics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQueryDatasetRejectsUnknownDataset(t *testing.T) {
	service := NewService(t.TempDir(), t.TempDir(), filepath.Join(t.TempDir(), "duckdb", "platform.duckdb"), t.TempDir())

	_, err := service.QueryDataset("unknown_dataset", QueryOptions{})
	if err == nil || !strings.Contains(err.Error(), "unknown curated dataset") {
		t.Fatalf("expected unknown dataset error, got %v", err)
	}
}

func TestQueryDatasetRejectsUnsupportedGroupBy(t *testing.T) {
	root := t.TempDir()
	writeAnalyticsSampleData(t, root)
	service := NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "missing-sql"))

	_, err := service.QueryDataset("mart_budget_vs_actual", QueryOptions{GroupBy: "owner"})
	if err == nil || !strings.Contains(err.Error(), "group_by") {
		t.Fatalf("expected unsupported group_by error, got %v", err)
	}
}

func TestQueryDatasetFallsBackToArtifactWhenDuckDBIsUnavailable(t *testing.T) {
	root := t.TempDir()
	dataRoot := filepath.Join(root, "materialized")
	writeArtifactRows(t, filepath.Join(dataRoot, "mart", "mart_monthly_cashflow.json"), `[{"month":"2026-01","income":5000,"expenses":2000,"savings_rate":0.6}]`)

	service := NewService(filepath.Join(root, "missing-sample"), dataRoot, filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "missing-sql"))
	result, err := service.QueryDataset("mart_monthly_cashflow", QueryOptions{})
	if err != nil {
		t.Fatalf("query dataset from artifact fallback: %v", err)
	}

	if len(result.Series) != 1 || result.Series[0]["month"] != "2026-01" {
		t.Fatalf("expected artifact-backed series, got %+v", result.Series)
	}
}

func TestQueryMetricFallsBackToSampleDataWhenDuckDBAndArtifactsAreUnavailable(t *testing.T) {
	root := t.TempDir()
	writeAnalyticsSampleData(t, root)

	service := NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "missing-sql"))
	result, err := service.QueryMetric("metrics_savings_rate", QueryOptions{})
	if err != nil {
		t.Fatalf("query metric from sample fallback: %v", err)
	}

	if result.Dataset != "metrics_savings_rate" {
		t.Fatalf("expected metric dataset id, got %+v", result)
	}
	if len(result.Series) == 0 {
		t.Fatalf("expected sample-backed metric rows, got %+v", result.Series)
	}
}

func TestFinalizeQueryResultHandlesEmptyDatasets(t *testing.T) {
	result, err := finalizeQueryResult("mart_monthly_cashflow", nil, QueryOptions{})
	if err != nil {
		t.Fatalf("finalize empty dataset: %v", err)
	}
	if len(result.Series) != 0 {
		t.Fatalf("expected no rows, got %+v", result.Series)
	}
	if len(result.AvailableDimensions) == 0 || len(result.AvailableMeasures) == 0 {
		t.Fatalf("expected schema metadata even for empty datasets, got %+v", result)
	}
}

func TestFinalizeQueryResultDoesNotOverTrimLargeLimits(t *testing.T) {
	rows := []map[string]any{
		{"month": "2026-01", "income": 5000.0, "expenses": 2000.0, "savings_rate": 0.6},
		{"month": "2026-02", "income": 5200.0, "expenses": 2100.0, "savings_rate": 0.596},
	}

	result, err := finalizeQueryResult("mart_monthly_cashflow", rows, QueryOptions{Limit: 5000})
	if err != nil {
		t.Fatalf("finalize large limit: %v", err)
	}
	if len(result.Series) != 2 {
		t.Fatalf("expected all rows to be preserved, got %+v", result.Series)
	}
}

func TestQueryDatasetSupportsMultiDimensionGroupBy(t *testing.T) {
	root := t.TempDir()
	writeAnalyticsSampleData(t, root)

	service := NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "missing-sql"))
	result, err := service.QueryDataset("mart_budget_vs_actual", QueryOptions{GroupBy: "month,category"})
	if err != nil {
		t.Fatalf("query multi-dimension group by: %v", err)
	}
	if len(result.Series) == 0 {
		t.Fatalf("expected grouped rows, got %+v", result.Series)
	}
	first := result.Series[0]
	if first["month"] == nil || first["category"] == nil {
		t.Fatalf("expected grouped dimensions to be preserved, got %+v", first)
	}
}

func writeAnalyticsSampleData(t *testing.T, sampleRoot string) {
	t.Helper()
	dataDir := filepath.Join(sampleRoot, "personal_finance")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir sample data dir: %v", err)
	}
	csv := "transaction_id,occurred_at,account_name,category,amount\n" +
		"tx_1,2026-01-03T09:00:00Z,Everyday,Salary,5000\n" +
		"tx_2,2026-01-05T09:00:00Z,Everyday,Rent,-2000\n" +
		"tx_3,2026-01-10T09:00:00Z,Everyday,Groceries,-250\n"
	if err := os.WriteFile(filepath.Join(dataDir, "transactions.csv"), []byte(csv), 0o644); err != nil {
		t.Fatalf("write transactions sample: %v", err)
	}
	budgets := `[{"category":"Rent","monthly_budget":2000},{"category":"Groceries","monthly_budget":300}]`
	if err := os.WriteFile(filepath.Join(dataDir, "budget_rules.json"), []byte(budgets), 0o644); err != nil {
		t.Fatalf("write budget rules sample: %v", err)
	}
}

func writeArtifactRows(t *testing.T, path string, payload string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir artifact dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("write artifact rows: %v", err)
	}
}
