// This file exposes the metadata catalog API surface. When PostgreSQL metadata
// persistence is available, the catalog is served from the database and
// operator annotations are written back there directly. The manifest loader is
// only used when the database path is unavailable.
package metadata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/authz"
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
	authz    *authz.Service
	audit    audit.Store
}

// NewCatalogHandler constructs the metadata API handler.
func NewCatalogHandler(loader AssetLoader, catalog *Catalog, dataRoot string, store Store, authService *authz.Service, auditStore audit.Store) http.Handler {
	return &CatalogHandler{
		loader:   loader,
		catalog:  catalog,
		dataRoot: dataRoot,
		store:    store,
		authz:    authService,
		audit:    auditStore,
	}
}

func (h *CatalogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveList(w)
	case http.MethodPatch:
		h.servePatch(w, r)
	default:
		w.Header().Set("Allow", "GET, PATCH")
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
	}
}

func (h *CatalogHandler) serveList(w http.ResponseWriter) {
	assets, err := h.assetsForResponse()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load assets",
		})
		return
	}

	enriched := h.enrichedAssets(assets)
	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"assets":  enriched,
		"summary": SummarizeAssets(enriched),
		"lineage": BuildEdges(enriched),
	})
}

func (h *CatalogHandler) servePatch(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		shared.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "metadata editing requires the postgres-backed control plane",
		})
		return
	}

	principal := anonymousPrincipal(h.authz, r)
	if !authz.Allowed(principal, authz.RoleEditor) {
		_ = h.appendAudit(audit.Event{
			ActorUserID:  principal.UserID,
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "update_metadata_annotations",
			Resource:     "unknown",
			Outcome:      "forbidden",
		})
		shared.WriteRoleError(w, string(authz.RoleEditor), string(principal.Role))
		return
	}

	var patch AssetAnnotationsPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid metadata patch payload",
		})
		return
	}

	if err := h.store.UpdateAnnotations(patch); err != nil {
		_ = h.appendAudit(audit.Event{
			ActorUserID:  principal.UserID,
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "update_metadata_annotations",
			Resource:     patch.AssetID,
			Outcome:      "failure",
			Details: map[string]any{
				"error": err.Error(),
			},
		})
		shared.WriteError(w, http.StatusBadRequest, "failed to update metadata annotations", err)
		return
	}

	assets, err := h.assetsForResponse()
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "failed to reload assets after update", err)
		return
	}
	enriched := h.enrichedAssets(assets)
	for _, asset := range enriched {
		if asset.ID == patch.AssetID {
			_ = h.appendAudit(audit.Event{
				ActorUserID:  principal.UserID,
				ActorSubject: principal.Subject,
				ActorRole:    string(principal.Role),
				Action:       "update_metadata_annotations",
				Resource:     patch.AssetID,
				Outcome:      "success",
			})
			shared.WriteJSON(w, http.StatusOK, map[string]any{
				"asset": asset,
			})
			return
		}
	}

	shared.WriteJSON(w, http.StatusNotFound, map[string]any{
		"error": "updated asset was not found after reload",
	})
}

func (h *CatalogHandler) assetsForResponse() ([]DataAsset, error) {
	if h.store != nil {
		return h.store.ListAssets()
	}
	return h.loader.LoadAssets()
}

func (h *CatalogHandler) enrichedAssets(assets []DataAsset) []DataAsset {
	h.catalog.ReplaceAssets(assets)
	enriched := h.catalog.ListAssets()
	for index := range enriched {
		enriched[index].FreshnessStatus = h.freshnessStatus(enriched[index])
	}
	return EnrichAssets(enriched)
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

func (h *CatalogHandler) appendAudit(event audit.Event) error {
	if h.audit == nil {
		return nil
	}
	return h.audit.Append(event)
}

func anonymousPrincipal(authService *authz.Service, r *http.Request) authz.Principal {
	if authService == nil {
		return authz.Principal{Subject: "anonymous", Role: authz.RoleAnonymous, AuthSource: "anonymous"}
	}
	return authService.ResolveRequest(r)
}
