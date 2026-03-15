package externaltools

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func TestDBTAdapterBuildsCommand(t *testing.T) {
	repoRoot := t.TempDir()
	projectRef := "packages/external_tools/dbt_finance_demo"
	configRef := "packages/external_tools/dbt_finance_demo/profiles"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))

	mustMkdirAll(t, filepath.Join(projectRoot, "profiles"))
	mustWriteFile(t, filepath.Join(projectRoot, "dbt_project.yml"), "name: finance_demo\n")
	mustWriteFile(t, filepath.Join(projectRoot, "profiles", "profiles.yml"), "finance_demo:\n")

	adapter := dbtAdapter{defaultBinary: "dbt-bin"}
	plan, err := adapter.Plan(AdapterRequest{
		RepoRoot: repoRoot,
		JobID:    "build_finance_dbt",
		Spec: orchestration.ExternalToolSpec{
			Tool:       "dbt",
			Action:     "build",
			ProjectRef: projectRef,
			ConfigRef:  configRef,
			Profile:    "dbt_finance_demo",
			Target:     "dev",
			Selector:   "monthly_cashflow",
			Args:       []string{"--fail-fast"},
			Vars: map[string]any{
				"schema": "analytics",
			},
			Artifacts: []orchestration.ExternalToolArtifact{
				{Path: "target/run_results.json", Required: true},
			},
		},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	expectedArgs := []string{
		"build",
		"--select", "monthly_cashflow",
		"--target", "dev",
		"--profile", "dbt_finance_demo",
		"--vars", `{"schema":"analytics"}`,
		"--fail-fast",
		"--project-dir", projectRoot,
		"--profiles-dir", filepath.Join(repoRoot, filepath.FromSlash(configRef)),
	}
	if !reflect.DeepEqual(plan.Args, expectedArgs) {
		t.Fatalf("expected args %v, got %v", expectedArgs, plan.Args)
	}
	if plan.Binary != "dbt-bin" {
		t.Fatalf("expected binary dbt-bin, got %q", plan.Binary)
	}
	if plan.WorkDir != projectRoot {
		t.Fatalf("expected workdir %q, got %q", projectRoot, plan.WorkDir)
	}
	if len(plan.Artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(plan.Artifacts))
	}
	if plan.Artifacts[0].RelativePath != "external_tools/build_finance_dbt/target/run_results.json" {
		t.Fatalf("unexpected artifact path %q", plan.Artifacts[0].RelativePath)
	}
}

func TestDBTAdapterRejectsMissingProjectFile(t *testing.T) {
	repoRoot := t.TempDir()
	projectRef := "packages/external_tools/dbt_finance_demo"
	configRef := "packages/external_tools/dbt_finance_demo/profiles"
	projectRoot := filepath.Join(repoRoot, filepath.FromSlash(projectRef))

	mustMkdirAll(t, filepath.Join(projectRoot, "profiles"))
	mustWriteFile(t, filepath.Join(projectRoot, "profiles", "profiles.yml"), "finance_demo:\n")

	adapter := dbtAdapter{defaultBinary: "dbt"}
	_, err := adapter.Plan(AdapterRequest{
		RepoRoot: repoRoot,
		Spec: orchestration.ExternalToolSpec{
			Tool:       "dbt",
			Action:     "build",
			ProjectRef: projectRef,
			ConfigRef:  configRef,
		},
	})
	if err == nil {
		t.Fatal("expected missing dbt_project.yml validation error")
	}
}
