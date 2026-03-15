// Package audit provides a durable append-only audit trail for privileged
// control-plane actions. The implementation favors explicit, inspectable
// storage so self-hosted operators can reason about who changed what and when.
package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Event describes one auditable platform action.
type Event struct {
	Time         time.Time      `json:"time"`
	ActorUserID  string         `json:"actor_user_id,omitempty"`
	ActorSubject string         `json:"actor_subject"`
	ActorRole    string         `json:"actor_role"`
	Action       string         `json:"action"`
	Resource     string         `json:"resource"`
	Outcome      string         `json:"outcome"`
	Details      map[string]any `json:"details,omitempty"`
}

// Store defines the audit persistence behavior used by handlers and services.
type Store interface {
	Append(Event) error
	ListRecent(limit int) ([]Event, error)
}

// MemoryStore provides a last-resort in-memory fallback.
type MemoryStore struct {
	mu     sync.RWMutex
	events []Event
}

// NewMemoryStore constructs an empty in-memory audit store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{events: []Event{}}
}

func (s *MemoryStore) Append(event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	s.events = append(s.events, event)
	return nil
}

func (s *MemoryStore) ListRecent(limit int) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return trimRecent(cloneEvents(s.events), limit), nil
}

// FileStore persists audit events under the platform data root as JSON lines.
type FileStore struct {
	mu   sync.RWMutex
	path string
}

// NewFileStore constructs a local-first audit store.
func NewFileStore(dataRoot string) (*FileStore, error) {
	path := filepath.Join(dataRoot, "control_plane", "audit_events.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create audit dir: %w", err)
	}
	return &FileStore{path: path}, nil
}

func (s *FileStore) Append(event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer file.Close()

	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode audit event: %w", err)
	}
	if _, err := file.Write(append(bytes, '\n')); err != nil {
		return fmt.Errorf("append audit event: %w", err)
	}
	return nil
}

func (s *FileStore) ListRecent(limit int) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, fmt.Errorf("open audit log: %w", err)
	}
	defer file.Close()

	events := []Event{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode audit event: %w", err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan audit log: %w", err)
	}
	return trimRecent(events, limit), nil
}

// MultiStore mirrors audit events into a primary and secondary store.
type MultiStore struct {
	primary   Store
	secondary Store
}

// NewMultiStore constructs a mirrored audit store.
func NewMultiStore(primary, secondary Store) Store {
	return &MultiStore{primary: primary, secondary: secondary}
}

func (s *MultiStore) Append(event Event) error {
	if err := s.primary.Append(event); err != nil {
		return err
	}
	return s.secondary.Append(event)
}

func (s *MultiStore) ListRecent(limit int) ([]Event, error) {
	events, err := s.primary.ListRecent(limit)
	if err == nil && len(events) > 0 {
		return events, nil
	}
	return s.secondary.ListRecent(limit)
}

func trimRecent(events []Event, limit int) []Event {
	if limit <= 0 || len(events) <= limit {
		return events
	}
	return append([]Event(nil), events[len(events)-limit:]...)
}

func cloneEvents(input []Event) []Event {
	out := make([]Event, len(input))
	copy(out, input)
	return out
}
