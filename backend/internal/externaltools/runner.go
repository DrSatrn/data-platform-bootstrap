package externaltools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Runner executes external pipeline tools as bounded subprocesses.
type Runner struct {
	settings Settings
	adapters map[string]Adapter
}

// NewRunner constructs a new external tool runner using the default adapter
// registry for this slice.
func NewRunner(settings Settings) *Runner {
	return newRunner(settings,
		newDBTAdapter(settings),
		reservedAdapter{tool: string(ToolDLT)},
		reservedAdapter{tool: string(ToolPySpark)},
	)
}

// NewRunnerWithAdapters constructs a runner with explicit adapters, which is
// mainly useful for package-level tests.
func NewRunnerWithAdapters(adapters ...Adapter) *Runner {
	return newRunner(Settings{}, adapters...)
}

func newRunner(settings Settings, adapters ...Adapter) *Runner {
	registry := make(map[string]Adapter, len(adapters))
	for _, adapter := range adapters {
		registry[strings.ToLower(strings.TrimSpace(adapter.Tool()))] = adapter
	}
	return &Runner{
		settings: settings,
		adapters: registry,
	}
}

// Validate checks whether a spec is structurally and adapter-wise valid.
func (r *Runner) Validate(spec RunRequest) error {
	_, err := r.plan(spec)
	return err
}

// Run executes one external tool job and returns structured outputs.
func (r *Runner) Run(ctx context.Context, request RunRequest) (Result, error) {
	plan, err := r.plan(request)
	if err != nil {
		return Result{
			Tool:         strings.ToLower(strings.TrimSpace(request.Spec.Tool)),
			Action:       strings.TrimSpace(request.Spec.Action),
			FailureClass: failureClass(err),
		}, err
	}

	runCtx := ctx
	cancel := func() {}
	timeout := request.Timeout
	if timeout <= 0 {
		timeout = r.settings.DefaultTimeout
	}
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	command := exec.CommandContext(runCtx, plan.Binary, plan.Args...)
	command.Dir = plan.WorkDir
	command.Env = append(command.Environ(), plan.Env...)
	commandLine := strings.Join(append([]string{plan.Binary}, plan.Args...), " ")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	result := Result{
		Tool:    plan.Tool,
		Action:  plan.Action,
		Command: append([]string{plan.Binary}, plan.Args...),
		Workdir: plan.WorkDir,
		Events: []Event{
			{
				Level:   "info",
				Message: "external tool selected",
				Fields: map[string]string{
					"tool":   plan.Tool,
					"job_id": request.JobID,
					"action": plan.Action,
				},
			},
			{
				Level:   "info",
				Message: "external tool command started",
				Fields: map[string]string{
					"tool":    plan.Tool,
					"job_id":  request.JobID,
					"command": commandLine,
				},
			},
		},
	}

	err = command.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.LogLines = collectLogLines(result.Stdout, result.Stderr)

	if err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			err = &ExecutionError{
				Kind:   FailureKindExecutionFailed,
				Tool:   plan.Tool,
				Action: plan.Action,
				Err:    runCtx.Err(),
			}
		} else {
			err = mapCommandError(plan.Tool, plan.Action, err)
		}
		result.ExitCode = exitCode(err)
		result.FailureClass = failureClass(err)
		result.Events = append(result.Events, Event{
			Level:   "error",
			Message: "external tool command finished",
			Fields: map[string]string{
				"tool":          plan.Tool,
				"job_id":        request.JobID,
				"failure_class": result.FailureClass,
				"exit_code":     fmt.Sprintf("%d", result.ExitCode),
			},
		})
		return result, err
	}

	artifacts, artifactErr := collectArtifacts(plan.Artifacts)
	if artifactErr != nil {
		result.FailureClass = failureClass(artifactErr)
		result.Events = append(result.Events, Event{
			Level:   "error",
			Message: "external tool artifact discovery failed",
			Fields: map[string]string{
				"tool":          plan.Tool,
				"job_id":        request.JobID,
				"failure_class": result.FailureClass,
			},
		})
		return result, artifactErr
	}

	result.ExitCode = 0
	result.Artifacts = artifacts
	result.Events = append(result.Events,
		Event{
			Level:   "info",
			Message: "external tool command finished",
			Fields: map[string]string{
				"tool":      plan.Tool,
				"job_id":    request.JobID,
				"exit_code": "0",
			},
		},
		Event{
			Level:   "info",
			Message: "external tool artifact discovery summary",
			Fields: map[string]string{
				"tool":           plan.Tool,
				"job_id":         request.JobID,
				"artifact_count": fmt.Sprintf("%d", len(artifacts)),
			},
		},
	)
	return result, nil
}

func (r *Runner) plan(request RunRequest) (ExecutionPlan, error) {
	tool := strings.ToLower(strings.TrimSpace(request.Spec.Tool))
	adapter, found := r.adapters[tool]
	if !found {
		return ExecutionPlan{}, &ExecutionError{
			Kind:   FailureKindUnsupportedTool,
			Tool:   tool,
			Action: request.Spec.Action,
			Err:    fmt.Errorf("no adapter registered"),
		}
	}
	return adapter.Plan(AdapterRequest{
		RepoRoot: resolveRepoRoot(request.RepoRoot, r.settings.Root),
		RunID:    request.RunID,
		JobID:    request.JobID,
		Settings: r.settings,
		Spec:     request.Spec,
	})
}

func collectArtifacts(planned []PlannedArtifact) ([]Artifact, error) {
	artifacts := make([]Artifact, 0, len(planned))
	for _, artifact := range planned {
		if err := requireArtifact(artifact); err != nil {
			return nil, err
		}
		if strings.TrimSpace(artifact.SourcePath) == "" {
			continue
		}
		if _, err := os.Stat(artifact.SourcePath); err == nil {
			artifacts = append(artifacts, Artifact{
				SourcePath:   artifact.SourcePath,
				RelativePath: filepathToSlash(artifact.RelativePath),
			})
		}
	}
	return artifacts, nil
}

func requireArtifact(artifact PlannedArtifact) error {
	if !artifact.Required {
		return nil
	}
	if strings.TrimSpace(artifact.SourcePath) == "" {
		return &ExecutionError{
			Kind: FailureKindArtifactMissing,
			Err:  fmt.Errorf("required artifact source path is empty"),
		}
	}
	if _, err := os.Stat(artifact.SourcePath); err != nil {
		return &ExecutionError{
			Kind: FailureKindArtifactMissing,
			Err:  err,
		}
	}
	return nil
}

func collectLogLines(stdout, stderr string) []string {
	lines := []string{}
	for _, chunk := range []string{stdout, stderr} {
		for _, line := range strings.Split(chunk, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func exitCode(err error) int {
	var execErr *ExecutionError
	if errors.As(err, &execErr) && execErr.ExitCode != 0 {
		return execErr.ExitCode
	}
	return 1
}

func failureClass(err error) string {
	var execErr *ExecutionError
	if errors.As(err, &execErr) {
		return string(execErr.Kind)
	}
	return "execution_failed"
}

func filepathToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
