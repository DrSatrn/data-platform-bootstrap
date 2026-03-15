// Package backup restores first-party recovery bundles back into a stopped
// runtime. The implementation is deliberately explicit so operators can see
// which filesystem roots and PostgreSQL tables are being reconstituted.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/authz"
	"github.com/streanor/data-platform/backend/internal/db"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/reporting"
)

// PostgresRestoreMode controls how aggressively restore should require the
// PostgreSQL control-plane path to be available.
type PostgresRestoreMode string

const (
	PostgresRestoreAuto     PostgresRestoreMode = "auto"
	PostgresRestoreRequired PostgresRestoreMode = "required"
	PostgresRestoreSkip     PostgresRestoreMode = "skip"
)

// RestoreOptions describes the operator-selected restore targets. Defaults
// come from the current runtime config so the CLI can stay compact.
type RestoreOptions struct {
	BundlePath         string
	Confirm            bool
	TargetDataRoot     string
	TargetArtifactRoot string
	TargetDuckDBPath   string
	ExtractRoot        string
	PostgresMode       PostgresRestoreMode
	PostgresDSN        string
	MigrationsRoot     string
}

// RestoreResult captures the concrete restore actions so docs and scripts can
// report what was actually rebuilt.
type RestoreResult struct {
	BundlePath            string         `json:"bundle_path"`
	Manifest              BackupManifest `json:"manifest"`
	DataRoot              string         `json:"data_root"`
	ArtifactRoot          string         `json:"artifact_root"`
	DuckDBPath            string         `json:"duckdb_path"`
	ExtractRoot           string         `json:"extract_root"`
	PostgresRestored      bool           `json:"postgres_restored"`
	QueueRequestsRequeued int            `json:"queue_requests_requeued"`
	Warnings              []string       `json:"warnings,omitempty"`
}

// Restore verifies a backup bundle, replaces the local runtime filesystem
// roots, and optionally rehydrates the PostgreSQL control plane.
func (s *Service) Restore(options RestoreOptions) (RestoreResult, error) {
	if !options.Confirm {
		return RestoreResult{}, fmt.Errorf("restore requires explicit confirmation; pass Confirm=true or --yes")
	}
	if strings.TrimSpace(options.BundlePath) == "" {
		return RestoreResult{}, fmt.Errorf("bundle path is required")
	}

	manifest, err := s.Verify(options.BundlePath)
	if err != nil {
		return RestoreResult{}, err
	}

	options = s.normalizeRestoreOptions(options)

	extractRoot := options.ExtractRoot
	if extractRoot == "" {
		extractRoot, err = os.MkdirTemp("", "data-platform-restore-*")
		if err != nil {
			return RestoreResult{}, fmt.Errorf("create restore extraction dir: %w", err)
		}
	}
	if err := extractBundle(options.BundlePath, extractRoot); err != nil {
		return RestoreResult{}, err
	}

	if err := restoreDirectory(filepath.Join(extractRoot, "files", "data"), options.TargetDataRoot); err != nil {
		return RestoreResult{}, err
	}
	if err := restoreDirectory(filepath.Join(extractRoot, "files", "artifacts"), options.TargetArtifactRoot); err != nil {
		return RestoreResult{}, err
	}
	if err := restoreFile(filepath.Join(extractRoot, "files", "duckdb", "platform.duckdb"), options.TargetDuckDBPath); err != nil {
		return RestoreResult{}, err
	}

	result := RestoreResult{
		BundlePath:   options.BundlePath,
		Manifest:     manifest,
		DataRoot:     options.TargetDataRoot,
		ArtifactRoot: options.TargetArtifactRoot,
		DuckDBPath:   options.TargetDuckDBPath,
		ExtractRoot:  extractRoot,
		Warnings:     append([]string{}, manifest.Warnings...),
	}

	switch options.PostgresMode {
	case "", PostgresRestoreAuto, PostgresRestoreRequired:
		postgresResult, err := restorePostgresControlPlane(options, extractRoot)
		if err != nil {
			if options.PostgresMode == PostgresRestoreAuto && errors.Is(err, errPostgresUnavailable) {
				result.Warnings = append(result.Warnings, err.Error())
				return result, nil
			}
			if options.PostgresMode == PostgresRestoreRequired && errors.Is(err, errPostgresUnavailable) {
				return RestoreResult{}, fmt.Errorf("postgres restore was required but the control-plane database was unavailable: %w", err)
			}
			return RestoreResult{}, err
		}
		result.PostgresRestored = postgresResult.restored
		result.QueueRequestsRequeued = postgresResult.requeued
		result.Warnings = append(result.Warnings, postgresResult.warnings...)
	case PostgresRestoreSkip:
		result.Warnings = append(result.Warnings, "postgres restore skipped by operator request; filesystem restore completed")
	default:
		return RestoreResult{}, fmt.Errorf("unsupported postgres restore mode %q", options.PostgresMode)
	}

	return result, nil
}

