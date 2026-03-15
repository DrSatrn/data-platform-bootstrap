// These tests cover the DuckDB transform engine because it sits on the
// analytical critical path for raw landing, transforms, quality queries, and
// metric materialization.
package transforms

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaterializeRawTablesHappyPath(t *testing.T) {
	engine, paths := newEngineFixture(t)

	if err := engine.MaterializeRawTables(paths.transactionsPath, paths.balancesPath, paths.budgetRulesPath); err != nil {
		t.Fatalf("materialize raw tables: %v", err)
	}

	rows, err := engine.QueryRows("select count(*) as transaction_count from raw_transactions")
	if err != nil {
		t.Fatalf("query raw_transactions: %v", err)
	}
	if len(rows) != 1 || rows[0]["transaction_count"] != int64(2) {
		t.Fatalf("unexpected raw transaction count: %#v", rows)
	}
}

func TestMaterializeRawTablesMissingFileReturnsError(t *testing.T) {
	engine, paths := newEngineFixture(t)

	err := engine.MaterializeRawTables(paths.transactionsPath, filepath.Join(t.TempDir(), "missing.json"), paths.budgetRulesPath)
	if err == nil {
		t.Fatal("expected missing file error, got nil")
	}
	if !strings.Contains(err.Error(), "raw_account_balances.sql") {
		t.Fatalf("expected raw_account_balances bootstrap error, got %v", err)
	}
}

func TestRunTransformValidAndInvalidReference(t *testing.T) {
	engine, paths := newEngineFixture(t)
	if err := engine.MaterializeRawTables(paths.transactionsPath, paths.balancesPath, paths.budgetRulesPath); err != nil {
		t.Fatalf("materialize raw tables: %v", err)
	}

	if err := engine.RunTransform("transform.monthly_cashflow"); err != nil {
		t.Fatalf("run valid transform: %v", err)
	}

	rows, err := engine.QueryRows("select month, savings_rate from mart_monthly_cashflow order by month")
	if err != nil {
		t.Fatalf("query mart_monthly_cashflow: %v", err)
	}
	if len(rows) != 1 || rows[0]["month"] != "2026-01" {
		t.Fatalf("unexpected mart rows: %#v", rows)
	}

	if err := engine.RunTransform("monthly_cashflow"); err == nil {
		t.Fatal("expected unsupported transform reference error, got nil")
	}
}

