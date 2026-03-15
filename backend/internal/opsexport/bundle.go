package opsexport

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/streanor/data-platform/backend/internal/opsview"
)

const bundleVersion = "opsexport.v1"

// BuildBundle creates a stable export bundle from pure opsview snapshots.
func BuildBundle(generatedAt time.Time, snapshots []opsview.RunOperatorSnapshot) Bundle {
	ordered := append([]opsview.RunOperatorSnapshot(nil), snapshots...)
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].UpdatedAt.Equal(ordered[right].UpdatedAt) {
			return ordered[left].RunID < ordered[right].RunID
		}
		return ordered[left].UpdatedAt.After(ordered[right].UpdatedAt)
	})
	return Bundle{
		Version:       bundleVersion,
		GeneratedAt:   generatedAt.UTC(),
		SnapshotCount: len(ordered),
		Rollup:        opsview.BuildAttentionRollup(ordered),
		Snapshots:     ordered,
	}
}

// MarshalBundle produces stable indented JSON for demos, review, or snapshot
// publication.
func MarshalBundle(bundle Bundle) ([]byte, error) {
	bytes, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal ops export bundle: %w", err)
	}
	return bytes, nil
}

// SuggestedFilename gives future export tooling a stable naming convention
// without requiring any runtime wiring today.
func SuggestedFilename(generatedAt time.Time) string {
	return "opsview_export_" + generatedAt.UTC().Format("20060102T150405Z") + ".json"
}
