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

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/python"
	"github.com/streanor/data-platform/backend/internal/storage"
	"github.com/streanor/data-platform/backend/internal/transforms"
)

// Runner executes one queued pipeline run.
type Runner struct {
	cfg    config.Settings
	loader manifests.Loader
	store  orchestration.Store
	files  *storage.Service
	sql    *transforms.Engine
	python *python.Runner
	logger *slog.Logger
}

// NewRunner constructs an execution runner.
func NewRunner(cfg config.Settings, loader manifests.Loader, store orchestration.Store, files *storage.Service, logger *slog.Logger) *Runner {
	return &Runner{
		cfg:    cfg,
		loader: loader,
		store:  store,
		files:  files,
		sql:    transforms.NewEngine(cfg.DuckDBPath, cfg.SQLRoot),
		python: python.NewRunner(cfg),
		logger: logger,
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
	jobRun.Attempts++
	jobRun.StartedAt = &now
	r.appendEvent(run, "info", "job started", map[string]string{"job_id": job.ID, "job_type": string(job.Type)})
	if err := r.store.SavePipelineRun(*run); err != nil {
		return err
	}

	var err error
	switch job.Type {
	case orchestration.JobTypeIngest:
		err = r.runIngest(run.ID, job)
	case orchestration.JobTypeTransformSQL:
		err = r.runMonthlyCashflowTransform(run.ID, job)
	case orchestration.JobTypeTransformPy:
		err = r.runPythonTransform(ctx, run.ID, run.PipelineID, job)
	case orchestration.JobTypeQualityCheck:
		err = r.runQualityCheck(run.ID, job)
	case orchestration.JobTypePublishMetric:
		err = r.runMetricsPublish(run.ID, job)
	default:
		err = fmt.Errorf("unsupported job type %s", job.Type)
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

func (r *Runner) runIngest(runID string, job orchestration.Job) error {
	switch job.ID {
	case "ingest_transactions_csv":
		return r.copySampleFile(runID, "personal_finance/transactions.csv", filepath.Join(r.cfg.DataRoot, "raw", "raw_transactions.csv"), "raw/raw_transactions.csv")
	case "ingest_account_balances_json":
		return r.copySampleFile(runID, "personal_finance/account_balances.json", filepath.Join(r.cfg.DataRoot, "raw", "raw_account_balances.json"), "raw/raw_account_balances.json")
	case "ingest_budget_rules_json":
		return r.copySampleFile(runID, "personal_finance/budget_rules.json", filepath.Join(r.cfg.DataRoot, "raw", "raw_budget_rules.json"), "raw/raw_budget_rules.json")
	default:
		return fmt.Errorf("unknown ingest job %s", job.ID)
	}
}

func (r *Runner) runMonthlyCashflowTransform(runID string, job orchestration.Job) error {
	if err := r.sql.MaterializeRawTables(
		filepath.Join(r.cfg.DataRoot, "raw", "raw_transactions.csv"),
		filepath.Join(r.cfg.DataRoot, "raw", "raw_account_balances.json"),
		filepath.Join(r.cfg.DataRoot, "raw", "raw_budget_rules.json"),
	); err != nil {
		return fmt.Errorf("materialize raw duckdb tables: %w", err)
	}
	stagingPath := filepath.Join(r.cfg.DataRoot, "staging", "staging_transactions_enriched.json")
	if _, err := os.Stat(stagingPath); err == nil {
		if err := r.sql.MaterializeStagingTransactions(stagingPath); err != nil {
			return fmt.Errorf("materialize staging duckdb table: %w", err)
		}
	}
	if err := r.sql.RunTransform(job.TransformRef); err != nil {
		return err
	}
	return r.writeTransformArtifact(runID, job.TransformRef)
}

func (r *Runner) runPythonTransform(ctx context.Context, runID, pipelineID string, job orchestration.Job) error {
	if job.Command == "" {
		return fmt.Errorf("python transform %s is missing a command", job.ID)
	}
	result, err := r.python.Run(ctx, pipelineID, job, python.TaskRequest{
		RunID:          runID,
		PipelineID:     pipelineID,
		JobID:          job.ID,
		Command:        job.Command,
		DataRoot:       r.cfg.DataRoot,
		ArtifactRoot:   r.cfg.ArtifactRoot,
		SampleDataRoot: r.cfg.SampleDataRoot,
		SQLRoot:        r.cfg.SQLRoot,
		Inputs:         job.Inputs,
		Outputs:        job.Outputs,
		Labels:         job.Labels,
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
	metricIDs := []string{"metrics_savings_rate", "metrics_category_variance"}
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

func (r *Runner) writeTransformArtifact(runID, transformRef string) error {
	var (
		query        string
		targetPath   string
		artifactPath string
	)

	switch transformRef {
	case "transform.monthly_cashflow":
		query = `
			select month, income, expenses, savings_rate
			from mart_monthly_cashflow
			order by month
		`
		targetPath = filepath.Join(r.cfg.DataRoot, "mart", "mart_monthly_cashflow.json")
		artifactPath = "mart/mart_monthly_cashflow.json"
	case "transform.category_spend":
		query = `
			select month, category, actual_spend
			from mart_category_spend
			order by month, category
		`
		targetPath = filepath.Join(r.cfg.DataRoot, "mart", "mart_category_spend.json")
		artifactPath = "mart/mart_category_spend.json"
	case "transform.intermediate_category_monthly_rollup":
		query = `
			select month, category, category_group, expense_total, transaction_count
			from intermediate_category_monthly_rollup
			order by month, category
		`
		targetPath = filepath.Join(r.cfg.DataRoot, "intermediate", "intermediate_category_monthly_rollup.json")
		artifactPath = "intermediate/intermediate_category_monthly_rollup.json"
	case "transform.budget_vs_actual":
		query = `
			select month, category, budget_amount, actual_spend, variance_amount
			from mart_budget_vs_actual
			order by month, category
		`
		targetPath = filepath.Join(r.cfg.DataRoot, "mart", "mart_budget_vs_actual.json")
		artifactPath = "mart/mart_budget_vs_actual.json"
	default:
		return fmt.Errorf("unsupported transform reference %s", transformRef)
	}

	rowsOut, err := r.sql.QueryRows(query)
	if err != nil {
		return fmt.Errorf("query transform output for %s: %w", transformRef, err)
	}
	return r.writeJSONArtifact(runID, targetPath, artifactPath, rowsOut)
}

func (r *Runner) writeMetricArtifact(runID, metricID string) error {
	var query string
	switch metricID {
	case "metrics_savings_rate":
		query = `
			select month, savings_rate
			from metrics_savings_rate
			order by month
		`
	case "metrics_category_variance":
		query = `
			select month, category, variance_amount
			from metrics_category_variance
			order by month, category
		`
	default:
		return fmt.Errorf("unsupported metric %s", metricID)
	}

	rows, err := r.sql.QueryRows(query)
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

func latestValue(row map[string]any) any {
	if value, present := row["savings_rate"]; present {
		return value
	}
	if value, present := row["variance_amount"]; present {
		return value
	}
	return nil
}
