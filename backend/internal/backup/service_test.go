// These tests focus on bundle integrity because operators should be able to
// trust that a green backup command really produced a recoverable archive.
package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/reporting"
)

type queueStub struct {
	requests []orchestration.QueueSnapshot
}

func (q queueStub) ListRequests() ([]orchestration.QueueSnapshot, error) {
	return q.requests, nil
}

type metadataStub struct {
	assets []metadata.DataAsset
}

func (s metadataStub) SyncAssets([]metadata.DataAsset) error {
	return nil
}

func (s metadataStub) ListAssets() ([]metadata.DataAsset, error) {
	return s.assets, nil
}

func TestCreateAndVerifyBackup(t *testing.T) {
	root := t.TempDir()
	dataRoot := filepath.Join(root, "data")
	artifactRoot := filepath.Join(root, "artifacts")
	manifestRoot := filepath.Join(root, "manifests")
	dashboardRoot := filepath.Join(root, "dashboards")
	sqlRoot := filepath.Join(root, "sql")
	duckDBPath := filepath.Join(root, "duckdb", "platform.duckdb")

	for _, dir := range []string{
		filepath.Join(dataRoot, "raw"),
		filepath.Join(dataRoot, "mart"),
		filepath.Join(dataRoot, "metrics"),
		filepath.Join(dataRoot, "quality"),
		filepath.Join(dataRoot, "control_plane"),
		filepath.Join(artifactRoot, "runs", "run_1"),
		manifestRoot,
		dashboardRoot,
		sqlRoot,
		filepath.Dir(duckDBPath),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(dataRoot, "raw", "raw_transactions.csv"):          "id,amount\n1,42\n",
		filepath.Join(dataRoot, "mart", "mart_monthly_cashflow.json"):   `[{"month":"2026-03","net_cashflow":42}]`,
		filepath.Join(dataRoot, "metrics", "metrics_savings_rate.json"): `[{"month":"2026-03","savings_rate":0.1}]`,
		filepath.Join(dataRoot, "quality", "check_duplicates.json"):     `{"status":"pass"}`,
		filepath.Join(artifactRoot, "runs", "run_1", "metrics.json"):    `{"status":"ok"}`,
		filepath.Join(manifestRoot, "asset.yaml"):                       "id: mart_test\nname: Test\n",
		filepath.Join(dashboardRoot, "finance.yaml"):                    "id: finance\nname: Finance\nwidgets: []\n",
		duckDBPath: "duckdb snapshot",
	}
	for path, contents := range files {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	runStore, err := orchestration.NewFileStore(dataRoot)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	run := orchestration.PipelineRun{
		ID:         "run_1",
		PipelineID: "personal_finance_pipeline",
		Status:     orchestration.RunStatusSucceeded,
		Trigger:    "manual",
		StartedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := runStore.SavePipelineRun(run); err != nil {
		t.Fatalf("save run: %v", err)
	}

	reportStore := reporting.NewMemoryStore()
	if err := reportStore.SaveDashboard(reporting.Dashboard{ID: "finance", Name: "Finance", Widgets: []reporting.DashboardWidget{}}); err != nil {
		t.Fatalf("save dashboard: %v", err)
	}

	auditStore := audit.NewMemoryStore()
	if err := auditStore.Append(audit.Event{ActorSubject: "admin", ActorRole: "admin", Action: "trigger", Resource: "pipeline", Outcome: "success"}); err != nil {
		t.Fatalf("append audit event: %v", err)
	}

	cfg := config.Settings{
		Environment:       "test",
		HTTPAddr:          "127.0.0.1:8080",
		APIBaseURL:        "http://127.0.0.1:8080",
		LogLevel:          "debug",
		DataRoot:          dataRoot,
		ArtifactRoot:      artifactRoot,
		DuckDBPath:        duckDBPath,
		ManifestRoot:      manifestRoot,
		DashboardRoot:     dashboardRoot,
		SQLRoot:           sqlRoot,
		SchedulerTick:     time.Second,
		WorkerPoll:        time.Second,
		MaxConcurrentJobs: 2,
	}

	service := NewService(
		cfg,
		manifests.NewLoader(manifestRoot),
		runStore,
		queueStub{requests: []orchestration.QueueSnapshot{{RunID: "run_queued", PipelineID: "personal_finance_pipeline", Status: "queued"}}},
		reportStore,
		auditStore,
		metadataStub{assets: []metadata.DataAsset{{ID: "mart_test", Name: "Test", Layer: "mart"}}},
	)

	result, err := service.Create(filepath.Join(root, "backup.tar.gz"))
	if err != nil {
		t.Fatalf("create backup: %v", err)
	}
	if result.Manifest.Counts.PipelineRuns != 1 {
		t.Fatalf("expected 1 pipeline run, got %d", result.Manifest.Counts.PipelineRuns)
	}
	if result.Manifest.Counts.BundleFiles == 0 {
		t.Fatalf("expected archived files to be recorded")
	}

	verified, err := service.Verify(result.Path)
	if err != nil {
		t.Fatalf("verify backup: %v", err)
	}
	if verified.FormatVersion == "" {
		t.Fatalf("expected manifest format version")
	}
	if verified.Counts.DataAssets != 1 {
		t.Fatalf("expected 1 data asset, got %d", verified.Counts.DataAssets)
	}
}
