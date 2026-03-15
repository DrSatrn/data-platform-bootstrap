package externaltools

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type dbtAdapter struct {
	defaultBinary string
}

func newDBTAdapter(settings Settings) Adapter {
	return dbtAdapter{defaultBinary: defaultBinary(settings.DBTBinary, "dbt")}
}

func (a dbtAdapter) Tool() string {
	return string(ToolDBT)
}

func (a dbtAdapter) Plan(request AdapterRequest) (ExecutionPlan, error) {
	spec := request.Spec
	if strings.TrimSpace(spec.Tool) == "" {
		return ExecutionPlan{}, invalidSpec(string(ToolDBT), spec.Action, fmt.Errorf("tool is required"))
	}
	if !strings.EqualFold(spec.Tool, string(ToolDBT)) {
		return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, fmt.Errorf("dbt adapter cannot plan tool %q", spec.Tool))
	}
	if strings.TrimSpace(spec.Action) == "" {
		return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, fmt.Errorf("action is required"))
	}

	repoRoot := resolveRepoRoot(request.RepoRoot, request.Settings.Root)
	projectRoot, err := resolveRepoRelativePath(repoRoot, spec.ProjectRef, "project_ref")
	if err != nil {
		return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, err)
	}
	if err := requireDirectory(projectRoot, "dbt project"); err != nil {
		return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, err)
	}
	if err := requireFile(filepath.Join(projectRoot, "dbt_project.yml"), "dbt project file"); err != nil {
		return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, err)
	}

	args := []string{strings.TrimSpace(spec.Action)}
	if strings.TrimSpace(spec.Selector) != "" {
		args = append(args, "--select", strings.TrimSpace(spec.Selector))
	}
	if strings.TrimSpace(spec.Target) != "" {
		args = append(args, "--target", strings.TrimSpace(spec.Target))
	}
	if strings.TrimSpace(spec.Profile) != "" {
		args = append(args, "--profile", strings.TrimSpace(spec.Profile))
	}
	if len(spec.Vars) > 0 {
		encodedVars, err := json.Marshal(spec.Vars)
		if err != nil {
			return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, fmt.Errorf("encode vars: %w", err))
		}
		args = append(args, "--vars", string(encodedVars))
	}
	args = append(args, spec.Args...)
	args = append(args, "--project-dir", projectRoot)

	if strings.TrimSpace(spec.ConfigRef) != "" {
		profilesDir, err := resolveRepoRelativePath(repoRoot, spec.ConfigRef, "config_ref")
		if err != nil {
			return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, err)
		}
		if err := requireDirectory(profilesDir, "dbt profiles dir"); err != nil {
			return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, err)
		}
		args = append(args, "--profiles-dir", profilesDir)
	}

	plannedArtifacts := make([]PlannedArtifact, 0, len(spec.Artifacts))
	for _, artifact := range spec.Artifacts {
		if !isRepoRelativeRef(artifact.Path) {
			return ExecutionPlan{}, invalidSpec(spec.Tool, spec.Action, fmt.Errorf("artifact path %q must be relative", artifact.Path))
		}
		plannedArtifacts = append(plannedArtifacts, PlannedArtifact{
			SourcePath:   filepath.Join(projectRoot, filepath.FromSlash(filepath.Clean(artifact.Path))),
			RelativePath: filepath.ToSlash(filepath.Join("external_tools", request.JobID, filepath.FromSlash(filepath.Clean(artifact.Path)))),
			Required:     artifact.Required,
		})
	}

	binary := strings.TrimSpace(spec.Binary)
	if binary == "" {
		binary = a.defaultBinary
	}
	return ExecutionPlan{
		Tool:      strings.ToLower(spec.Tool),
		Action:    strings.TrimSpace(spec.Action),
		WorkDir:   projectRoot,
		Binary:    binary,
		Args:      args,
		Artifacts: plannedArtifacts,
	}, nil
}

func invalidSpec(tool, action string, err error) error {
	return &ExecutionError{
		Kind:   FailureKindInvalidSpec,
		Tool:   tool,
		Action: action,
		Err:    err,
	}
}
