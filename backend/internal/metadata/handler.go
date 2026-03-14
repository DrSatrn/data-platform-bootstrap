// This file exposes the metadata catalog API surface. The handler loads asset
// manifests and merges them with the current in-memory catalog snapshot.
package metadata

import (
	"net/http"

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
	loader  AssetLoader
	catalog *Catalog
}

// NewCatalogHandler constructs the metadata API handler.
func NewCatalogHandler(loader AssetLoader, catalog *Catalog) http.Handler {
	return &CatalogHandler{
		loader:  loader,
		catalog: catalog,
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

	h.catalog.ReplaceAssets(assets)
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"assets": h.catalog.ListAssets(),
	})
}
