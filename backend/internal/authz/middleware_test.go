// These tests cover the shared auth middleware so role-gated handlers fail in
// a predictable way rather than relying on ad hoc checks spread across the API.
package authz

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRoleRejectsAnonymousRequests(t *testing.T) {
	service, err := NewService("admin-token", "viewer-token:viewer:alice")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := RequireRole(service, RoleViewer, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/catalog", nil)
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", recorder.Code)
	}
}

func TestRequireRoleAllowsViewerRequests(t *testing.T) {
	service, err := NewService("admin-token", "viewer-token:viewer:alice")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := RequireRole(service, RoleViewer, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/catalog", nil)
	request.Header.Set("Authorization", "Bearer viewer-token")
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected success, got %d", recorder.Code)
	}
}
