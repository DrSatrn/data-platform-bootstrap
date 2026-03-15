// This file centralizes how logical asset identifiers map to local
// materialization paths. Keeping the mapping in one place avoids drift between
// freshness checks, profiling, and future artifact inspection features.
package metadata

import "path/filepath"

// MaterializationPath returns the expected local file path for an asset.
func MaterializationPath(dataRoot, assetID string) string {
	switch {
	case assetID == "raw_transactions":
		return filepath.Join(dataRoot, "raw", "raw_transactions.csv")
	case assetID == "raw_account_balances":
		return filepath.Join(dataRoot, "raw", "raw_account_balances.json")
	case assetID == "raw_budget_rules":
		return filepath.Join(dataRoot, "raw", "raw_budget_rules.json")
	case len(assetID) > 8 && assetID[:8] == "staging_":
		return filepath.Join(dataRoot, "staging", assetID+".json")
	case len(assetID) > 13 && assetID[:13] == "intermediate_":
		return filepath.Join(dataRoot, "intermediate", assetID+".json")
	case len(assetID) > 5 && assetID[:5] == "mart_":
		return filepath.Join(dataRoot, "mart", assetID+".json")
	case len(assetID) > 8 && assetID[:8] == "metrics_":
		return filepath.Join(dataRoot, "metrics", assetID+".json")
	default:
		return filepath.Join(dataRoot, assetID+".json")
	}
}
