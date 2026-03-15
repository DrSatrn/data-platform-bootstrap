// This file exposes first-party diagnostics APIs for system overview and log
// inspection. These endpoints back the built-in admin portal rather than
// relying on external metrics dashboards.
package observability

import (
	"net/http"
	"time"

	"github.com/streanor/data-platform/backend/internal/backup"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/shared"
)

// PipelineLoader describes the pipeline loading dependency needed for system
// overview responses.
type PipelineLoader interface {
	LoadPipelines() ([]orchestration.Pipeline, error)
}

// AssetLoader describes the asset loading dependency needed for system
// overview responses.
type AssetLoader interface {
	LoadAssets() ([]metadata.DataAsset, error)
}

// QueueSnapshotter describes the queue visibility needed for operational
// summary responses.
type QueueSnapshotter interface {
	ListRequests() ([]orchestration.QueueSnapshot, error)
}

// BackupInventory describes the recovery inventory behavior needed for the
// system overview.
type BackupInventory interface {
	ListBundles() ([]backup.BundleFile, error)
}

// OverviewHandler returns a system summary payload.
type OverviewHandler struct {
	cfg        config.Settings
	service    *Service
	pipelines  PipelineLoader
	assets     AssetLoader
	runHistory orchestration.Store
	queue      QueueSnapshotter
	backups    BackupInventory
}

// NewOverviewHandler constructs the system overview handler.
func NewOverviewHandler(
	cfg config.Settings,
	service *Service,
	pipelines PipelineLoader,
	assets AssetLoader,
	runHistory orchestration.Store,
	queue QueueSnapshotter,
	backups BackupInventory,
) http.Handler {
	return &OverviewHandler{
		cfg:        cfg,
		service:    service,
		pipelines:  pipelines,
		assets:     assets,
		runHistory: runHistory,
		queue:      queue,
		backups:    backups,
	}
}

func (h *OverviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pipelines, err := h.pipelines.LoadPipelines()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load pipelines for system overview",
		})
		return
	}

	assets, err := h.assets.LoadAssets()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load assets for system overview",
		})
		return
	}

	runHistory, err := h.runHistory.ListPipelineRuns()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load run history for system overview",
		})
		return
	}

	queueSummary := summarizeQueue(nil)
	if h.queue != nil {
		if requests, err := h.queue.ListRequests(); err == nil {
			queueSummary = summarizeQueue(requests)
		}
	}

	backupSummary := summarizeBackups(nil)
	if h.backups != nil {
		if bundles, err := h.backups.ListBundles(); err == nil {
			backupSummary = summarizeBackups(bundles)
		}
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"environment":     h.cfg.Environment,
		"http_addr":       h.cfg.HTTPAddr,
		"web_addr":        h.cfg.WebAddr,
		"known_pipelines": len(pipelines),
		"known_assets":    len(assets),
		"run_history":     len(runHistory),
		"run_summary":     summarizeRuns(runHistory),
		"queue_summary":   queueSummary,
		"backup_summary":  backupSummary,
		"telemetry": h.service.Snapshot(map[string]string{
			"environment": h.cfg.Environment,
			"api_base":    h.cfg.APIBaseURL,
		}),
	})
}

// RecentLogsHandler returns the in-memory log buffer.
type RecentLogsHandler struct {
	service *Service
}

// NewRecentLogsHandler constructs the recent log API.
func NewRecentLogsHandler(service *Service) http.Handler {
	return &RecentLogsHandler{service: service}
}

func (h *RecentLogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"logs": h.service.RecentLogs(),
	})
}

type runSummary struct {
	TotalRuns              int    `json:"total_runs"`
	QueuedRuns             int    `json:"queued_runs"`
	RunningRuns            int    `json:"running_runs"`
	SucceededRuns          int    `json:"succeeded_runs"`
	FailedRuns             int    `json:"failed_runs"`
	CompletedLast24Hours   int    `json:"completed_last_24_hours"`
	FailedLast24Hours      int    `json:"failed_last_24_hours"`
	AverageDurationSeconds int64  `json:"average_duration_seconds"`
	LatestFailureRunID     string `json:"latest_failure_run_id,omitempty"`
	LatestFailureMessage   string `json:"latest_failure_message,omitempty"`
}

type queueSummary struct {
	Queued    int `json:"queued"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

type backupSummary struct {
	BundleCount       int    `json:"bundle_count"`
	LatestBundlePath  string `json:"latest_bundle_path,omitempty"`
	LatestBundleBytes int64  `json:"latest_bundle_bytes,omitempty"`
}

func summarizeRuns(runs []orchestration.PipelineRun) runSummary {
	summary := runSummary{TotalRuns: len(runs)}
	var (
		durationTotal time.Duration
		durationCount int
		cutoff        = time.Now().UTC().Add(-24 * time.Hour)
	)
	for _, run := range runs {
		switch run.Status {
		case orchestration.RunStatusQueued:
			summary.QueuedRuns++
		case orchestration.RunStatusRunning:
			summary.RunningRuns++
		case orchestration.RunStatusSucceeded:
			summary.SucceededRuns++
		case orchestration.RunStatusFailed:
			summary.FailedRuns++
			if summary.LatestFailureRunID == "" {
				summary.LatestFailureRunID = run.ID
				summary.LatestFailureMessage = run.Error
			}
		}
		if run.FinishedAt != nil {
			if run.FinishedAt.After(cutoff) {
				summary.CompletedLast24Hours++
				if run.Status == orchestration.RunStatusFailed {
					summary.FailedLast24Hours++
				}
			}
			if run.FinishedAt.After(run.StartedAt) {
				durationTotal += run.FinishedAt.Sub(run.StartedAt)
				durationCount++
			}
		}
	}
	if durationCount > 0 {
		summary.AverageDurationSeconds = int64((durationTotal / time.Duration(durationCount)).Seconds())
	}
	return summary
}

func summarizeQueue(requests []orchestration.QueueSnapshot) queueSummary {
	summary := queueSummary{Total: len(requests)}
	for _, request := range requests {
		switch request.Status {
		case "queued":
			summary.Queued++
		case "active":
			summary.Active++
		case "completed":
			summary.Completed++
		}
	}
	return summary
}

func summarizeBackups(bundles []backup.BundleFile) backupSummary {
	summary := backupSummary{BundleCount: len(bundles)}
	if len(bundles) == 0 {
		return summary
	}
	summary.LatestBundlePath = bundles[0].Path
	summary.LatestBundleBytes = bundles[0].SizeBytes
	return summary
}
