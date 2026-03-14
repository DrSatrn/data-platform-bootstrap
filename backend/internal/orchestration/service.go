// This file provides the orchestration control service used by the API and
// admin terminal. It creates queued runs with durable state so the worker can
// execute them asynchronously.
package orchestration

import (
	"context"
	"fmt"
	"time"
)

// ControlService creates and tracks queued pipeline runs.
type ControlService struct {
	loader PipelineLoader
	store  Store
	queue  *Queue
}

// NewControlService constructs a control-plane service.
func NewControlService(loader PipelineLoader, store Store, queue *Queue) *ControlService {
	return &ControlService{
		loader: loader,
		store:  store,
		queue:  queue,
	}
}

// TriggerPipeline validates the target pipeline, persists an initial queued
// run, and enqueues it for worker execution.
func (s *ControlService) TriggerPipeline(ctx context.Context, pipelineID, trigger string) (PipelineRun, error) {
	pipelines, err := s.loader.LoadPipelines()
	if err != nil {
		return PipelineRun{}, err
	}

	var target *Pipeline
	for index := range pipelines {
		if pipelines[index].ID == pipelineID {
			target = &pipelines[index]
			break
		}
	}
	if target == nil {
		return PipelineRun{}, fmt.Errorf("unknown pipeline %q", pipelineID)
	}
	if err := ValidatePipeline(*target); err != nil {
		return PipelineRun{}, fmt.Errorf("pipeline %q is invalid: %w", pipelineID, err)
	}

	now := time.Now().UTC()
	run := PipelineRun{
		ID:         newRunID(now),
		PipelineID: pipelineID,
		Status:     RunStatusQueued,
		Trigger:    trigger,
		StartedAt:  now,
		UpdatedAt:  now,
		JobRuns:    buildInitialJobRuns(target.Jobs),
		Events: []RunEvent{
			{
				Time:    now,
				Level:   "info",
				Message: "pipeline run queued",
				Fields: map[string]string{
					"pipeline_id": pipelineID,
					"trigger":     trigger,
				},
			},
		},
	}
	if err := s.store.SavePipelineRun(run); err != nil {
		return PipelineRun{}, err
	}
	if err := s.queue.Enqueue(RunRequest{
		RunID:       run.ID,
		PipelineID:  pipelineID,
		Trigger:     trigger,
		RequestedAt: now,
	}); err != nil {
		return PipelineRun{}, err
	}
	return run, nil
}

func buildInitialJobRuns(jobs []Job) []JobRun {
	out := make([]JobRun, 0, len(jobs))
	for _, job := range jobs {
		out = append(out, JobRun{
			ID:       job.ID,
			JobID:    job.ID,
			Status:   RunStatusPending,
			Attempts: 0,
		})
	}
	return out
}

func newRunID(now time.Time) string {
	return "run_" + now.Format("20060102T150405.000000000")
}
