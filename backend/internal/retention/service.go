// Package retention enforces repo-declared data retention windows. The purge
// workflow is designed for cron or operator CLI use so it can safely clean up
// stale files and mirrored control-plane rows without expanding the core API.
package retention

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// DBExecutor captures the narrow SQL behavior needed to remove mirrored
// control-plane rows during retention purges.
type DBExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// Settings declares the runtime paths and fallback windows for one purge run.
type Settings struct {
	DataRoot              string
	ArtifactRoot          string
	Now                   time.Time
	DefaultRunHistoryTTL  time.Duration
	DefaultRunArtifactTTL time.Duration
}

// Report describes what one purge invocation removed.
type Report struct {
	StartedAt                   time.Time `json:"started_at"`
	MaterializationsRemoved     []string  `json:"materializations_removed"`
	RunSnapshotFilesRemoved     []string  `json:"run_snapshot_files_removed"`
	RunArtifactDirsRemoved      []string  `json:"run_artifact_dirs_removed"`
	PostgresRunRowsRemoved      int       `json:"postgres_run_rows_removed"`
	PostgresQueueRowsRemoved    int       `json:"postgres_queue_rows_removed"`
	PostgresArtifactRowsRemoved int       `json:"postgres_artifact_rows_removed"`
}

// Service enforces retention windows against the local filesystem and optional
// PostgreSQL mirror tables.
type Service struct {
	settings Settings
	db       DBExecutor
}

// NewService constructs a purge service.
func NewService(settings Settings, db DBExecutor) *Service {
	if settings.Now.IsZero() {
		settings.Now = time.Now().UTC()
	}
	if settings.DefaultRunHistoryTTL <= 0 {
		settings.DefaultRunHistoryTTL = 7 * 24 * time.Hour
	}
	if settings.DefaultRunArtifactTTL <= 0 {
		settings.DefaultRunArtifactTTL = 7 * 24 * time.Hour
	}
	return &Service{settings: settings, db: db}
}

// Purge removes stale asset materializations and completed run history derived
// from manifest retention policies.
func (s *Service) Purge(ctx context.Context, assets []metadata.DataAsset, pipelines []orchestration.Pipeline) (Report, error) {
	report := Report{StartedAt: s.settings.Now}

	assetPolicies, err := parseAssetPolicies(assets)
	if err != nil {
		return report, err
	}
	for _, asset := range assets {
		policy := assetPolicies[asset.ID]
		if policy.Materializations <= 0 {
			continue
		}
		path := metadata.MaterializationPath(s.settings.DataRoot, asset.ID)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return report, fmt.Errorf("stat materialization %s: %w", path, err)
		}
		if info.ModTime().UTC().After(s.settings.Now.Add(-policy.Materializations)) {
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return report, fmt.Errorf("remove materialization %s: %w", path, err)
		}
		report.MaterializationsRemoved = append(report.MaterializationsRemoved, path)
	}

	runPolicies := derivePipelinePolicies(pipelines, assetPolicies, s.settings.DefaultRunHistoryTTL, s.settings.DefaultRunArtifactTTL)
	runIDs, runFiles, artifactDirs, err := s.purgeRunFiles(runPolicies)
	if err != nil {
		return report, err
	}
	report.RunSnapshotFilesRemoved = append(report.RunSnapshotFilesRemoved, runFiles...)
	report.RunArtifactDirsRemoved = append(report.RunArtifactDirsRemoved, artifactDirs...)

	if s.db != nil && len(runIDs) > 0 {
		runRows, queueRows, artifactRows, err := s.purgePostgresRows(ctx, runIDs)
		if err != nil {
			return report, err
		}
		report.PostgresRunRowsRemoved = runRows
		report.PostgresQueueRowsRemoved = queueRows
		report.PostgresArtifactRowsRemoved = artifactRows
	}

	return report, nil
}

type assetPolicy struct {
	Materializations time.Duration
	RunArtifacts     time.Duration
	RunHistory       time.Duration
}

type pipelinePolicy struct {
	RunArtifacts time.Duration
	RunHistory   time.Duration
}

func parseAssetPolicies(assets []metadata.DataAsset) (map[string]assetPolicy, error) {
	out := make(map[string]assetPolicy, len(assets))
	for _, asset := range assets {
		policy := assetPolicy{}
		var err error
		if strings.TrimSpace(asset.Retention.Materializations) != "" {
			policy.Materializations, err = time.ParseDuration(asset.Retention.Materializations)
			if err != nil {
				return nil, fmt.Errorf("parse retention.materializations for %s: %w", asset.ID, err)
			}
		}
		if strings.TrimSpace(asset.Retention.RunArtifacts) != "" {
			policy.RunArtifacts, err = time.ParseDuration(asset.Retention.RunArtifacts)
			if err != nil {
				return nil, fmt.Errorf("parse retention.run_artifacts for %s: %w", asset.ID, err)
			}
		}
		if strings.TrimSpace(asset.Retention.RunHistory) != "" {
			policy.RunHistory, err = time.ParseDuration(asset.Retention.RunHistory)
			if err != nil {
				return nil, fmt.Errorf("parse retention.run_history for %s: %w", asset.ID, err)
			}
		}
		out[asset.ID] = policy
	}
	return out, nil
}

