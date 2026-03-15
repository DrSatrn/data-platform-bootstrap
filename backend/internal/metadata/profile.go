// This file implements runtime dataset profiling for the catalog. Profiles are
// cached on disk so the UI can expose row counts and column-level trust signals
// without re-running Python inspection on every request.
package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AssetProfiler describes the bounded profiling capability required by the
// catalog. The concrete implementation is backed by the Python utility runner.
type AssetProfiler interface {
	Profile(ctx context.Context, assetID, sourcePath string) (AssetProfile, error)
}

// ProfileService resolves assets, manages cache freshness, and returns the
// latest available profile for operator-facing dataset detail pages.
type ProfileService struct {
	loader   AssetLoader
	profiler AssetProfiler
	dataRoot string
}

// NewProfileService constructs a profile service.
func NewProfileService(loader AssetLoader, profiler AssetProfiler, dataRoot string) *ProfileService {
	return &ProfileService{
		loader:   loader,
		profiler: profiler,
		dataRoot: dataRoot,
	}
}

// GenerateProfile returns the latest profile for one asset, regenerating it if
// the cache is missing or older than the source file.
func (s *ProfileService) GenerateProfile(ctx context.Context, assetID string) (AssetProfile, error) {
	assets, err := s.loader.LoadAssets()
	if err != nil {
		return AssetProfile{}, fmt.Errorf("load assets: %w", err)
	}

	var asset *DataAsset
	for index := range assets {
		if assets[index].ID == assetID {
			asset = &assets[index]
			break
		}
	}
	if asset == nil {
		return AssetProfile{}, fmt.Errorf("asset %s not found", assetID)
	}

	sourcePath := MaterializationPath(s.dataRoot, asset.ID)
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return AssetProfile{}, fmt.Errorf("inspect materialized asset %s: %w", asset.ID, err)
	}

	cachePath := filepath.Join(s.dataRoot, "profiles", asset.ID+".json")
	if profile, ok := s.readCachedProfile(cachePath, sourceInfo.ModTime()); ok {
		return profile, nil
	}

	profile, err := s.profiler.Profile(ctx, asset.ID, sourcePath)
	if err != nil {
		return AssetProfile{}, err
	}
	profile.AssetID = asset.ID
	profile.Path = sourcePath
	profile.FileBytes = sourceInfo.Size()
	profile.ObservedAt = sourceInfo.ModTime().UTC().Format(time.RFC3339)
	profile.ProfileState = "fresh"
	if profile.GeneratedAt == "" {
		profile.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if err := s.writeCachedProfile(cachePath, profile); err != nil {
		return AssetProfile{}, err
	}
	return profile, nil
}

func (s *ProfileService) readCachedProfile(path string, sourceUpdatedAt time.Time) (AssetProfile, bool) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return AssetProfile{}, false
	}
	info, err := os.Stat(path)
	if err != nil || info.ModTime().Before(sourceUpdatedAt) {
		return AssetProfile{}, false
	}

	var profile AssetProfile
	if err := json.Unmarshal(bytes, &profile); err != nil {
		return AssetProfile{}, false
	}
	profile.ProfileState = "cached"
	return profile, true
}

func (s *ProfileService) writeCachedProfile(path string, profile AssetProfile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create profile cache dir: %w", err)
	}
	bytes, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile cache: %w", err)
	}
	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return fmt.Errorf("write profile cache: %w", err)
	}
	return nil
}
