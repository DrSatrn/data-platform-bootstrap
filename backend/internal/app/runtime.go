// Package app wires together runtime configuration, logging, repositories, and
// bounded-context services into executable processes. The goal is to keep each
// binary small while ensuring the runtime graph remains explicit and testable.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/streanor/data-platform/backend/internal/admin"
	"github.com/streanor/data-platform/backend/internal/analytics"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/db"
	"github.com/streanor/data-platform/backend/internal/execution"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/observability"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/quality"
	"github.com/streanor/data-platform/backend/internal/reporting"
	"github.com/streanor/data-platform/backend/internal/scheduler"
	"github.com/streanor/data-platform/backend/internal/storage"
)

type runtimePersistence struct {
	store     orchestration.Store
	queue     orchestration.RunQueue
	artifacts *storage.Service
}

// RunAPI starts the HTTP control-plane server.
func RunAPI(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	telemetry := observability.NewService()
	logger := observability.NewLogger(cfg.LogLevel, telemetry)
	persistence, err := buildRuntimePersistence(ctx, cfg, logger)
	if err != nil {
		return err
	}
	router := newRouter(logger, cfg, telemetry, persistence)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting API server", slog.String("addr", cfg.HTTPAddr))
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		logger.Info("shutting down API server")
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == nil || err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

// RunScheduler starts the lightweight scheduling loop.
func RunScheduler(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := observability.NewLogger(cfg.LogLevel, nil)
	loader := manifests.NewLoader(cfg.ManifestRoot)
	persistence, err := buildRuntimePersistence(ctx, cfg, logger)
	if err != nil {
		return err
	}
	catalog := metadata.NewCatalog()
	control := orchestration.NewControlService(loader, persistence.store, persistence.queue)
	service := scheduler.NewService(cfg.SchedulerTick, loader, persistence.store, control, catalog, logger, cfg.DataRoot)

	logger.Info("starting scheduler loop", slog.Duration("tick", cfg.SchedulerTick))
	return service.Run(ctx)
}

// RunWorker starts the execution loop placeholder.
func RunWorker(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := observability.NewLogger(cfg.LogLevel, nil)
	loader := manifests.NewLoader(cfg.ManifestRoot)
	persistence, err := buildRuntimePersistence(ctx, cfg, logger)
	if err != nil {
		return err
	}
	runner := execution.NewRunner(cfg, loader, persistence.store, persistence.artifacts, logger)
	worker := execution.NewWorker(persistence.queue, runner, logger, cfg.WorkerPoll)
	logger.Info("starting worker loop", slog.Duration("poll", cfg.WorkerPoll))
	return worker.Run(ctx)
}

func newRouter(logger *slog.Logger, cfg config.Settings, telemetry *observability.Service, persistence runtimePersistence) http.Handler {
	loader := manifests.NewLoader(cfg.ManifestRoot)
	catalog := metadata.NewCatalog()
	qualityService := quality.NewService(cfg.SampleDataRoot, cfg.DataRoot, cfg.DuckDBPath, cfg.SQLRoot)
	var reportStore reporting.Store
	reportStore, err := reporting.NewFileStore(cfg.DataRoot, cfg.DashboardRoot)
	if err != nil {
		logger.Warn("file-backed dashboard store unavailable, falling back to memory store", slog.String("reason", err.Error()))
		reportStore = reporting.NewMemoryStore()
	}
	analyticsService := analytics.NewService(cfg.SampleDataRoot, cfg.DataRoot, cfg.DuckDBPath, cfg.SQLRoot)
	controlService := orchestration.NewControlService(loader, persistence.store, persistence.queue)
	adminService := admin.NewService(cfg, loader, persistence.store, controlService, qualityService, reportStore, persistence.artifacts, telemetry)

	mux := http.NewServeMux()
	mux.Handle("/healthz", observability.HealthHandler(cfg))
	mux.Handle("/api/v1/pipelines", orchestration.NewPipelineHandler(loader, persistence.store, controlService, logger))
	mux.Handle("/api/v1/catalog", metadata.NewCatalogHandler(loader, catalog))
	mux.Handle("/api/v1/quality", quality.NewHandler(qualityService))
	mux.Handle("/api/v1/analytics", analytics.NewHandler(analyticsService))
	mux.Handle("/api/v1/reports", reporting.NewHandler(reportStore))
	mux.Handle("/api/v1/artifacts", storage.NewHandler(persistence.artifacts))
	mux.Handle("/api/v1/system/overview", observability.NewOverviewHandler(cfg, telemetry, loader, loader, persistence.store))
	mux.Handle("/api/v1/system/logs", observability.NewRecentLogsHandler(telemetry))
	mux.Handle("/api/v1/admin/terminal/execute", admin.NewHandler(cfg, adminService))

	return observability.RequestLoggingMiddleware(logger, telemetry, mux)
}

func buildRuntimePersistence(ctx context.Context, cfg config.Settings, logger *slog.Logger) (runtimePersistence, error) {
	fileStore, err := orchestration.NewFileStore(cfg.DataRoot)
	if err != nil {
		return runtimePersistence{}, err
	}
	fileQueue, err := orchestration.NewQueue(cfg.DataRoot)
	if err != nil {
		return runtimePersistence{}, err
	}
	fileArtifacts := storage.NewService(cfg.ArtifactRoot, nil)

	controlPlane, err := db.NewControlPlane(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Warn("postgres control plane disabled", slog.String("reason", err.Error()))
		return runtimePersistence{
			store:     fileStore,
			queue:     fileQueue,
			artifacts: fileArtifacts,
		}, nil
	}

	logger.Info("postgres control plane enabled")
	return runtimePersistence{
		store:     orchestration.NewMultiStore(controlPlane.RunStore, fileStore),
		queue:     controlPlane.RunQueue,
		artifacts: storage.NewService(cfg.ArtifactRoot, controlPlane.ArtifactIdx),
	}, nil
}

// ExplainConfig returns a compact human-readable configuration summary that is
// useful for diagnostics and onboarding docs.
func ExplainConfig(cfg config.Settings) string {
	return fmt.Sprintf(
		"env=%s http=%s manifests=%s data_root=%s",
		cfg.Environment,
		cfg.HTTPAddr,
		cfg.ManifestRoot,
		cfg.DataRoot,
	)
}
