// Package scheduler implements the local-first scheduling loop used by the
// platform. The current implementation refreshes metadata and enqueues due
// pipeline runs for the subset of cron expressions used by the sample slice.
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// Service owns the scheduling loop.
type Service struct {
	tick      time.Duration
	loader    manifests.Loader
	store     orchestration.Store
	control   *orchestration.ControlService
	catalog   *metadata.Catalog
	logger    *slog.Logger
	statePath string
}

// NewService creates a scheduler with explicit dependencies.
func NewService(
	tick time.Duration,
	loader manifests.Loader,
	store orchestration.Store,
	control *orchestration.ControlService,
	catalog *metadata.Catalog,
	logger *slog.Logger,
	dataRoot string,
) *Service {
	return &Service{
		tick:      tick,
		loader:    loader,
		store:     store,
		control:   control,
		catalog:   catalog,
		logger:    logger,
		statePath: filepath.Join(dataRoot, "control_plane", "scheduler_state.json"),
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
	if err := s.enqueueScheduledRuns(pipelines); err != nil {
		return err
	}
	s.logger.Info(
		"scheduler refresh complete",
		slog.Int("pipeline_count", len(pipelines)),
		slog.Int("asset_count", len(assets)),
	)
	return nil
}

func (s *Service) enqueueScheduledRuns(pipelines []orchestration.Pipeline) error {
	state, err := s.loadState()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, pipeline := range pipelines {
		if pipeline.Schedule.IsPaused {
			continue
		}

		slot, supported, err := currentScheduleSlot(now, pipeline.Schedule.Cron)
		if err != nil {
			s.logger.Warn("failed to evaluate pipeline schedule", slog.String("pipeline_id", pipeline.ID), slog.String("error", err.Error()))
			continue
		}
		if !supported || slot.IsZero() {
			continue
		}
		lastScheduled := state[pipeline.ID]
		if !slot.After(lastScheduled) {
			continue
		}

		if _, err := s.control.TriggerPipeline(context.Background(), pipeline.ID, "scheduled"); err != nil {
			s.logger.Warn("failed to enqueue scheduled pipeline run", slog.String("pipeline_id", pipeline.ID), slog.String("error", err.Error()))
			continue
		}
		state[pipeline.ID] = slot
		s.logger.Info("scheduler queued pipeline run", slog.String("pipeline_id", pipeline.ID), slog.Time("slot", slot))
	}

	return s.saveState(state)
}

func (s *Service) loadState() (map[string]time.Time, error) {
	if err := os.MkdirAll(filepath.Dir(s.statePath), 0o755); err != nil {
		return nil, fmt.Errorf("create scheduler state dir: %w", err)
	}
	bytes, err := os.ReadFile(s.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]time.Time{}, nil
		}
		return nil, fmt.Errorf("read scheduler state: %w", err)
	}
	if len(bytes) == 0 {
		return map[string]time.Time{}, nil
	}
	raw := map[string]string{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, fmt.Errorf("decode scheduler state: %w", err)
	}

	out := make(map[string]time.Time, len(raw))
	for pipelineID, value := range raw {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, fmt.Errorf("parse scheduler state for %s: %w", pipelineID, err)
		}
		out[pipelineID] = parsed
	}
	return out, nil
}

func (s *Service) saveState(state map[string]time.Time) error {
	raw := make(map[string]string, len(state))
	for pipelineID, value := range state {
		raw[pipelineID] = value.Format(time.RFC3339)
	}
	bytes, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("encode scheduler state: %w", err)
	}
	return os.WriteFile(s.statePath, bytes, 0o644)
}

func currentScheduleSlot(now time.Time, cron string) (time.Time, bool, error) {
	fields := strings.Fields(strings.TrimSpace(cron))
	if len(fields) != 5 {
		return time.Time{}, false, nil
	}
	if fields[2] != "*" || fields[3] != "*" || fields[4] != "*" {
		return time.Time{}, false, nil
	}

	minute, err := strconv.Atoi(fields[0])
	if err != nil || minute < 0 || minute > 59 {
		return time.Time{}, false, fmt.Errorf("unsupported cron minute field %q", fields[0])
	}

	hourField := fields[1]
	switch {
	case hourField == "*":
		slot := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minute, 0, 0, time.UTC)
		if slot.After(now) {
			slot = slot.Add(-1 * time.Hour)
		}
		return slot, true, nil
	case strings.HasPrefix(hourField, "*/"):
		intervalHours, err := strconv.Atoi(strings.TrimPrefix(hourField, "*/"))
		if err != nil || intervalHours <= 0 || intervalHours > 24 {
			return time.Time{}, false, fmt.Errorf("unsupported cron hour interval %q", hourField)
		}
		hour := (now.Hour() / intervalHours) * intervalHours
		slot := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
		if slot.After(now) {
			slot = slot.Add(-1 * time.Duration(intervalHours) * time.Hour)
		}
		return slot, true, nil
	default:
		hour, err := strconv.Atoi(hourField)
		if err != nil || hour < 0 || hour > 23 {
			return time.Time{}, false, fmt.Errorf("unsupported cron hour field %q", hourField)
		}
		slot := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
		if slot.After(now) {
			slot = slot.Add(-24 * time.Hour)
		}
		return slot, true, nil
	}
}
