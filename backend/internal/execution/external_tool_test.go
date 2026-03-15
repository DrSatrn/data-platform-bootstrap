package execution

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestRunExternalToolMirrorsLogsAndArtifacts(t *testing.T) {
	repoRoot := t.TempDir()
	projectRef := "packages/external_tools/dbt_finance_demo"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))
	profilesRef := "packages/external_tools/dbt_finance_demo/profiles"
	profilesRoot := filepath.Join(repoRoot, filepath.FromSlash(profilesRef))
	artifactRoot := filepath.Join(repoRoot, "var", "artifacts")

	mustMkdirAll(t, profilesRoot)
	mustWriteFile(t, filepath.Join(projectRoot, "dbt_project.yml"), "name: dbt_finance_demo\n")
	mustWriteFile(t, filepath.Join(profilesRoot, "profiles.yml"), "dbt_finance_demo:\n")

	scriptPath := filepath.Join(repoRoot, "fake-dbt.sh")
	mustWriteFile(t, scriptPath, "#!/bin/sh\nproject_dir=\"\"\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"--project-dir\" ]; then\n    shift\n    project_dir=\"$1\"\n    break\n  fi\n  shift\ndone\necho 'dbt started'\necho 'dbt warning' >&2\nmkdir -p \"$project_dir/target\"\nprintf '{\"ok\":true}' > \"$project_dir/target/run_results.json\"\nprintf '{\"nodes\":[]}' > \"$project_dir/target/manifest.json\"\n")
	mustChmod(t, scriptPath)

	cfg := config.Settings{
		ArtifactRoot:        artifactRoot,
		ManifestRoot:        filepath.Join(repoRoot, "packages", "manifests"),
		ExternalToolRoot:    repoRoot,
		DBTBinary:           scriptPath,
		ExternalToolTimeout: 0,
	}
	runner := NewRunner(cfg, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))
	run := &orchestration.PipelineRun{
		ID:         "run_1",
		PipelineID: "personal_finance_dbt_pipeline",
	}
	job := orchestration.Job{
		ID:      "build_finance_dbt",
		Type:    orchestration.JobTypeExternalTool,
		Timeout: "30s",
		ExternalTool: &orchestration.ExternalToolSpec{
			Tool:       "dbt",
			Action:     "build",
			ProjectRef: projectRef,
			ConfigRef:  profilesRef,
			Artifacts: []orchestration.ExternalToolArtifact{
				{Path: "target/run_results.json", Required: true},
				{Path: "target/manifest.json", Required: true},
			},
		},
	}

	if err := runner.runExternalTool(context.Background(), run, job, "run_1:build_finance_dbt:1"); err != nil {
		t.Fatalf("runExternalTool returned error: %v", err)
	}

	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_1", "external_tools", "build_finance_dbt", "logs", "stdout.log"))
	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_1", "external_tools", "build_finance_dbt", "logs", "stderr.log"))
	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_1", "external_tools", "build_finance_dbt", "target", "run_results.json"))
	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_1", "external_tools", "build_finance_dbt", "target", "manifest.json"))
	if len(run.Events) == 0 {
		t.Fatal("expected external tool run events to be recorded")
	}
}

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	mustMkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustChmod(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("chmod %s: %v", path, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}
