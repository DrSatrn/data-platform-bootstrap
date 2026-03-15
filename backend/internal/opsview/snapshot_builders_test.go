package opsview

import (
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestBuildRunOperatorSnapshotBuildsOperatorFacingSummary(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	run := orchestration.PipelineRun{
		ID:         "run_1",
		PipelineID: "personal_finance_dbt_pipeline",
		Status:     orchestration.RunStatusFailed,
		Trigger:    "manual_api",
		UpdatedAt:  now,
		JobRuns: []orchestration.JobRun{
			{JobID: "build_finance_dbt", Status: orchestration.RunStatusFailed},
			{JobID: "publish_metrics", Status: orchestration.RunStatusPending},
		},
		Events: []orchestration.RunEvent{
			{
				Time:    now.Add(-2 * time.Minute),
				Level:   "info",
				Message: "external tool command started",
				Fields: map[string]string{
					"job_id": "build_finance_dbt",
					"tool":   "dbt",
					"action": "build",
				},
			},
			{
				Time:    now.Add(-time.Minute),
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
		{RunID: "run_1", RelativePath: "external_tools/build_finance_dbt/target/run_results.json"},
		{RunID: "run_1", RelativePath: "metrics/metrics_savings_rate.json"},
	}

	snapshot := BuildRunOperatorSnapshot(run, artifacts)
	if snapshot.RunID != "run_1" {
		t.Fatalf("expected run_1, got %q", snapshot.RunID)
	}
	if snapshot.FailedJobCount != 1 {
		t.Fatalf("expected 1 failed job, got %d", snapshot.FailedJobCount)
	}
	if !snapshot.HasExternalToolRuns {
		t.Fatal("expected external tool runs to be detected")
	}
	if len(snapshot.ExternalToolRuns) != 1 {
		t.Fatalf("expected 1 external tool summary, got %d", len(snapshot.ExternalToolRuns))
	}
	if len(snapshot.EvidenceGroups) != 3 {
		t.Fatalf("expected 3 evidence groups, got %d", len(snapshot.EvidenceGroups))
	}
	if snapshot.Attention.FailedJobs != 1 {
		t.Fatalf("expected attention failed jobs 1, got %d", snapshot.Attention.FailedJobs)
	}
}

func TestBuildEvidenceGroupsSeparatesLogsOutputsAndOtherArtifacts(t *testing.T) {
	artifacts := []storage.Artifact{
		{RelativePath: "external_tools/build_finance_dbt/logs/stdout.log"},
		{RelativePath: "external_tools/build_finance_dbt/logs/stderr.log"},
		{RelativePath: "external_tools/build_finance_dbt/target/run_results.json"},
		{RelativePath: "quality/check_duplicate_transactions.json"},
	}

	groups := BuildEvidenceGroups(artifacts)
	if len(groups) != 3 {
		t.Fatalf("expected 3 evidence groups, got %d", len(groups))
	}
	if groups[0].Kind != "artifacts" && groups[1].Kind != "external_tool_logs" && groups[2].Kind != "external_tool_outputs" {
		t.Fatalf("unexpected group kinds %+v", groups)
	}
}

func TestBuildAttentionRollupCountsRunLevelStates(t *testing.T) {
	rollup := BuildAttentionRollup([]RunOperatorSnapshot{
		{Status: "failed", FailedJobCount: 1, ExternalToolRuns: []ExternalToolRunSummary{{JobID: "job_a"}}, EvidenceGroups: []EvidenceGroup{{Key: "a"}}},
		{Status: "running", ExternalToolRuns: []ExternalToolRunSummary{{JobID: "job_b"}}, EvidenceGroups: nil},
		{Status: "succeeded", ExternalToolRuns: nil, EvidenceGroups: []EvidenceGroup{{Key: "b"}}},
	})

	if rollup.TotalRuns != 3 {
		t.Fatalf("expected 3 runs, got %d", rollup.TotalRuns)
	}
	if rollup.FailedRuns != 1 || rollup.RunningRuns != 1 || rollup.SucceededRuns != 1 {
		t.Fatalf("unexpected status counts %+v", rollup)
	}
	if rollup.RunsWithExternalToolFailures != 1 {
		t.Fatalf("expected 1 run with external tool failures, got %d", rollup.RunsWithExternalToolFailures)
	}
	if rollup.RunsMissingEvidence != 1 {
		t.Fatalf("expected 1 run missing evidence, got %d", rollup.RunsMissingEvidence)
	}
	if rollup.ExternalToolJobCount != 2 {
		t.Fatalf("expected 2 external tool jobs, got %d", rollup.ExternalToolJobCount)
	}
}

func TestBuildRunOperatorSnapshotHandlesEmptyCaseSafely(t *testing.T) {
	snapshot := BuildRunOperatorSnapshot(orchestration.PipelineRun{}, nil)
	if snapshot.RunID != "" {
		t.Fatalf("expected empty run id, got %q", snapshot.RunID)
	}
	if snapshot.HasExternalToolRuns {
		t.Fatal("expected no external tool runs")
	}
	if len(snapshot.EvidenceGroups) != 0 {
		t.Fatalf("expected no evidence groups, got %d", len(snapshot.EvidenceGroups))
	}
	if snapshot.Attention.TotalJobs != 0 {
		t.Fatalf("expected zero attention jobs, got %d", snapshot.Attention.TotalJobs)
	}
}
