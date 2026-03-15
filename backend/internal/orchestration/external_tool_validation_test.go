package orchestration

import (
	"strings"
	"testing"
)

func TestValidatePipelineAcceptsExternalToolDBTJob(t *testing.T) {
	pipeline := Pipeline{
		ID: "pipeline",
		Jobs: []Job{
			{
				ID:   "build_models",
				Type: JobTypeExternalTool,
				ExternalTool: &ExternalToolSpec{
					Tool:       "dbt",
					Action:     "build",
					ProjectRef: "packages/external_tools/dbt/finance",
					ConfigRef:  "packages/external_tools/dbt/profiles",
					Artifacts: []ExternalToolArtifact{
						{Path: "target/run_results.json", Required: true},
					},
				},
			},
		},
	}

	if err := ValidatePipeline(pipeline); err != nil {
		t.Fatalf("expected valid pipeline, got error: %v", err)
	}
}

func TestValidatePipelineRejectsMissingExternalToolBlock(t *testing.T) {
	pipeline := Pipeline{
		ID:   "pipeline",
		Jobs: []Job{{ID: "tool_job", Type: JobTypeExternalTool}},
	}

	err := ValidatePipeline(pipeline)
	if err == nil || !strings.Contains(err.Error(), "external_tool block") {
		t.Fatalf("expected missing external_tool block error, got %v", err)
	}
}

func TestValidatePipelineRejectsUnimplementedExternalTool(t *testing.T) {
	pipeline := Pipeline{
		ID: "pipeline",
		Jobs: []Job{
			{
				ID:   "dlt_job",
				Type: JobTypeExternalTool,
				ExternalTool: &ExternalToolSpec{
					Tool:       "dlt",
					Action:     "run",
					ProjectRef: "packages/external_tools/dlt/example",
					ConfigRef:  "packages/external_tools/dlt/config",
					Artifacts: []ExternalToolArtifact{
						{Path: "artifacts/state.json", Required: true},
					},
				},
			},
		},
	}

	err := ValidatePipeline(pipeline)
	if err == nil || !strings.Contains(err.Error(), "not implemented yet") {
		t.Fatalf("expected gated external tool error, got %v", err)
	}
}

func TestValidatePipelineRejectsNonRelativeExternalToolRefs(t *testing.T) {
	pipeline := Pipeline{
		ID: "pipeline",
		Jobs: []Job{
			{
				ID:   "tool_job",
				Type: JobTypeExternalTool,
				ExternalTool: &ExternalToolSpec{
					Tool:       "dbt",
					Action:     "build",
					ProjectRef: "/tmp/project",
					ConfigRef:  "packages/external_tools/dbt/profiles",
					Artifacts: []ExternalToolArtifact{
						{Path: "target/run_results.json", Required: true},
					},
				},
			},
		},
	}

	err := ValidatePipeline(pipeline)
	if err == nil || !strings.Contains(err.Error(), "project_ref") {
		t.Fatalf("expected repo-relative project_ref error, got %v", err)
	}
}
