// This file defines the manifest-backed quality-check model so validation and
// future control-plane features can reason about quality definitions directly
// instead of treating them as untyped YAML blobs.
package quality

// Definition describes one repo-managed quality check contract.
type Definition struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	Severity    string    `json:"severity" yaml:"severity"`
	AssetRef    string    `json:"asset_ref" yaml:"asset_ref"`
	Type        string    `json:"type" yaml:"type"`
	ColumnRef   string    `json:"column_ref" yaml:"column_ref"`
	Threshold   Threshold `json:"threshold" yaml:"threshold"`
}

// Threshold defines the simple threshold-based contract used by the current
// sample quality manifests.
type Threshold struct {
	Operator string `json:"operator" yaml:"operator"`
	Value    int    `json:"value" yaml:"value"`
}
