// This file provides an in-memory orchestration store so the scaffold has a
// working control-plane surface before PostgreSQL persistence is wired in.
package orchestration

import (
	"sync"
	"time"
)

// Store exposes the orchestration operations needed by the handlers and the
// scheduler. The interface is intentionally small so the future Postgres-backed
// implementation can replace the in-memory version cleanly.
type Store interface {
	ListPipelineRuns() []PipelineRun
	RecordPipelineRun(PipelineRun)
}

// InMemoryStore is a concurrency-safe placeholder implementation.
type InMemoryStore struct {
	mu   sync.RWMutex
	runs []PipelineRun
}

// NewInMemoryStore creates an empty store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{runs: []PipelineRun{}}
}

// ListPipelineRuns returns a snapshot of known runs.
func (s *InMemoryStore) ListPipelineRuns() []PipelineRun {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]PipelineRun, len(s.runs))
	copy(out, s.runs)
	return out
}

// RecordPipelineRun appends a run snapshot for diagnostics and handler output.
func (s *InMemoryStore) RecordPipelineRun(run PipelineRun) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	run.UpdatedAt = time.Now().UTC()
	s.runs = append(s.runs, run)
}
