package execution

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestExternalToolOutputsAreInspectableThroughStorageService(t *testing.T) {
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

	service := storage.NewService(artifactRoot, nil)
	cfg := config.Settings{
		ArtifactRoot:        artifactRoot,
		ManifestRoot:        filepath.Join(repoRoot, "packages", "manifests"),
		ExternalToolRoot:    repoRoot,
		DBTBinary:           scriptPath,
		ExternalToolTimeout: 0,
	}
	runner := NewRunner(cfg, nil, orchestration.NewInMemoryStore(), service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	run := &orchestration.PipelineRun{
		ID:         "run_inspectable",
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

	if err := runner.runExternalTool(context.Background(), run, job); err != nil {
		t.Fatalf("runExternalTool returned error: %v", err)
	}

	artifacts, err := service.ListRunArtifacts(run.ID)
	if err != nil {
		t.Fatalf("ListRunArtifacts returned error: %v", err)
	}
	if len(artifacts) != 4 {
		t.Fatalf("expected 4 inspectable artifacts, got %d", len(artifacts))
	}

	stdoutBytes, err := service.ReadRunArtifact(run.ID, "external_tools/build_finance_dbt/logs/stdout.log")
	if err != nil {
		t.Fatalf("ReadRunArtifact stdout.log returned error: %v", err)
	}
	if string(stdoutBytes) != "dbt started\n" {
		t.Fatalf("unexpected stdout artifact contents %q", string(stdoutBytes))
	}

	runResultsBytes, err := service.ReadRunArtifact(run.ID, "external_tools/build_finance_dbt/target/run_results.json")
	if err != nil {
		t.Fatalf("ReadRunArtifact run_results.json returned error: %v", err)
	}
	if string(runResultsBytes) != "{\"ok\":true}" {
		t.Fatalf("unexpected run_results artifact contents %q", string(runResultsBytes))
	}
}
