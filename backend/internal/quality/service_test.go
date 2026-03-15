// This test verifies that the first-party quality logic detects duplicates and
// uncategorized transactions from the sample dataset.
package quality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListStatusesComputesDuplicateAndUncategorizedCounts(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "personal_finance")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir sample data dir: %v", err)
	}

	csv := "transaction_id,occurred_at,account_name,category,amount\n" +
		"tx_1,2026-01-03T09:00:00Z,Everyday,Salary,5000\n" +
		"tx_1,2026-01-05T09:00:00Z,Everyday,, -2000\n"
	if err := os.WriteFile(filepath.Join(dataDir, "transactions.csv"), []byte(csv), 0o644); err != nil {
		t.Fatalf("write transactions sample: %v", err)
	}

	service := NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "sql"))
	statuses, err := service.ListStatuses()
	if err != nil {
		t.Fatalf("list statuses: %v", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("expected 2 quality statuses, got %d", len(statuses))
	}
	if statuses[0].Status != "warning" {
		t.Fatalf("expected duplicate warning, got %s", statuses[0].Status)
	}
	if statuses[1].Status != "warning" {
		t.Fatalf("expected uncategorized warning, got %s", statuses[1].Status)
	}
}

func TestListStatusesFallsBackWhenDuckDBQueriesReturnNoRows(t *testing.T) {
	root := t.TempDir()
	sqlRoot := filepath.Join(root, "sql")
	writeQualitySQLFixture(t, sqlRoot, "quality/check_duplicate_transactions.sql", "select 1 where false;")
	writeQualitySQLFixture(t, sqlRoot, "quality/check_uncategorized_transactions.sql", "select 1 where false;")

	sampleRoot := filepath.Join(root, "sample")
	writeTransactionsSample(t, sampleRoot, "transaction_id,occurred_at,account_name,category,amount\n"+
		"tx_1,2026-01-03T09:00:00Z,Everyday,Salary,5000\n")

	service := NewService(sampleRoot, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), sqlRoot)
	statuses, err := service.ListStatuses()
	if err != nil {
		t.Fatalf("list statuses fallback: %v", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses after fallback, got %d", len(statuses))
	}
	if statuses[0].Status != "passing" || statuses[1].Status != "passing" {
		t.Fatalf("expected passing fallback statuses, got %+v", statuses)
	}
}

func TestListStatusesFromDuckDBReturnsFormattedErrorForMissingSQLFiles(t *testing.T) {
	root := t.TempDir()
	service := NewService(filepath.Join(root, "sample"), filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "missing-sql"))

	_, err := service.listStatusesFromDuckDB()
	if err == nil {
		t.Fatal("expected missing sql file error")
	}
	if !strings.Contains(err.Error(), "read sql file") {
		t.Fatalf("expected wrapped sql read error, got %v", err)
	}
	if !strings.Contains(err.Error(), "check_duplicate_transactions.sql") {
		t.Fatalf("expected missing duplicate check path in error, got %v", err)
	}
}

func TestListStatusesRejectsEmptyInputsWhenNoFallbackDataExists(t *testing.T) {
	service := NewService("", "", filepath.Join(t.TempDir(), "duckdb", "platform.duckdb"), filepath.Join(t.TempDir(), "sql"))

	_, err := service.ListStatuses()
	if err == nil {
		t.Fatal("expected error for empty sample data root")
	}
	if !strings.Contains(err.Error(), "open transactions sample") {
		t.Fatalf("expected open sample error, got %v", err)
	}
}

func TestListStatusesFromDuckDBEvaluatesMultipleChecksInSequence(t *testing.T) {
	root := t.TempDir()
	sqlRoot := filepath.Join(root, "sql")
	writeQualitySQLFixture(t, sqlRoot, "quality/check_duplicate_transactions.sql", "select 2 as duplicate_count;")
	writeQualitySQLFixture(t, sqlRoot, "quality/check_uncategorized_transactions.sql", "select 1 as uncategorized_count;")

	service := NewService(filepath.Join(root, "sample"), filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), sqlRoot)
	statuses, err := service.listStatusesFromDuckDB()
	if err != nil {
		t.Fatalf("list statuses from duckdb: %v", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if statuses[0].ID != "check_duplicate_transactions" || statuses[0].Status != "warning" {
		t.Fatalf("expected duplicate warning first, got %+v", statuses[0])
	}
	if statuses[1].ID != "check_uncategorized_transactions" || statuses[1].Status != "warning" {
		t.Fatalf("expected uncategorized warning second, got %+v", statuses[1])
	}
}

func TestListStatusesReturnsArtifactFallbackBeforeSampleData(t *testing.T) {
	root := t.TempDir()
	sqlRoot := filepath.Join(root, "sql")
	writeQualitySQLFixture(t, sqlRoot, "quality/check_duplicate_transactions.sql", "select 1 where false;")
	writeQualitySQLFixture(t, sqlRoot, "quality/check_uncategorized_transactions.sql", "select 1 where false;")

	dataRoot := filepath.Join(root, "materialized")
	if err := os.MkdirAll(filepath.Join(dataRoot, "quality"), 0o755); err != nil {
		t.Fatalf("mkdir quality artifact dir: %v", err)
	}
	artifact := `{"status":"warning","uncategorized_count":3}`
	if err := os.WriteFile(filepath.Join(dataRoot, "quality", "check_uncategorized_transactions.json"), []byte(artifact), 0o644); err != nil {
		t.Fatalf("write quality artifact: %v", err)
	}

	sampleRoot := filepath.Join(root, "sample")
	writeTransactionsSample(t, sampleRoot, "transaction_id,occurred_at,account_name,category,amount\n"+
		"tx_1,2026-01-03T09:00:00Z,Everyday,Salary,5000\n"+
		"tx_2,2026-01-05T09:00:00Z,Everyday,, -2000\n")

	service := NewService(sampleRoot, dataRoot, filepath.Join(root, "duckdb", "platform.duckdb"), sqlRoot)
	statuses, err := service.ListStatuses()
	if err != nil {
		t.Fatalf("list statuses: %v", err)
	}

	if !strings.Contains(statuses[1].Message, "latest worker run artifact") {
		t.Fatalf("expected artifact-backed message, got %q", statuses[1].Message)
	}
	if !strings.Contains(statuses[1].Message, "3 uncategorized") {
		t.Fatalf("expected artifact count in message, got %q", statuses[1].Message)
	}
}

func writeTransactionsSample(t *testing.T, sampleRoot string, content string) {
	t.Helper()
	dataDir := filepath.Join(sampleRoot, "personal_finance")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir sample data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "transactions.csv"), []byte(content), 0o644); err != nil {
		t.Fatalf("write transactions sample: %v", err)
	}
}

func writeQualitySQLFixture(t *testing.T, sqlRoot string, relativePath string, sql string) {
	t.Helper()
	path := filepath.Join(sqlRoot, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir sql dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(sql), 0o644); err != nil {
		t.Fatalf("write sql fixture: %v", err)
	}
}
