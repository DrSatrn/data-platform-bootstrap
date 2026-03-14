// This file provides orchestration run stores used by the API and worker. The
// file-backed implementation is the default local-first store because it gives
// us durable cross-process run history without introducing extra dependencies
// before the PostgreSQL repositories are wired in.
package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Store exposes the orchestration operations needed by the handlers and the
// scheduler. The interface is intentionally small so the future Postgres-backed
// implementation can replace the in-memory version cleanly.
type Store interface {
	ListPipelineRuns() ([]PipelineRun, error)
	SavePipelineRun(PipelineRun) error
	GetPipelineRun(id string) (PipelineRun, bool, error)
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
func (s *InMemoryStore) ListPipelineRuns() ([]PipelineRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]PipelineRun, len(s.runs))
	copy(out, s.runs)
	return out, nil
}

// SavePipelineRun records a run snapshot for diagnostics and handler output.
func (s *InMemoryStore) SavePipelineRun(run PipelineRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	run.UpdatedAt = time.Now().UTC()
	for index, existing := range s.runs {
		if existing.ID == run.ID {
			s.runs[index] = run
			return nil
		}
	}
	s.runs = append(s.runs, run)
	return nil
}

// GetPipelineRun returns one run by identifier.
func (s *InMemoryStore) GetPipelineRun(id string) (PipelineRun, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, run := range s.runs {
		if run.ID == id {
			return run, true, nil
		}
	}
	return PipelineRun{}, false, nil
}

// FileStore persists runs as JSON files under the local data directory. This
// keeps the local platform restart-safe while staying easy to inspect.
type FileStore struct {
	mu   sync.Mutex
	root string
}

// NewFileStore constructs a file-backed run store.
func NewFileStore(dataRoot string) (*FileStore, error) {
	root := filepath.Join(dataRoot, "control_plane", "runs")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create run store root: %w", err)
	}
	return &FileStore{root: root}, nil
}

// ListPipelineRuns loads all persisted run snapshots sorted by most recent
// update time first so operators see the freshest activity at the top.
func (s *FileStore) ListPipelineRuns() ([]PipelineRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, fmt.Errorf("read run store: %w", err)
	}

	runs := make([]PipelineRun, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		run, err := s.readRun(filepath.Join(s.root, entry.Name()))
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	sort.Slice(runs, func(left, right int) bool {
		return runs[left].UpdatedAt.After(runs[right].UpdatedAt)
	})
	return runs, nil
}

// SavePipelineRun persists a single run snapshot atomically.
func (s *FileStore) SavePipelineRun(run PipelineRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	run.UpdatedAt = time.Now().UTC()

	target := filepath.Join(s.root, run.ID+".json")
	temp := target + ".tmp"
	bytes, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return fmt.Errorf("encode run %s: %w", run.ID, err)
	}
	if err := os.WriteFile(temp, bytes, 0o644); err != nil {
		return fmt.Errorf("write run temp file %s: %w", run.ID, err)
	}
	if err := os.Rename(temp, target); err != nil {
		return fmt.Errorf("rename run file %s: %w", run.ID, err)
	}
	return nil
}

// GetPipelineRun loads one persisted run by identifier.
func (s *FileStore) GetPipelineRun(id string) (PipelineRun, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.root, id+".json")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return PipelineRun{}, false, nil
		}
		return PipelineRun{}, false, fmt.Errorf("stat run %s: %w", id, err)
	}
	run, err := s.readRun(path)
	if err != nil {
		return PipelineRun{}, false, err
	}
	return run, true, nil
}

func (s *FileStore) readRun(path string) (PipelineRun, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return PipelineRun{}, fmt.Errorf("read run file %s: %w", path, err)
	}

	var run PipelineRun
	if err := json.Unmarshal(bytes, &run); err != nil {
		return PipelineRun{}, fmt.Errorf("decode run file %s: %w", path, err)
	}
	return run, nil
}
