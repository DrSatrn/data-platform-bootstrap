// Package admin provides a command-oriented management surface for the web
// portal and remote CLI. The commands are intentionally product-specific and
// read-only in this first iteration.
package admin

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	quality   *quality.Service
	reports   *reporting.MemoryStore
	artifacts *storage.Service
	telemetry *observability.Service
}

// NewService constructs a new admin command service.
func NewService(
	cfg config.Settings,
	loader manifests.Loader,
	store orchestration.Store,
	control *orchestration.ControlService,
	quality *quality.Service,
	reports *reporting.MemoryStore,
	artifacts *storage.Service,
	telemetry *observability.Service,
) *Service {
	return &Service{
		cfg:       cfg,
		loader:    loader,
		store:     store,
		control:   control,
		quality:   quality,
		reports:   reports,
		artifacts: artifacts,
		telemetry: telemetry,
	}
}

// Execute runs a supported platform command and returns textual output suited
// for both terminal UIs and local CLI use.
func (s *Service) Execute(command string) Result {
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
		dashboards := s.reports.ListDashboards()
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
	s.telemetry.RecordCommand(command, result.Success, preview)
	return result
}
