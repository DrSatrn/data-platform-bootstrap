// These tests cover zero-data and degraded-path observability behavior so the
// operator diagnostics surfaces remain stable under partial failures.
package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func TestHealthHandlerReturnsJSONForMinimalConfig(t *testing.T) {
	handler := HealthHandler(config.Settings{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected health payload, got %s", recorder.Body.String())
	}
}

func TestRecentLogsHandlerReturnsEmptyLogBuffer(t *testing.T) {
	handler := NewRecentLogsHandler(NewService())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/logs", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"logs":[]`) {
		t.Fatalf("expected empty logs array, got %s", recorder.Body.String())
	}
}

func TestSnapshotStartsWithZeroObservations(t *testing.T) {
	snapshot := NewService().Snapshot(map[string]string{"environment": "test"})

	if snapshot.TotalRequests != 0 || snapshot.TotalErrors != 0 || snapshot.TotalCommands != 0 {
		t.Fatalf("expected zeroed counters, got %+v", snapshot)
	}
	if len(snapshot.RecentRequests) != 0 || len(snapshot.RecentCommands) != 0 || len(snapshot.RecentLogSummary) != 0 {
		t.Fatalf("expected empty recent buffers, got %+v", snapshot)
	}
}

func TestOverviewHandlerHandlesPartialDataWithoutQueueOrBackups(t *testing.T) {
	handler := NewOverviewHandler(
		config.Settings{Environment: "test", DataRoot: t.TempDir()},
		NewService(),
		pipelineLoaderStub{pipelines: []orchestration.Pipeline{{ID: "finance"}}},
		assetLoaderStub{assets: []metadata.DataAsset{{ID: "mart_cashflow"}}},
		runStoreStub{runs: []orchestration.PipelineRun{{ID: "run_1", Status: orchestration.RunStatusQueued}}},
		nil,
		nil,
		nil,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/overview", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"queue_summary":{"queued":0,"active":0,"completed":0,"total":0}`) {
		t.Fatalf("expected zeroed queue summary, got %s", recorder.Body.String())
	}
}

func TestOverviewHandlerReturnsJSONErrorWhenPipelinesFail(t *testing.T) {
	handler := NewOverviewHandler(
		config.Settings{Environment: "test", DataRoot: t.TempDir()},
		NewService(),
		failingPipelineLoader{},
		assetLoaderStub{},
		runStoreStub{},
		nil,
		nil,
		nil,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/overview", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"error":"failed to load pipelines for system overview"`) {
		t.Fatalf("expected JSON error body, got %s", recorder.Body.String())
	}
}

type failingPipelineLoader struct{}

func (failingPipelineLoader) LoadPipelines() ([]orchestration.Pipeline, error) {
	return nil, http.ErrHandlerTimeout
}
