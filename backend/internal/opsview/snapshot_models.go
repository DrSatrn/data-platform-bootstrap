package opsview

import "time"

// EvidenceGroup is a compact, link-preserving grouping of artifacts suitable
// for operator inspection panes and future evidence boards.
type EvidenceGroup struct {
	Key           string   `json:"key"`
	Kind          string   `json:"kind"`
	Label         string   `json:"label"`
	ArtifactCount int      `json:"artifact_count"`
	ArtifactPaths []string `json:"artifact_paths"`
}

// RunOperatorSnapshot is the backend-only read model for one pipeline run from
// an operator perspective.
type RunOperatorSnapshot struct {
	RunID               string                   `json:"run_id"`
	PipelineID          string                   `json:"pipeline_id"`
	Status              string                   `json:"status"`
	Trigger             string                   `json:"trigger"`
	UpdatedAt           time.Time                `json:"updated_at"`
	JobCount            int                      `json:"job_count"`
	FailedJobCount      int                      `json:"failed_job_count"`
	ExternalToolRuns    []ExternalToolRunSummary `json:"external_tool_runs"`
	EvidenceGroups      []EvidenceGroup          `json:"evidence_groups"`
	Attention           AttentionSummary         `json:"attention"`
	HasExternalToolRuns bool                     `json:"has_external_tool_runs"`
}

// AttentionRollup is a compact cross-run summary suitable for future
// management-console overview panels.
type AttentionRollup struct {
	TotalRuns                    int `json:"total_runs"`
	FailedRuns                   int `json:"failed_runs"`
	RunningRuns                  int `json:"running_runs"`
	SucceededRuns                int `json:"succeeded_runs"`
	RunsWithExternalToolFailures int `json:"runs_with_external_tool_failures"`
	RunsMissingEvidence          int `json:"runs_missing_evidence"`
	ExternalToolJobCount         int `json:"external_tool_job_count"`
}
