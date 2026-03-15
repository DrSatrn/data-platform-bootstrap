// Package admin provides a command-oriented management surface for the web
// portal and remote CLI. The commands are intentionally product-specific so the
// platform can expose high-value operational workflows without generic shell
// access.
package admin

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/streanor/data-platform/backend/internal/authz"
	"github.com/streanor/data-platform/backend/internal/backup"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/observability"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/quality"
	"github.com/streanor/data-platform/backend/internal/reporting"
	"github.com/streanor/data-platform/backend/internal/storage"
)

// Result is the structured response returned by admin commands.
type Result struct {
	Command string   `json:"command"`
	Success bool     `json:"success"`
	Output  []string `json:"output"`
}

// Service owns admin-terminal command execution.
type Service struct {
	cfg       config.Settings
	loader    manifests.Loader
	store     orchestration.Store
	control   *orchestration.ControlService
	identity  *authz.Service
	quality   *quality.Service
	reports   reporting.Store
	artifacts *storage.Service
	telemetry *observability.Service
	backup    *backup.Service
}

// NewService constructs a new admin command service.
func NewService(
	cfg config.Settings,
	loader manifests.Loader,
	store orchestration.Store,
	control *orchestration.ControlService,
	identity *authz.Service,
	quality *quality.Service,
	reports reporting.Store,
	artifacts *storage.Service,
	telemetry *observability.Service,
	backupService *backup.Service,
) *Service {
	return &Service{
		cfg:       cfg,
		loader:    loader,
		store:     store,
		control:   control,
		identity:  identity,
		quality:   quality,
		reports:   reports,
		artifacts: artifacts,
		telemetry: telemetry,
		backup:    backupService,
	}
}

