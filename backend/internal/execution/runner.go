// Package execution runs queued pipeline jobs and materializes local artifacts
// for the first end-to-end finance slice. The execution path is intentionally
// explicit so each step is easy to inspect, test, and debug.
package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/streanor/data-platform/backend/internal/alerting"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/externaltools"
	"github.com/streanor/data-platform/backend/internal/ingestion"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/python"
	"github.com/streanor/data-platform/backend/internal/storage"
	"github.com/streanor/data-platform/backend/internal/transforms"
)

// Runner executes one queued pipeline run.
type Runner struct {
	cfg                    config.Settings
	loader                 manifests.Loader
	store                  orchestration.Store
	files                  *storage.Service
	ingest                 *ingestion.Exporter
	sql                    *transforms.Engine
	python                 *python.Runner
	tools                  *externaltools.Runner
	alerts                 *alerting.Dispatcher
	logger                 *slog.Logger
	repo                   string
	sleep                  func(context.Context, time.Duration) error
	executeAttemptOverride func(context.Context, *orchestration.PipelineRun, orchestration.Job, string) error
}

// NewRunner constructs an execution runner.
func NewRunner(cfg config.Settings, loader manifests.Loader, store orchestration.Store, files *storage.Service, logger *slog.Logger) *Runner {
	return &Runner{
		cfg:    cfg,
		loader: loader,
		store:  store,
		files:  files,
		ingest: ingestion.NewExporter(),
		sql:    transforms.NewEngine(cfg.DuckDBPath, cfg.SQLRoot),
		python: python.NewRunner(cfg),
		tools: externaltools.NewRunner(externaltools.Settings{
			Root:           cfg.ExternalToolRoot,
			DBTBinary:      cfg.DBTBinary,
			DLTBinary:      cfg.DLTBinary,
			PySparkBinary:  cfg.PySparkBinary,
			DefaultTimeout: cfg.ExternalToolTimeout,
		}),
		alerts: alerting.NewDispatcher(alerting.Settings{
			Environment:    cfg.Environment,
			RunFailureURLs: cfg.RunFailureWebhookURLs,
			WebhookTimeout: cfg.AlertWebhookTimeout,
		}, nil),
		logger: logger,
		repo:   externalToolRepoRoot(cfg),
		sleep:  sleepWithContext,
	}
}

// Execute processes all jobs in dependency order and materializes finance
// outputs to the local data root.
func (r *Runner) Execute(ctx context.Context, request orchestration.RunRequest) error {
	run, found, err := r.store.GetPipelineRun(request.RunID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("run %s not found", request.RunID)
	}

	pipelines, err := r.loader.LoadPipelines()
	if err != nil {
		return err
	}

	var pipeline *orchestration.Pipeline
	for index := range pipelines {
		if pipelines[index].ID == request.PipelineID {
			pipeline = &pipelines[index]
			break
		}
	}
	if pipeline == nil {
		return fmt.Errorf("pipeline %s not found", request.PipelineID)
	}

	run.Status = orchestration.RunStatusRunning
	r.appendEvent(&run, "info", "worker started pipeline run", map[string]string{"pipeline_id": pipeline.ID})
	if err := r.store.SavePipelineRun(run); err != nil {
		return err
	}

	completed := map[string]bool{}
	for len(completed) < len(pipeline.Jobs) {
		progressed := false
		for _, job := range pipeline.Jobs {
			if completed[job.ID] || !depsSatisfied(job, completed) {
				continue
			}
			progressed = true
			if err := r.executeJob(ctx, &run, job); err != nil {
				run.Status = orchestration.RunStatusFailed
				run.Error = err.Error()
				now := time.Now().UTC()
				run.FinishedAt = &now
				r.appendEvent(&run, "error", "pipeline run failed", map[string]string{"job_id": job.ID, "error": err.Error()})
				_ = r.store.SavePipelineRun(run)
				r.notifyRunFailure(ctx, *pipeline, run, job.ID)
				return err
			}
			completed[job.ID] = true
		}
		if !progressed {
			return fmt.Errorf("pipeline %s could not make progress due to unresolved dependencies", pipeline.ID)
		}
	}

	run.Status = orchestration.RunStatusSucceeded
	now := time.Now().UTC()
	run.FinishedAt = &now
	r.appendEvent(&run, "info", "pipeline run succeeded", map[string]string{"pipeline_id": pipeline.ID})
	return r.store.SavePipelineRun(run)
}

