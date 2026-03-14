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

	"github.com/streanor/data-platform/backend/internal/analytics"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/observability"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/quality"
	"github.com/streanor/data-platform/backend/internal/reporting"
	"github.com/streanor/data-platform/backend/internal/scheduler"
)

// RunAPI starts the HTTP control-plane server.
func RunAPI(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := observability.NewLogger(cfg.LogLevel)
	router := newRouter(logger, cfg)

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

	logger := observability.NewLogger(cfg.LogLevel)
	loader := manifests.NewLoader(cfg.ManifestRoot)
	store := orchestration.NewInMemoryStore()
	catalog := metadata.NewCatalog()
	service := scheduler.NewService(cfg.SchedulerTick, loader, store, catalog, logger)

	logger.Info("starting scheduler loop", slog.Duration("tick", cfg.SchedulerTick))
	return service.Run(ctx)
}

// RunWorker starts the execution loop placeholder.
func RunWorker(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := observability.NewLogger(cfg.LogLevel)
	logger.Info("starting worker loop", slog.Duration("poll", cfg.WorkerPoll))

	<-ctx.Done()
	logger.Info("worker shutdown complete")
	return nil
}

func newRouter(logger *slog.Logger, cfg config.Settings) http.Handler {
	loader := manifests.NewLoader(cfg.ManifestRoot)
	store := orchestration.NewInMemoryStore()
	catalog := metadata.NewCatalog()
	qualityService := quality.NewService()
	reportStore := reporting.NewMemoryStore()
	analyticsService := analytics.NewService()

	mux := http.NewServeMux()
	mux.Handle("/healthz", observability.HealthHandler(cfg))
	mux.Handle("/api/v1/pipelines", orchestration.NewPipelineHandler(loader, store, logger))
	mux.Handle("/api/v1/catalog", metadata.NewCatalogHandler(loader, catalog))
	mux.Handle("/api/v1/quality", quality.NewHandler(qualityService))
	mux.Handle("/api/v1/analytics", analytics.NewHandler(analyticsService))
	mux.Handle("/api/v1/reports", reporting.NewHandler(reportStore))

	return observability.RequestLoggingMiddleware(logger, mux)
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