func (s *Service) normalizeRestoreOptions(options RestoreOptions) RestoreOptions {
	if options.TargetDataRoot == "" {
		options.TargetDataRoot = s.cfg.DataRoot
	}
	if options.TargetArtifactRoot == "" {
		options.TargetArtifactRoot = s.cfg.ArtifactRoot
	}
	if options.TargetDuckDBPath == "" {
		options.TargetDuckDBPath = s.cfg.DuckDBPath
	}
	if options.PostgresMode == "" {
		options.PostgresMode = PostgresRestoreAuto
	}
	if options.PostgresDSN == "" {
		options.PostgresDSN = s.cfg.PostgresDSN
	}
	if options.MigrationsRoot == "" {
		options.MigrationsRoot = s.cfg.MigrationsRoot
	}
	return options
}

func extractBundle(bundlePath, targetRoot string) error {
	file, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("open backup bundle: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open backup gzip: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read backup archive: %w", err)
		}
		targetPath := filepath.Join(targetRoot, filepath.FromSlash(header.Name))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create restore dir %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("create restore parent %s: %w", targetPath, err)
			}
			file, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create restore file %s: %w", targetPath, err)
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return fmt.Errorf("write restore file %s: %w", targetPath, err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("close restore file %s: %w", targetPath, err)
			}
			if err := os.Chtimes(targetPath, header.ModTime, header.ModTime); err != nil {
				return fmt.Errorf("restore modtime for %s: %w", targetPath, err)
			}
		}
	}
}

func restoreDirectory(sourceRoot, targetRoot string) error {
	if err := os.RemoveAll(targetRoot); err != nil {
		return fmt.Errorf("clear restore target dir %s: %w", targetRoot, err)
	}
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return fmt.Errorf("create restore target dir %s: %w", targetRoot, err)
	}

	info, err := os.Stat(sourceRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat restore source dir %s: %w", sourceRoot, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("restore source %s is not a directory", sourceRoot)
	}

	return filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk restore source %s: %w", path, walkErr)
		}
		relative, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return fmt.Errorf("derive restore relative path for %s: %w", path, err)
		}
		if relative == "." {
			return nil
		}
		targetPath := filepath.Join(targetRoot, relative)
		if entry.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create restore dir %s: %w", targetPath, err)
			}
			return nil
		}
		return copyFile(path, targetPath)
	})
}

func restoreFile(sourcePath, targetPath string) error {
	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("clear restore target file %s: %w", targetPath, err)
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat restore source file %s: %w", sourcePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("restore source %s is a directory, expected a file", sourcePath)
	}
	return copyFile(sourcePath, targetPath)
}

func copyFile(sourcePath, targetPath string) error {
	bytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read restore source %s: %w", sourcePath, err)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create restore target parent %s: %w", targetPath, err)
	}
	if err := os.WriteFile(targetPath, bytes, 0o644); err != nil {
		return fmt.Errorf("write restore target %s: %w", targetPath, err)
	}
	info, err := os.Stat(sourcePath)
	if err == nil {
		if err := os.Chtimes(targetPath, info.ModTime(), info.ModTime()); err != nil {
			return fmt.Errorf("restore modtime for %s: %w", targetPath, err)
		}
	}
	return nil
}

var errPostgresUnavailable = errors.New("postgres restore skipped because the control-plane database was unavailable")

type postgresRestoreResult struct {
	restored bool
	requeued int
	warnings []string
}

