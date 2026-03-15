// These tests verify that the router enforces the documented access model on
// the most sensitive operator-facing routes. They intentionally assert policy
// at the HTTP boundary so docs, frontend capability messaging, and backend
// behavior do not drift apart.
package app

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/observability"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/reporting"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestRouterProtectsViewerAndAdminEndpoints(t *testing.T) {
	root := t.TempDir()
	manifestRoot := filepath.Join(root, "manifests")
	for _, dir := range []string{
		filepath.Join(manifestRoot, "pipelines"),
		filepath.Join(manifestRoot, "assets"),
		filepath.Join(manifestRoot, "metrics"),
		filepath.Join(root, "dashboards"),
		filepath.Join(root, "sql"),
		filepath.Join(root, "python"),
		filepath.Join(root, "sample_data", "personal_finance"),
		filepath.Join(root, "data"),
		filepath.Join(root, "artifacts", "runs", "run_1"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	for path, contents := range map[string]string{
		filepath.Join(root, "sample_data", "personal_finance", "transactions.csv"):  "transaction_id,date,description,category,amount\n1,2026-03-01,Coffee,Food,-4.50\n",
		filepath.Join(root, "sample_data", "personal_finance", "budget_rules.json"): `[{"category":"Food","monthly_budget":100}]`,
	} {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	runStore, err := orchestration.NewFileStore(filepath.Join(root, "data"))
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	queue, err := orchestration.NewQueue(filepath.Join(root, "data"))
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}

	router := newRouter(
		slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		config.Settings{
			Environment:         "test",
			HTTPAddr:            ":8080",
			WebAddr:             ":3000",
			APIBaseURL:          "http://127.0.0.1:8080",
			LogLevel:            "debug",
			DataRoot:            filepath.Join(root, "data"),
			ArtifactRoot:        filepath.Join(root, "artifacts"),
			DuckDBPath:          filepath.Join(root, "duckdb", "platform.duckdb"),
			ManifestRoot:        manifestRoot,
			DashboardRoot:       filepath.Join(root, "dashboards"),
			SQLRoot:             filepath.Join(root, "sql"),
			PythonTaskRoot:      filepath.Join(root, "python"),
			SampleDataRoot:      filepath.Join(root, "sample_data"),
			AdminToken:          "admin-token",
			AccessTokens:        "viewer-token:viewer:viewer-user,editor-token:editor:editor-user",
			SchedulerTick:       time.Second,
			WorkerPoll:          time.Second,
			ExternalToolTimeout: 5 * time.Minute,
			MaxConcurrentJobs:   1,
		},
		observability.NewService(),
		runtimePersistence{
			store:     runStore,
			queue:     queue,
			artifacts: storage.NewService(filepath.Join(root, "artifacts"), nil),
			reports:   reporting.NewMemoryStore(),
			audit:     audit.NewMemoryStore(),
			metadata:  nil,
			modes:     map[string]observability.PersistenceMode{},
		},
	)

	testCases := []struct {
		name          string
		method        string
		target        string
		token         string
		body          []byte
		expectedCodes []int
	}{
		{name: "catalog rejects anonymous", method: http.MethodGet, target: "/api/v1/catalog", expectedCodes: []int{http.StatusForbidden}},
		{name: "catalog allows viewer", method: http.MethodGet, target: "/api/v1/catalog", token: "viewer-token", expectedCodes: []int{http.StatusOK}},
		{name: "analytics rejects anonymous", method: http.MethodGet, target: "/api/v1/analytics", expectedCodes: []int{http.StatusForbidden}},
		{name: "analytics allows viewer", method: http.MethodGet, target: "/api/v1/analytics", token: "viewer-token", expectedCodes: []int{http.StatusOK}},
		{name: "analytics export rejects anonymous", method: http.MethodGet, target: "/api/v1/analytics/export?dataset=mart_monthly_cashflow", expectedCodes: []int{http.StatusForbidden}},
		{name: "analytics export allows viewer", method: http.MethodGet, target: "/api/v1/analytics/export?dataset=mart_monthly_cashflow", token: "viewer-token", expectedCodes: []int{http.StatusOK}},
		{name: "artifacts rejects anonymous", method: http.MethodGet, target: "/api/v1/artifacts?run_id=run_1", expectedCodes: []int{http.StatusForbidden}},
		{name: "artifacts allows viewer", method: http.MethodGet, target: "/api/v1/artifacts?run_id=run_1", token: "viewer-token", expectedCodes: []int{http.StatusOK}},
		{name: "audit rejects anonymous", method: http.MethodGet, target: "/api/v1/system/audit", expectedCodes: []int{http.StatusForbidden}},
		{name: "audit allows viewer", method: http.MethodGet, target: "/api/v1/system/audit", token: "viewer-token", expectedCodes: []int{http.StatusOK}},
		{name: "logs allows viewer", method: http.MethodGet, target: "/api/v1/system/logs", token: "viewer-token", expectedCodes: []int{http.StatusOK}},
		{name: "overview allows viewer", method: http.MethodGet, target: "/api/v1/system/overview", token: "viewer-token", expectedCodes: []int{http.StatusOK}},
		{name: "metrics allows viewer", method: http.MethodGet, target: "/api/v1/system/metrics", token: "viewer-token", expectedCodes: []int{http.StatusOK}},
		{name: "admin terminal rejects viewer", method: http.MethodPost, target: "/api/v1/admin/terminal/execute", token: "viewer-token", body: []byte(`{"command":"status"}`), expectedCodes: []int{http.StatusForbidden}},
		{name: "admin terminal allows admin", method: http.MethodPost, target: "/api/v1/admin/terminal/execute", token: "admin-token", body: []byte(`{"command":"status"}`), expectedCodes: []int{http.StatusOK}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(testCase.method, testCase.target, bytes.NewReader(testCase.body))
			if testCase.token != "" {
				request.Header.Set("Authorization", "Bearer "+testCase.token)
			}
			if len(testCase.body) > 0 {
				request.Header.Set("Content-Type", "application/json")
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			for _, expectedCode := range testCase.expectedCodes {
				if recorder.Code == expectedCode {
					return
				}
			}
			t.Fatalf("expected one of %v, got %d with body %s", testCase.expectedCodes, recorder.Code, recorder.Body.String())
		})
	}
}
