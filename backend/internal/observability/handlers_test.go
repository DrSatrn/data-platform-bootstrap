// These tests keep the system overview summaries trustworthy because the
// frontend relies on them for operator-facing diagnostics.
package observability

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/backup"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/metadata"
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

type pipelineLoaderStub struct {
	pipelines []orchestration.Pipeline
}

func (s pipelineLoaderStub) LoadPipelines() ([]orchestration.Pipeline, error) {
	return s.pipelines, nil
}

type assetLoaderStub struct {
	assets []metadata.DataAsset
}

func (s assetLoaderStub) LoadAssets() ([]metadata.DataAsset, error) {
	return s.assets, nil
}

type runStoreStub struct {
	runs []orchestration.PipelineRun
}

func (s runStoreStub) ListPipelineRuns() ([]orchestration.PipelineRun, error) {
	return s.runs, nil
}

func (s runStoreStub) SavePipelineRun(orchestration.PipelineRun) error {
	return nil
}

func (s runStoreStub) GetPipelineRun(string) (orchestration.PipelineRun, bool, error) {
	return orchestration.PipelineRun{}, false, nil
}

type queueStub struct {
	requests []orchestration.QueueSnapshot
}

func (s queueStub) ListRequests() ([]orchestration.QueueSnapshot, error) {
	return s.requests, nil
}

type backupInventoryStub struct {
	bundles []backup.BundleFile
}

func (s backupInventoryStub) ListBundles() ([]backup.BundleFile, error) {
	return s.bundles, nil
}

func TestOverviewHandlerIncludesPersistenceModes(t *testing.T) {
	dataRoot := t.TempDir()
	statusPath := filepath.Join(dataRoot, "control_plane", "scheduler_status.json")
	if err := os.MkdirAll(filepath.Dir(statusPath), 0o755); err != nil {
		t.Fatalf("mkdir status dir: %v", err)
	}
	if err := os.WriteFile(statusPath, []byte(`{"refreshed_at":"2026-03-15T01:00:00Z","pipeline_count":1,"asset_count":1}`), 0o644); err != nil {
		t.Fatalf("write scheduler status: %v", err)
	}

	handler := NewOverviewHandler(
		config.Settings{Environment: "test", HTTPAddr: ":8080", WebAddr: ":3000", APIBaseURL: "http://127.0.0.1:8080", DataRoot: dataRoot},
		NewService(),
		pipelineLoaderStub{pipelines: []orchestration.Pipeline{{ID: "finance"}}},
		assetLoaderStub{assets: []metadata.DataAsset{{ID: "mart_cashflow"}}},
		runStoreStub{runs: []orchestration.PipelineRun{{ID: "run_1", Status: orchestration.RunStatusSucceeded, StartedAt: time.Now().UTC()}}},
		queueStub{requests: []orchestration.QueueSnapshot{{Status: "queued"}}},
		backupInventoryStub{bundles: []backup.BundleFile{{Path: "/tmp/backup.tar.gz", SizeBytes: 42}}},
		map[string]PersistenceMode{
			"runs": {
				SourceOfTruth: "postgres",
				ReadPath:      "postgres run_snapshots",
				WritePath:     "postgres first, filesystem mirror",
			},
		},
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/overview", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"persistence_modes"`) || !strings.Contains(body, `"source_of_truth":"postgres"`) {
		t.Fatalf("expected persistence modes in response, got %s", body)
	}
	if !strings.Contains(body, `"scheduler_summary"`) || !strings.Contains(body, `"pipeline_count":1`) {
		t.Fatalf("expected scheduler summary in response, got %s", body)
	}
}
