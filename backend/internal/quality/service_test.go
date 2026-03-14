// This test verifies that the first-party quality logic detects duplicates and
// uncategorized transactions from the sample dataset.
package quality

import (
	"os"
	"path/filepath"
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
