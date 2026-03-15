package execution

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestExecutePostsWebhookWhenRunFails(t *testing.T) {
	repoRoot := t.TempDir()
	manifestRoot := filepath.Join(repoRoot, "packages", "manifests")
	projectRef := "packages/external_tools/dbt_finance_demo"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))
	profilesRef := "packages/external_tools/dbt_finance_demo/profiles"
	profilesRoot := filepath.Join(repoRoot, filepath.FromSlash(profilesRef))
	artifactRoot := filepath.Join(repoRoot, "var", "artifacts")
	dataRoot := filepath.Join(repoRoot, "var", "data")

	mustWriteManifest(t, filepath.Join(manifestRoot, "pipelines", "external_tool.yaml"), ""+
		"id: external_tool_pipeline\n"+
		"name: External Tool Pipeline\n"+
		"description: Example pipeline.\n"+
		"owner: platform-team\n"+
		"jobs:\n"+
		"  - id: run_finance_dbt\n"+
		"    name: Build Finance Models In DBT\n"+
		"    type: external_tool\n"+
		"    timeout: 30s\n"+
		"    external_tool:\n"+
		"      tool: dbt\n"+
		"      action: build\n"+
		"      project_ref: packages/external_tools/dbt_finance_demo\n"+
		"      config_ref: packages/external_tools/dbt_finance_demo/profiles\n"+
		"      artifacts:\n"+
		"        - path: target/run_results.json\n"+
		"          required: true\n")
	mustMkdirAll(t, profilesRoot)
	mustWriteFile(t, filepath.Join(projectRoot, "dbt_project.yml"), "name: dbt_finance_demo\n")
	mustWriteFile(t, filepath.Join(profilesRoot, "profiles.yml"), "dbt_finance_demo:\n")
	scriptPath := filepath.Join(repoRoot, "fake-dbt.sh")
	mustWriteFile(t, scriptPath, "#!/bin/sh\necho 'dbt failed' >&2\nexit 7\n")
	mustChmod(t, scriptPath)

	var (
		mu      sync.Mutex
		payload map[string]any
		calls   int
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		mu.Lock()
		defer mu.Unlock()
		calls++
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode webhook payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	cfg := config.Settings{
		Environment:           "test",
		DataRoot:              dataRoot,
		ArtifactRoot:          artifactRoot,
		ManifestRoot:          manifestRoot,
		ExternalToolRoot:      repoRoot,
		DBTBinary:             scriptPath,
		ExternalToolTimeout:   30 * time.Second,
		RunFailureWebhookURLs: []string{server.URL},
		AlertWebhookTimeout:   5 * time.Second,
	}
	store := orchestration.NewInMemoryStore()
	loader := manifests.NewLoader(manifestRoot)
	runner := NewRunner(cfg, loader, store, storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))

	now := time.Now().UTC()
	run := orchestration.PipelineRun{
		ID:         "run_failure_alert",
		PipelineID: "external_tool_pipeline",
		Status:     orchestration.RunStatusQueued,
		Trigger:    "manual",
		StartedAt:  now,
		UpdatedAt:  now,
		JobRuns: []orchestration.JobRun{
			{ID: "run_finance_dbt", JobID: "run_finance_dbt", Status: orchestration.RunStatusPending},
		},
	}
	if err := store.SavePipelineRun(run); err != nil {
		t.Fatalf("save initial run: %v", err)
	}

	err := runner.Execute(context.Background(), orchestration.RunRequest{
		RunID:       run.ID,
		PipelineID:  run.PipelineID,
		Trigger:     run.Trigger,
		RequestedAt: now,
	})
	if err == nil {
		t.Fatal("expected runner to fail")
	}

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("expected 1 webhook call, got %d", calls)
	}
	if payload["event_type"] != "pipeline_run_failed" {
		t.Fatalf("expected pipeline_run_failed event type, got %#v", payload["event_type"])
	}
	runPayload, ok := payload["run"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested run payload, got %#v", payload["run"])
	}
	if runPayload["id"] != "run_failure_alert" || runPayload["job_id"] != "run_finance_dbt" {
		t.Fatalf("unexpected run webhook payload: %#v", runPayload)
	}
}

func mustWriteManifest(t *testing.T, path, contents string) {
	t.Helper()
	mustMkdirAll(t, filepath.Dir(path))
	mustWriteFile(t, path, contents)
}
