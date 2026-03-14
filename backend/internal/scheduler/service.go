// Package scheduler implements the lightweight scheduling loop used by the
// first platform slice. It currently focuses on manifest discovery and run
// registration so the system has a visible control-plane heartbeat.
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// Service owns the scheduling loop.
type Service struct {
	tick    time.Duration
	loader  manifests.Loader
	store   orchestration.Store
	catalog *metadata.Catalog
	logger  *slog.Logger
}

// NewService creates a scheduler with explicit dependencies.
func NewService(
	tick time.Duration,
	loader manifests.Loader,
	store orchestration.Store,
	catalog *metadata.Catalog,
	logger *slog.Logger,
) *Service {
	return &Service{
		tick:    tick,
		loader:  loader,
		store:   store,
		catalog: catalog,
		logger:  logger,
	}
}

// Run executes the periodic scheduler loop until shutdown.
func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()

	if err := s.refreshCatalog(); err != nil {
		s.logger.Warn("initial scheduler refresh failed", slog.String("error", err.Error()))
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler shutdown complete")
			return nil
		case <-ticker.C:
			if err := s.refreshCatalog(); err != nil {
				s.logger.Warn("scheduled refresh failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (s *Service) refreshCatalog() error {
	pipelines, err := s.loader.LoadPipelines()
	if err != nil {
		return err
	}

	assets, err := s.loader.LoadAssets()
	if err != nil {
		return err
	}

	s.catalog.ReplaceAssets(assets)
	s.logger.Info(
		"scheduler refresh complete",
		slog.Int("pipeline_count", len(pipelines)),
		slog.Int("asset_count", len(assets)),
	)
	return nil
}