func (r *Runner) executeJob(ctx context.Context, run *orchestration.PipelineRun, job orchestration.Job) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	jobRun := findJobRun(run, job.ID)
	now := time.Now().UTC()
	jobRun.Status = orchestration.RunStatusRunning
	jobRun.StartedAt = &now
	r.appendEvent(run, "info", "job started", map[string]string{"job_id": job.ID, "job_type": string(job.Type)})
	if err := r.store.SavePipelineRun(*run); err != nil {
		return err
	}

	var err error
	totalAttempts := job.Retries + 1
	if totalAttempts < 1 {
		totalAttempts = 1
	}
	for attempt := 1; attempt <= totalAttempts; attempt++ {
		jobRun.Attempts = attempt
		idempotencyKey := jobIdempotencyKey(run.ID, job.ID, attempt)
		r.appendEvent(run, "info", "job attempt started", map[string]string{
			"job_id":          job.ID,
			"attempt":         strconv.Itoa(attempt),
			"idempotency_key": idempotencyKey,
		})
		err = r.executeJobAttempt(ctx, run, job, idempotencyKey)
		if err == nil {
			break
		}
		if attempt >= totalAttempts {
			break
		}
		backoff := retryBackoff(r.cfg.JobRetryBaseDelay, attempt)
		r.appendEvent(run, "warn", "job attempt failed; scheduling retry", map[string]string{
			"job_id":          job.ID,
			"attempt":         strconv.Itoa(attempt),
			"retry_in":        backoff.String(),
			"idempotency_key": idempotencyKey,
			"error":           err.Error(),
		})
		if saveErr := r.store.SavePipelineRun(*run); saveErr != nil {
			return saveErr
		}
		if sleepErr := r.sleep(ctx, backoff); sleepErr != nil {
			return sleepErr
		}
	}

	finished := time.Now().UTC()
	jobRun.EndedAt = &finished
	if err != nil {
		jobRun.Status = orchestration.RunStatusFailed
		jobRun.Error = err.Error()
		r.appendEvent(run, "error", "job failed", map[string]string{"job_id": job.ID, "error": err.Error()})
		if saveErr := r.store.SavePipelineRun(*run); saveErr != nil {
			return saveErr
		}
		return err
	}

	jobRun.Status = orchestration.RunStatusSucceeded
	r.appendEvent(run, "info", "job succeeded", map[string]string{"job_id": job.ID})
	return r.store.SavePipelineRun(*run)
}

func (r *Runner) executeJobAttempt(ctx context.Context, run *orchestration.PipelineRun, job orchestration.Job, idempotencyKey string) error {
	if r.executeAttemptOverride != nil {
		return r.executeAttemptOverride(ctx, run, job, idempotencyKey)
	}
	switch job.Type {
	case orchestration.JobTypeIngest:
		return r.runIngest(ctx, run.ID, job)
	case orchestration.JobTypeTransformSQL:
		return r.runSQLTransform(run.ID, job)
	case orchestration.JobTypeTransformPy:
		return r.runPythonTransform(ctx, run.ID, run.PipelineID, job, idempotencyKey)
	case orchestration.JobTypeQualityCheck:
		return r.runQualityCheck(run.ID, job)
	case orchestration.JobTypePublishMetric:
		return r.runMetricsPublish(run.ID, job)
	case orchestration.JobTypeExternalTool:
		return r.runExternalTool(ctx, run, job, idempotencyKey)
	default:
		return fmt.Errorf("unsupported job type %s", job.Type)
	}
}

