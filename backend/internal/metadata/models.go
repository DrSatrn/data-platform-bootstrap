// Package metadata defines the catalog model used by the platform to describe
// datasets, columns, lineage, freshness, and ownership. These types are shared
// across manifest loading and API responses.
package metadata

// DataAsset describes a curated or intermediate platform asset.
type DataAsset struct {
	ID                string    `json:"id" yaml:"id"`
	Name              string    `json:"name" yaml:"name"`
	Layer             string    `json:"layer" yaml:"layer"`
	Description       string    `json:"description" yaml:"description"`
	Owner             string    `json:"owner" yaml:"owner"`
	Kind              string    `json:"kind" yaml:"kind"`
	SourceRefs        []string  `json:"source_refs" yaml:"source_refs"`
	Columns           []Column  `json:"columns" yaml:"columns"`
	Freshness         Freshness `json:"freshness" yaml:"freshness"`
	FreshnessStatus   Status    `json:"freshness_status"`
	Coverage          Coverage  `json:"coverage"`
	Lineage           Lineage   `json:"lineage"`
	QualityCheckRefs  []string  `json:"quality_check_refs" yaml:"quality_check_refs"`
	DocumentationRefs []string  `json:"documentation_refs" yaml:"documentation_refs"`
}

// Column describes one field in a data asset.
type Column struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"`
	Description string `json:"description" yaml:"description"`
	IsPII       bool   `json:"is_pii" yaml:"is_pii"`
}

// Freshness captures the SLA-style expectation for asset recency.
type Freshness struct {
	ExpectedWithin string `json:"expected_within" yaml:"expected_within"`
	WarnAfter      string `json:"warn_after" yaml:"warn_after"`
}

// Status describes the current observed freshness state for an asset. It is
// computed at runtime from local materializations rather than coming directly
// from the manifest.
type Status struct {
	State       string `json:"state"`
	LastUpdated string `json:"last_updated,omitempty"`
	LagSeconds  int64  `json:"lag_seconds,omitempty"`
	Message     string `json:"message"`
}

// Owner describes a responsible individual or team.
type Owner struct {
	ID          string `json:"id" yaml:"id"`
	DisplayName string `json:"display_name" yaml:"display_name"`
	Email       string `json:"email" yaml:"email"`
	Team        string `json:"team" yaml:"team"`
}

// Coverage describes how well a catalog asset is documented and governed.
type Coverage struct {
	DocumentedColumns int  `json:"documented_columns"`
	TotalColumns      int  `json:"total_columns"`
	HasDocumentation  bool `json:"has_documentation"`
	HasQualityChecks  bool `json:"has_quality_checks"`
	ContainsPII       bool `json:"contains_pii"`
}

// Lineage captures immediate upstream and downstream relationships for an
// asset so the UI can render useful context without a separate graph service.
type Lineage struct {
	Upstream   []string `json:"upstream"`
	Downstream []string `json:"downstream"`
}

// Edge describes one lineage relationship between two assets or sources.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Summary provides high-signal aggregate catalog metrics for operators and
// validation tooling.
type Summary struct {
	TotalAssets          int            `json:"total_assets"`
	ByLayer              map[string]int `json:"by_layer"`
	ByFreshness          map[string]int `json:"by_freshness"`
	AssetsMissingDocs    int            `json:"assets_missing_docs"`
	AssetsMissingQuality int            `json:"assets_missing_quality"`
	AssetsContainingPII  int            `json:"assets_containing_pii"`
	DocumentedColumns    int            `json:"documented_columns"`
	TotalColumns         int            `json:"total_columns"`
	LineageEdges         int            `json:"lineage_edges"`
}

// AssetProfile captures runtime-observed shape information about a materialized
// asset. The profile intentionally focuses on operator-facing trust signals
// rather than exhaustive statistics so the UI remains fast and readable.
type AssetProfile struct {
	AssetID      string          `json:"asset_id"`
	Path         string          `json:"path"`
	Format       string          `json:"format"`
	RowCount     int             `json:"row_count"`
	FileBytes    int64           `json:"file_bytes"`
	GeneratedAt  string          `json:"generated_at"`
	ObservedAt   string          `json:"observed_at,omitempty"`
	ProfileState string          `json:"profile_state"`
	Columns      []ColumnProfile `json:"columns"`
}

// ColumnProfile summarizes the observed values for one column in a materialized
// asset. Values are stored as strings to keep the JSON format simple across CSV
// and JSON-backed assets.
type ColumnProfile struct {
	Name         string   `json:"name"`
	ObservedType string   `json:"observed_type"`
	NullCount    int      `json:"null_count"`
	UniqueCount  int      `json:"unique_count"`
	SampleValues []string `json:"sample_values"`
	MinValue     string   `json:"min_value,omitempty"`
	MaxValue     string   `json:"max_value,omitempty"`
}

// Store defines the persistence behavior for the synchronized metadata catalog.
type Store interface {
	SyncAssets([]DataAsset) error
	ListAssets() ([]DataAsset, error)
}
