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

// Owner describes a responsible individual or team.
type Owner struct {
	ID          string `json:"id" yaml:"id"`
	DisplayName string `json:"display_name" yaml:"display_name"`
	Email       string `json:"email" yaml:"email"`
	Team        string `json:"team" yaml:"team"`
}
