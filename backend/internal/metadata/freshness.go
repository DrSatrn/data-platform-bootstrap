package metadata

import (
	"fmt"
	"os"
	"time"
)

// ResolveFreshnessStatus derives the runtime freshness status for one asset
// from its local materialization timestamp and configured SLA windows.
func ResolveFreshnessStatus(dataRoot string, asset DataAsset, now time.Time) Status {
	path := MaterializationPath(dataRoot, asset.ID)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Status{
				State:   "missing",
				Message: "No local materialization has been recorded yet.",
			}
		}
		return Status{
			State:   "unknown",
			Message: fmt.Sprintf("Unable to inspect local materialization: %v", err),
		}
	}

	updatedAt := info.ModTime().UTC()
	lag := now.Sub(updatedAt)
	if lag < 0 {
		lag = 0
	}

	expectedWithin, expectedErr := time.ParseDuration(asset.Freshness.ExpectedWithin)
	warnAfter, warnErr := time.ParseDuration(asset.Freshness.WarnAfter)
	switch {
	case expectedErr != nil || warnErr != nil:
		return Status{
			State:       "fresh",
			LastUpdated: updatedAt.Format(time.RFC3339),
			LagSeconds:  int64(lag.Seconds()),
			Message:     "Freshness SLA is configured but could not be parsed; using raw local timestamp only.",
		}
	case lag > warnAfter:
		return Status{
			State:       "stale",
			LastUpdated: updatedAt.Format(time.RFC3339),
			LagSeconds:  int64(lag.Seconds()),
			Message:     fmt.Sprintf("Asset is past its warning SLA of %s.", asset.Freshness.WarnAfter),
		}
	case lag > expectedWithin:
		return Status{
			State:       "late",
			LastUpdated: updatedAt.Format(time.RFC3339),
			LagSeconds:  int64(lag.Seconds()),
			Message:     fmt.Sprintf("Asset is past its expected freshness target of %s.", asset.Freshness.ExpectedWithin),
		}
	default:
		return Status{
			State:       "fresh",
			LastUpdated: updatedAt.Format(time.RFC3339),
			LagSeconds:  int64(lag.Seconds()),
			Message:     "Asset is within its expected freshness window.",
		}
	}
}
