// This test covers the self-built analytics path that computes curated finance
// metrics from repo-managed sample data.
package analytics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSampleDashboardBuildsMonthlySeries(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "personal_finance")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir sample data dir: %v", err)
	}

	csv := "transaction_id,occurred_at,account_name,category,amount\n" +
		"tx_1,2026-01-03T09:00:00Z,Everyday,Salary,5000\n" +
		"tx_2,2026-01-05T09:00:00Z,Everyday,Rent,-2000\n" +
		"tx_3,2026-02-03T09:00:00Z,Everyday,Salary,5000\n"
	if err := os.WriteFile(filepath.Join(dataDir, "transactions.csv"), []byte(csv), 0o644); err != nil {
		t.Fatalf("write transactions sample: %v", err)
	}
	budgets := `[{"category":"Rent","monthly_budget":2000}]`
	if err := os.WriteFile(filepath.Join(dataDir, "budget_rules.json"), []byte(budgets), 0o644); err != nil {
		t.Fatalf("write budget rules sample: %v", err)
	}

	service := NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "sql"))
	result, err := service.SampleDashboard()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if result.Dataset != "mart_monthly_cashflow" {
		t.Fatalf("unexpected dataset: %s", result.Dataset)
	}
	if len(result.Series) != 2 {
		t.Fatalf("expected 2 monthly rows, got %d", len(result.Series))
	}
}

func TestQueryCategorySpendBuildsRowsFromSampleData(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "personal_finance")
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

	service := NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "sql"))
	result, err := service.QueryDataset("mart_category_spend", QueryOptions{Category: "Groceries"})
	if err != nil {
		t.Fatalf("query category spend: %v", err)
	}

	if len(result.Series) != 1 {
		t.Fatalf("expected 1 filtered row, got %d", len(result.Series))
	}
	if result.Series[0]["category"] != "Groceries" {
		t.Fatalf("unexpected category row: %#v", result.Series[0])
	}
}

func TestQueryBudgetVarianceSupportsGroupingAndDrilldown(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "personal_finance")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir sample data dir: %v", err)
	}

	csv := "transaction_id,occurred_at,account_name,category,amount\n" +
		"tx_1,2026-01-03T09:00:00Z,Everyday,Salary,5000\n" +
		"tx_2,2026-01-05T09:00:00Z,Everyday,Rent,-2000\n" +
		"tx_3,2026-01-10T09:00:00Z,Everyday,Groceries,-250\n" +
		"tx_4,2026-02-10T09:00:00Z,Everyday,Groceries,-300\n"
	if err := os.WriteFile(filepath.Join(dataDir, "transactions.csv"), []byte(csv), 0o644); err != nil {
		t.Fatalf("write transactions sample: %v", err)
	}
	budgets := `[{"category":"Rent","monthly_budget":2000},{"category":"Groceries","monthly_budget":300}]`
	if err := os.WriteFile(filepath.Join(dataDir, "budget_rules.json"), []byte(budgets), 0o644); err != nil {
		t.Fatalf("write budget rules sample: %v", err)
	}

	service := NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "sql"))
	result, err := service.QueryDataset("mart_budget_vs_actual", QueryOptions{
		GroupBy:        "category",
		DrillDimension: "month",
		DrillValue:     "2026-01",
		SortBy:         "actual_spend",
		SortDirection:  "desc",
	})
	if err != nil {
		t.Fatalf("query grouped budget variance: %v", err)
	}

	if result.GroupBy != "category" || result.DrillDimension != "month" || result.DrillValue != "2026-01" {
		t.Fatalf("expected drilldown metadata, got %+v", result)
	}
	if len(result.Series) != 2 {
		t.Fatalf("expected 2 grouped rows for January categories, got %d", len(result.Series))
	}
	if result.Series[0]["category"] != "Rent" {
		t.Fatalf("expected rows sorted by spend desc, got %#v", result.Series)
	}
}

func TestQueryMetricExposesAvailableDimensions(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "personal_finance")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir sample data dir: %v", err)
	}

	csv := "transaction_id,occurred_at,account_name,category,amount\n" +
		"tx_1,2026-01-03T09:00:00Z,Everyday,Salary,5000\n" +
		"tx_2,2026-01-05T09:00:00Z,Everyday,Rent,-2000\n"
	if err := os.WriteFile(filepath.Join(dataDir, "transactions.csv"), []byte(csv), 0o644); err != nil {
		t.Fatalf("write transactions sample: %v", err)
	}
	budgets := `[{"category":"Rent","monthly_budget":2000}]`
	if err := os.WriteFile(filepath.Join(dataDir, "budget_rules.json"), []byte(budgets), 0o644); err != nil {
		t.Fatalf("write budget rules sample: %v", err)
	}

	service := NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "sql"))
	result, err := service.QueryMetric("metrics_category_variance", QueryOptions{})
	if err != nil {
		t.Fatalf("query metric: %v", err)
	}
	if len(result.AvailableDimensions) == 0 || len(result.AvailableMeasures) == 0 {
		t.Fatalf("expected available dimensions and measures, got %+v", result)
	}
}
