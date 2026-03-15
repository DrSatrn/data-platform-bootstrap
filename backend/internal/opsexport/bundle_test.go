package opsexport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/opsview"
)

func TestBuildBundleOrdersSnapshotsAndBuildsRollup(t *testing.T) {
	snapshots := []opsview.RunOperatorSnapshot{
		{
			RunID:            "run_older",
			Status:           "running",
			UpdatedAt:        time.Date(2026, 3, 15, 11, 0, 0, 0, time.UTC),
			FailedJobCount:   0,
			ExternalToolRuns: []opsview.ExternalToolRunSummary{{JobID: "job_running"}},
		},
		{
			RunID:            "run_newer",
			Status:           "failed",
			UpdatedAt:        time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
			FailedJobCount:   1,
			ExternalToolRuns: []opsview.ExternalToolRunSummary{{JobID: "job_failed"}},
		},
	}

	bundle := BuildBundle(time.Date(2026, 3, 15, 13, 0, 0, 0, time.UTC), snapshots)
	if bundle.Version != bundleVersion {
		t.Fatalf("expected version %q, got %q", bundleVersion, bundle.Version)
	}
	if bundle.SnapshotCount != 2 {
		t.Fatalf("expected 2 snapshots, got %d", bundle.SnapshotCount)
	}
	if bundle.Snapshots[0].RunID != "run_newer" {
		t.Fatalf("expected newest snapshot first, got %q", bundle.Snapshots[0].RunID)
	}
	if bundle.Rollup.TotalRuns != 2 {
		t.Fatalf("expected total runs 2, got %d", bundle.Rollup.TotalRuns)
	}
	if bundle.Rollup.RunsWithExternalToolFailures != 1 {
		t.Fatalf("expected 1 run with external tool failures, got %d", bundle.Rollup.RunsWithExternalToolFailures)
	}
}

func TestBuildBundleHandlesEmptyInputSafely(t *testing.T) {
	bundle := BuildBundle(time.Date(2026, 3, 15, 13, 0, 0, 0, time.UTC), nil)
	if bundle.SnapshotCount != 0 {
		t.Fatalf("expected zero snapshot count, got %d", bundle.SnapshotCount)
	}
	if len(bundle.Snapshots) != 0 {
		t.Fatalf("expected no snapshots, got %d", len(bundle.Snapshots))
	}
	if bundle.Rollup.TotalRuns != 0 {
		t.Fatalf("expected zero rollup runs, got %d", bundle.Rollup.TotalRuns)
	}
}

func TestMarshalBundleMatchesGoldenJSON(t *testing.T) {
	bundle := BuildBundle(time.Date(2026, 3, 15, 13, 0, 0, 0, time.UTC), []opsview.RunOperatorSnapshot{
		{
			RunID:          "run_fixture",
			PipelineID:     "personal_finance_dbt_pipeline",
			Status:         "failed",
			Trigger:        "manual_api",
			UpdatedAt:      time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
			JobCount:       1,
			FailedJobCount: 1,
			ExternalToolRuns: []opsview.ExternalToolRunSummary{
				{
					RunID:        "run_fixture",
					PipelineID:   "personal_finance_dbt_pipeline",
					JobID:        "build_finance_dbt",
					Tool:         "dbt",
					Action:       "build",
					Status:       "failed",
					FailureClass: "execution_failed",
				},
			},
			EvidenceGroups: []opsview.EvidenceGroup{
				{
					Key:           "external_tool_logs:build_finance_dbt",
					Kind:          "external_tool_logs",
					Label:         "External Tool Logs: build_finance_dbt",
					ArtifactCount: 2,
					ArtifactPaths: []string{
						"external_tools/build_finance_dbt/logs/stderr.log",
						"external_tools/build_finance_dbt/logs/stdout.log",
					},
				},
			},
			Attention: opsview.AttentionSummary{
				TotalJobs:          1,
				FailedJobs:         1,
				RunningJobs:        0,
				SucceededJobs:      0,
				JobsMissingLogs:    0,
				JobsMissingOutputs: 1,
			},
			HasExternalToolRuns: true,
		},
	})

	bytes, err := MarshalBundle(bundle)
	if err != nil {
		t.Fatalf("MarshalBundle returned error: %v", err)
	}

	expected, err := os.ReadFile(filepath.Join("testdata", "bundle.golden.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if strings.TrimSpace(string(bytes)) != strings.TrimSpace(string(expected)) {
		t.Fatalf("bundle did not match golden file\nexpected:\n%s\n\nactual:\n%s", string(expected), string(bytes))
	}
}

func TestSuggestedFilenameIsStable(t *testing.T) {
	got := SuggestedFilename(time.Date(2026, 3, 15, 13, 0, 5, 0, time.UTC))
	if got != "opsview_export_20260315T130005Z.json" {
		t.Fatalf("unexpected filename %q", got)
	}
}
