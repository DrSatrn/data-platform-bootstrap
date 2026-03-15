package storage

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandlerListsExternalToolArtifactsAndLogs(t *testing.T) {
	root := t.TempDir()
	mustWriteArtifact(t, filepath.Join(root, "runs", "run_1", "external_tools", "build_finance_dbt", "logs", "stdout.log"), "dbt started")
	mustWriteArtifact(t, filepath.Join(root, "runs", "run_1", "external_tools", "build_finance_dbt", "logs", "stderr.log"), "dbt warning")
	mustWriteArtifact(t, filepath.Join(root, "runs", "run_1", "external_tools", "build_finance_dbt", "target", "run_results.json"), `{"ok":true}`)

	handler := NewHandler(NewService(root, nil))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/artifacts?run_id=run_1", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var payload struct {
		Artifacts []Artifact `json:"artifacts"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Artifacts) != 3 {
		t.Fatalf("expected 3 artifacts, got %d", len(payload.Artifacts))
	}

	paths := map[string]bool{}
	for _, artifact := range payload.Artifacts {
		paths[artifact.RelativePath] = true
	}
	for _, expected := range []string{
		"external_tools/build_finance_dbt/logs/stdout.log",
		"external_tools/build_finance_dbt/logs/stderr.log",
		"external_tools/build_finance_dbt/target/run_results.json",
	} {
		if !paths[expected] {
			t.Fatalf("expected artifact listing to include %q", expected)
		}
	}
}

func mustWriteArtifact(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
