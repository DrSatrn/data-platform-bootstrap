// These tests verify the native session and admin user-management endpoints so
// the auth flow remains runnable without relying on manual browser checks.
package authz

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/audit"
)

func TestSessionHandlerLoginAndLogout(t *testing.T) {
	repository := newRepositoryStub()
	service, err := NewService("bootstrap-token", "", repository, time.Hour)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CreateUser("operator", "Operator", RoleEditor, "secret-password"); err != nil {
		t.Fatalf("create user: %v", err)
	}

	handler := NewSessionHandler(service, audit.NewMemoryStore())

	body, _ := json.Marshal(map[string]string{"username": "operator", "password": "secret-password"})
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader(body))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginRequest)
	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected login success, got %d: %s", loginRecorder.Code, loginRecorder.Body.String())
	}

	var loginPayload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginRecorder.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("decode login payload: %v", err)
	}
	if loginPayload.Token == "" {
		t.Fatalf("expected session token")
	}

	logoutRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/session", nil)
	logoutRequest.Header.Set("Authorization", "Bearer "+loginPayload.Token)
	logoutRecorder := httptest.NewRecorder()
	handler.ServeHTTP(logoutRecorder, logoutRequest)
	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("expected logout success, got %d: %s", logoutRecorder.Code, logoutRecorder.Body.String())
	}
}

func TestUserHandlerRequiresAdmin(t *testing.T) {
	repository := newRepositoryStub()
	service, err := NewService("bootstrap-token", "viewer-token:viewer:alice", repository, time.Hour)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	handler := NewUserHandler(service, audit.NewMemoryStore())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	request.Header.Set("Authorization", "Bearer viewer-token")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected admin-only endpoint to reject viewer, got %d", recorder.Code)
	}
}
