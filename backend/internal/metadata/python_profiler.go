// This file adapts the shared Python runner into the metadata profiling
// interface. The Go control plane decides when profiling is allowed; Python
// handles the row/column inspection where text and mixed-type handling is
// easier to express.
package metadata

import (
	"context"

	"github.com/streanor/data-platform/backend/internal/python"
)

type pythonProfiler struct {
	runner *python.Runner
}

type profileRequest struct {
	AssetID    string `json:"asset_id"`
	SourcePath string `json:"source_path"`
}

type profileResult struct {
	Format      string          `json:"format"`
	RowCount    int             `json:"row_count"`
	GeneratedAt string          `json:"generated_at"`
	Columns     []ColumnProfile `json:"columns"`
}

// NewPythonProfiler constructs an AssetProfiler backed by the repo-owned
// Python utility script.
func NewPythonProfiler(runner *python.Runner) AssetProfiler {
	return &pythonProfiler{runner: runner}
}

func (p *pythonProfiler) Profile(ctx context.Context, assetID, sourcePath string) (AssetProfile, error) {
	var result profileResult
	if err := p.runner.RunUtility(ctx, "tasks/profile_asset.py", profileRequest{
		AssetID:    assetID,
		SourcePath: sourcePath,
	}, &result, nil); err != nil {
		return AssetProfile{}, err
	}

	return AssetProfile{
		AssetID:     assetID,
		Path:        sourcePath,
		Format:      result.Format,
		RowCount:    result.RowCount,
		GeneratedAt: result.GeneratedAt,
		Columns:     result.Columns,
	}, nil
}
