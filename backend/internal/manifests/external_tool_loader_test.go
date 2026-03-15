package manifests

import (
	"path/filepath"
	"testing"
)

func TestFileLoaderReadsExternalToolPipelineShape(t *testing.T) {
	root := t.TempDir()
	mustWriteManifest(t, filepath.Join(root, "pipelines", "external_tool.yaml"), ""+
		"id: external_tool_pipeline\n"+
		"name: External Tool Pipeline\n"+
		"description: Example pipeline.\n"+
		"owner: platform-team\n"+
		"jobs:\n"+
		"  - id: run_finance_dbt\n"+
		"    name: Build Finance Models In DBT\n"+
		"    type: external_tool\n"+
		"    external_tool:\n"+
		"      tool: dbt\n"+
		"      action: build\n"+
		"      project_ref: packages/external_tools/dbt_finance_demo\n"+
		"      config_ref: packages/external_tools/dbt_finance_demo/profiles\n"+
		"      profile: dbt_finance_demo\n"+
		"      target: dev\n"+
		"      selector: monthly_cashflow\n"+
		"      artifacts:\n"+
		"        - path: target/run_results.json\n"+
		"          required: true\n")

	loader := NewLoader(root)
	pipelines, err := loader.LoadPipelines()
	if err != nil {
		t.Fatalf("LoadPipelines returned error: %v", err)
	}
	if len(pipelines) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(pipelines))
	}
	job := pipelines[0].Jobs[0]
	if job.Type != "external_tool" {
		t.Fatalf("expected external_tool job type, got %q", job.Type)
	}
	if job.ExternalTool == nil {
		t.Fatal("expected external_tool block to decode")
	}
	if job.ExternalTool.Action != "build" {
		t.Fatalf("expected external tool action build, got %q", job.ExternalTool.Action)
	}
	if len(job.ExternalTool.Artifacts) != 1 || job.ExternalTool.Artifacts[0].Path != "target/run_results.json" {
		t.Fatalf("unexpected external tool artifacts %+v", job.ExternalTool.Artifacts)
	}
}