func (r *Runner) runIngest(ctx context.Context, runID string, job orchestration.Job) error {
	if job.Ingest == nil {
		return fmt.Errorf("ingest job %s is missing an ingest block", job.ID)
	}
	switch normalizeIngestKind(job.Ingest.SourceKind) {
	case "postgres", "mysql":
		targetPath := filepath.Join(r.cfg.DataRoot, filepath.FromSlash(job.Ingest.TargetPath))
		if err := r.ingest.ExportQueryToCSV(ctx, ingestion.DatabaseSpec{
			Driver:        job.Ingest.SourceKind,
			ConnectionEnv: job.Ingest.ConnectionEnv,
			Query:         job.Ingest.Query,
			TargetPath:    targetPath,
		}); err != nil {
			return err
		}
		bytes, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("read database ingest target %s: %w", targetPath, err)
		}
		return r.writeRunScopedArtifact(runID, artifactPathOrDefault(job.Ingest.ArtifactPath, job.Ingest.TargetPath), bytes)
	default:
		return r.copySampleFile(
			runID,
			job.Ingest.SourceRef,
			filepath.Join(r.cfg.DataRoot, filepath.FromSlash(job.Ingest.TargetPath)),
			artifactPathOrDefault(job.Ingest.ArtifactPath, job.Ingest.TargetPath),
		)
	}
}

func (r *Runner) runSQLTransform(runID string, job orchestration.Job) error {
	for _, bootstrap := range job.Bootstrap {
		sourcePath := filepath.Join(r.cfg.DataRoot, filepath.FromSlash(bootstrap.SourcePath))
		if _, err := os.Stat(sourcePath); err != nil {
			if os.IsNotExist(err) && !bootstrap.Required {
				continue
			}
			return fmt.Errorf("bootstrap source %s: %w", bootstrap.SourcePath, err)
		}
		if err := r.sql.ExecFile(sqlPathFromRef(bootstrap.SQLRef), map[string]string{
			bootstrap.Placeholder: quotedSQLString(sourcePath),
		}); err != nil {
			return fmt.Errorf("bootstrap %s: %w", bootstrap.SQLRef, err)
		}
	}
	if err := r.sql.RunTransform(job.TransformRef); err != nil {
		return err
	}
	return r.writeTransformArtifacts(runID, job.Outputs)
}

func (r *Runner) runPythonTransform(ctx context.Context, runID, pipelineID string, job orchestration.Job, idempotencyKey string) error {
	if job.Command == "" {
		return fmt.Errorf("python transform %s is missing a command", job.ID)
	}
	labels := make(map[string]string, len(job.Labels)+1)
	for key, value := range job.Labels {
		labels[key] = value
	}
	labels["platform_idempotency_key"] = idempotencyKey
	result, err := r.python.Run(ctx, pipelineID, job, python.TaskRequest{
		RunID:          runID,
		PipelineID:     pipelineID,
		JobID:          job.ID,
		IdempotencyKey: idempotencyKey,
		Command:        job.Command,
		DataRoot:       r.cfg.DataRoot,
		ArtifactRoot:   r.cfg.ArtifactRoot,
		SampleDataRoot: r.cfg.SampleDataRoot,
		SQLRoot:        r.cfg.SQLRoot,
		Inputs:         job.Inputs,
		Outputs:        job.Outputs,
		Labels:         labels,
	})
	if err != nil {
		return err
	}
	for _, line := range result.LogLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		r.logger.Info("python task", slog.String("job_id", job.ID), slog.String("message", line))
	}
	for _, output := range result.Outputs {
		if err := r.mirrorPythonOutput(runID, output); err != nil {
			return err
		}
	}
	if result.Message != "" {
		r.logger.Info("python task completed", slog.String("job_id", job.ID), slog.String("summary", result.Message))
	}
	return nil
}

func (r *Runner) runQualityCheck(runID string, job orchestration.Job) error {
	queryPath := filepath.Join("quality", job.ID+".sql")
	rows, err := r.sql.QueryRowsFromFile(queryPath, nil)
	if err != nil {
		return fmt.Errorf("run quality sql for %s: %w", job.ID, err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("quality query %s returned no rows", job.ID)
	}

	key, count := qualityResult(job.ID, rows[0])
	return r.writeJSONArtifact(runID, filepath.Join(r.cfg.DataRoot, "quality", job.ID+".json"), filepath.ToSlash(filepath.Join("quality", job.ID+".json")), map[string]any{
		"check_id":        job.ID,
		"status":          qualityStatus(count),
		key:               count,
		"evaluated_at":    time.Now().UTC(),
		"source_artifact": "duckdb:" + queryPath,
	})
}

func (r *Runner) runMetricsPublish(runID string, job orchestration.Job) error {
	metricIDs := metricRefsForJob(job)
	for _, metricID := range metricIDs {
		if err := r.sql.RunMetric(metricID); err != nil {
			return err
		}
		if err := r.writeMetricArtifact(runID, metricID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) copySampleFile(runID, relativeSource, target, runArtifactPath string) error {
	source := filepath.Join(r.cfg.SampleDataRoot, relativeSource)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create target dir for %s: %w", target, err)
	}

	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open sample source %s: %w", source, err)
	}
	defer input.Close()

	output, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create target %s: %w", target, err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("copy %s to %s: %w", source, target, err)
	}
	bytes, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("read copied target %s for run artifact mirror: %w", target, err)
	}
	return r.writeRunScopedArtifact(runID, runArtifactPath, bytes)
}

