package execution

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestRunExternalToolFailsForMissingBinary(t *testing.T) {
	repoRoot, projectRef, profilesRef, artifactRoot := prepareExternalToolProject(t)
	cfg := config.Settings{
		ArtifactRoot:        artifactRoot,
		ManifestRoot:        filepath.Join(repoRoot, "packages", "manifests"),
		ExternalToolRoot:    repoRoot,
		DBTBinary:           filepath.Join(repoRoot, "missing-dbt"),
		ExternalToolTimeout: 0,
	}
	runner := NewRunner(cfg, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))
	run := &orchestration.PipelineRun{ID: "run_missing_binary", PipelineID: "personal_finance_dbt_pipeline"}
	job := externalToolJob(projectRef, profilesRef)

	err := runner.runExternalTool(context.Background(), run, job, "run_missing_binary:build_finance_dbt:1")
	if err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("expected missing binary failure, got %v", err)
	}
	assertRunEventMessage(t, run, "external tool failed")
	assertRunEventField(t, run, "failure_class", "tool_unavailable")
}

func TestRunExternalToolFailsForNonZeroExitAndMirrorsLogs(t *testing.T) {
	repoRoot, projectRef, profilesRef, artifactRoot := prepareExternalToolProject(t)
	scriptPath := filepath.Join(repoRoot, "fake-dbt.sh")
	mustWriteFile(t, scriptPath, "#!/bin/sh\necho 'dbt started'\necho 'dbt warning' >&2\nexit 7\n")
	mustChmod(t, scriptPath)

	cfg := config.Settings{
		ArtifactRoot:        artifactRoot,
		ManifestRoot:        filepath.Join(repoRoot, "packages", "manifests"),
		ExternalToolRoot:    repoRoot,
		DBTBinary:           scriptPath,
		ExternalToolTimeout: 0,
	}
	runner := NewRunner(cfg, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))
	run := &orchestration.PipelineRun{ID: "run_non_zero", PipelineID: "personal_finance_dbt_pipeline"}
	job := externalToolJob(projectRef, profilesRef)

	err := runner.runExternalTool(context.Background(), run, job, "run_command_failure:build_finance_dbt:1")
	if err == nil || !strings.Contains(err.Error(), "exit code 7") {
		t.Fatalf("expected non-zero exit failure, got %v", err)
	}
	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_non_zero", "external_tools", "build_finance_dbt", "logs", "stdout.log"))
	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_non_zero", "external_tools", "build_finance_dbt", "logs", "stderr.log"))
	assertRunEventField(t, run, "failure_class", "execution_failed")
}

func TestRunExternalToolFailsWhenRequiredArtifactIsMissing(t *testing.T) {
	repoRoot, projectRef, profilesRef, artifactRoot := prepareExternalToolProject(t)
	scriptPath := filepath.Join(repoRoot, "fake-dbt.sh")
	mustWriteFile(t, scriptPath, "#!/bin/sh\necho 'dbt ran'\nexit 0\n")
	mustChmod(t, scriptPath)

	cfg := config.Settings{
		ArtifactRoot:        artifactRoot,
		ManifestRoot:        filepath.Join(repoRoot, "packages", "manifests"),
		ExternalToolRoot:    repoRoot,
		DBTBinary:           scriptPath,
		ExternalToolTimeout: 0,
	}
	runner := NewRunner(cfg, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))
	run := &orchestration.PipelineRun{ID: "run_missing_artifact", PipelineID: "personal_finance_dbt_pipeline"}
	job := externalToolJob(projectRef, profilesRef)

	err := runner.runExternalTool(context.Background(), run, job, "run_missing_artifact:build_finance_dbt:1")
	if err == nil || !strings.Contains(err.Error(), "declared artifact") {
		t.Fatalf("expected missing artifact failure, got %v", err)
	}
	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_missing_artifact", "external_tools", "build_finance_dbt", "logs", "stdout.log"))
	assertRunEventField(t, run, "failure_class", "artifact_missing")
}

func TestRunExternalToolFailsForInvalidRepoRelativeRefs(t *testing.T) {
	repoRoot, _, profilesRef, artifactRoot := prepareExternalToolProject(t)
	cfg := config.Settings{
		ArtifactRoot:        artifactRoot,
		ManifestRoot:        filepath.Join(repoRoot, "packages", "manifests"),
		ExternalToolRoot:    repoRoot,
		DBTBinary:           "dbt",
		ExternalToolTimeout: 0,
	}
	runner := NewRunner(cfg, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))
	run := &orchestration.PipelineRun{ID: "run_invalid_ref", PipelineID: "personal_finance_dbt_pipeline"}
	job := externalToolJob("../unsafe-project", profilesRef)

	err := runner.runExternalTool(context.Background(), run, job, "run_invalid_timeout:build_finance_dbt:1")
	if err == nil || !strings.Contains(err.Error(), "repo-relative") {
		t.Fatalf("expected repo-relative validation failure, got %v", err)
	}
	assertRunEventField(t, run, "failure_class", "invalid_spec")
}

func prepareExternalToolProject(t *testing.T) (string, string, string, string) {
	t.Helper()
	repoRoot := t.TempDir()
	projectRef := "packages/external_tools/dbt_finance_demo"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))
	profilesRef := "packages/external_tools/dbt_finance_demo/profiles"
	profilesRoot := filepath.Join(repoRoot, filepath.FromSlash(profilesRef))
	artifactRoot := filepath.Join(repoRoot, "var", "artifacts")

	mustMkdirAll(t, profilesRoot)
	mustWriteFile(t, filepath.Join(projectRoot, "dbt_project.yml"), "name: dbt_finance_demo\n")
	mustWriteFile(t, filepath.Join(profilesRoot, "profiles.yml"), "dbt_finance_demo:\n")
	return repoRoot, projectRef, profilesRef, artifactRoot
}

func externalToolJob(projectRef, profilesRef string) orchestration.Job {
	return orchestration.Job{
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
}

func assertRunEventMessage(t *testing.T, run *orchestration.PipelineRun, message string) {
	t.Helper()
	for _, event := range run.Events {
		if event.Message == message {
			return
		}
	}
	t.Fatalf("expected run event %q", message)
}

func assertRunEventField(t *testing.T, run *orchestration.PipelineRun, field, value string) {
	t.Helper()
	for _, event := range run.Events {
		if event.Fields[field] == value {
			return
		}
	}
	t.Fatalf("expected run event field %s=%q", field, value)
}
