// These tests cover the dataset profiling cache so the catalog detail page can
// rely on stable row-count and column-shape summaries without re-running Python
// work unnecessarily.
package metadata

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type stubAssetLoader struct {
	assets []DataAsset
}

func (s stubAssetLoader) LoadAssets() ([]DataAsset, error) {
	return s.assets, nil
}

type stubProfiler struct {
	calls int
}

func (s *stubProfiler) Profile(_ context.Context, assetID, sourcePath string) (AssetProfile, error) {
	s.calls++
	return AssetProfile{
		AssetID:     assetID,
		Path:        sourcePath,
		Format:      "json",
		RowCount:    2,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Columns: []ColumnProfile{
			{Name: "category", ObservedType: "string", NullCount: 0, UniqueCount: 2, SampleValues: []string{"groceries", "rent"}},
		},
	}, nil
}

func TestProfileServiceCachesFreshProfiles(t *testing.T) {
	root := t.TempDir()
	martPath := filepath.Join(root, "mart", "mart_budget_vs_actual.json")
	if err := os.MkdirAll(filepath.Dir(martPath), 0o755); err != nil {
		t.Fatalf("mkdir mart dir: %v", err)
	}
	if err := os.WriteFile(martPath, []byte(`[{"category":"groceries"}]`), 0o644); err != nil {
		t.Fatalf("write mart asset: %v", err)
	}

	profiler := &stubProfiler{}
	service := NewProfileService(stubAssetLoader{
		assets: []DataAsset{{ID: "mart_budget_vs_actual"}},
	}, profiler, root)

	first, err := service.GenerateProfile(context.Background(), "mart_budget_vs_actual")
	if err != nil {
		t.Fatalf("generate first profile: %v", err)
	}
	second, err := service.GenerateProfile(context.Background(), "mart_budget_vs_actual")
	if err != nil {
		t.Fatalf("generate cached profile: %v", err)
	}

	if profiler.calls != 1 {
		t.Fatalf("expected one profiler call, got %d", profiler.calls)
	}
	if first.RowCount != second.RowCount {
		t.Fatalf("expected cached profile row count to match, got %d and %d", first.RowCount, second.RowCount)
	}
	if second.ProfileState != "cached" {
		t.Fatalf("expected cached profile state, got %s", second.ProfileState)
	}
}

func TestProfileServiceRebuildsStaleCache(t *testing.T) {
	root := t.TempDir()
	martPath := filepath.Join(root, "mart", "mart_category_spend.json")
	cachePath := filepath.Join(root, "profiles", "mart_category_spend.json")
	if err := os.MkdirAll(filepath.Dir(martPath), 0o755); err != nil {
		t.Fatalf("mkdir mart dir: %v", err)
	}
	if err := os.WriteFile(martPath, []byte(`[{"category":"groceries"}]`), 0o644); err != nil {
		t.Fatalf("write mart asset: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir profile dir: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte(`{"asset_id":"mart_category_spend","row_count":1,"profile_state":"cached"}`), 0o644); err != nil {
		t.Fatalf("write cached profile: %v", err)
	}

	old := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(cachePath, old, old); err != nil {
		t.Fatalf("set cache time: %v", err)
	}
	if err := os.Chtimes(martPath, newer, newer); err != nil {
		t.Fatalf("set source time: %v", err)
	}

	profiler := &stubProfiler{}
	service := NewProfileService(stubAssetLoader{
		assets: []DataAsset{{ID: "mart_category_spend"}},
	}, profiler, root)

	profile, err := service.GenerateProfile(context.Background(), "mart_category_spend")
	if err != nil {
		t.Fatalf("rebuild stale profile: %v", err)
	}
	if profiler.calls != 1 {
		t.Fatalf("expected profiler to rebuild stale cache, got %d calls", profiler.calls)
	}
	if profile.ProfileState != "fresh" {
		t.Fatalf("expected fresh profile state after rebuild, got %s", profile.ProfileState)
	}
}
