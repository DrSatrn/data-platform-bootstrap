// These tests cover the lightweight RBAC layer so handler protections and UI
// capability flags stay aligned as the product evolves.
package authz

import (
	"net/http/httptest"
	"testing"
)

func TestResolveRequestAndCapabilities(t *testing.T) {
	service, err := NewService("legacy-admin", "viewer-token:viewer:alice,editor-token:editor:bob")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	request := httptest.NewRequest("GET", "/api/v1/session", nil)
	request.Header.Set("Authorization", "Bearer editor-token")
	session := service.SessionForRequest(request)
	if session.Principal.Subject != "bob" || session.Principal.Role != RoleEditor {
		t.Fatalf("unexpected principal: %#v", session.Principal)
	}
	if !session.Capabilities["edit_dashboards"] || session.Capabilities["run_admin_terminal"] {
		t.Fatalf("unexpected capabilities: %#v", session.Capabilities)
	}
	if !session.Capabilities["view_platform"] {
		t.Fatalf("expected editor to be able to view platform: %#v", session.Capabilities)
	}

	adminRequest := httptest.NewRequest("GET", "/api/v1/session", nil)
	adminRequest.Header.Set("Authorization", "Bearer legacy-admin")
	adminSession := service.SessionForRequest(adminRequest)
	if !adminSession.Capabilities["run_admin_terminal"] {
		t.Fatalf("expected admin capability: %#v", adminSession.Capabilities)
	}

	anonymousRequest := httptest.NewRequest("GET", "/api/v1/session", nil)
	anonymousSession := service.SessionForRequest(anonymousRequest)
	if anonymousSession.Capabilities["view_platform"] {
		t.Fatalf("expected anonymous session to be denied product access: %#v", anonymousSession.Capabilities)
	}
}
