// These tests cover the benchmark math helpers because benchmark summaries are
// only useful if their aggregates stay deterministic and easy to trust.
package main

import "testing"

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
