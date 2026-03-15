// Package python defines the subprocess contract used to run bounded Python
// data tasks from the Go worker. The contract is explicit and file-based so
// operators can inspect failed requests and task outputs without needing a
// hidden RPC layer.
package python

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// TaskRequest is the JSON payload made available to Python subprocesses.
type TaskRequest struct {
	RunID          string            `json:"run_id"`
	PipelineID     string            `json:"pipeline_id"`
	JobID          string            `json:"job_id"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Command        string            `json:"command"`
	DataRoot       string            `json:"data_root"`
	ArtifactRoot   string            `json:"artifact_root"`
	SampleDataRoot string            `json:"sample_data_root"`
	SQLRoot        string            `json:"sql_root"`
	Inputs         []string          `json:"inputs"`
	Outputs        []string          `json:"outputs"`
	Labels         map[string]string `json:"labels"`
}

// TaskOutput describes one file produced by the Python task that should be
// mirrored into the run artifact namespace.
type TaskOutput struct {
	RelativePath string `json:"relative_path"`
	SourcePath   string `json:"source_path"`
	ContentType  string `json:"content_type,omitempty"`
}

// TaskResult is the structured response read back from the Python task.
type TaskResult struct {
	Message  string         `json:"message"`
	Outputs  []TaskOutput   `json:"outputs"`
	Metadata map[string]any `json:"metadata,omitempty"`
	LogLines []string       `json:"log_lines,omitempty"`
}

// Runner executes Python tasks behind the Go worker.
type Runner struct {
	taskRoot string
	binary   string
}

// NewRunner constructs a Python subprocess runner.
func NewRunner(cfg config.Settings) *Runner {
	return &Runner{
		taskRoot: cfg.PythonTaskRoot,
		binary:   cfg.PythonBinary,
	}
}

// Run executes a manifest-declared Python command and returns its structured
// result.
func (r *Runner) Run(ctx context.Context, pipelineID string, job orchestration.Job, request TaskRequest) (TaskResult, error) {
	tokens := strings.Fields(strings.TrimSpace(job.Command))
	if len(tokens) < 2 {
		return TaskResult{}, fmt.Errorf("python job %s must declare a script command such as 'python3 tasks/example.py'", job.ID)
	}

	scriptPath := r.resolveScriptPath(tokens[1])
	workDir := r.taskRoot
	requestBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return TaskResult{}, fmt.Errorf("encode python task request: %w", err)
	}

	tempDir, err := os.MkdirTemp(filepath.Join(request.DataRoot, "control_plane"), "python-task-")
	if err != nil {
		return TaskResult{}, fmt.Errorf("create python task temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	requestPath := filepath.Join(tempDir, "request.json")
	resultPath := filepath.Join(tempDir, "result.json")
	if err := os.WriteFile(requestPath, requestBytes, 0o644); err != nil {
		return TaskResult{}, fmt.Errorf("write python task request: %w", err)
	}

	args := append([]string{scriptPath}, tokens[2:]...)
	command := exec.CommandContext(ctx, r.binary, args...)
	command.Dir = workDir
	command.Env = append(os.Environ(),
		"PLATFORM_TASK_REQUEST_PATH="+requestPath,
		"PLATFORM_TASK_RESULT_PATH="+resultPath,
		"PLATFORM_PIPELINE_ID="+pipelineID,
		"PLATFORM_IDEMPOTENCY_KEY="+request.IdempotencyKey,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		return TaskResult{}, fmt.Errorf("python task %s failed: %w\n%s", job.ID, err, strings.TrimSpace(string(output)))
	}

	resultBytes, err := os.ReadFile(resultPath)
	if err != nil {
		return TaskResult{}, fmt.Errorf("read python task result for %s: %w", job.ID, err)
	}
	var result TaskResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return TaskResult{}, fmt.Errorf("decode python task result for %s: %w", job.ID, err)
	}
	if len(result.LogLines) == 0 && len(output) > 0 {
		result.LogLines = strings.Split(strings.TrimSpace(string(output)), "\n")
	}
	return result, nil
}

// RunUtility executes a repo-owned Python utility script that is not tied to a
// pipeline job. Metadata profiling uses this path so Go can keep the control
// plane while delegating data-shape inspection to Python.
func (r *Runner) RunUtility(ctx context.Context, scriptPath string, request any, result any, env map[string]string) error {
	requestBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return fmt.Errorf("encode python utility request: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "platform-python-utility-")
	if err != nil {
		return fmt.Errorf("create python utility temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	requestPath := filepath.Join(tempDir, "request.json")
	resultPath := filepath.Join(tempDir, "result.json")
	if err := os.WriteFile(requestPath, requestBytes, 0o644); err != nil {
		return fmt.Errorf("write python utility request: %w", err)
	}

	command := exec.CommandContext(ctx, r.binary, r.resolveScriptPath(scriptPath))
	command.Dir = r.taskRoot
	command.Env = append(os.Environ(),
		"PLATFORM_TASK_REQUEST_PATH="+requestPath,
		"PLATFORM_TASK_RESULT_PATH="+resultPath,
	)
	for key, value := range env {
		command.Env = append(command.Env, key+"="+value)
	}
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python utility %s failed: %w\n%s", scriptPath, err, strings.TrimSpace(string(output)))
	}

	resultBytes, err := os.ReadFile(resultPath)
	if err != nil {
		return fmt.Errorf("read python utility result for %s: %w", scriptPath, err)
	}
	if err := json.Unmarshal(resultBytes, result); err != nil {
		return fmt.Errorf("decode python utility result for %s: %w", scriptPath, err)
	}
	return nil
}

func (r *Runner) resolveScriptPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(r.taskRoot, filepath.Clean(path))
}
