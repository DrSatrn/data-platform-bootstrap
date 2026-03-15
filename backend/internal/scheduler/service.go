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

	"github.com/streanor/data-platform/backend/internal/alerting"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// Service owns the scheduling loop.
type Service struct {
	tick       time.Duration
	loader     manifests.Loader
	store      orchestration.Store
	control    *orchestration.ControlService
	catalog    *metadata.Catalog
	metaStore  metadata.Store
	alerts     *alerting.Dispatcher
	logger     *slog.Logger
	dataRoot   string
	statePath  string
	statusPath string
}

// Status captures the latest scheduler refresh heartbeat so operators and
// benchmarks can verify that scheduled execution is still alive.
type Status struct {
	RefreshedAt   time.Time  `json:"refreshed_at"`
	PipelineCount int        `json:"pipeline_count"`
	AssetCount    int        `json:"asset_count"`
	LastEnqueueAt *time.Time `json:"last_enqueue_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
}

// NewService creates a scheduler with explicit dependencies.
func NewService(
	tick time.Duration,
	loader manifests.Loader,
	store orchestration.Store,
	control *orchestration.ControlService,
	catalog *metadata.Catalog,
	metaStore metadata.Store,
	alerts *alerting.Dispatcher,
	logger *slog.Logger,
	dataRoot string,
) *Service {
	return &Service{
		tick:       tick,
		loader:     loader,
		store:      store,
		control:    control,
		catalog:    catalog,
		metaStore:  metaStore,
		alerts:     alerts,
		logger:     logger,
		dataRoot:   dataRoot,
		statePath:  filepath.Join(dataRoot, "control_plane", "scheduler_state.json"),
		statusPath: filepath.Join(dataRoot, "control_plane", "scheduler_status.json"),
	}
}

// Run executes the periodic scheduler loop until shutdown.
func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()

	if err := s.refreshCatalog(ctx); err != nil {
		_ = s.writeStatus(Status{RefreshedAt: time.Now().UTC(), LastError: err.Error()})
		s.logger.Warn("initial scheduler refresh failed", slog.String("error", err.Error()))
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler shutdown complete")
			return nil
		case <-ticker.C:
			if err := s.refreshCatalog(ctx); err != nil {
				_ = s.writeStatus(Status{RefreshedAt: time.Now().UTC(), LastError: err.Error()})
				s.logger.Warn("scheduled refresh failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (s *Service) refreshCatalog(ctx context.Context) error {
	pipelines, err := s.loader.LoadPipelines()
	if err != nil {
		return err
	}

	assets, err := s.loader.LoadAssets()
	if err != nil {
		return err
	}

	if err := metadata.ProjectStore(s.loader, s.metaStore); err != nil {
		s.logger.Warn("metadata projection refresh failed", slog.String("error", err.Error()))
	}

	now := time.Now().UTC()
	for index := range assets {
		assets[index].FreshnessStatus = metadata.ResolveFreshnessStatus(s.dataRoot, assets[index], now)
		s.alertAssetStatus(ctx, assets[index])
	}
	s.catalog.ReplaceAssets(assets)
	lastEnqueueAt, err := s.enqueueScheduledRuns(pipelines)
	if err != nil {
		return err
	}
	if err := s.writeStatus(Status{
		RefreshedAt:   time.Now().UTC(),
		PipelineCount: len(pipelines),
		AssetCount:    len(assets),
		LastEnqueueAt: lastEnqueueAt,
	}); err != nil {
		s.logger.Warn("failed to write scheduler heartbeat", slog.String("error", err.Error()))
	}
	s.logger.Info(
		"scheduler refresh complete",
		slog.Int("pipeline_count", len(pipelines)),
		slog.Int("asset_count", len(assets)),
	)
	return nil
}

func (s *Service) alertAssetStatus(ctx context.Context, asset metadata.DataAsset) {
	if s.alerts == nil {
		return
	}
	if err := s.alerts.ObserveAssetWarning(ctx, alerting.AssetWarningEvent{
		AssetID:        asset.ID,
		AssetName:      asset.Name,
		State:          asset.FreshnessStatus.State,
		Message:        asset.FreshnessStatus.Message,
		LastUpdated:    asset.FreshnessStatus.LastUpdated,
		LagSeconds:     asset.FreshnessStatus.LagSeconds,
		ExpectedWithin: asset.Freshness.ExpectedWithin,
		WarnAfter:      asset.Freshness.WarnAfter,
		ObservedAt:     time.Now().UTC(),
	}); err != nil {
		s.logger.Warn("failed to post asset freshness webhook", slog.String("asset_id", asset.ID), slog.String("error", err.Error()))
	}
}

func (s *Service) enqueueScheduledRuns(pipelines []orchestration.Pipeline) (*time.Time, error) {
	state, err := s.loadState()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var lastEnqueueAt *time.Time
	for _, pipeline := range pipelines {
		if pipeline.Schedule.IsPaused {
			continue
		}

		slot, supported, err := currentScheduleSlot(now, pipeline.Schedule)
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
		slotCopy := now
		lastEnqueueAt = &slotCopy
		s.logger.Info("scheduler queued pipeline run", slog.String("pipeline_id", pipeline.ID), slog.Time("slot", slot))
	}

	return lastEnqueueAt, s.saveState(state)
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

func (s *Service) writeStatus(status Status) error {
	if err := os.MkdirAll(filepath.Dir(s.statusPath), 0o755); err != nil {
		return fmt.Errorf("create scheduler status dir: %w", err)
	}
	bytes, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("encode scheduler status: %w", err)
	}
	return os.WriteFile(s.statusPath, bytes, 0o644)
}

func currentScheduleSlot(now time.Time, schedule orchestration.Schedule) (time.Time, bool, error) {
	location := time.UTC
	if schedule.Timezone != "" {
		resolved, err := time.LoadLocation(schedule.Timezone)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("load timezone %q: %w", schedule.Timezone, err)
		}
		location = resolved
	}

	minutes, err := parseCronField(fieldSpec{
		value: cronFieldValue(schedule.Cron, 0),
		min:   0,
		max:   59,
	})
	if err != nil {
		return time.Time{}, false, err
	}
	hours, err := parseCronField(fieldSpec{
		value: cronFieldValue(schedule.Cron, 1),
		min:   0,
		max:   23,
	})
	if err != nil {
		return time.Time{}, false, err
	}
	daysOfMonth, err := parseCronField(fieldSpec{
		value: cronFieldValue(schedule.Cron, 2),
		min:   1,
		max:   31,
	})
	if err != nil {
		return time.Time{}, false, err
	}
	months, err := parseCronField(fieldSpec{
		value: cronFieldValue(schedule.Cron, 3),
		min:   1,
		max:   12,
	})
	if err != nil {
		return time.Time{}, false, err
	}
	daysOfWeek, err := parseCronField(fieldSpec{
		value:       cronFieldValue(schedule.Cron, 4),
		min:         0,
		max:         6,
		allowSeven:  true,
		convertZero: true,
	})
	if err != nil {
		return time.Time{}, false, err
	}

	localNow := now.In(location).Truncate(time.Minute)
	for attempts := 0; attempts < 60*24*32; attempts++ {
		if matchesCron(localNow, minutes, hours, daysOfMonth, months, daysOfWeek) {
			return localNow.UTC(), true, nil
		}
		localNow = localNow.Add(-1 * time.Minute)
	}
	return time.Time{}, false, nil
}

type fieldSpec struct {
	value       string
	min         int
	max         int
	allowSeven  bool
	convertZero bool
}

type cronField map[int]struct{}

func parseCronField(spec fieldSpec) (cronField, error) {
	value := strings.TrimSpace(spec.value)
	if value == "" {
		return nil, fmt.Errorf("cron field cannot be empty")
	}
	if value == "*" {
		field := cronField{}
		for item := spec.min; item <= spec.max; item++ {
			field[item] = struct{}{}
		}
		return field, nil
	}

	field := cronField{}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, "*/"):
			step, err := strconv.Atoi(strings.TrimPrefix(part, "*/"))
			if err != nil || step <= 0 {
				return nil, fmt.Errorf("invalid cron step %q", part)
			}
			for item := spec.min; item <= spec.max; item += step {
				field[normalizeCronValue(item, spec)] = struct{}{}
			}
		default:
			parsed, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid cron field %q", part)
			}
			if parsed < spec.min || (parsed > spec.max && !(spec.allowSeven && parsed == 7)) {
				return nil, fmt.Errorf("cron field %q out of range", part)
			}
			field[normalizeCronValue(parsed, spec)] = struct{}{}
		}
	}
	return field, nil
}

func normalizeCronValue(value int, spec fieldSpec) int {
	if spec.allowSeven && spec.convertZero && value == 7 {
		return 0
	}
	return value
}

func matchesCron(now time.Time, minutes, hours, daysOfMonth, months, daysOfWeek cronField) bool {
	if _, ok := minutes[now.Minute()]; !ok {
		return false
	}
	if _, ok := hours[now.Hour()]; !ok {
		return false
	}
	if _, ok := daysOfMonth[now.Day()]; !ok {
		return false
	}
	if _, ok := months[int(now.Month())]; !ok {
		return false
	}
	if _, ok := daysOfWeek[int(now.Weekday())]; !ok {
		return false
	}
	return true
}

func cronFieldValue(cron string, index int) string {
	fields := strings.Fields(strings.TrimSpace(cron))
	if len(fields) != 5 {
		return ""
	}
	return fields[index]
}
