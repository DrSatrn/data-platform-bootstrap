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
