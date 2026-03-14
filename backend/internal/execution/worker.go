// This file provides the worker loop that polls the local queue and executes
// queued pipeline runs. The loop is intentionally straightforward because local
// reliability matters more than premature queue abstraction.
package execution

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// Worker polls for queued runs and executes them.
type Worker struct {
	queue  *orchestration.Queue
	runner *Runner
	logger *slog.Logger
	poll   time.Duration
}

// NewWorker constructs a local queue worker.
func NewWorker(queue *orchestration.Queue, runner *Runner, logger *slog.Logger, poll time.Duration) *Worker {
	return &Worker{
		queue:  queue,
		runner: runner,
		logger: logger,
		poll:   poll,
	}
}

// Run executes the worker polling loop until shutdown.
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.poll)
	defer ticker.Stop()

	for {
		if err := w.processOne(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				w.logger.Error("worker loop iteration failed", slog.String("error", err.Error()))
			}
		}

		select {
		case <-ctx.Done():
			w.logger.Info("worker shutdown complete")
			return nil
		case <-ticker.C:
		}
	}
}

func (w *Worker) processOne(ctx context.Context) error {
	claimed, err := w.queue.ClaimNext()
	if err != nil {
		return err
	}
	if claimed == nil {
		return nil
	}

	w.logger.Info("worker claimed pipeline run", slog.String("run_id", claimed.Request.RunID), slog.String("pipeline_id", claimed.Request.PipelineID))
	defer func() {
		if err := w.queue.Complete(claimed); err != nil {
			w.logger.Error("failed to complete queue item", slog.String("error", err.Error()))
		}
	}()
	return w.runner.Execute(ctx, claimed.Request)
}
