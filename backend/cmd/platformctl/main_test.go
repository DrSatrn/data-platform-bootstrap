// These tests cover the benchmark math helpers because benchmark summaries are
// only useful if their aggregates stay deterministic and easy to trust.
package main

import (
	"testing"
	"time"
)

func TestAverage(t *testing.T) {
	if got := average([]float64{10, 20, 30}); got != 20 {
		t.Fatalf("expected average 20, got %v", got)
	}
}

func TestPercentile(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}
	if got := percentile(values, 50); got != 30 {
		t.Fatalf("expected p50 30, got %v", got)
	}
	if got := percentile(values, 95); got < 40 || got > 50 {
		t.Fatalf("expected p95 between 40 and 50, got %v", got)
	}
}

func TestBuildBenchmarkAssertionsIncludesQueueAndSchedulerChecks(t *testing.T) {
	assertions := buildBenchmarkAssertions(
		[]benchmarkResult{{Name: "health", Iterations: 5, Successes: 5}},
		&benchmarkLoadScenario{
			RequestedTriggers: 3,
			AcceptedTriggers:  3,
			QueueTotalBefore:  10,
			QueueTotalAfter:   13,
			QueueVisibleMS:    400,
		},
		benchmarkSchedulerSummary{
			RefreshedAt: time.Now().UTC(),
			LagSeconds:  5,
		},
		2*time.Second,
		30*time.Second,
	)

	if len(assertions) < 3 {
		t.Fatalf("expected queue and scheduler assertions, got %#v", assertions)
	}
	for _, assertion := range assertions {
		if assertion.Status == "fail" {
			t.Fatalf("expected all assertions to pass, got %#v", assertions)
		}
	}
}

func TestBuildBenchmarkAssertionsFailsStaleScheduler(t *testing.T) {
	assertions := buildBenchmarkAssertions(
		[]benchmarkResult{{Name: "health", Iterations: 5, Successes: 5}},
		nil,
		benchmarkSchedulerSummary{
			RefreshedAt: time.Now().UTC().Add(-10 * time.Minute),
			LagSeconds:  600,
		},
		2*time.Second,
		30*time.Second,
	)

	foundFailure := false
	for _, assertion := range assertions {
		if assertion.Name == "scheduler_heartbeat_freshness" && assertion.Status == "fail" {
			foundFailure = true
		}
	}
	if !foundFailure {
		t.Fatalf("expected stale scheduler assertion failure, got %#v", assertions)
	}
}
