// This test suite checks the core manifest invariants that protect the
// scheduler and worker from invalid DAG definitions.
package orchestration

import "testing"

func TestValidatePipeline(t *testing.T) {
	valid := Pipeline{
		ID: "pipeline",
		Jobs: []Job{
			{ID: "extract"},
			{ID: "transform", DependsOn: []string{"extract"}},
		},
	}

	if err := ValidatePipeline(valid); err != nil {
		t.Fatalf("expected valid pipeline, got error: %v", err)
	}
}

func TestValidatePipelineRejectsUnknownDependency(t *testing.T) {
	pipeline := Pipeline{
		ID: "pipeline",
		Jobs: []Job{
			{ID: "transform", DependsOn: []string{"missing"}},
		},
	}

	if err := ValidatePipeline(pipeline); err == nil {
		t.Fatal("expected unknown dependency validation error")
	}
}

func TestValidatePipelineRejectsDuplicateJobs(t *testing.T) {
	pipeline := Pipeline{
		ID: "pipeline",
		Jobs: []Job{
			{ID: "duplicate"},
			{ID: "duplicate"},
		},
	}

	if err := ValidatePipeline(pipeline); err == nil {
		t.Fatal("expected duplicate job validation error")
	}
}
