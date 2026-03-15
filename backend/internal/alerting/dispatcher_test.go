package alerting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNotifyRunFailurePostsStructuredWebhook(t *testing.T) {
	var (
		mu      sync.Mutex
		payload webhookEnvelope
		calls   int
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		mu.Lock()
		defer mu.Unlock()
		calls++
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	dispatcher := NewDispatcher(Settings{
		Environment:    "test",
		RunFailureURLs: []string{server.URL},
	}, nil)

	err := dispatcher.NotifyRunFailure(context.Background(), RunFailureEvent{
		RunID:      "run_123",
		PipelineID: "finance",
		Pipeline:   "Finance Pipeline",
		Trigger:    "manual",
		JobID:      "build_dbt",
		Error:      "exit status 7",
		FailedAt:   time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("notify run failure: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("expected 1 webhook call, got %d", calls)
	}
	if payload.EventType != "pipeline_run_failed" {
		t.Fatalf("expected pipeline_run_failed event type, got %q", payload.EventType)
	}
	if payload.Run == nil || payload.Run.ID != "run_123" || payload.Run.JobID != "build_dbt" {
		t.Fatalf("unexpected run payload: %+v", payload.Run)
	}
	if payload.Run.Error != "exit status 7" {
		t.Fatalf("expected run error in payload, got %+v", payload.Run)
	}
}

func TestObserveAssetWarningSuppressesRepeatedStaleAlerts(t *testing.T) {
	var (
		mu       sync.Mutex
		payloads []webhookEnvelope
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload webhookEnvelope
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		mu.Lock()
		payloads = append(payloads, payload)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewDispatcher(Settings{
		Environment:      "test",
		AssetWarningURLs: []string{server.URL},
	}, nil)

	event := AssetWarningEvent{
		AssetID:        "mart_cashflow",
		AssetName:      "Monthly Cashflow",
		State:          "stale",
		Message:        "Asset is past its warning SLA of 48h.",
		LastUpdated:    "2026-03-10T00:00:00Z",
		LagSeconds:     360000,
		ExpectedWithin: "24h",
		WarnAfter:      "48h",
		ObservedAt:     time.Date(2026, 3, 15, 13, 0, 0, 0, time.UTC),
	}
	if err := dispatcher.ObserveAssetWarning(context.Background(), event); err != nil {
		t.Fatalf("first stale alert: %v", err)
	}
	if err := dispatcher.ObserveAssetWarning(context.Background(), event); err != nil {
		t.Fatalf("second stale alert: %v", err)
	}
	if err := dispatcher.ObserveAssetWarning(context.Background(), AssetWarningEvent{AssetID: "mart_cashflow", State: "fresh"}); err != nil {
		t.Fatalf("fresh transition: %v", err)
	}
	if err := dispatcher.ObserveAssetWarning(context.Background(), event); err != nil {
		t.Fatalf("stale after recovery: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(payloads) != 2 {
		t.Fatalf("expected 2 stale alert webhooks, got %d", len(payloads))
	}
	if payloads[0].EventType != "asset_warning_sla_breached" {
		t.Fatalf("unexpected asset event type %q", payloads[0].EventType)
	}
	if payloads[0].Asset == nil || payloads[0].Asset.ID != "mart_cashflow" {
		t.Fatalf("unexpected asset payload: %+v", payloads[0].Asset)
	}
}