func derivePipelinePolicies(pipelines []orchestration.Pipeline, assetPolicies map[string]assetPolicy, defaultRunHistory, defaultRunArtifacts time.Duration) map[string]pipelinePolicy {
	out := make(map[string]pipelinePolicy, len(pipelines))
	for _, pipeline := range pipelines {
		policy := pipelinePolicy{
			RunArtifacts: defaultRunArtifacts,
			RunHistory:   defaultRunHistory,
		}
		for _, assetID := range pipelineAssetRefs(pipeline) {
			assetPolicy, present := assetPolicies[assetID]
			if !present {
				continue
			}
			if assetPolicy.RunArtifacts > 0 && assetPolicy.RunArtifacts < policy.RunArtifacts {
				policy.RunArtifacts = assetPolicy.RunArtifacts
			}
			if assetPolicy.RunHistory > 0 && assetPolicy.RunHistory < policy.RunHistory {
				policy.RunHistory = assetPolicy.RunHistory
			}
		}
		out[pipeline.ID] = policy
	}
	return out
}

func pipelineAssetRefs(pipeline orchestration.Pipeline) []string {
	seen := map[string]struct{}{}
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		seen[value] = struct{}{}
	}
	for _, job := range pipeline.Jobs {
		if job.Ingest != nil {
			add(assetIDFromTargetPath(job.Ingest.TargetPath))
		}
		for _, output := range job.Outputs {
			add(output)
		}
		for _, metricRef := range job.MetricRefs {
			add(metricRef)
		}
	}
	refs := make([]string, 0, len(seen))
	for ref := range seen {
		refs = append(refs, ref)
	}
	return refs
}

func assetIDFromTargetPath(path string) string {
	base := filepath.Base(strings.TrimSpace(path))
	extension := filepath.Ext(base)
	return strings.TrimSuffix(base, extension)
}

func (s *Service) purgeRunFiles(policies map[string]pipelinePolicy) ([]string, []string, []string, error) {
	runRoot := filepath.Join(s.settings.DataRoot, "control_plane", "runs")
	entries, err := os.ReadDir(runRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, fmt.Errorf("read run store %s: %w", runRoot, err)
	}

	var (
		runIDs       []string
		runFiles     []string
		artifactDirs []string
	)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(runRoot, entry.Name())
		run, err := readRun(path)
		if err != nil {
			return nil, nil, nil, err
		}
		policy, present := policies[run.PipelineID]
		if !present {
			policy = pipelinePolicy{
				RunArtifacts: s.settings.DefaultRunArtifactTTL,
				RunHistory:   s.settings.DefaultRunHistoryTTL,
			}
		}
		referenceTime := run.UpdatedAt.UTC()
		if run.FinishedAt != nil && !run.FinishedAt.IsZero() {
			referenceTime = run.FinishedAt.UTC()
		}
		if !referenceTime.Before(s.settings.Now.Add(-policy.RunHistory)) {
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, nil, nil, fmt.Errorf("remove run snapshot %s: %w", path, err)
		}
		runIDs = append(runIDs, run.ID)
		runFiles = append(runFiles, path)

		artifactDir := filepath.Join(s.settings.ArtifactRoot, "runs", run.ID)
		if err := os.RemoveAll(artifactDir); err != nil {
			return nil, nil, nil, fmt.Errorf("remove run artifact dir %s: %w", artifactDir, err)
		}
		artifactDirs = append(artifactDirs, artifactDir)
	}
	return runIDs, runFiles, artifactDirs, nil
}

func (s *Service) purgePostgresRows(ctx context.Context, runIDs []string) (int, int, int, error) {
	var (
		runRows      int
		queueRows    int
		artifactRows int
	)
	for _, runID := range runIDs {
		if affected, err := execDelete(ctx, s.db, `delete from artifact_snapshots where run_id = $1`, runID); err != nil {
			return 0, 0, 0, fmt.Errorf("delete artifact snapshots for %s: %w", runID, err)
		} else {
			artifactRows += affected
		}
		if affected, err := execDelete(ctx, s.db, `delete from queue_requests where run_id = $1 and status = 'completed'`, runID); err != nil {
			return 0, 0, 0, fmt.Errorf("delete queue request for %s: %w", runID, err)
		} else {
			queueRows += affected
		}
		if affected, err := execDelete(ctx, s.db, `delete from run_snapshots where run_id = $1`, runID); err != nil {
			return 0, 0, 0, fmt.Errorf("delete run snapshot for %s: %w", runID, err)
		} else {
			runRows += affected
		}
		_, _ = execDelete(ctx, s.db, `delete from run_events where pipeline_run_id = $1`, runID)
		_, _ = execDelete(ctx, s.db, `delete from job_runs where pipeline_run_id = $1`, runID)
		_, _ = execDelete(ctx, s.db, `delete from pipeline_runs where id = $1`, runID)
	}
	return runRows, queueRows, artifactRows, nil
}

func execDelete(ctx context.Context, db DBExecutor, query string, args ...any) (int, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

func readRun(path string) (orchestration.PipelineRun, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return orchestration.PipelineRun{}, fmt.Errorf("read run snapshot %s: %w", path, err)
	}
	var run orchestration.PipelineRun
	if err := json.Unmarshal(bytes, &run); err != nil {
		return orchestration.PipelineRun{}, fmt.Errorf("decode run snapshot %s: %w", path, err)
	}
	return run, nil
}
