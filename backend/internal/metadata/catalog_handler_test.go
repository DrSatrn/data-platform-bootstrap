// These tests verify that the catalog handler behaves like a database-first
// runtime surface when a metadata store is present and that editor-driven
// annotations persist through the handler contract.
package metadata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/authz"
)

type assetLoaderStub struct {
	assets []DataAsset
}

func (s assetLoaderStub) LoadAssets() ([]DataAsset, error) {
	return s.assets, nil
}

type metadataStoreStub struct {
	assets       []DataAsset
	lastPatch    AssetAnnotationsPatch
	updateCalled bool
}

func (s *metadataStoreStub) SeedAssets([]DataAsset) error {
	return nil
}

func (s *metadataStoreStub) ListAssets() ([]DataAsset, error) {
	out := make([]DataAsset, len(s.assets))
	copy(out, s.assets)
	return out, nil
}

func (s *metadataStoreStub) UpdateAnnotations(patch AssetAnnotationsPatch) error {
	s.lastPatch = patch
	s.updateCalled = true
	for index := range s.assets {
		if s.assets[index].ID != patch.AssetID {
			continue
		}
		if patch.Owner != nil {
			s.assets[index].Owner = *patch.Owner
		}
		if patch.Description != nil {
			s.assets[index].Description = *patch.Description
		}
		if patch.DocumentationRefs != nil {
			s.assets[index].DocumentationRefs = append([]string{}, (*patch.DocumentationRefs)...)
		}
		if patch.QualityCheckRefs != nil {
			s.assets[index].QualityCheckRefs = append([]string{}, (*patch.QualityCheckRefs)...)
		}
		for _, columnPatch := range patch.ColumnDescriptions {
			for columnIndex := range s.assets[index].Columns {
				if s.assets[index].Columns[columnIndex].Name == columnPatch.Name && columnPatch.Description != nil {
					s.assets[index].Columns[columnIndex].Description = *columnPatch.Description
				}
			}
		}
	}
	return nil
}

func TestCatalogHandlerPrefersStoreOverLoader(t *testing.T) {
	store := &metadataStoreStub{
		assets: []DataAsset{{
			ID:          "mart_from_store",
			Name:        "Store Asset",
			Layer:       "mart",
			Owner:       "platform-team",
			Kind:        "table",
			Description: "served from the database-backed store",
		}},
	}
	handler := NewCatalogHandler(
		assetLoaderStub{assets: []DataAsset{{ID: "manifest_only", Name: "Manifest Asset", Layer: "raw", Owner: "manifest", Kind: "table", Description: "loader fallback"}}},
		NewCatalog(),
		t.TempDir(),
		store,
		nil,
		nil,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/catalog", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "manifest_only") {
		t.Fatalf("expected database-backed catalog response, got loader asset payload: %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "mart_from_store") {
		t.Fatalf("expected store asset in response: %s", recorder.Body.String())
	}
}

func TestCatalogHandlerPatchUpdatesAnnotations(t *testing.T) {
	store := &metadataStoreStub{
		assets: []DataAsset{{
			ID:                "mart_budget_vs_actual",
			Name:              "Budget vs Actual",
			Layer:             "mart",
			Owner:             "finance-team",
			Kind:              "table",
			Description:       "Original description",
			DocumentationRefs: []string{"docs/original.md"},
			QualityCheckRefs:  []string{"quality_original"},
			Columns: []Column{
				{Name: "month", Type: "text", Description: "Original month"},
			},
		}},
	}
	authService, err := authz.NewService("", "editor-token:editor:alice", nil, 0)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	handler := NewCatalogHandler(
		assetLoaderStub{},
		NewCatalog(),
		t.TempDir(),
		store,
		authService,
		audit.NewMemoryStore(),
	)

	body := `{
		"asset_id": "mart_budget_vs_actual",
		"owner": "platform-governance",
		"description": "Operator override",
		"documentation_refs": ["docs/runtime.md"],
		"quality_check_refs": ["quality_runtime"],
		"column_descriptions": [{"name": "month", "description": "Runtime month documentation"}]
	}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/catalog", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer editor-token")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !store.updateCalled {
		t.Fatalf("expected metadata update to be persisted")
	}
	if store.assets[0].Owner != "platform-governance" || store.assets[0].Description != "Operator override" {
		t.Fatalf("expected asset annotations to update, got %+v", store.assets[0])
	}
	if store.assets[0].Columns[0].Description != "Runtime month documentation" {
		t.Fatalf("expected column annotation to update, got %+v", store.assets[0].Columns[0])
	}

	var payload struct {
		Asset DataAsset `json:"asset"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Asset.Owner != "platform-governance" {
		t.Fatalf("expected updated asset in response, got %+v", payload.Asset)
	}
}
