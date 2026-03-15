package opsview

import (
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestBuildExternalToolRunSummariesGroupsFailedRunEvidence(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	run := orchestration.PipelineRun{
		ID:         "run_1",
		PipelineID: "personal_finance_dbt_pipeline",
		Events: []orchestration.RunEvent{
			{
				Time:    now,
				Level:   "info",
				Message: "external tool command started",
				Fields: map[string]string{
					"job_id": "build_finance_dbt",
					"tool":   "dbt",
					"action": "build",
				},
			},
			{
				Time:    now.Add(2 * time.Second),
				Level:   "error",
				Message: "external tool failed",
				Fields: map[string]string{
					"job_id":        "build_finance_dbt",
					"tool":          "dbt",
					"action":        "build",
					"failure_class": "execution_failed",
				},
			},
		},
	}
	artifacts := []storage.Artifact{
		{RunID: "run_1", RelativePath: "external_tools/build_finance_dbt/logs/stdout.log"},
		{RunID: "run_1", RelativePath: "external_tools/build_finance_dbt/logs/stderr.log"},
		{RunID: "run_1", RelativePath: "external_tools/build_finance_dbt/target/run_results.json"},
	}

	summaries := BuildExternalToolRunSummaries(run, artifacts)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	summary := summaries[0]
	if summary.JobID != "build_finance_dbt" {
		t.Fatalf("expected job build_finance_dbt, got %q", summary.JobID)
	}
	if summary.Tool != "dbt" {
		t.Fatalf("expected tool dbt, got %q", summary.Tool)
	}
	if summary.Status != "failed" {
		t.Fatalf("expected failed status, got %q", summary.Status)
	}
	if summary.FailureClass != "execution_failed" {
		t.Fatalf("expected failure_class execution_failed, got %q", summary.FailureClass)
	}
	if len(summary.LogArtifacts) != 2 {
		t.Fatalf("expected 2 log artifacts, got %d", len(summary.LogArtifacts))
	}
	if len(summary.OutputArtifacts) != 1 {
		t.Fatalf("expected 1 output artifact, got %d", len(summary.OutputArtifacts))
	}
	if summary.Evidence.TotalArtifacts != 3 {
		t.Fatalf("expected evidence total 3, got %d", summary.Evidence.TotalArtifacts)
	}
}

func TestBuildExternalToolRunSummariesSeparatesLogsAndOutputs(t *testing.T) {
	run := orchestration.PipelineRun{ID: "run_1", PipelineID: "pipeline"}
	artifacts := []storage.Artifact{
		{RunID: "run_1", RelativePath: "external_tools/job_a/logs/stdout.log"},
		{RunID: "run_1", RelativePath: "external_tools/job_a/target/manifest.json"},
		{RunID: "run_1", RelativePath: "external_tools/job_a/target/run_results.json"},
	}

	summaries := BuildExternalToolRunSummaries(run, artifacts)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if got := summaries[0].LogArtifacts[0].RelativePath; got != "external_tools/job_a/logs/stdout.log" {
		t.Fatalf("expected stdout.log in logs, got %q", got)
	}
	if len(summaries[0].OutputArtifacts) != 2 {
		t.Fatalf("expected 2 output artifacts, got %d", len(summaries[0].OutputArtifacts))
	}
}

func TestBuildExternalToolRunSummariesHandlesEmptyInputSafely(t *testing.T) {
	summaries := BuildExternalToolRunSummaries(orchestration.PipelineRun{}, nil)
	if len(summaries) != 0 {
		t.Fatalf("expected empty summaries, got %d", len(summaries))
	}

	evidence := BuildOperatorEvidenceSummary("run_1", nil, nil)
	if evidence.TotalArtifacts != 0 {
		t.Fatalf("expected 0 evidence artifacts, got %d", evidence.TotalArtifacts)
	}

	attention := BuildAttentionSummary(nil)
	if attention.TotalJobs != 0 {
		t.Fatalf("expected 0 total jobs, got %d", attention.TotalJobs)
	}
}

func TestBuildOperatorEvidenceSummaryPreservesArtifactPaths(t *testing.T) {
	evidence := BuildOperatorEvidenceSummary("run_1",
		[]storage.Artifact{
			{RunID: "run_1", RelativePath: "external_tools/job_a/logs/stderr.log"},
		},
		[]storage.Artifact{
			{RunID: "run_1", RelativePath: "external_tools/job_a/target/run_results.json"},
		},
	)

	if evidence.LogPaths[0] != "external_tools/job_a/logs/stderr.log" {
		t.Fatalf("expected stderr path preserved, got %q", evidence.LogPaths[0])
	}
	if evidence.OutputPaths[0] != "external_tools/job_a/target/run_results.json" {
		t.Fatalf("expected output path preserved, got %q", evidence.OutputPaths[0])
	}
	if len(evidence.ArtifactPaths) != 2 {
		t.Fatalf("expected 2 flattened artifact paths, got %d", len(evidence.ArtifactPaths))
	}
}

func TestBuildAttentionSummaryCountsMissingLogsAndOutputs(t *testing.T) {
	attention := BuildAttentionSummary([]ExternalToolRunSummary{
		{
			JobID:           "job_failed",
			Status:          "failed",
			LogArtifacts:    nil,
			OutputArtifacts: []storage.Artifact{{RelativePath: "external_tools/job_failed/target/run_results.json"}},
		},
		{
			JobID:           "job_running",
			Status:          "running",
			LogArtifacts:    []storage.Artifact{{RelativePath: "external_tools/job_running/logs/stdout.log"}},
			OutputArtifacts: nil,
		},
	})

	if attention.TotalJobs != 2 {
		t.Fatalf("expected 2 total jobs, got %d", attention.TotalJobs)
	}
	if attention.FailedJobs != 1 {
		t.Fatalf("expected 1 failed job, got %d", attention.FailedJobs)
	}
	if attention.RunningJobs != 1 {
		t.Fatalf("expected 1 running job, got %d", attention.RunningJobs)
	}
	if attention.JobsMissingLogs != 1 {
		t.Fatalf("expected 1 job missing logs, got %d", attention.JobsMissingLogs)
	}
	if attention.JobsMissingOutputs != 1 {
		t.Fatalf("expected 1 job missing outputs, got %d", attention.JobsMissingOutputs)
	}
}
