package analytics

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportHandlerReturnsCSV(t *testing.T) {
	root := t.TempDir()
	writeAnalyticsSampleData(t, root)
	handler := NewExportHandler(NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "missing-sql")))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/export?dataset=mart_budget_vs_actual&group_by=month,category", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/csv") {
		t.Fatalf("expected text/csv content type, got %q", contentType)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "month,category,budget_amount,actual_spend,variance_amount") {
		t.Fatalf("expected csv header, got %s", body)
	}
}

func TestHandlerRejectsInvalidMultiGroupBy(t *testing.T) {
	root := t.TempDir()
	writeAnalyticsSampleData(t, root)
	handler := NewHandler(NewService(root, filepath.Join(root, "materialized"), filepath.Join(root, "duckdb", "platform.duckdb"), filepath.Join(root, "missing-sql")))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/analytics?dataset=mart_budget_vs_actual&group_by=month,owner", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid group_by, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "unsupported group_by") {
		t.Fatalf("expected unsupported group_by response, got %s", recorder.Body.String())
	}
}
