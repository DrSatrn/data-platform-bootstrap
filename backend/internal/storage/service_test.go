package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadRunArtifactRejectsTraversal(t *testing.T) {
	service := NewService(t.TempDir(), nil)

	if _, err := service.ReadRunArtifact("run_1", "../secrets.txt"); err == nil {
		t.Fatal("expected path traversal attempt to be rejected")
	}
}

func TestListRunArtifactsFallsBackToFilesystemWhenIndexIsEmpty(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "runs", "run_1", "nested")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", runDir, err)
	}
	target := filepath.Join(runDir, "result.json")
	if err := os.WriteFile(target, []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	service := NewService(root, stubIndex{})
	artifacts, err := service.ListRunArtifacts("run_1")
	if err != nil {
		t.Fatalf("ListRunArtifacts returned error: %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].RelativePath != "nested/result.json" {
		t.Fatalf("expected relative path nested/result.json, got %q", artifacts[0].RelativePath)
	}
}

func TestRecordRunArtifactPersistsMetadataToIndex(t *testing.T) {
	root := t.TempDir()
	targetDir := filepath.Join(root, "runs", "run_1", "metrics")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", targetDir, err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "metric.json"), []byte(`{"value":1}`), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	index := &capturingIndex{}
	service := NewService(root, index)
	if err := service.RecordRunArtifact("run_1", "metrics/metric.json"); err != nil {
		t.Fatalf("RecordRunArtifact returned error: %v", err)
	}
	if index.recorded == nil {
		t.Fatal("expected metadata index to receive an artifact record")
	}
	if index.recorded.RunID != "run_1" {
		t.Fatalf("expected run_id run_1, got %q", index.recorded.RunID)
	}
	if index.recorded.RelativePath != "metrics/metric.json" {
		t.Fatalf("expected relative path metrics/metric.json, got %q", index.recorded.RelativePath)
	}
	if index.recorded.ContentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", index.recorded.ContentType)
	}
}

type stubIndex struct{}

func (stubIndex) ListRunArtifacts(string) ([]Artifact, error) {
	return nil, errors.New("index unavailable")
}

func (stubIndex) RecordArtifact(Artifact) error {
	return nil
}

type capturingIndex struct {
	recorded *Artifact
}

func (c *capturingIndex) ListRunArtifacts(string) ([]Artifact, error) {
	return nil, nil
}

func (c *capturingIndex) RecordArtifact(artifact Artifact) error {
	copyArtifact := artifact
	c.recorded = &copyArtifact
	return nil
}
