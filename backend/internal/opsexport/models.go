// Package opsexport converts backend-only opsview read models into stable
// export bundles suitable for demos, review, snapshots, or future API
// publication work.
package opsexport

import (
	"time"

	"github.com/streanor/data-platform/backend/internal/opsview"
)

// Bundle is the stable export envelope for one set of opsview snapshots.
type Bundle struct {
	Version       string                        `json:"version"`
	GeneratedAt   time.Time                     `json:"generated_at"`
	SnapshotCount int                           `json:"snapshot_count"`
	Rollup        opsview.AttentionRollup       `json:"rollup"`
	Snapshots     []opsview.RunOperatorSnapshot `json:"snapshots"`
}