func restorePostgresControlPlane(options RestoreOptions, extractRoot string) (postgresRestoreResult, error) {
	conn, err := db.Open(options.PostgresDSN)
	if err != nil {
		return postgresRestoreResult{}, fmt.Errorf("%w: %v", errPostgresUnavailable, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		return postgresRestoreResult{}, fmt.Errorf("%w: %v", errPostgresUnavailable, err)
	}
	if err := db.ApplyMigrations(ctx, conn, options.MigrationsRoot); err != nil {
		return postgresRestoreResult{}, fmt.Errorf("apply migrations before restore: %w", err)
	}

	runs, err := readRestoreJSON[[]orchestration.PipelineRun](extractRoot, "exports/pipeline_runs.json")
	if err != nil {
		return postgresRestoreResult{}, err
	}
	queueRequests, err := readRestoreJSON[[]orchestration.QueueSnapshot](extractRoot, "exports/queue_requests.json")
	if err != nil {
		return postgresRestoreResult{}, err
	}
	dashboards, err := readRestoreJSON[[]reporting.Dashboard](extractRoot, "exports/dashboards.json")
	if err != nil {
		return postgresRestoreResult{}, err
	}
	auditEvents, err := readRestoreJSON[[]audit.Event](extractRoot, "exports/audit_events.json")
	if err != nil {
		return postgresRestoreResult{}, err
	}
	assets, err := readRestoreJSON[[]metadata.DataAsset](extractRoot, "exports/data_assets.json")
	if err != nil {
		return postgresRestoreResult{}, err
	}
	users, err := readRestoreJSON[[]authz.StoredUser](extractRoot, "exports/platform_users.json")
	if err != nil {
		return postgresRestoreResult{}, err
	}

	requeued, warnings, err := restorePostgresState(ctx, conn, runs, queueRequests, dashboards, auditEvents, assets, users)
	if err != nil {
		return postgresRestoreResult{}, err
	}
	return postgresRestoreResult{
		restored: true,
		requeued: requeued,
		warnings: warnings,
	}, nil
}

func readRestoreJSON[T any](extractRoot, relativePath string) (T, error) {
	var payload T
	path := filepath.Join(extractRoot, filepath.FromSlash(relativePath))
	bytes, err := os.ReadFile(path)
	if err != nil {
		return payload, fmt.Errorf("read restore export %s: %w", relativePath, err)
	}
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return payload, fmt.Errorf("decode restore export %s: %w", relativePath, err)
	}
	return payload, nil
}

