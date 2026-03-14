// Package manifests loads repo-managed YAML definitions for pipelines, assets,
// metrics, quality checks, and owners. The loader intentionally stays strict so
// malformed configuration fails loudly instead of creating hidden runtime drift.
package manifests

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"gopkg.in/yaml.v3"
)

// Loader defines the manifest access required by platform services.
type Loader interface {
	LoadPipelines() ([]orchestration.Pipeline, error)
	LoadAssets() ([]metadata.DataAsset, error)
}

// FileLoader reads manifests from the repo filesystem.
type FileLoader struct {
	root string
}

// NewLoader creates a filesystem-backed loader.
func NewLoader(root string) Loader {
	return &FileLoader{root: root}
}

// LoadPipelines reads all pipeline manifests in the pipelines directory.
func (l *FileLoader) LoadPipelines() ([]orchestration.Pipeline, error) {
	pattern := filepath.Join(l.root, "pipelines", "*.yaml")
	return loadFiles[orchestration.Pipeline](pattern)
}

// LoadAssets reads all asset manifests in the assets directory.
func (l *FileLoader) LoadAssets() ([]metadata.DataAsset, error) {
	pattern := filepath.Join(l.root, "assets", "*.yaml")
	return loadFiles[metadata.DataAsset](pattern)
}

func loadFiles[T any](pattern string) ([]T, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob manifests: %w", err)
	}

	out := make([]T, 0, len(matches))
	for _, match := range matches {
		bytes, err := os.ReadFile(match)
		if err != nil {
			return nil, fmt.Errorf("read manifest %s: %w", match, err)
		}

		var item T
		if err := yaml.Unmarshal(bytes, &item); err != nil {
			return nil, fmt.Errorf("decode manifest %s: %w", match, err)
		}
		out = append(out, item)
	}

	return out, nil
}
