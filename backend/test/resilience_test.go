package test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/transforms"
)

func TestWorkerRestartReclaimsActiveQueueRequest(t *testing.T) {
	root := t.TempDir()
	queue, err := orchestration.NewQueue(root)
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}
	request := orchestration.RunRequest{
		RunID:      "run_restart",
		PipelineID: "inventory_operations_pipeline",
		Trigger:    "test",
	}
	if err := queue.Enqueue(request); err != nil {
		t.Fatalf("enqueue request: %v", err)
	}

	claimed, err := queue.ClaimNext()
	if err != nil {
		t.Fatalf("claim request before restart: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected an active queue claim before restart")
	}

	restartedQueue, err := orchestration.NewQueue(root)
	if err != nil {
		t.Fatalf("new queue after restart: %v", err)
	}
	reclaimed, err := restartedQueue.ClaimNext()
	if err != nil {
		t.Fatalf("claim request after restart: %v", err)
	}
	if reclaimed == nil || reclaimed.Request.RunID != request.RunID {
		t.Fatalf("expected active request %s to be reclaimed after restart, got %#v", request.RunID, reclaimed)
	}
}

func TestCorruptDuckDBReturnsClearError(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "duckdb", "platform.duckdb")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatalf("mkdir duckdb dir: %v", err)
	}
	if err := os.WriteFile(dbPath, []byte("not a duckdb database"), 0o644); err != nil {
		t.Fatalf("write corrupt duckdb: %v", err)
	}

	engine := transforms.NewEngine(dbPath, filepath.Join(root, "sql"))
	_, err := engine.QueryRows("select 1 as ok")
	if err == nil {
		t.Fatal("expected corrupt duckdb query to fail")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "duckdb") && !strings.Contains(strings.ToLower(err.Error()), "database") {
		t.Fatalf("expected a clear duckdb/database error, got %v", err)
	}
}

func TestConfigLoadStillSupportsResilienceDrillsEnvExample(t *testing.T) {
	t.Setenv("PLATFORM_ENV_FILE", filepath.Clean("../.env.example"))
	if _, err := config.Load(); err != nil {
		t.Fatalf("load config for resilience drills: %v", err)
	}
	_ = context.Background()
}
