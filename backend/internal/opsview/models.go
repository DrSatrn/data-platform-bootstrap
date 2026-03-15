// Package opsview builds operator-facing read models from existing run events
// and artifacts. The package is intentionally pure so future handlers or UIs
// can consume stable summaries without taking on orchestration logic.
package opsview

import (
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

// ExternalToolRunSummary groups one external-tool job's lifecycle and evidence
// into a management-console-friendly view model.
type ExternalToolRunSummary struct {
	RunID           string                   `json:"run_id"`
	PipelineID      string                   `json:"pipeline_id"`
	JobID           string                   `json:"job_id"`
	Tool            string                   `json:"tool"`
	Action          string                   `json:"action"`
	Status          string                   `json:"status"`
	FailureClass    string                   `json:"failure_class,omitempty"`
	LastEventAt     *time.Time               `json:"last_event_at,omitempty"`
	Events          []orchestration.RunEvent `json:"events"`
	LogArtifacts    []storage.Artifact       `json:"log_artifacts"`
	OutputArtifacts []storage.Artifact       `json:"output_artifacts"`
	Evidence        OperatorEvidenceSummary  `json:"evidence"`
}

// OperatorEvidenceSummary preserves the inspectable artifact paths and simple
// counts that an operator view can later link to directly.
type OperatorEvidenceSummary struct {
	RunID               string   `json:"run_id"`
	TotalArtifacts      int      `json:"total_artifacts"`
	LogArtifactCount    int      `json:"log_artifact_count"`
	OutputArtifactCount int      `json:"output_artifact_count"`
	ArtifactPaths       []string `json:"artifact_paths"`
	LogPaths            []string `json:"log_paths"`
	OutputPaths         []string `json:"output_paths"`
}

// AttentionSummary gives the future management console a compact status view
// across one run's external-tool job summaries.
type AttentionSummary struct {
	TotalJobs          int `json:"total_jobs"`
	FailedJobs         int `json:"failed_jobs"`
	RunningJobs        int `json:"running_jobs"`
	SucceededJobs      int `json:"succeeded_jobs"`
	JobsMissingLogs    int `json:"jobs_missing_logs"`
	JobsMissingOutputs int `json:"jobs_missing_outputs"`
}
