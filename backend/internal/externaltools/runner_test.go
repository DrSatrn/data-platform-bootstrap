package externaltools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func TestRunnerCapturesDeclaredArtifacts(t *testing.T) {
	repoRoot := t.TempDir()
	projectRef := "packages/external_tools/dbt_finance_demo"
	configRef := "packages/external_tools/dbt_finance_demo/profiles"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))
	binaryPath := filepath.Join(repoRoot, "fake-dbt")

	mustMkdirAll(t, filepath.Join(projectRoot, "profiles"))
	mustWriteFile(t, filepath.Join(projectRoot, "dbt_project.yml"), "name: finance_demo\n")
	mustWriteFile(t, filepath.Join(projectRoot, "profiles", "profiles.yml"), "finance_demo:\n")
	mustWriteExecutable(t, binaryPath, `#!/bin/sh
project_dir=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--project-dir" ]; then
    project_dir="$2"
    shift 2
    continue
  fi
  shift
done
mkdir -p "$project_dir/target"
printf '{"status":"ok"}' > "$project_dir/target/run_results.json"
printf '{"nodes":{}}' > "$project_dir/target/manifest.json"
`)

	runner := NewRunner(Settings{Root: repoRoot, DBTBinary: binaryPath})
	result, err := runner.Run(context.Background(), RunRequest{
		RunID:    "run_123",
		JobID:    "build_finance_dbt",
		RepoRoot: repoRoot,
		Spec: orchestration.ExternalToolSpec{
			Tool:       "dbt",
			Action:     "build",
			ProjectRef: projectRef,
			ConfigRef:  configRef,
			Artifacts: []orchestration.ExternalToolArtifact{
				{Path: "target/run_results.json", Required: true},
				{Path: "target/manifest.json", Required: false},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected successful run, got %v", err)
	}
	if len(result.Artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(result.Artifacts))
	}
	if result.Artifacts[0].RelativePath != "external_tools/build_finance_dbt/target/run_results.json" {
		t.Fatalf("unexpected artifact path %q", result.Artifacts[0].RelativePath)
	}
}

func TestRunnerMapsMissingBinaryFailure(t *testing.T) {
	repoRoot := t.TempDir()
	projectRef := "packages/external_tools/dbt_finance_demo"
	configRef := "packages/external_tools/dbt_finance_demo/profiles"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))

	mustMkdirAll(t, filepath.Join(projectRoot, "profiles"))
	mustWriteFile(t, filepath.Join(projectRoot, "dbt_project.yml"), "name: finance_demo\n")
	mustWriteFile(t, filepath.Join(projectRoot, "profiles", "profiles.yml"), "finance_demo:\n")

	runner := NewRunner(Settings{Root: repoRoot, DBTBinary: filepath.Join(repoRoot, "missing-dbt")})
	result, err := runner.Run(context.Background(), RunRequest{
		JobID: "build_finance_dbt",
		Spec: orchestration.ExternalToolSpec{
			Tool:       "dbt",
			Action:     "build",
			ProjectRef: projectRef,
			ConfigRef:  configRef,
		},
	})
	if err == nil {
		t.Fatal("expected missing binary error")
	}
	if result.FailureClass != "tool_unavailable" {
		t.Fatalf("expected tool_unavailable, got %q", result.FailureClass)
	}
}

func TestRunnerMapsExitFailureAndCapturesLogs(t *testing.T) {
	repoRoot := t.TempDir()
	projectRef := "packages/external_tools/dbt_finance_demo"
	configRef := "packages/external_tools/dbt_finance_demo/profiles"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))
	binaryPath := filepath.Join(repoRoot, "fake-dbt")

	mustMkdirAll(t, filepath.Join(projectRoot, "profiles"))
	mustWriteFile(t, filepath.Join(projectRoot, "dbt_project.yml"), "name: finance_demo\n")
	mustWriteFile(t, filepath.Join(projectRoot, "profiles", "profiles.yml"), "finance_demo:\n")
	mustWriteExecutable(t, binaryPath, `#!/bin/sh
echo first line
echo second line
exit 7
`)

	runner := NewRunner(Settings{Root: repoRoot, DBTBinary: binaryPath})
	result, err := runner.Run(context.Background(), RunRequest{
		JobID: "build_finance_dbt",
		Spec: orchestration.ExternalToolSpec{
			Tool:       "dbt",
			Action:     "build",
			ProjectRef: projectRef,
			ConfigRef:  configRef,
		},
	})
	if err == nil {
		t.Fatal("expected exit error")
	}
	if result.FailureClass != "execution_failed" {
		t.Fatalf("expected execution_failed, got %q", result.FailureClass)
	}
	if result.ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", result.ExitCode)
	}
	if strings.Join(result.LogLines, "|") != "first line|second line" {
		t.Fatalf("unexpected log lines %v", result.LogLines)
	}
}

func TestRunnerFailsWhenRequiredArtifactIsMissing(t *testing.T) {
	repoRoot := t.TempDir()
	projectRef := "packages/external_tools/dbt_finance_demo"
	configRef := "packages/external_tools/dbt_finance_demo/profiles"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))
	binaryPath := filepath.Join(repoRoot, "fake-dbt")

	mustMkdirAll(t, filepath.Join(projectRoot, "profiles"))
	mustWriteFile(t, filepath.Join(projectRoot, "dbt_project.yml"), "name: finance_demo\n")
	mustWriteFile(t, filepath.Join(projectRoot, "profiles", "profiles.yml"), "finance_demo:\n")
	mustWriteExecutable(t, binaryPath, "#!/bin/sh\nexit 0\n")

	runner := NewRunner(Settings{Root: repoRoot, DBTBinary: binaryPath})
	result, err := runner.Run(context.Background(), RunRequest{
		JobID: "build_finance_dbt",
		Spec: orchestration.ExternalToolSpec{
			Tool:       "dbt",
			Action:     "build",
			ProjectRef: projectRef,
			ConfigRef:  configRef,
			Artifacts: []orchestration.ExternalToolArtifact{
				{Path: "target/run_results.json", Required: true},
			},
		},
	})
	if err == nil {
		t.Fatal("expected artifact capture failure")
	}
	if result.FailureClass != "artifact_missing" {
		t.Fatalf("expected artifact_missing, got %q", result.FailureClass)
	}
}

func mustWriteExecutable(t *testing.T, path, content string) {
	t.Helper()
	mustWriteFile(t, path, content)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("chmod %s: %v", path, err)
	}
}
