// This file implements a small file-backed run queue for the local worker. The
// queue is restart-safe and easy to inspect, which makes it a good fit for the
// platform's initial localhost execution path.
package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RunRequest describes one queued pipeline run request.
type RunRequest struct {
	RunID       string    `json:"run_id"`
	PipelineID  string    `json:"pipeline_id"`
	Trigger     string    `json:"trigger"`
	RequestedAt time.Time `json:"requested_at"`
}

// RunQueue defines the durable queue behavior shared by the API, scheduler,
// and worker. Keeping the interface narrow makes it straightforward to support
// both local filesystem and PostgreSQL-backed control-plane queues.
type RunQueue interface {
	Enqueue(RunRequest) error
	ClaimNext() (*ClaimedRequest, error)
	Complete(*ClaimedRequest) error
}

// Queue owns pending and in-flight run request files.
type Queue struct {
	root      string
	queuedDir string
	activeDir string
}

// ClaimedRequest represents a request file claimed by a worker.
type ClaimedRequest struct {
	Request RunRequest
	Receipt string
}

// NewQueue constructs a durable local queue.
func NewQueue(dataRoot string) (*Queue, error) {
	root := filepath.Join(dataRoot, "control_plane", "queue")
	queuedDir := filepath.Join(root, "queued")
	activeDir := filepath.Join(root, "active")
	for _, dir := range []string{queuedDir, activeDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create queue dir %s: %w", dir, err)
		}
	}
	return &Queue{
		root:      root,
		queuedDir: queuedDir,
		activeDir: activeDir,
	}, nil
}

// Enqueue persists a run request for worker pickup.
func (q *Queue) Enqueue(request RunRequest) error {
	path := filepath.Join(q.queuedDir, request.RunID+".json")
	bytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return fmt.Errorf("encode run request: %w", err)
	}
	return os.WriteFile(path, bytes, 0o644)
}

// ClaimNext returns the next active or queued request for processing.
func (q *Queue) ClaimNext() (*ClaimedRequest, error) {
	if claimed, err := q.claimFromDir(q.activeDir, false); err != nil || claimed != nil {
		return claimed, err
	}
	return q.claimFromDir(q.queuedDir, true)
}

// Complete removes a successfully or unsuccessfully processed request file.
func (q *Queue) Complete(claimed *ClaimedRequest) error {
	if claimed == nil {
		return nil
	}
	if err := os.Remove(claimed.Receipt); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove claimed request %s: %w", claimed.Receipt, err)
	}
	return nil
}

// ListRequests returns a point-in-time snapshot of queued and active requests.
// The file-backed queue does not retain completed requests, so this export is
// intentionally scoped to currently pending work.
func (q *Queue) ListRequests() ([]QueueSnapshot, error) {
	requests := []QueueSnapshot{}
	for _, source := range []struct {
		dir    string
		status string
	}{
		{dir: q.activeDir, status: "active"},
		{dir: q.queuedDir, status: "queued"},
	} {
		entries, err := os.ReadDir(source.dir)
		if err != nil {
			return nil, fmt.Errorf("read queue dir %s: %w", source.dir, err)
		}
		sort.Slice(entries, func(left, right int) bool {
			return entries[left].Name() < entries[right].Name()
		})
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			bytes, err := os.ReadFile(filepath.Join(source.dir, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("read queue request %s: %w", entry.Name(), err)
			}
			var request RunRequest
			if err := json.Unmarshal(bytes, &request); err != nil {
				return nil, fmt.Errorf("decode queue request %s: %w", entry.Name(), err)
			}
			snapshot := QueueSnapshot{
				RunID:       request.RunID,
				PipelineID:  request.PipelineID,
				Trigger:     request.Trigger,
				Status:      source.status,
				RequestedAt: request.RequestedAt,
			}
			requests = append(requests, snapshot)
		}
	}
	return requests, nil
}

func (q *Queue) claimFromDir(dir string, moveToActive bool) (*ClaimedRequest, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read queue dir %s: %w", dir, err)
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].Name() < entries[right].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		source := filepath.Join(dir, entry.Name())
		target := source
		if moveToActive {
			target = filepath.Join(q.activeDir, entry.Name())
			if err := os.Rename(source, target); err != nil {
				continue
			}
		}

		bytes, err := os.ReadFile(target)
		if err != nil {
			return nil, fmt.Errorf("read claimed request %s: %w", target, err)
		}
		var request RunRequest
		if err := json.Unmarshal(bytes, &request); err != nil {
			return nil, fmt.Errorf("decode claimed request %s: %w", target, err)
		}
		return &ClaimedRequest{Request: request, Receipt: target}, nil
	}
	return nil, nil
}
