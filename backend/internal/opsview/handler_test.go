// These tests keep the opsview HTTP seam trustworthy so the frontend
// management console can depend on backend-owned operator summaries.
package opsview

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

type handlerRunStoreStub struct {
	runs []orchestration.PipelineRun
}

func (s handlerRunStoreStub) ListPipelineRuns() ([]orchestration.PipelineRun, error) {
	return s.runs, nil
}

type handlerArtifactStoreStub struct {
	artifacts map[string][]storage.Artifact
}

func (s handlerArtifactStoreStub) ListRunArtifacts(runID string) ([]storage.Artifact, error) {
	return s.artifacts[runID], nil
}

func TestHandlerBuildsSnapshotsAndAttention(t *testing.T) {
	now := time.Now().UTC()
	run := orchestration.PipelineRun{
		ID:         "run_1",
		PipelineID: "finance",
		Status:     orchestration.RunStatusFailed,
		Trigger:    "manual_api",
		UpdatedAt:  now,
		JobRuns: []orchestration.JobRun{
			{JobID: "run_finance_dbt", Status: orchestration.RunStatusFailed},
		},
		Events: []orchestration.RunEvent{
			{
				Time:    now,
				Level:   "info",
				Message: "external tool command started",
				Fields:  map[string]string{"job_id": "run_finance_dbt", "tool": "dbt", "action": "build"},
			},
			{
				Time:    now.Add(time.Second),
				Level:   "error",
				Message: "external tool failed",
				Fields:  map[string]string{"job_id": "run_finance_dbt", "tool": "dbt", "failure_class": "missing_artifact"},
			},
		},
	}

	handler := NewHandler(
		handlerRunStoreStub{runs: []orchestration.PipelineRun{run}},
		handlerArtifactStoreStub{artifacts: map[string][]storage.Artifact{
			"run_1": {
				{RunID: "run_1", RelativePath: "external_tools/run_finance_dbt/logs/stdout.log"},
				{RunID: "run_1", RelativePath: "external_tools/run_finance_dbt/outputs/run_results.json"},
			},
		}},
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/opsview?limit=1", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"attention_rollup"`) {
		t.Fatalf("expected attention rollup in response, got %s", body)
	}
	if !strings.Contains(body, `"external_tool_attention"`) {
		t.Fatalf("expected external tool attention in response, got %s", body)
	}
	if !strings.Contains(body, `"missing_artifact"`) {
		t.Fatalf("expected failure class in response, got %s", body)
	}
	if !strings.Contains(body, `"external_tools/run_finance_dbt/logs/stdout.log"`) {
		t.Fatalf("expected artifact path in response, got %s", body)
	}
}
