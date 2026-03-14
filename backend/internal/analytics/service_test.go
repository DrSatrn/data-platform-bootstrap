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
