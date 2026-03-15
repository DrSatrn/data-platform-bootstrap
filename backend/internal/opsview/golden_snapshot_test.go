package opsview

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestRunOperatorSnapshotMatchesGoldenJSON(t *testing.T) {
	run := orchestration.PipelineRun{
		ID:         "run_fixture",
		PipelineID: "personal_finance_dbt_pipeline",
		Status:     orchestration.RunStatusFailed,
		Trigger:    "manual_api",
		UpdatedAt:  time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC),
		JobRuns: []orchestration.JobRun{
			{JobID: "build_finance_dbt", Status: orchestration.RunStatusFailed},
		},
		Events: []orchestration.RunEvent{
			{
				Time:    time.Date(2026, 3, 15, 13, 59, 0, 0, time.UTC),
				Level:   "info",
				Message: "external tool command started",
				Fields: map[string]string{
					"job_id": "build_finance_dbt",
					"tool":   "dbt",
					"action": "build",
				},
			},
			{
				Time:    time.Date(2026, 3, 15, 13, 59, 30, 0, time.UTC),
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
		{RunID: "run_fixture", RelativePath: "external_tools/build_finance_dbt/logs/stdout.log"},
		{RunID: "run_fixture", RelativePath: "external_tools/build_finance_dbt/logs/stderr.log"},
		{RunID: "run_fixture", RelativePath: "external_tools/build_finance_dbt/target/run_results.json"},
		{RunID: "run_fixture", RelativePath: "metrics/metrics_savings_rate.json"},
	}

	snapshot := BuildRunOperatorSnapshot(run, artifacts)
	bytes, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	goldenPath := filepath.Join("testdata", "run_operator_snapshot.golden.json")
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v", goldenPath, err)
	}
	if strings.TrimSpace(string(bytes)) != strings.TrimSpace(string(expected)) {
		t.Fatalf("snapshot did not match golden file\nexpected:\n%s\n\nactual:\n%s", string(expected), string(bytes))
	}
}