func TestQueryRowsReturnsRowsAndEmptyResults(t *testing.T) {
	engine, paths := newEngineFixture(t)
	if err := engine.MaterializeRawTables(paths.transactionsPath, paths.balancesPath, paths.budgetRulesPath); err != nil {
		t.Fatalf("materialize raw tables: %v", err)
	}

	rows, err := engine.QueryRows("select category from raw_transactions order by category")
	if err != nil {
		t.Fatalf("query rows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	emptyRows, err := engine.QueryRows("select category from raw_transactions where 1 = 0")
	if err != nil {
		t.Fatalf("query empty rows: %v", err)
	}
	if len(emptyRows) != 0 {
		t.Fatalf("expected empty result, got %#v", emptyRows)
	}
}

func TestQueryRowsFromFileValidPathAndMissingFile(t *testing.T) {
	engine, paths := newEngineFixture(t)
	if err := engine.MaterializeRawTables(paths.transactionsPath, paths.balancesPath, paths.budgetRulesPath); err != nil {
		t.Fatalf("materialize raw tables: %v", err)
	}

	rows, err := engine.QueryRowsFromFile(filepath.Join("quality", "check_duplicate_transactions.sql"), nil)
	if err != nil {
		t.Fatalf("query rows from file: %v", err)
	}
	if len(rows) != 1 || rows[0]["duplicate_count"] != int64(0) {
		t.Fatalf("unexpected duplicate query result: %#v", rows)
	}

	if _, err := engine.QueryRowsFromFile(filepath.Join("quality", "missing.sql"), nil); err == nil {
		t.Fatal("expected missing SQL file error, got nil")
	}
}

func TestRunMetricValidAndInvalidMetric(t *testing.T) {
	engine, paths := newEngineFixture(t)
	if err := engine.MaterializeRawTables(paths.transactionsPath, paths.balancesPath, paths.budgetRulesPath); err != nil {
		t.Fatalf("materialize raw tables: %v", err)
	}
	if err := engine.RunTransform("transform.monthly_cashflow"); err != nil {
		t.Fatalf("run transform: %v", err)
	}

	if err := engine.RunMetric("metrics_savings_rate"); err != nil {
		t.Fatalf("run valid metric: %v", err)
	}
	rows, err := engine.QueryRows("select savings_rate from metrics_savings_rate")
	if err != nil {
		t.Fatalf("query metric rows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 metric row, got %#v", rows)
	}

	if err := engine.RunMetric("missing_metric"); err == nil {
		t.Fatal("expected invalid metric error, got nil")
	}
}

type engineFixturePaths struct {
	transactionsPath string
	balancesPath     string
	budgetRulesPath  string
}

func newEngineFixture(t *testing.T) (*Engine, engineFixturePaths) {
	t.Helper()

	root := t.TempDir()
	sqlRoot := filepath.Join(root, "sql")
	writeSQLFixture(t, sqlRoot, filepath.Join("bootstrap", "raw_transactions.sql"), `
create or replace table raw_transactions as
select transaction_id, cast(occurred_at as timestamp) as occurred_at, account_name, nullif(trim(category), '') as category, cast(amount as double) as amount
from read_csv_auto({{RAW_TRANSACTIONS_PATH}}, header = true);
`)
	writeSQLFixture(t, sqlRoot, filepath.Join("bootstrap", "raw_account_balances.sql"), `
create or replace table raw_account_balances as
select account_id, cast(captured_at as timestamp) as captured_at, cast(balance as double) as balance
from read_json_auto({{RAW_ACCOUNT_BALANCES_PATH}});
`)
	writeSQLFixture(t, sqlRoot, filepath.Join("bootstrap", "raw_budget_rules.sql"), `
create or replace table raw_budget_rules as
select category, cast(monthly_budget as double) as monthly_budget
from read_json_auto({{RAW_BUDGET_RULES_PATH}});
`)
	writeSQLFixture(t, sqlRoot, filepath.Join("transforms", "monthly_cashflow.sql"), `
create or replace table mart_monthly_cashflow as
with monthly as (
  select
    strftime(date_trunc('month', occurred_at), '%Y-%m') as month,
    sum(case when amount >= 0 then amount else 0 end) as income,
    sum(case when amount < 0 then -amount else 0 end) as expenses
  from raw_transactions
  group by 1
)
select
  month,
  income,
  expenses,
  case when income > 0 then (income - expenses) / income else 0 end as savings_rate
from monthly
order by month;
`)
	writeSQLFixture(t, sqlRoot, filepath.Join("metrics", "metrics_savings_rate.sql"), `
create or replace table metrics_savings_rate as
select month, savings_rate
from mart_monthly_cashflow
order by month;
`)
	writeSQLFixture(t, sqlRoot, filepath.Join("quality", "check_duplicate_transactions.sql"), `
select count(*) as duplicate_count
from (
  select transaction_id
  from raw_transactions
  group by transaction_id
  having count(*) > 1
) duplicates;
`)

	transactionsPath := filepath.Join(root, "transactions.csv")
	balancesPath := filepath.Join(root, "balances.json")
	budgetRulesPath := filepath.Join(root, "budget_rules.json")
	if err := os.WriteFile(transactionsPath, []byte("transaction_id,occurred_at,account_name,category,amount\n"+
		"tx_1,2026-01-02T09:00:00Z,Everyday,Salary,5000\n"+
		"tx_2,2026-01-03T09:00:00Z,Everyday,Groceries,-120\n"), 0o644); err != nil {
		t.Fatalf("write transactions fixture: %v", err)
	}
	if err := os.WriteFile(balancesPath, []byte(`[{"account_id":"acct_1","captured_at":"2026-01-03T09:00:00Z","balance":1000}]`), 0o644); err != nil {
		t.Fatalf("write balances fixture: %v", err)
	}
	if err := os.WriteFile(budgetRulesPath, []byte(`[{"category":"Groceries","monthly_budget":500}]`), 0o644); err != nil {
		t.Fatalf("write budget rules fixture: %v", err)
	}

	engine := NewEngine(filepath.Join(root, "duckdb", "platform.duckdb"), sqlRoot)
	return engine, engineFixturePaths{
		transactionsPath: transactionsPath,
		balancesPath:     balancesPath,
		budgetRulesPath:  budgetRulesPath,
	}
}

func writeSQLFixture(t *testing.T, root, relativePath, contents string) {
	t.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir sql fixture dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)), 0o644); err != nil {
		t.Fatalf("write sql fixture %s: %v", relativePath, err)
	}
}
