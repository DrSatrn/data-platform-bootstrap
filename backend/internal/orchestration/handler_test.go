package orchestration

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPipelineHandlerSanitizesValidationErrors(t *testing.T) {
	handler := NewPipelineHandler(
		staticPipelineLoader{pipelines: []Pipeline{
			{
				ID: "broken_pipeline",
				Jobs: []Job{
					{
						ID:        "job_a",
						DependsOn: []string{"missing_job"},
					},
				},
			},
		}},
		NewInMemoryStore(),
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
		ValidationErrors map[string]string `json:"validation_errors"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	got := payload.ValidationErrors["broken_pipeline"]
	if got == "" {
		t.Fatal("expected validation error entry for broken pipeline")
	}
	if strings.Contains(strings.ToLower(got), "missing_job") {
		t.Fatalf("expected sanitized validation error, got %q", got)
	}
	if got != "pipeline definition is invalid; inspect local validation logs for details" {
		t.Fatalf("unexpected validation error message %q", got)
	}
}