func (r *Runner) writeJSONArtifact(runID, path, runArtifactPath string, payload any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create artifact dir for %s: %w", path, err)
	}
	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode artifact %s: %w", path, err)
	}
	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return err
	}
	return r.writeRunScopedArtifact(runID, runArtifactPath, bytes)
}

func (r *Runner) mirrorPythonOutput(runID string, output python.TaskOutput) error {
	bytes, err := os.ReadFile(output.SourcePath)
	if err != nil {
		return fmt.Errorf("read python output %s: %w", output.SourcePath, err)
	}
	return r.writeRunScopedArtifact(runID, output.RelativePath, bytes)
}

func (r *Runner) writeRunScopedArtifact(runID, relativePath string, bytes []byte) error {
	path := filepath.Join(r.cfg.ArtifactRoot, "runs", runID, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create run artifact dir for %s: %w", path, err)
	}
	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return err
	}
	if r.files == nil {
		return nil
	}
	return r.files.RecordRunArtifact(runID, relativePath)
}

func (r *Runner) appendEvent(run *orchestration.PipelineRun, level, message string, fields map[string]string) {
	run.Events = append(run.Events, orchestration.RunEvent{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: message,
		Fields:  fields,
	})
	r.logger.Info(message, slog.String("level", level), slog.Any("fields", fields))
}

func (r *Runner) notifyRunFailure(ctx context.Context, pipeline orchestration.Pipeline, run orchestration.PipelineRun, jobID string) {
	if r.alerts == nil {
		return
	}
	if err := r.alerts.NotifyRunFailure(ctx, alerting.RunFailureEvent{
		RunID:      run.ID,
		PipelineID: run.PipelineID,
		Pipeline:   pipeline.Name,
		Trigger:    run.Trigger,
		JobID:      jobID,
		Error:      run.Error,
		FailedAt:   derefTime(run.FinishedAt),
	}); err != nil {
		r.logger.Warn("failed to post run failure webhook", slog.String("run_id", run.ID), slog.String("pipeline_id", run.PipelineID), slog.String("error", err.Error()))
	}
}

func findJobRun(run *orchestration.PipelineRun, jobID string) *orchestration.JobRun {
	for index := range run.JobRuns {
		if run.JobRuns[index].JobID == jobID {
			return &run.JobRuns[index]
		}
	}
	run.JobRuns = append(run.JobRuns, orchestration.JobRun{ID: jobID, JobID: jobID, Status: orchestration.RunStatusPending})
	return &run.JobRuns[len(run.JobRuns)-1]
}

func depsSatisfied(job orchestration.Job, completed map[string]bool) bool {
	for _, dependency := range job.DependsOn {
		if !completed[dependency] {
			return false
		}
	}
	return true
}

func qualityResult(jobID string, row map[string]any) (string, int) {
	key := "count"
	switch jobID {
	case "check_uncategorized_transactions":
		key = "uncategorized_count"
	case "check_duplicate_transactions":
		key = "duplicate_count"
	}
	return key, intFromRow(row[key])
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.UTC()
}

func intFromRow(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}
	return 0
}

func qualityStatus(count int) string {
	if count == 0 {
		return "passing"
	}
	return "warning"
}

