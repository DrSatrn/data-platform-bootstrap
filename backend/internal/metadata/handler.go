// This file exposes the metadata catalog API surface. The handler loads asset
// manifests and merges them with the current in-memory catalog snapshot.
package metadata

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// AssetLoader defines the manifest-loading behavior the catalog API needs.
// Keeping the dependency inverted here prevents import cycles with the manifest
// package, which already depends on metadata models.
type AssetLoader interface {
	LoadAssets() ([]DataAsset, error)
}

// CatalogHandler serves dataset and metadata endpoints.
type CatalogHandler struct {
	loader   AssetLoader
	catalog  *Catalog
	dataRoot string
	store    Store
}

// NewCatalogHandler constructs the metadata API handler.
func NewCatalogHandler(loader AssetLoader, catalog *Catalog, dataRoot string, store Store) http.Handler {
	return &CatalogHandler{
		loader:   loader,
		catalog:  catalog,
		dataRoot: dataRoot,
		store:    store,
	}
}

func (h *CatalogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assets, err := h.loader.LoadAssets()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load assets",
		})
		return
	}
	if h.store != nil {
		if err := h.store.SyncAssets(assets); err == nil {
			if storedAssets, listErr := h.store.ListAssets(); listErr == nil && len(storedAssets) > 0 {
				assets = storedAssets
			}
		}
	}

	h.catalog.ReplaceAssets(assets)
	enriched := h.catalog.ListAssets()
	for index := range enriched {
		enriched[index].FreshnessStatus = h.freshnessStatus(enriched[index])
	}
	enriched = EnrichAssets(enriched)
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"assets":  enriched,
		"summary": SummarizeAssets(enriched),
		"lineage": BuildEdges(enriched),
	})
}

func (h *CatalogHandler) freshnessStatus(asset DataAsset) Status {
	path := MaterializationPath(h.dataRoot, asset.ID)
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
	lag := time.Since(updatedAt)
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