// Execute runs a supported platform command and returns textual output suited
// for both terminal UIs and local CLI use.
func (s *Service) Execute(command string, actor authz.Principal) Result {
	command = strings.TrimSpace(command)
	if command == "" {
		return Result{Command: command, Success: false, Output: []string{"command cannot be empty"}}
	}

	tokens := strings.Fields(command)
	head := tokens[0]
	args := tokens[1:]

	var result Result
	switch head {
	case "help":
		result = Result{Command: command, Success: true, Output: []string{
			"help",
			"status",
			"pipelines",
			"assets",
			"quality",
			"runs",
			"dashboards",
			"metrics",
			"logs [limit]",
			"trigger <pipeline_id>",
			"artifacts <run_id>",
			"users",
			"user create <username> <role> <password> [display_name]",
			"user password <username> <password>",
			"user activate <username>",
			"user deactivate <username>",
			"backups",
			"backup create",
			"backup verify <bundle-name-or-path>",
		}}
	case "status":
		snapshot := s.telemetry.Snapshot(map[string]string{"environment": s.cfg.Environment})
		result = Result{Command: command, Success: true, Output: []string{
			fmt.Sprintf("environment: %s", s.cfg.Environment),
			fmt.Sprintf("api_base_url: %s", s.cfg.APIBaseURL),
			fmt.Sprintf("uptime_seconds: %d", snapshot.UptimeSeconds),
			fmt.Sprintf("total_requests: %d", snapshot.TotalRequests),
			fmt.Sprintf("total_errors: %d", snapshot.TotalErrors),
			fmt.Sprintf("total_commands: %d", snapshot.TotalCommands),
		}}
	case "pipelines":
		pipelines, err := s.loader.LoadPipelines()
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		lines := make([]string, 0, len(pipelines))
		for _, pipeline := range pipelines {
			lines = append(lines, fmt.Sprintf("%s | jobs=%d | owner=%s", pipeline.ID, len(pipeline.Jobs), pipeline.Owner))
		}
		sort.Strings(lines)
		result = Result{Command: command, Success: true, Output: lines}
	case "assets":
		assets, err := s.loader.LoadAssets()
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		lines := make([]string, 0, len(assets))
		for _, asset := range assets {
			lines = append(lines, fmt.Sprintf("%s | layer=%s | owner=%s", asset.ID, asset.Layer, asset.Owner))
		}
		sort.Strings(lines)
		result = Result{Command: command, Success: true, Output: lines}
	case "quality":
		statuses, err := s.quality.ListStatuses()
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		lines := make([]string, 0, len(statuses))
		for _, check := range statuses {
			lines = append(lines, fmt.Sprintf("%s | %s | %s", check.ID, check.Status, check.Message))
		}
		result = Result{Command: command, Success: true, Output: lines}
	case "runs":
		runs, err := s.store.ListPipelineRuns()
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		lines := make([]string, 0, len(runs))
		for _, run := range runs {
			lines = append(lines, fmt.Sprintf("%s | pipeline=%s | status=%s", run.ID, run.PipelineID, run.Status))
		}
		if len(lines) == 0 {
			lines = []string{"no pipeline runs recorded yet"}
		}
		result = Result{Command: command, Success: true, Output: lines}
	case "dashboards":
		dashboards, err := s.reports.ListDashboards()
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		lines := make([]string, 0, len(dashboards))
		for _, dashboard := range dashboards {
			lines = append(lines, fmt.Sprintf("%s | widgets=%d", dashboard.Name, len(dashboard.Widgets)))
		}
		result = Result{Command: command, Success: true, Output: lines}
	case "trigger":
		if len(args) == 0 {
			result = Result{Command: command, Success: false, Output: []string{"usage: trigger <pipeline_id>"}}
			break
		}
		run, err := s.control.TriggerPipeline(context.Background(), args[0], "admin_terminal")
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		result = Result{Command: command, Success: true, Output: []string{
			fmt.Sprintf("queued run %s for pipeline %s", run.ID, run.PipelineID),
		}}
	case "metrics":
		snapshot := s.telemetry.Snapshot(map[string]string{"environment": s.cfg.Environment})
		lines := []string{
			fmt.Sprintf("total_requests=%d", snapshot.TotalRequests),
			fmt.Sprintf("total_errors=%d", snapshot.TotalErrors),
		}
		keys := make([]string, 0, len(snapshot.RequestCounts))
		for path := range snapshot.RequestCounts {
			keys = append(keys, path)
		}
		sort.Strings(keys)
		for _, path := range keys {
			lines = append(lines, fmt.Sprintf("%s=%d", path, snapshot.RequestCounts[path]))
		}
		result = Result{Command: command, Success: true, Output: lines}
	case "logs":
		limit := 10
		if len(args) > 0 {
			if parsed, err := strconv.Atoi(args[0]); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		logs := s.telemetry.RecentLogs()
		if len(logs) == 0 {
			result = Result{Command: command, Success: true, Output: []string{"no logs recorded yet"}}
			break
		}
		if limit > len(logs) {
			limit = len(logs)
		}
		lines := make([]string, 0, limit)
		for _, entry := range logs[len(logs)-limit:] {
			lines = append(lines, fmt.Sprintf("%s | %s | %s", entry.Time.Format("15:04:05"), entry.Level, entry.Message))
		}
		result = Result{Command: command, Success: true, Output: lines}
	case "artifacts":
		if len(args) == 0 {
			result = Result{Command: command, Success: false, Output: []string{"usage: artifacts <run_id>"}}
			break
		}
		artifacts, err := s.artifacts.ListRunArtifacts(args[0])
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		if len(artifacts) == 0 {
			result = Result{Command: command, Success: true, Output: []string{"no artifacts found for run"}}
			break
		}
		lines := make([]string, 0, len(artifacts))
		for _, artifact := range artifacts {
			lines = append(lines, fmt.Sprintf("%s | %d bytes", artifact.RelativePath, artifact.SizeBytes))
		}
		result = Result{Command: command, Success: true, Output: lines}
	case "backups":
		if s.backup == nil {
			result = Result{Command: command, Success: false, Output: []string{"backup service is unavailable"}}
			break
		}
		bundles, err := s.backup.ListBundles()
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		if len(bundles) == 0 {
			result = Result{Command: command, Success: true, Output: []string{"no backup bundles recorded yet"}}
			break
		}
		lines := make([]string, 0, len(bundles))
		for _, bundle := range bundles {
			lines = append(lines, fmt.Sprintf("%s | %d bytes", bundle.Path, bundle.SizeBytes))
		}
		result = Result{Command: command, Success: true, Output: lines}
	case "backup":
		if s.backup == nil {
			result = Result{Command: command, Success: false, Output: []string{"backup service is unavailable"}}
			break
		}
		if len(args) == 0 {
			result = Result{Command: command, Success: false, Output: []string{"usage: backup create | backup verify <bundle-name-or-path>"}}
			break
		}
		switch args[0] {
		case "create":
			created, err := s.backup.Create("")
			if err != nil {
				result = Result{Command: command, Success: false, Output: []string{err.Error()}}
				break
			}
			result = Result{Command: command, Success: true, Output: []string{
				fmt.Sprintf("bundle_path: %s", created.Path),
				fmt.Sprintf("pipeline_runs: %d", created.Manifest.Counts.PipelineRuns),
				fmt.Sprintf("dashboards: %d", created.Manifest.Counts.Dashboards),
				fmt.Sprintf("data_assets: %d", created.Manifest.Counts.DataAssets),
				fmt.Sprintf("bundle_files: %d", created.Manifest.Counts.BundleFiles),
			}}
		case "verify":
			if len(args) < 2 {
				result = Result{Command: command, Success: false, Output: []string{"usage: backup verify <bundle-name-or-path>"}}
				break
			}
			path, err := s.resolveBackupPath(args[1])
			if err != nil {
				result = Result{Command: command, Success: false, Output: []string{err.Error()}}
				break
			}
			manifest, err := s.backup.Verify(path)
			if err != nil {
				result = Result{Command: command, Success: false, Output: []string{err.Error()}}
				break
			}
			result = Result{Command: command, Success: true, Output: []string{
				fmt.Sprintf("bundle_path: %s", path),
				fmt.Sprintf("generated_at: %s", manifest.GeneratedAt.Format(time.RFC3339)),
				fmt.Sprintf("pipeline_runs: %d", manifest.Counts.PipelineRuns),
				fmt.Sprintf("queue_requests: %d", manifest.Counts.QueueRequests),
				fmt.Sprintf("bundle_files: %d", manifest.Counts.BundleFiles),
			}}
		default:
			result = Result{Command: command, Success: false, Output: []string{"usage: backup create | backup verify <bundle-name-or-path>"}}
		}
	case "users":
		if s.identity == nil {
			result = Result{Command: command, Success: false, Output: []string{"native identity store is unavailable"}}
			break
		}
		users, err := s.identity.ListUsers()
		if err != nil {
			result = Result{Command: command, Success: false, Output: []string{err.Error()}}
			break
		}
		if len(users) == 0 {
			result = Result{Command: command, Success: true, Output: []string{"no platform users recorded yet"}}
			break
		}
		lines := make([]string, 0, len(users))
		for _, user := range users {
			lines = append(lines, fmt.Sprintf("%s | role=%s | active=%t | bootstrap=%t", user.Username, user.Role, user.IsActive, user.IsBootstrap))
		}
		result = Result{Command: command, Success: true, Output: lines}
	case "user":
		if s.identity == nil {
			result = Result{Command: command, Success: false, Output: []string{"native identity store is unavailable"}}
			break
		}
		if len(args) == 0 {
			result = Result{Command: command, Success: false, Output: []string{"usage: user create|password|activate|deactivate ..."}}
			break
		}
		switch args[0] {
		case "create":
			if len(args) < 4 {
				result = Result{Command: command, Success: false, Output: []string{"usage: user create <username> <role> <password> [display_name]"}}
				break
			}
			displayName := args[1]
			if len(args) > 4 {
				displayName = strings.Join(args[4:], " ")
			}
			user, err := s.identity.CreateUser(args[1], displayName, authz.Role(args[2]), args[3])
			if err != nil {
				result = Result{Command: command, Success: false, Output: []string{err.Error()}}
				break
			}
			result = Result{Command: command, Success: true, Output: []string{fmt.Sprintf("created user %s with role %s", user.Username, user.Role)}}
		case "password":
			if len(args) < 3 {
				result = Result{Command: command, Success: false, Output: []string{"usage: user password <username> <password>"}}
				break
			}
			user, err := s.identity.ResetPassword(args[1], args[2])
			if err != nil {
				result = Result{Command: command, Success: false, Output: []string{err.Error()}}
				break
			}
			result = Result{Command: command, Success: true, Output: []string{fmt.Sprintf("reset password for %s", user.Username)}}
		case "activate":
			if len(args) < 2 {
				result = Result{Command: command, Success: false, Output: []string{"usage: user activate <username>"}}
				break
			}
			user, err := s.identity.SetUserActive(args[1], true)
			if err != nil {
				result = Result{Command: command, Success: false, Output: []string{err.Error()}}
				break
			}
			result = Result{Command: command, Success: true, Output: []string{fmt.Sprintf("activated %s", user.Username)}}
		case "deactivate":
			if len(args) < 2 {
				result = Result{Command: command, Success: false, Output: []string{"usage: user deactivate <username>"}}
				break
			}
			user, err := s.identity.SetUserActive(args[1], false)
			if err != nil {
				result = Result{Command: command, Success: false, Output: []string{err.Error()}}
				break
			}
			result = Result{Command: command, Success: true, Output: []string{fmt.Sprintf("deactivated %s", user.Username)}}
		default:
			result = Result{Command: command, Success: false, Output: []string{"usage: user create|password|activate|deactivate ..."}}
		}
	default:
		result = Result{Command: command, Success: false, Output: []string{
			fmt.Sprintf("unknown command %q", head),
			"try: help",
		}}
	}

	preview := ""
	if len(result.Output) > 0 {
		preview = result.Output[0]
	}
	s.telemetry.RecordCommand(fmt.Sprintf("%s (%s)", command, actor.Subject), result.Success, preview)
	return result
}

func (s *Service) resolveBackupPath(value string) (string, error) {
	backupRoot := filepath.Clean(filepath.Join(s.cfg.DataRoot, "backups"))
	if value == "" {
		return "", fmt.Errorf("backup bundle path is required")
	}
	if filepath.IsAbs(value) {
		clean := filepath.Clean(value)
		if !strings.HasPrefix(clean, backupRoot+string(filepath.Separator)) && clean != backupRoot {
			return "", fmt.Errorf("backup verification is limited to the configured backup directory")
		}
		return clean, nil
	}
	clean := filepath.Clean(value)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("invalid backup bundle path")
	}
	return filepath.Join(backupRoot, clean), nil
}