func (r *Runner) writeTransformArtifacts(runID string, outputs []string) error {
	for _, output := range outputs {
		if !isSQLIdentifier(output) {
			return fmt.Errorf("unsupported transform output identifier %q", output)
		}
		targetPath, artifactPath, err := dataArtifactLocation(output)
		if err != nil {
			return err
		}
		rowsOut, err := r.sql.QueryRows(fmt.Sprintf("select * from %s", output))
		if err != nil {
			return fmt.Errorf("query transform output for %s: %w", output, err)
		}
		if err := r.writeJSONArtifact(runID, filepath.Join(r.cfg.DataRoot, filepath.FromSlash(targetPath)), artifactPath, rowsOut); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) writeMetricArtifact(runID, metricID string) error {
	if !isSQLIdentifier(metricID) {
		return fmt.Errorf("unsupported metric %s", metricID)
	}

	rows, err := r.sql.QueryRows(fmt.Sprintf("select * from %s", metricID))
	if err != nil {
		return fmt.Errorf("query metric rows for %s: %w", metricID, err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("%s is empty", metricID)
	}
	latest := rows[len(rows)-1]
	return r.writeJSONArtifact(runID, filepath.Join(r.cfg.DataRoot, "metrics", metricID+".json"), filepath.ToSlash(filepath.Join("metrics", metricID+".json")), map[string]any{
		"metric_id":       metricID,
		"latest_month":    latest["month"],
		"latest_value":    latestValue(latest),
		"series":          rows,
		"generated_at":    time.Now().UTC(),
		"source_artifact": "duckdb:" + metricID,
	})
}

func (r *Runner) runExternalTool(ctx context.Context, run *orchestration.PipelineRun, job orchestration.Job, idempotencyKey string) error {
	if job.ExternalTool == nil {
		return fmt.Errorf("external tool job %s is missing configuration", job.ID)
	}

	spec := *job.ExternalTool
	timeout, err := parseJobTimeout(job.Timeout)
	if err != nil {
		return err
	}

	result, err := r.tools.Run(ctx, externaltools.RunRequest{
		RunID:          run.ID,
		RepoRoot:       r.repo,
		PipelineID:     run.PipelineID,
		JobID:          job.ID,
		IdempotencyKey: idempotencyKey,
		Timeout:        timeout,
		Spec:           spec,
	})
	for _, event := range result.Events {
		r.appendEvent(run, event.Level, event.Message, mergeEventFields(event.Fields, map[string]string{"job_id": job.ID}))
	}
	for _, line := range result.LogLines {
		r.logger.Info("external tool", slog.String("job_id", job.ID), slog.String("tool", firstNonEmpty(result.Tool, spec.Tool)), slog.String("message", line))
	}
	if err := r.writeExternalToolLogArtifacts(run.ID, job.ID, result); err != nil {
		return err
	}
	if err != nil {
		r.appendEvent(run, "error", "external tool failed", map[string]string{
			"job_id":          job.ID,
			"tool":            firstNonEmpty(result.Tool, spec.Tool),
			"action":          firstNonEmpty(result.Action, spec.Action),
			"failure_class":   result.FailureClass,
			"idempotency_key": idempotencyKey,
			"error":           err.Error(),
		})
		return fmt.Errorf("external tool %s %s failed: %w", spec.Tool, spec.Action, err)
	}
	for _, artifact := range result.Artifacts {
		bytes, readErr := os.ReadFile(artifact.SourcePath)
		if readErr != nil {
			return fmt.Errorf("read external tool artifact %s: %w", artifact.SourcePath, readErr)
		}
		if writeErr := r.writeRunScopedArtifact(run.ID, artifact.RelativePath, bytes); writeErr != nil {
			return writeErr
		}
	}
	r.appendEvent(run, "info", "external tool finished", map[string]string{
		"job_id":          job.ID,
		"tool":            firstNonEmpty(result.Tool, spec.Tool),
		"action":          firstNonEmpty(result.Action, spec.Action),
		"artifact_count":  strconv.Itoa(len(result.Artifacts)),
		"idempotency_key": idempotencyKey,
	})
	return nil
}

func latestValue(row map[string]any) any {
	if value, present := row["savings_rate"]; present {
		return value
	}
	if value, present := row["variance_amount"]; present {
		return value
	}
	if value, present := row["net_quantity"]; present {
		return value
	}
	for key, value := range row {
		switch key {
		case "month", "category", "warehouse", "sku":
			continue
		default:
			return value
		}
	}
	return nil
}

func artifactPathOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return filepath.ToSlash(filepath.Clean(value))
	}
	return filepath.ToSlash(filepath.Clean(fallback))
}