func restorePostgresState(
	ctx context.Context,
	conn *sql.DB,
	runs []orchestration.PipelineRun,
	queueRequests []orchestration.QueueSnapshot,
	dashboards []reporting.Dashboard,
	auditEvents []audit.Event,
	assets []metadata.DataAsset,
	users []authz.StoredUser,
) (int, []string, error) {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("begin postgres restore transaction: %w", err)
	}
	defer tx.Rollback()

	for _, statement := range []string{
		`delete from platform_sessions`,
		`delete from platform_users where is_bootstrap = false`,
		`delete from artifact_snapshots`,
		`delete from queue_requests`,
		`delete from run_snapshots`,
		`delete from dashboards`,
		`delete from audit_events`,
		`delete from asset_columns`,
		`delete from data_assets`,
	} {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return 0, nil, fmt.Errorf("reset postgres restore state: %w", err)
		}
	}

	for _, run := range runs {
		payload, err := json.Marshal(run)
		if err != nil {
			return 0, nil, fmt.Errorf("encode restored run %s: %w", run.ID, err)
		}
		if run.UpdatedAt.IsZero() {
			run.UpdatedAt = time.Now().UTC()
		}
		if _, err := tx.ExecContext(ctx, `
			insert into run_snapshots (run_id, pipeline_id, status, trigger_source, payload, updated_at)
			values ($1, $2, $3, $4, $5::jsonb, $6)
		`, run.ID, run.PipelineID, string(run.Status), run.Trigger, string(payload), run.UpdatedAt); err != nil {
			return 0, nil, fmt.Errorf("restore run snapshot %s: %w", run.ID, err)
		}
	}

	requeued := 0
	warnings := []string{}
	for _, request := range queueRequests {
		status := strings.TrimSpace(request.Status)
		claimedAt := request.ClaimedAt
		completedAt := request.CompletedAt
		if status == "" {
			status = "queued"
		}
		if status == "active" {
			// Backup exports do not carry the worker claim token, so a restored
			// in-flight row must be released back to the queue instead of being
			// presented as resumable active work.
			status = "queued"
			claimedAt = nil
			completedAt = nil
			requeued++
		}
		if _, err := tx.ExecContext(ctx, `
			insert into queue_requests (
				run_id, pipeline_id, trigger_source, requested_at, status, claim_token, claimed_at, completed_at
			)
			values ($1, $2, $3, $4, $5, '', $6, $7)
		`, request.RunID, request.PipelineID, request.Trigger, request.RequestedAt, status, claimedAt, completedAt); err != nil {
			return 0, nil, fmt.Errorf("restore queue request %s: %w", request.RunID, err)
		}
	}
	if requeued > 0 {
		warnings = append(warnings, fmt.Sprintf("%d queue requests were restored as queued because active claim tokens are not exported in backup bundles", requeued))
	}

	for _, dashboard := range dashboards {
		definition, err := marshalDashboardDefinition(dashboard)
		if err != nil {
			return 0, nil, err
		}
		if _, err := tx.ExecContext(ctx, `
			insert into dashboards (id, name, description, definition)
			values ($1, $2, $3, $4::jsonb)
		`, dashboard.ID, dashboard.Name, dashboard.Description, string(definition)); err != nil {
			return 0, nil, fmt.Errorf("restore dashboard %s: %w", dashboard.ID, err)
		}
	}

	for _, event := range auditEvents {
		details, err := json.Marshal(event.Details)
		if err != nil {
			return 0, nil, fmt.Errorf("encode audit event %s: %w", event.Action, err)
		}
		if event.Time.IsZero() {
			event.Time = time.Now().UTC()
		}
		if _, err := tx.ExecContext(ctx, `
			insert into audit_events (event_time, actor_user_id, actor_subject, actor_role, action, resource, outcome, details)
			values ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
		`, event.Time, event.ActorUserID, event.ActorSubject, event.ActorRole, event.Action, event.Resource, event.Outcome, string(details)); err != nil {
			return 0, nil, fmt.Errorf("restore audit event %s: %w", event.Action, err)
		}
	}

	for _, asset := range assets {
		sourceRefs, err := json.Marshal(asset.SourceRefs)
		if err != nil {
			return 0, nil, fmt.Errorf("encode source refs for restored asset %s: %w", asset.ID, err)
		}
		qualityRefs, err := json.Marshal(asset.QualityCheckRefs)
		if err != nil {
			return 0, nil, fmt.Errorf("encode quality refs for restored asset %s: %w", asset.ID, err)
		}
		docRefs, err := json.Marshal(asset.DocumentationRefs)
		if err != nil {
			return 0, nil, fmt.Errorf("encode documentation refs for restored asset %s: %w", asset.ID, err)
		}
		if _, err := tx.ExecContext(ctx, `
			insert into data_assets (
				id, name, layer, owner_id, kind, description,
				source_refs, freshness_expected_within, freshness_warn_after,
				quality_check_refs, documentation_refs,
				annotation_owner_id, annotation_description,
				annotation_quality_check_refs, annotation_documentation_refs,
				annotation_updated_at, manifest_synced_at
			)
			values (
				$1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10::jsonb, $11::jsonb,
				$12, $13, $14::jsonb, $15::jsonb, $16, $17
			)
		`,
			asset.ID,
			asset.Name,
			asset.Layer,
			asset.Owner,
			asset.Kind,
			asset.Description,
			string(sourceRefs),
			asset.Freshness.ExpectedWithin,
			asset.Freshness.WarnAfter,
			string(qualityRefs),
			string(docRefs),
			asset.Owner,
			asset.Description,
			string(qualityRefs),
			string(docRefs),
			time.Now().UTC(),
			time.Now().UTC(),
		); err != nil {
			return 0, nil, fmt.Errorf("restore metadata asset %s: %w", asset.ID, err)
		}
		for _, column := range asset.Columns {
			if _, err := tx.ExecContext(ctx, `
				insert into asset_columns (
					asset_id, column_name, column_type, description, is_pii,
					annotation_description, annotation_updated_at
				)
				values ($1, $2, $3, $4, $5, $6, $7)
			`, asset.ID, column.Name, column.Type, column.Description, column.IsPII, column.Description, time.Now().UTC()); err != nil {
				return 0, nil, fmt.Errorf("restore metadata column %s.%s: %w", asset.ID, column.Name, err)
			}
		}
	}

	for _, user := range users {
		if _, err := tx.ExecContext(ctx, `
			insert into platform_users (
				id, username, display_name, role, password_hash, password_salt,
				is_active, is_bootstrap, created_at, updated_at
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			on conflict (id) do update set
				username = excluded.username,
				display_name = excluded.display_name,
				role = excluded.role,
				password_hash = excluded.password_hash,
				password_salt = excluded.password_salt,
				is_active = excluded.is_active,
				is_bootstrap = excluded.is_bootstrap,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at
		`,
			user.ID,
			user.Username,
			user.DisplayName,
			string(user.Role),
			user.PasswordHash,
			user.PasswordSalt,
			user.IsActive,
			user.IsBootstrap,
			user.CreatedAt,
			user.UpdatedAt,
		); err != nil {
			return 0, nil, fmt.Errorf("restore platform user %s: %w", user.Username, err)
		}
	}
	if len(users) > 0 {
		warnings = append(warnings, "native users were restored, but active login sessions were intentionally cleared so operators must sign in again")
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, fmt.Errorf("commit postgres restore transaction: %w", err)
	}
	return requeued, warnings, nil
}

func marshalDashboardDefinition(dashboard reporting.Dashboard) ([]byte, error) {
	definition, err := json.Marshal(map[string]any{
		"owner":           dashboard.Owner,
		"tags":            dashboard.Tags,
		"shared_role":     dashboard.SharedRole,
		"default_filters": dashboard.DefaultFilters,
		"presets":         dashboard.Presets,
		"widgets":         dashboard.Widgets,
		"updated_at":      dashboard.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("encode restored dashboard %s: %w", dashboard.ID, err)
	}
	return definition, nil
}
