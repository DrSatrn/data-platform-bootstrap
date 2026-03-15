package observability

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
)

// NewMetricsHandler exposes built-in telemetry in the Prometheus text format.
func NewMetricsHandler(service *Service, queue QueueSnapshotter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queueSummary := queueSummary{}
		if queue != nil {
			requests, err := queue.ListRequests()
			if err != nil {
				http.Error(w, "failed to load queue state for metrics", http.StatusInternalServerError)
				return
			}
			queueSummary = summarizeQueue(requests)
		}

		metrics := service.MetricsSnapshot()
		memStats := runtime.MemStats{}
		runtime.ReadMemStats(&memStats)

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		writePrometheusMetric(w, "go_memstats_alloc_bytes", "gauge", float64(memStats.Alloc))
		writePrometheusMetric(w, "go_memstats_heap_alloc_bytes", "gauge", float64(memStats.HeapAlloc))
		writePrometheusMetric(w, "go_memstats_heap_sys_bytes", "gauge", float64(memStats.HeapSys))
		writePrometheusMetric(w, "go_memstats_heap_objects", "gauge", float64(memStats.HeapObjects))
		writePrometheusMetric(w, "go_goroutines", "gauge", float64(runtime.NumGoroutine()))
		writePrometheusMetric(w, "platform_workers_active", "gauge", float64(queueSummary.Active))
		writePrometheusMetric(w, "platform_queue_depth", "gauge", float64(queueSummary.Queued))
		writePrometheusMetric(w, "platform_http_requests_total", "counter", float64(metrics.TotalRequests))
		writePrometheusMetric(w, "platform_http_request_errors_total", "counter", float64(metrics.TotalErrors))
		writeHistogram(w, "platform_http_request_duration_seconds", metrics)
	})
}

func writePrometheusMetric(w http.ResponseWriter, name, kind string, value float64) {
	fmt.Fprintf(w, "# TYPE %s %s\n", name, kind)
	fmt.Fprintf(w, "%s %s\n", name, strconv.FormatFloat(value, 'f', -1, 64))
}

func writeHistogram(w http.ResponseWriter, name string, snapshot MetricsSnapshot) {
	fmt.Fprintf(w, "# TYPE %s histogram\n", name)
	for _, bucket := range snapshot.HTTPRequestDurationBuckets {
		fmt.Fprintf(w, "%s_bucket{le=\"%s\"} %d\n", name, strconv.FormatFloat(bucket.UpperBoundSeconds, 'f', -1, 64), bucket.Count)
	}
	fmt.Fprintf(w, "%s_bucket{le=\"+Inf\"} %d\n", name, snapshot.HTTPRequestDurationCount)
	fmt.Fprintf(w, "%s_sum %s\n", name, strconv.FormatFloat(snapshot.HTTPRequestDurationSum, 'f', -1, 64))
	fmt.Fprintf(w, "%s_count %d\n", name, snapshot.HTTPRequestDurationCount)
}