func metricRefsForJob(job orchestration.Job) []string {
	if len(job.MetricRefs) > 0 {
		return append([]string{}, job.MetricRefs...)
	}
	if len(job.Outputs) > 0 {
		metricRefs := make([]string, 0, len(job.Outputs))
		for _, output := range job.Outputs {
			if strings.HasPrefix(output, "metrics_") {
				metricRefs = append(metricRefs, output)
			}
		}
		if len(metricRefs) > 0 {
			return metricRefs
		}
	}
	return []string{"metrics_savings_rate", "metrics_category_variance"}
}

func sqlPathFromRef(ref string) string {
	name := strings.TrimPrefix(strings.TrimSpace(ref), "bootstrap.")
	return filepath.Join("bootstrap", name+".sql")
}

func dataArtifactLocation(output string) (string, string, error) {
	switch {
	case strings.HasPrefix(output, "raw_"):
		return filepath.ToSlash(filepath.Join("raw", output+".json")), filepath.ToSlash(filepath.Join("raw", output+".json")), nil
	case strings.HasPrefix(output, "staging_"):
		return filepath.ToSlash(filepath.Join("staging", output+".json")), filepath.ToSlash(filepath.Join("staging", output+".json")), nil
	case strings.HasPrefix(output, "intermediate_"):
		return filepath.ToSlash(filepath.Join("intermediate", output+".json")), filepath.ToSlash(filepath.Join("intermediate", output+".json")), nil
	case strings.HasPrefix(output, "mart_"):
		return filepath.ToSlash(filepath.Join("mart", output+".json")), filepath.ToSlash(filepath.Join("mart", output+".json")), nil
	default:
		return "", "", fmt.Errorf("unsupported transform output %q", output)
	}
}

func isSQLIdentifier(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	for index, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9' && index > 0) || (r == '_' && index > 0) {
			continue
		}
		if index == 0 && r >= 'a' && r <= 'z' {
			continue
		}
		return false
	}
	return true
}

func quotedSQLString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func repoRootFromManifestRoot(manifestRoot string) string {
	return filepath.Clean(filepath.Join(manifestRoot, "..", ".."))
}

func externalToolRepoRoot(cfg config.Settings) string {
	if strings.TrimSpace(cfg.ExternalToolRoot) != "" {
		return filepath.Clean(cfg.ExternalToolRoot)
	}
	return repoRootFromManifestRoot(cfg.ManifestRoot)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func parseJobTimeout(value string) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse job timeout %q: %w", value, err)
	}
	return timeout, nil
}

func retryBackoff(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		base = 250 * time.Millisecond
	}
	if attempt < 1 {
		attempt = 1
	}
	delay := base
	for index := 1; index < attempt; index++ {
		if delay >= 8*time.Second {
			return 8 * time.Second
		}
		delay *= 2
	}
	if delay > 8*time.Second {
		return 8 * time.Second
	}
	return delay
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func jobIdempotencyKey(runID, jobID string, attempt int) string {
	return fmt.Sprintf("%s:%s:%d", runID, jobID, attempt)
}

func normalizeIngestKind(value string) string {
	kind := strings.ToLower(strings.TrimSpace(value))
	if kind == "" {
		return "file"
	}
	return kind
}

func mergeEventFields(primary, defaults map[string]string) map[string]string {
	if len(primary) == 0 && len(defaults) == 0 {
		return nil
	}
	merged := make(map[string]string, len(primary)+len(defaults))
	for key, value := range defaults {
		merged[key] = value
	}
	for key, value := range primary {
		merged[key] = value
	}
	return merged
}

func (r *Runner) writeExternalToolLogArtifacts(runID, jobID string, result externaltools.Result) error {
	logRoot := filepath.ToSlash(filepath.Join("external_tools", jobID, "logs"))
	if strings.TrimSpace(result.Stdout) != "" {
		if err := r.writeRunScopedArtifact(runID, filepath.ToSlash(filepath.Join(logRoot, "stdout.log")), []byte(result.Stdout)); err != nil {
			return err
		}
	}
	if strings.TrimSpace(result.Stderr) != "" {
		if err := r.writeRunScopedArtifact(runID, filepath.ToSlash(filepath.Join(logRoot, "stderr.log")), []byte(result.Stderr)); err != nil {
			return err
		}
	}
	return nil
}
