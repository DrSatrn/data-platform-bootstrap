package orchestration

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPipelineHandlerListsExternalToolRunEvents(t *testing.T) {
	store := NewInMemoryStore()
	now := time.Now().UTC()
	if err := store.SavePipelineRun(PipelineRun{
		ID:         "run_1",
		PipelineID: "personal_finance_dbt_pipeline",
		Status:     RunStatusFailed,
		Events: []RunEvent{
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
				Time:    now.Add(time.Second),
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
	}); err != nil {
		t.Fatalf("save pipeline run: %v", err)
	}

	handler := NewPipelineHandler(
		staticPipelineLoader{pipelines: []Pipeline{
			{
				ID: "personal_finance_dbt_pipeline",
				Jobs: []Job{
					{
						ID:   "build_finance_dbt",
						Type: JobTypeExternalTool,
						ExternalTool: &ExternalToolSpec{
							Tool:       "dbt",
							Action:     "build",
							ProjectRef: "packages/external_tools/dbt_finance_demo",
							ConfigRef:  "packages/external_tools/dbt_finance_demo/profiles",
							Artifacts: []ExternalToolArtifact{
								{Path: "target/run_results.json", Required: true},
							},
						},
					},
				},
			},
		}},
		store,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
		nil,
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/pipelines", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var payload struct {
		Runs []struct {
			ID     string `json:"id"`
			Events []struct {
				Message string            `json:"message"`
				Fields  map[string]string `json:"fields"`
			} `json:"events"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(payload.Runs))
	}
	if len(payload.Runs[0].Events) != 2 {
		t.Fatalf("expected 2 run events, got %d", len(payload.Runs[0].Events))
	}
	if payload.Runs[0].Events[0].Message != "external tool command started" {
		t.Fatalf("unexpected first event %q", payload.Runs[0].Events[0].Message)
	}
	if payload.Runs[0].Events[1].Fields["failure_class"] != "execution_failed" {
		t.Fatalf("expected failure_class execution_failed, got %q", payload.Runs[0].Events[1].Fields["failure_class"])
	}
}

type staticPipelineLoader struct {
	pipelines []Pipeline
}

func (s staticPipelineLoader) LoadPipelines() ([]Pipeline, error) {
	return s.pipelines, nil
}
