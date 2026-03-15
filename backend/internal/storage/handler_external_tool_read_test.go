package storage

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerReadsExternalToolLogArtifactContent(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "runs", "run_1", "external_tools", "build_finance_dbt", "logs", "stderr.log")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(target), err)
	}
	if err := os.WriteFile(target, []byte("dbt warning"), 0o644); err != nil {
		t.Fatalf("write %s: %v", target, err)
	}

	handler := NewHandler(NewService(root, nil))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/artifacts?run_id=run_1&path=external_tools/build_finance_dbt/logs/stderr.log", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "dbt warning") {
		t.Fatalf("expected artifact body to include log contents, got %q", recorder.Body.String())
	}
}
