// Package orchestration defines the durable and in-memory control-plane models
// for pipelines, jobs, runs, and run events. The types are intentionally
// explicit so state transitions remain easy to validate and document.
package orchestration

import "time"

// Pipeline describes a logical workflow made of dependency-linked jobs.
type Pipeline struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	Owner       string    `json:"owner" yaml:"owner"`
	Tags        []string  `json:"tags" yaml:"tags"`
	Schedule    Schedule  `json:"schedule" yaml:"schedule"`
	Jobs        []Job     `json:"jobs" yaml:"jobs"`
	CreatedAt   time.Time `json:"created_at"`
}

// Schedule captures how and when a pipeline should be released.
type Schedule struct {
	Cron        string `json:"cron" yaml:"cron"`
	Timezone    string `json:"timezone" yaml:"timezone"`
	Catchup     bool   `json:"catchup" yaml:"catchup"`
	IsPaused    bool   `json:"is_paused" yaml:"is_paused"`
	Description string `json:"description" yaml:"description"`
}

// Job is the atomic execution unit inside a pipeline DAG.
type Job struct {
	ID           string            `json:"id" yaml:"id"`
	Name         string            `json:"name" yaml:"name"`
	Type         JobType           `json:"type" yaml:"type"`
	DependsOn    []string          `json:"depends_on" yaml:"depends_on"`
	Retries      int               `json:"retries" yaml:"retries"`
	Timeout      string            `json:"timeout" yaml:"timeout"`
	Command      string            `json:"command" yaml:"command"`
	TransformRef string            `json:"transform_ref" yaml:"transform_ref"`
	Inputs       []string          `json:"inputs" yaml:"inputs"`
	Outputs      []string          `json:"outputs" yaml:"outputs"`
	Labels       map[string]string `json:"labels" yaml:"labels"`
	Ingest       *IngestSpec       `json:"ingest,omitempty" yaml:"ingest,omitempty"`
	Bootstrap    []BootstrapSpec   `json:"bootstrap,omitempty" yaml:"bootstrap,omitempty"`
	MetricRefs   []string          `json:"metric_refs,omitempty" yaml:"metric_refs,omitempty"`
	ExternalTool *ExternalToolSpec `json:"external_tool,omitempty" yaml:"external_tool,omitempty"`
}

// JobType distinguishes the supported execution strategies.
type JobType string

const (
	JobTypeIngest        JobType = "ingest"
	JobTypeTransformSQL  JobType = "transform_sql"
	JobTypeTransformPy   JobType = "transform_python"
	JobTypeQualityCheck  JobType = "quality_check"
	JobTypePublishMetric JobType = "publish_metric"
	JobTypeExternalTool  JobType = "external_tool"
)

// IngestSpec declares the physical source and landing target for an ingest
// job. Paths stay explicit in the manifest so new domains do not require
// runner code changes.
type IngestSpec struct {
	SourceRef     string `json:"source_ref" yaml:"source_ref"`
	TargetPath    string `json:"target_path" yaml:"target_path"`
	ArtifactPath  string `json:"artifact_path,omitempty" yaml:"artifact_path,omitempty"`
	SourceKind    string `json:"source_kind,omitempty" yaml:"source_kind,omitempty"`
	ConnectionEnv string `json:"connection_env,omitempty" yaml:"connection_env,omitempty"`
	Query         string `json:"query,omitempty" yaml:"query,omitempty"`
	Format        string `json:"format,omitempty" yaml:"format,omitempty"`
}

// BootstrapSpec declares one SQL bootstrap step that loads a landed file into
// DuckDB before a transform runs.
type BootstrapSpec struct {
	SQLRef      string `json:"sql_ref" yaml:"sql_ref"`
	Placeholder string `json:"placeholder" yaml:"placeholder"`
	SourcePath  string `json:"source_path" yaml:"source_path"`
	Required    bool   `json:"required,omitempty" yaml:"required,omitempty"`
}

// ExternalToolSpec declares a bounded invocation of an optional repo-local
// tool such as dbt. The worker remains the control plane and only delegates a
// single execution step plus artifact collection to the adapter.
type ExternalToolSpec struct {
	Tool       string                 `json:"tool" yaml:"tool"`
	Action     string                 `json:"action" yaml:"action"`
	ProjectRef string                 `json:"project_ref" yaml:"project_ref"`
	ConfigRef  string                 `json:"config_ref" yaml:"config_ref"`
	Profile    string                 `json:"profile,omitempty" yaml:"profile,omitempty"`
	Target     string                 `json:"target,omitempty" yaml:"target,omitempty"`
	Selector   string                 `json:"selector,omitempty" yaml:"selector,omitempty"`
	Binary     string                 `json:"binary,omitempty" yaml:"binary,omitempty"`
	Args       []string               `json:"args,omitempty" yaml:"args,omitempty"`
	Vars       map[string]any         `json:"vars,omitempty" yaml:"vars,omitempty"`
	Artifacts  []ExternalToolArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

// ExternalToolArtifact declares one file that should be mirrored into the
// run-scoped artifact namespace after the tool finishes.
type ExternalToolArtifact struct {
	Path     string `json:"path" yaml:"path"`
	Required bool   `json:"required,omitempty" yaml:"required,omitempty"`
}

// RunStatus models the explicit state machine used for pipeline and job runs.
type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusQueued    RunStatus = "queued"
	RunStatusRunning   RunStatus = "running"
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCanceled  RunStatus = "canceled"
)

// PipelineRun captures one attempt to execute a pipeline.
type PipelineRun struct {
	ID         string     `json:"id"`
	PipelineID string     `json:"pipeline_id"`
	Status     RunStatus  `json:"status"`
	Trigger    string     `json:"trigger"`
	StartedAt  time.Time  `json:"started_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	JobRuns    []JobRun   `json:"job_runs"`
	Events     []RunEvent `json:"events"`
	Error      string     `json:"error,omitempty"`
}

// JobRun captures execution state for one job within a pipeline run.
type JobRun struct {
	ID        string     `json:"id"`
	JobID     string     `json:"job_id"`
	Status    RunStatus  `json:"status"`
	Attempts  int        `json:"attempts"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// RunEvent gives operators a durable, auditable view of state transitions.
type RunEvent struct {
	Time    time.Time         `json:"time"`
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

// QueueSnapshot describes one durable queue record for backup/export and
// diagnostics. The runtime queue interface stays intentionally small, so this
// snapshot type is consumed only by tooling that needs a fuller view.
type QueueSnapshot struct {
	RunID       string     `json:"run_id"`
	PipelineID  string     `json:"pipeline_id"`
	Trigger     string     `json:"trigger"`
	Status      string     `json:"status"`
	RequestedAt time.Time  `json:"requested_at"`
	ClaimedAt   *time.Time `json:"claimed_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
