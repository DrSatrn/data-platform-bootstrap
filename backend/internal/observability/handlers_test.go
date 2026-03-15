// These tests keep the system overview summaries trustworthy because the
// frontend relies on them for operator-facing diagnostics.
package observability

import (
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/backup"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func TestSummarizeRuns(t *testing.T) {
	now := time.Now().UTC()
	finished := now.Add(-2 * time.Minute)
	finishedAt := now.Add(-time.Minute)
	summary := summarizeRuns([]orchestration.PipelineRun{
		{ID: "run_failed", Status: orchestration.RunStatusFailed, StartedAt: finished, FinishedAt: &finishedAt, Error: "boom"},
		{ID: "run_queued", Status: orchestration.RunStatusQueued, StartedAt: now},
		{ID: "run_succeeded", Status: orchestration.RunStatusSucceeded, StartedAt: finished, FinishedAt: &finishedAt},
	})
	if summary.TotalRuns != 3 {
		t.Fatalf("expected 3 runs, got %d", summary.TotalRuns)
	}
	if summary.FailedRuns != 1 || summary.SucceededRuns != 1 || summary.QueuedRuns != 1 {
		t.Fatalf("unexpected run breakdown: %#v", summary)
	}
	if summary.LatestFailureRunID != "run_failed" {
		t.Fatalf("expected latest failure run_failed, got %s", summary.LatestFailureRunID)
	}
	if summary.AverageDurationSeconds <= 0 {
		t.Fatalf("expected non-zero average duration, got %d", summary.AverageDurationSeconds)
	}
}

func TestSummarizeQueueAndBackups(t *testing.T) {
	queue := summarizeQueue([]orchestration.QueueSnapshot{
		{Status: "queued"},
		{Status: "active"},
		{Status: "completed"},
	})
	if queue.Total != 3 || queue.Queued != 1 || queue.Active != 1 || queue.Completed != 1 {
		t.Fatalf("unexpected queue summary: %#v", queue)
	}

	backups := summarizeBackups([]backup.BundleFile{{Path: "/tmp/latest.tar.gz", SizeBytes: 42}})
	if backups.BundleCount != 1 || backups.LatestBundlePath == "" || backups.LatestBundleBytes != 42 {
		t.Fatalf("unexpected backup summary: %#v", backups)
	}
}
