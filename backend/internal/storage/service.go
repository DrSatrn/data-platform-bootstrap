// Package storage exposes safe inspection of materialized run artifacts. The
// service only serves files rooted under the configured artifact directory so
// callers can inspect results without gaining arbitrary filesystem access.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Artifact describes one run-scoped materialized file.
type Artifact struct {
	RunID        string    `json:"run_id"`
	RelativePath string    `json:"relative_path"`
	SizeBytes    int64     `json:"size_bytes"`
	ModifiedAt   time.Time `json:"modified_at"`
	ContentType  string    `json:"content_type"`
}

// Service exposes artifact discovery and reading rooted under the local
// artifact directory.
type Service struct {
	root string
}

// NewService constructs an artifact inspection service.
func NewService(root string) *Service {
	return &Service{root: root}
}

// ListRunArtifacts returns all files for a run-scoped artifact directory.
func (s *Service) ListRunArtifacts(runID string) ([]Artifact, error) {
	runRoot := filepath.Join(s.root, "runs", runID)
	entries := []Artifact{}
	err := filepath.Walk(runRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(runRoot, path)
		if err != nil {
			return err
		}
		entries = append(entries, Artifact{
			RunID:        runID,
			RelativePath: filepath.ToSlash(relative),
			SizeBytes:    info.Size(),
			ModifiedAt:   info.ModTime().UTC(),
			ContentType:  detectContentType(relative),
		})
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return []Artifact{}, nil
		}
		return nil, fmt.Errorf("list run artifacts for %s: %w", runID, err)
	}
	return entries, nil
}

// ReadRunArtifact returns the bytes for a run-scoped artifact after validating
// the requested path stays within the run artifact directory.
func (s *Service) ReadRunArtifact(runID, relativePath string) ([]byte, error) {
	clean := filepath.Clean(relativePath)
	if clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return nil, fmt.Errorf("invalid artifact path")
	}
	target := filepath.Join(s.root, "runs", runID, clean)
	bytes, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("read run artifact %s/%s: %w", runID, clean, err)
	}
	return bytes, nil
}

func detectContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return "application/json"
	case ".csv":
		return "text/csv"
	default:
		return "application/octet-stream"
	}
}
