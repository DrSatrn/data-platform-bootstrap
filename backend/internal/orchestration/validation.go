// This file contains basic pipeline validation logic. Validating manifests
// early prevents entire classes of scheduler and worker failures that are
// otherwise expensive to diagnose at runtime.
package orchestration

import "fmt"

// ValidatePipeline checks for the most important DAG and metadata invariants.
func ValidatePipeline(pipeline Pipeline) error {
	if pipeline.ID == "" {
		return fmt.Errorf("pipeline id is required")
	}

	if len(pipeline.Jobs) == 0 {
		return fmt.Errorf("pipeline must contain at least one job")
	}

	seen := make(map[string]struct{}, len(pipeline.Jobs))
	for _, job := range pipeline.Jobs {
		if job.ID == "" {
			return fmt.Errorf("job id is required")
		}
		if err := validateJob(job); err != nil {
			return fmt.Errorf("job %q is invalid: %w", job.ID, err)
		}
		if _, exists := seen[job.ID]; exists {
			return fmt.Errorf("duplicate job id %q", job.ID)
		}
		seen[job.ID] = struct{}{}
	}

	for _, job := range pipeline.Jobs {
		for _, dependency := range job.DependsOn {
			if _, exists := seen[dependency]; !exists {
				return fmt.Errorf("job %q depends on unknown job %q", job.ID, dependency)
			}
			if dependency == job.ID {
				return fmt.Errorf("job %q cannot depend on itself", job.ID)
			}
		}
	}

	return nil
}
