package orchestration

import (
	"strings"
	"testing"
)

func TestValidatePipelineRejectsDuplicateExternalToolArtifacts(t *testing.T) {
	pipeline := Pipeline{
		ID: "pipeline",
		Jobs: []Job{
			{
				ID:   "tool_job",
				Type: JobTypeExternalTool,
				ExternalTool: &ExternalToolSpec{
					Tool:       "dbt",
					Action:     "build",
					ProjectRef: "packages/external_tools/dbt_finance_demo",
					ConfigRef:  "packages/external_tools/dbt_finance_demo/profiles",
					Artifacts: []ExternalToolArtifact{
						{Path: "target/run_results.json", Required: true},
						{Path: "target/run_results.json", Required: false},
					},
				},
			},
		},
	}

	err := ValidatePipeline(pipeline)
	if err == nil || !strings.Contains(err.Error(), "duplicates") {
		t.Fatalf("expected duplicate artifact validation error, got %v", err)
	}
}
