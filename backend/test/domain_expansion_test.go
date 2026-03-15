package test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/streanor/data-platform/backend/internal/analytics"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/execution"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestBothDomainPipelinesExecuteAndServeAnalytics(t *testing.T) {
	repoRoot := repoRootFromTestFile(t)
	runtimeRoot := t.TempDir()
	dataRoot := filepath.Join(runtimeRoot, "data")
	artifactRoot := filepath.Join(runtimeRoot, "artifacts")
	duckDBPath := filepath.Join(runtimeRoot, "duckdb", "platform.duckdb")

	cfg := config.Settings{
		DataRoot:         dataRoot,
		ArtifactRoot:     artifactRoot,
		DuckDBPath:       duckDBPath,
		ManifestRoot:     filepath.Join(repoRoot, "packages", "manifests"),
		SQLRoot:          filepath.Join(repoRoot, "packages", "sql"),
		PythonTaskRoot:   filepath.Join(repoRoot, "packages", "python"),
		PythonBinary:     "python3",
		SampleDataRoot:   filepath.Join(repoRoot, "packages", "sample_data"),
		ExternalToolRoot: repoRoot,
	}

	loader := manifests.NewLoader(cfg.ManifestRoot)
	store := orchestration.NewInMemoryStore()
	queue, err := orchestration.NewQueue(dataRoot)
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}
	control := orchestration.NewControlService(loader, store, queue)
	runner := execution.NewRunner(cfg, loader, store, storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))

	for _, pipelineID := range []string{"personal_finance_pipeline", "inventory_operations_pipeline"} {
		run, err := control.TriggerPipeline(context.Background(), pipelineID, "test")
		if err != nil {
			t.Fatalf("trigger %s: %v", pipelineID, err)
		}
		claimed, err := queue.ClaimNext()
		if err != nil {
			t.Fatalf("claim queued request for %s: %v", pipelineID, err)
		}
		if claimed == nil {
			t.Fatalf("expected queued request for %s", pipelineID)
		}
		if err := runner.Execute(context.Background(), claimed.Request); err != nil {
			t.Fatalf("execute %s run %s: %v", pipelineID, run.ID, err)
		}
		if err := queue.Complete(claimed); err != nil {
			t.Fatalf("complete queue item for %s: %v", pipelineID, err)
		}
	}

	service := analytics.NewService(cfg.SampleDataRoot, cfg.DataRoot, cfg.DuckDBPath, cfg.SQLRoot)

	financeResult, err := service.QueryDataset("mart_budget_vs_actual", analytics.QueryOptions{})
	if err != nil {
		t.Fatalf("query finance dataset: %v", err)
	}
	if len(financeResult.Series) == 0 {
		t.Fatal("expected finance analytics rows after pipeline execution")
	}

	inventoryResult, err := service.QueryDataset("mart_inventory_monthly_summary", analytics.QueryOptions{})
	if err != nil {
		t.Fatalf("query inventory dataset: %v", err)
	}
	if len(inventoryResult.Series) == 0 {
		t.Fatal("expected inventory analytics rows after pipeline execution")
	}

	metricResult, err := service.QueryMetric("metrics_inventory_net_change", analytics.QueryOptions{})
	if err != nil {
		t.Fatalf("query inventory metric: %v", err)
	}
	if len(metricResult.Series) == 0 {
		t.Fatal("expected inventory metric rows after pipeline execution")
	}
}

func repoRootFromTestFile(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
