// Package alerting posts structured operational events to configured webhook
// destinations so operators can detect failures without polling the UI.
package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const defaultTimeout = 5 * time.Second

// Settings captures the environment-driven webhook destinations for operator
// alerts. Empty URL lists leave alerting disabled without affecting execution.
type Settings struct {
	Environment      string
	RunFailureURLs   []string
	AssetWarningURLs []string
	WebhookTimeout   time.Duration
}

// HTTPDoer matches the subset of http.Client used by the dispatcher.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Dispatcher de-duplicates operator alerts and posts JSON payloads to the
// configured webhook endpoints.
type Dispatcher struct {
	client           HTTPDoer
	environment      string
	runFailureURLs   []string
	assetWarningURLs []string
	timeout          time.Duration

	mu            sync.Mutex
	notifiedRuns  map[string]struct{}
	assetStatuses map[string]string
}

// RunFailureEvent captures the operator-facing facts for a failed pipeline run.
type RunFailureEvent struct {
	RunID      string
	PipelineID string
	Pipeline   string
	Trigger    string
	JobID      string
	Error      string
	FailedAt   time.Time
}

// AssetWarningEvent captures a freshness state transition for one asset.
type AssetWarningEvent struct {
	AssetID        string
	AssetName      string
	State          string
	Message        string
	LastUpdated    string
	LagSeconds     int64
	ExpectedWithin string
	WarnAfter      string
	ObservedAt     time.Time
}

type webhookEnvelope struct {
	EventType   string        `json:"event_type"`
	Environment string        `json:"environment"`
	Severity    string        `json:"severity"`
	OccurredAt  string        `json:"occurred_at"`
	Summary     string        `json:"summary"`
	Run         *runPayload   `json:"run,omitempty"`
	Asset       *assetPayload `json:"asset,omitempty"`
}

type runPayload struct {
	ID         string `json:"id"`
	PipelineID string `json:"pipeline_id"`
	Pipeline   string `json:"pipeline,omitempty"`
	Trigger    string `json:"trigger,omitempty"`
	JobID      string `json:"job_id,omitempty"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

type assetPayload struct {
	ID             string `json:"id"`
	Name           string `json:"name,omitempty"`
	State          string `json:"state"`
	Message        string `json:"message"`
	LastUpdated    string `json:"last_updated,omitempty"`
	LagSeconds     int64  `json:"lag_seconds,omitempty"`
	ExpectedWithin string `json:"expected_within,omitempty"`
	WarnAfter      string `json:"warn_after,omitempty"`
}

// NewDispatcher constructs a webhook dispatcher. Nil clients default to a
// standard http.Client with the configured timeout.
func NewDispatcher(settings Settings, client HTTPDoer) *Dispatcher {
	timeout := settings.WebhookTimeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	return &Dispatcher{
		client:           client,
		environment:      strings.TrimSpace(settings.Environment),
		runFailureURLs:   cloneStrings(settings.RunFailureURLs),
		assetWarningURLs: cloneStrings(settings.AssetWarningURLs),
		timeout:          timeout,
		notifiedRuns:     map[string]struct{}{},
		assetStatuses:    map[string]string{},
	}
}

// NotifyRunFailure posts a failure alert once per run ID.
func (d *Dispatcher) NotifyRunFailure(ctx context.Context, event RunFailureEvent) error {
	if len(d.runFailureURLs) == 0 || strings.TrimSpace(event.RunID) == "" {
		return nil
	}

	d.mu.Lock()
	if _, exists := d.notifiedRuns[event.RunID]; exists {
		d.mu.Unlock()
		return nil
	}
	d.notifiedRuns[event.RunID] = struct{}{}
	d.mu.Unlock()

	occurredAt := event.FailedAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	payload := webhookEnvelope{
		EventType:   "pipeline_run_failed",
		Environment: d.environment,
		Severity:    "error",
		OccurredAt:  occurredAt.Format(time.RFC3339),
		Summary:     fmt.Sprintf("Pipeline %s run %s failed", firstNonEmpty(event.PipelineID, event.Pipeline), event.RunID),
		Run: &runPayload{
			ID:         event.RunID,
			PipelineID: event.PipelineID,
			Pipeline:   event.Pipeline,
			Trigger:    event.Trigger,
			JobID:      event.JobID,
			Status:     "failed",
			Error:      event.Error,
		},
	}

	if err := d.postJSON(ctx, d.runFailureURLs, payload); err != nil {
		d.mu.Lock()
		delete(d.notifiedRuns, event.RunID)
		d.mu.Unlock()
		return err
	}
	return nil
}

// ObserveAssetWarning posts a webhook when an asset first enters the stale
// freshness state and suppresses repeated alerts until the asset recovers.
func (d *Dispatcher) ObserveAssetWarning(ctx context.Context, event AssetWarningEvent) error {
	if len(d.assetWarningURLs) == 0 || strings.TrimSpace(event.AssetID) == "" {
		return nil
	}

	state := strings.TrimSpace(event.State)
	d.mu.Lock()
	previous := d.assetStatuses[event.AssetID]
	d.assetStatuses[event.AssetID] = state
	d.mu.Unlock()

	if state != "stale" || previous == "stale" {
		return nil
	}

	occurredAt := event.ObservedAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	payload := webhookEnvelope{
		EventType:   "asset_warning_sla_breached",
		Environment: d.environment,
		Severity:    "warning",
		OccurredAt:  occurredAt.Format(time.RFC3339),
		Summary:     fmt.Sprintf("Asset %s exceeded its warning SLA", event.AssetID),
		Asset: &assetPayload{
			ID:             event.AssetID,
			Name:           event.AssetName,
			State:          state,
			Message:        event.Message,
			LastUpdated:    event.LastUpdated,
			LagSeconds:     event.LagSeconds,
			ExpectedWithin: event.ExpectedWithin,
			WarnAfter:      event.WarnAfter,
		},
	}

	if err := d.postJSON(ctx, d.assetWarningURLs, payload); err != nil {
		d.mu.Lock()
		d.assetStatuses[event.AssetID] = previous
		d.mu.Unlock()
		return err
	}
	return nil
}

func (d *Dispatcher) postJSON(parent context.Context, urls []string, payload webhookEnvelope) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(parent, d.timeout)
	defer cancel()

	var postErrs []error
	for _, rawURL := range urls {
		target := strings.TrimSpace(rawURL)
		if target == "" {
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
		if err != nil {
			postErrs = append(postErrs, fmt.Errorf("build webhook request for %s: %w", target, err))
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := d.client.Do(req)
		if err != nil {
			postErrs = append(postErrs, fmt.Errorf("post webhook to %s: %w", target, err))
			continue
		}
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			postErrs = append(postErrs, fmt.Errorf("post webhook to %s returned status %d", target, resp.StatusCode))
		}
	}
	return errors.Join(postErrs...)
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
