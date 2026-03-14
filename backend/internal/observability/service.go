// Package observability owns built-in telemetry for the platform. This service
// keeps recent logs, request metrics, and admin command history in process so
// the product can expose first-party diagnostics without relying on external
// monitoring systems.
package observability

import (
	"sync"
	"time"
)

const (
	maxRecentLogs     = 200
	maxRecentRequests = 200
	maxRecentCommands = 100
)

// LogEntry captures one structured application log event.
type LogEntry struct {
	Time    time.Time         `json:"time"`
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

// RequestEntry captures an HTTP request observation.
type RequestEntry struct {
	Time       time.Time `json:"time"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	DurationMS int64     `json:"duration_ms"`
}

// CommandEntry captures one admin terminal command execution.
type CommandEntry struct {
	Time    time.Time `json:"time"`
	Command string    `json:"command"`
	Success bool      `json:"success"`
	Preview string    `json:"preview"`
}

// Snapshot summarizes the current built-in telemetry state for API consumers.
type Snapshot struct {
	StartedAt        time.Time         `json:"started_at"`
	UptimeSeconds    int64             `json:"uptime_seconds"`
	TotalRequests    int               `json:"total_requests"`
	TotalErrors      int               `json:"total_errors"`
	TotalCommands    int               `json:"total_commands"`
	RequestCounts    map[string]int    `json:"request_counts"`
	RecentRequests   []RequestEntry    `json:"recent_requests"`
	RecentCommands   []CommandEntry    `json:"recent_commands"`
	RecentLogSummary []LogEntry        `json:"recent_log_summary"`
	RuntimeLabels    map[string]string `json:"runtime_labels"`
}

// Service is the in-process telemetry registry.
type Service struct {
	mu             sync.RWMutex
	startedAt      time.Time
	totalRequests  int
	totalErrors    int
	totalCommands  int
	requestCounts  map[string]int
	recentRequests []RequestEntry
	recentLogs     []LogEntry
	recentCommands []CommandEntry
}

// NewService creates a telemetry registry with startup time captured.
func NewService() *Service {
	return &Service{
		startedAt:     time.Now().UTC(),
		requestCounts: map[string]int{},
	}
}

// RecordLog keeps a ring buffer of recent log entries.
func (s *Service) RecordLog(level, message string, fields map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.recentLogs = appendTrimmed(
		s.recentLogs,
		maxRecentLogs,
		LogEntry{
			Time:    time.Now().UTC(),
			Level:   level,
			Message: message,
			Fields:  cloneMap(fields),
		},
	)
}

// RecordRequest tracks request counts and recent request samples.
func (s *Service) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalRequests++
	if statusCode >= 400 {
		s.totalErrors++
	}
	s.requestCounts[path]++
	s.recentRequests = appendTrimmed(
		s.recentRequests,
		maxRecentRequests,
		RequestEntry{
			Time:       time.Now().UTC(),
			Method:     method,
			Path:       path,
			StatusCode: statusCode,
			DurationMS: duration.Milliseconds(),
		},
	)
}

// RecordCommand tracks admin-terminal usage and outcome.
func (s *Service) RecordCommand(command string, success bool, preview string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalCommands++
	s.recentCommands = appendTrimmed(
		s.recentCommands,
		maxRecentCommands,
		CommandEntry{
			Time:    time.Now().UTC(),
			Command: command,
			Success: success,
			Preview: preview,
		},
	)
}

// Snapshot returns a stable copy of the telemetry state.
func (s *Service) Snapshot(runtimeLabels map[string]string) Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Snapshot{
		StartedAt:        s.startedAt,
		UptimeSeconds:    int64(time.Since(s.startedAt).Seconds()),
		TotalRequests:    s.totalRequests,
		TotalErrors:      s.totalErrors,
		TotalCommands:    s.totalCommands,
		RequestCounts:    cloneIntMap(s.requestCounts),
		RecentRequests:   cloneRequests(s.recentRequests),
		RecentCommands:   cloneCommands(s.recentCommands),
		RecentLogSummary: cloneLogs(s.recentLogs),
		RuntimeLabels:    cloneMap(runtimeLabels),
	}
}

// RecentLogs returns a copy of the recent log buffer.
func (s *Service) RecentLogs() []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return cloneLogs(s.recentLogs)
}

func appendTrimmed[T any](items []T, limit int, next T) []T {
	items = append(items, next)
	if len(items) <= limit {
		return items
	}
	return append([]T(nil), items[len(items)-limit:]...)
}

func cloneMap(input map[string]string) map[string]string {
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneIntMap(input map[string]int) map[string]int {
	out := make(map[string]int, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneLogs(input []LogEntry) []LogEntry {
	out := make([]LogEntry, len(input))
	copy(out, input)
	return out
}

func cloneRequests(input []RequestEntry) []RequestEntry {
	out := make([]RequestEntry, len(input))
	copy(out, input)
	return out
}

func cloneCommands(input []CommandEntry) []CommandEntry {
	out := make([]CommandEntry, len(input))
	copy(out, input)
	return out
}
