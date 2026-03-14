// Package execution runs queued pipeline jobs and materializes local artifacts
// for the first end-to-end finance slice. The execution path is intentionally
// explicit so each step is easy to inspect, test, and debug.
package execution

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// Runner executes one queued pipeline run.
type Runner struct {
	cfg    config.Settings
	loader manifests.Loader
	store  orchestration.Store
	logger *slog.Logger
}

// NewRunner constructs an execution runner.
func NewRunner(cfg config.Settings, loader manifests.Loader, store orchestration.Store, logger *slog.Logger) *Runner {
	return &Runner{
		cfg:    cfg,
		loader: loader,
		store:  store,
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
		err = r.runIngest(job)
	case orchestration.JobTypeTransformSQL:
		err = r.runMonthlyCashflowTransform(job)
	case orchestration.JobTypeQualityCheck:
		err = r.runQualityCheck(job)
	case orchestration.JobTypePublishMetric:
		err = r.runMetricsPublish(job)
	default:
		err = fmt.Errorf("unsupported job type %s", job.Type)
	}

	finished := time.Now().UTC()
	jobRun.EndedAt = &finished
	if err != nil {
		jobRun.Status = orchestration.RunStatusFailed
		jobRun.Error = err.Error()
		r.appendEvent(run, "error", "job failed", map[string]string{"job_id": job.ID, "error": err.Error()})
	} else {
		jobRun.Status = orchestration.RunStatusSucceeded
		r.appendEvent(run, "info", "job succeeded", map[string]string{"job_id": job.ID})
	}
	return r.store.SavePipelineRun(*run)
}

func (r *Runner) runIngest(job orchestration.Job) error {
	switch job.ID {
	case "ingest_transactions_csv":
		return r.copySampleFile("personal_finance/transactions.csv", filepath.Join(r.cfg.DataRoot, "raw", "raw_transactions.csv"))
	case "ingest_account_balances_json":
		return r.copySampleFile("personal_finance/account_balances.json", filepath.Join(r.cfg.DataRoot, "raw", "raw_account_balances.json"))
	default:
		return fmt.Errorf("unknown ingest job %s", job.ID)
	}
}

func (r *Runner) runMonthlyCashflowTransform(job orchestration.Job) error {
	source := filepath.Join(r.cfg.DataRoot, "raw", "raw_transactions.csv")
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open raw transactions: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("read raw transactions: %w", err)
	}

	type summary struct {
		Income      float64 `json:"income"`
		Expenses    float64 `json:"expenses"`
		SavingsRate float64 `json:"savings_rate"`
	}

	summaries := map[string]summary{}
	for index, row := range rows {
		if index == 0 || len(row) < 5 {
			continue
		}
		month := row[1][:7]
		amount, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		if err != nil {
			return fmt.Errorf("parse transaction amount: %w", err)
		}
		current := summaries[month]
		if amount >= 0 {
			current.Income += amount
		} else {
			current.Expenses += amount * -1
		}
		summaries[month] = current
	}

	months := make([]string, 0, len(summaries))
	for month := range summaries {
		months = append(months, month)
	}
	sort.Strings(months)

	rowsOut := make([]map[string]any, 0, len(months))
	for _, month := range months {
		current := summaries[month]
		if current.Income > 0 {
			current.SavingsRate = (current.Income - current.Expenses) / current.Income
		}
		rowsOut = append(rowsOut, map[string]any{
			"month":        month,
			"income":       current.Income,
			"expenses":     current.Expenses,
			"savings_rate": current.SavingsRate,
		})
	}

	return r.writeJSONArtifact(filepath.Join(r.cfg.DataRoot, "mart", "mart_monthly_cashflow.json"), rowsOut)
}

func (r *Runner) runQualityCheck(job orchestration.Job) error {
	source := filepath.Join(r.cfg.DataRoot, "raw", "raw_transactions.csv")
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open raw transactions for quality check: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("read raw transactions for quality check: %w", err)
	}

	uncategorized := 0
	for index, row := range rows {
		if index == 0 || len(row) < 4 {
			continue
		}
		if strings.TrimSpace(row[3]) == "" {
			uncategorized++
		}
	}

	return r.writeJSONArtifact(filepath.Join(r.cfg.DataRoot, "quality", job.ID+".json"), map[string]any{
		"check_id":        job.ID,
		"status":          qualityStatus(uncategorized),
		"uncategorized":   uncategorized,
		"evaluated_at":    time.Now().UTC(),
		"source_artifact": source,
	})
}

func (r *Runner) runMetricsPublish(job orchestration.Job) error {
	source := filepath.Join(r.cfg.DataRoot, "mart", "mart_monthly_cashflow.json")
	bytes, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read mart artifact: %w", err)
	}

	var rows []map[string]any
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return fmt.Errorf("decode mart artifact: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("mart artifact is empty")
	}
	latest := rows[len(rows)-1]
	return r.writeJSONArtifact(filepath.Join(r.cfg.DataRoot, "metrics", "metrics_savings_rate.json"), map[string]any{
		"metric_id":       "metrics_savings_rate",
		"latest_month":    latest["month"],
		"latest_value":    latest["savings_rate"],
		"series":          rows,
		"generated_at":    time.Now().UTC(),
		"source_artifact": source,
	})
}

func (r *Runner) copySampleFile(relativeSource string, target string) error {
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
	return nil
}

func (r *Runner) writeJSONArtifact(path string, payload any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create artifact dir for %s: %w", path, err)
	}
	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode artifact %s: %w", path, err)
	}
	return os.WriteFile(path, bytes, 0o644)
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

func qualityStatus(uncategorized int) string {
	if uncategorized == 0 {
		return "passing"
	}
	return "warning"
}
