package externaltools

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// FailureKind classifies external tool failures into operator-meaningful
// categories so the worker can report them consistently.
type FailureKind string

const (
	FailureKindUnsupportedTool FailureKind = "unsupported_tool"
	FailureKindInvalidSpec     FailureKind = "invalid_spec"
	FailureKindUnavailable     FailureKind = "tool_unavailable"
	FailureKindExecutionFailed FailureKind = "execution_failed"
	FailureKindArtifactMissing FailureKind = "artifact_missing"
)

const (
	// Public failure classes are the stable operator-facing values surfaced in
	// run events, logs, and docs. They intentionally preserve the wording used
	// in the runbook so operators do not need to translate internal error kinds.
	publicFailureConfigError   = "config_error"
	publicFailureMissingBinary = "missing_binary"
	publicFailureExitNonZero   = "exit_non_zero"
	publicFailureTimeout       = "timeout"
	publicFailureNotImplemented = "not_implemented"
)

// ExecutionError is the structured error returned by the generic runner.
type ExecutionError struct {
	Kind     FailureKind
	Tool     string
	Action   string
	ExitCode int
	Err      error
}

func (e *ExecutionError) Error() string {
	if e == nil {
		return ""
	}
	message := fmt.Sprintf("external tool %s", e.Tool)
	if e.Action != "" {
		message += " " + e.Action
	}
	switch e.Kind {
	case FailureKindUnsupportedTool:
		return message + " is unsupported"
	case FailureKindInvalidSpec:
		return message + " has an invalid configuration: " + e.Err.Error()
	case FailureKindUnavailable:
		return message + " is unavailable: " + e.Err.Error()
	case FailureKindExecutionFailed:
		if e.ExitCode > 0 {
			return fmt.Sprintf("%s failed with exit code %d: %s", message, e.ExitCode, e.Err.Error())
		}
		return message + " failed: " + e.Err.Error()
	case FailureKindArtifactMissing:
		return message + " did not produce a declared artifact: " + e.Err.Error()
	default:
		return message + " failed: " + e.Err.Error()
	}
}

func (e *ExecutionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func mapCommandError(tool, action string, err error) error {
	var execErr *exec.Error
	if errors.As(err, &execErr) {
		return &ExecutionError{
			Kind:   FailureKindUnavailable,
			Tool:   tool,
			Action: action,
			Err:    execErr,
		}
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return &ExecutionError{
			Kind:   FailureKindUnavailable,
			Tool:   tool,
			Action: action,
			Err:    pathErr,
		}
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &ExecutionError{
			Kind:     FailureKindExecutionFailed,
			Tool:     tool,
			Action:   action,
			ExitCode: exitErr.ExitCode(),
			Err:      exitErr,
		}
	}

	return &ExecutionError{
		Kind:   FailureKindExecutionFailed,
		Tool:   tool,
		Action: action,
		Err:    err,
	}
}
