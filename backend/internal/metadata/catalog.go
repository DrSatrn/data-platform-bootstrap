// This file provides a small concurrency-safe in-memory metadata catalog. The
// in-memory form is enough for the scaffold and keeps the API useful while the
// Postgres-backed implementation is still under construction.
package metadata

import "sync"

// Catalog holds the currently loaded asset catalog.
type Catalog struct {
	mu     sync.RWMutex
	assets []DataAsset
}

// NewCatalog returns an empty catalog.
func NewCatalog() *Catalog {
	return &Catalog{assets: []DataAsset{}}
}

// ReplaceAssets swaps the full asset snapshot. This full replacement strategy
// keeps the in-memory implementation simple and avoids partial update bugs.
func (c *Catalog) ReplaceAssets(assets []DataAsset) {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]DataAsset, len(assets))
	copy(out, assets)
	c.assets = out
}

// ListAssets returns a stable snapshot for API responses.
func (c *Catalog) ListAssets() []DataAsset {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]DataAsset, len(c.assets))
	copy(out, c.assets)
	return out
}

// EnrichAssets computes derived trust and lineage fields from the current
// asset snapshot. Keeping this logic centralized makes the catalog API and
// tests share the same business rules.
func EnrichAssets(assets []DataAsset) []DataAsset {
	downstreamBySource := map[string][]string{}
	for _, asset := range assets {
		for _, source := range asset.SourceRefs {
			downstreamBySource[source] = appendIfMissing(downstreamBySource[source], asset.ID)
		}
	}

	enriched := make([]DataAsset, len(assets))
	for index, asset := range assets {
		asset.Coverage = deriveCoverage(asset)
		asset.Lineage = Lineage{
			Upstream:   dedupeStrings(asset.SourceRefs),
			Downstream: dedupeStrings(downstreamBySource[asset.ID]),
		}
		enriched[index] = asset
	}
	return enriched
}

// SummarizeAssets aggregates the catalog into operator-friendly coverage and
// trust counts used by the UI and future validation tooling.
func SummarizeAssets(assets []DataAsset) Summary {
	summary := Summary{
		TotalAssets: len(assets),
		ByLayer:     map[string]int{},
		ByFreshness: map[string]int{},
	}

	for _, asset := range assets {
		summary.ByLayer[asset.Layer]++
		summary.ByFreshness[asset.FreshnessStatus.State]++
		summary.TotalColumns += asset.Coverage.TotalColumns
		summary.DocumentedColumns += asset.Coverage.DocumentedColumns
		if !asset.Coverage.HasDocumentation {
			summary.AssetsMissingDocs++
		}
		if !asset.Coverage.HasQualityChecks {
			summary.AssetsMissingQuality++
		}
		if asset.Coverage.ContainsPII {
			summary.AssetsContainingPII++
		}
		summary.LineageEdges += len(asset.Lineage.Upstream)
	}

	return summary
}

// BuildEdges converts asset source references into graph edges for the UI.
func BuildEdges(assets []DataAsset) []Edge {
	edges := make([]Edge, 0)
	seen := map[string]struct{}{}
	for _, asset := range assets {
		for _, upstream := range asset.Lineage.Upstream {
			key := upstream + "->" + asset.ID
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			edges = append(edges, Edge{From: upstream, To: asset.ID})
		}
	}
	return edges
}

func deriveCoverage(asset DataAsset) Coverage {
	coverage := Coverage{
		TotalColumns:     len(asset.Columns),
		HasDocumentation: len(asset.DocumentationRefs) > 0,
		HasQualityChecks: len(asset.QualityCheckRefs) > 0,
	}
	for _, column := range asset.Columns {
		if column.Description != "" {
			coverage.DocumentedColumns++
		}
		if column.IsPII {
			coverage.ContainsPII = true
		}
	}
	return coverage
}

func dedupeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func appendIfMissing(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}
