package externaltools

import (
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// ToolName identifies a supported external pipeline tool adapter.
type ToolName string

const (
	ToolDBT     ToolName = "dbt"
	ToolDLT     ToolName = "dlt"
	ToolPySpark ToolName = "pyspark"
)

// Settings contains the optional runtime configuration for external tools.
type Settings struct {
	Root           string
	DBTBinary      string
	DLTBinary      string
	PySparkBinary  string
	DefaultTimeout time.Duration
}

// RunRequest contains the context needed to execute one external tool job.
type RunRequest struct {
	RepoRoot       string
	RunID          string
	PipelineID     string
	JobID          string
	IdempotencyKey string
	Timeout        time.Duration
	Spec           orchestration.ExternalToolSpec
}

// Artifact describes one discovered output file after a successful run.
type Artifact struct {
	SourcePath   string
	RelativePath string
}

// Event describes a structured execution event suitable for run histories.
type Event struct {
	Level   string
	Message string
	Fields  map[string]string
}

// Result captures the bounded process outcome and discovered artifacts.
type Result struct {
	Tool         string
	Action       string
	Command      []string
	Workdir      string
	Stdout       string
	Stderr       string
	LogLines     []string
	ExitCode     int
	Artifacts    []Artifact
	Events       []Event
	FailureClass string
}

// AdapterRequest is the adapter-facing execution contract.
type AdapterRequest struct {
	RepoRoot string
	RunID    string
	JobID    string
	Settings Settings
	Spec     orchestration.ExternalToolSpec
}

// PlannedArtifact is an adapter-declared file that should be mirrored after a
// successful command finishes.
type PlannedArtifact struct {
	SourcePath   string
	RelativePath string
	Required     bool
}

// ExecutionPlan is the subprocess-ready output of one adapter.
type ExecutionPlan struct {
	Tool      string
	Action    string
	WorkDir   string
	Binary    string
	Args      []string
	Env       []string
	Artifacts []PlannedArtifact
}

// Adapter builds an execution plan for one supported tool.
type Adapter interface {
	Tool() string
	Plan(AdapterRequest) (ExecutionPlan, error)
}
