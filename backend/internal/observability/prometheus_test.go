package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func TestMetricsHandlerRendersPrometheusPayload(t *testing.T) {
	service := NewService()
	service.RecordRequest(http.MethodGet, "/api/v1/catalog", http.StatusOK, 12*time.Millisecond)
	service.RecordRequest(http.MethodGet, "/api/v1/catalog", http.StatusInternalServerError, 180*time.Millisecond)

	handler := NewMetricsHandler(service, queueStub{
		requests: []orchestration.QueueSnapshot{
			{Status: "queued"},
			{Status: "queued"},
			{Status: "active"},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/metrics", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		t.Fatalf("expected text/plain content type, got %q", contentType)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		"go_memstats_alloc_bytes",
		"go_goroutines",
		"platform_workers_active 1",
		"platform_queue_depth 2",
		"platform_http_requests_total 2",
		"platform_http_request_errors_total 1",
		"platform_http_request_duration_seconds_bucket",
		"platform_http_request_duration_seconds_count 2",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected metrics output to contain %q, got %s", expected, body)
		}
	}
}
