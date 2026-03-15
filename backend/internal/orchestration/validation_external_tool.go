package orchestration

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validateExternalToolJob(job Job) error {
	spec := job.ExternalTool
	if spec == nil {
		return fmt.Errorf("external_tool jobs must declare an external_tool block")
	}
	if strings.TrimSpace(job.Command) != "" {
		return fmt.Errorf("external_tool jobs must not set command")
	}
	if strings.TrimSpace(job.TransformRef) != "" {
		return fmt.Errorf("external_tool jobs must not set transform_ref")
	}
	if strings.TrimSpace(spec.Tool) == "" {
		return fmt.Errorf("external_tool.tool is required")
	}
	if strings.TrimSpace(spec.Action) == "" {
		return fmt.Errorf("external_tool.action is required")
	}
	if !isRepoRelativeRef(spec.ProjectRef) {
		return fmt.Errorf("external_tool.project_ref must be repo-relative")
	}
	if !isRepoRelativeRef(spec.ConfigRef) {
		return fmt.Errorf("external_tool.config_ref must be repo-relative")
	}
	if len(spec.Artifacts) == 0 {
		return fmt.Errorf("external_tool.artifacts must declare at least one artifact")
	}
	seenArtifacts := make(map[string]struct{}, len(spec.Artifacts))
	for index, artifact := range spec.Artifacts {
		if !isRepoRelativeRef(artifact.Path) {
			return fmt.Errorf("external_tool.artifacts[%d].path must be relative", index)
		}
		cleanPath := filepath.Clean(strings.TrimSpace(artifact.Path))
		if _, exists := seenArtifacts[cleanPath]; exists {
			return fmt.Errorf("external_tool.artifacts[%d].path duplicates %q", index, cleanPath)
		}
		seenArtifacts[cleanPath] = struct{}{}
	}

	switch strings.ToLower(strings.TrimSpace(spec.Tool)) {
	case "dbt":
		return nil
	case "dlt", "pyspark":
		return fmt.Errorf("external tool %q is declared in the contract but not implemented yet", spec.Tool)
	default:
		return fmt.Errorf("external tool %q is not supported", spec.Tool)
	}
}

func isRepoRelativeRef(path string) bool {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "." || clean == "" {
		return false
	}
	if filepath.IsAbs(clean) {
		return false
	}
	return !strings.HasPrefix(clean, "..")
}
