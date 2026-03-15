// These tests cover the native identity and compatibility auth paths so
// handler protections and UI capability flags stay aligned as the platform
// transitions away from static environment tokens.
package authz

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"testing"
	"time"
)

type repositoryStub struct {
	usersByUsername map[string]StoredUser
	usersByID       map[string]User
	sessionsByHash  map[string]sessionWithUser
	touchErr        error
	expiredSweeps   int
}

type sessionWithUser struct {
	session SessionRecord
	user    User
}

func newRepositoryStub() *repositoryStub {
	return &repositoryStub{
		usersByUsername: map[string]StoredUser{},
		usersByID:       map[string]User{},
		sessionsByHash:  map[string]sessionWithUser{},
	}
}

func (s *repositoryStub) EnsureBootstrapUser(username, displayName string) (User, error) {
	user := User{
		ID:          "user_bootstrap_admin",
		Username:    username,
		DisplayName: displayName,
		Role:        RoleAdmin,
		IsActive:    true,
		IsBootstrap: true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	s.usersByUsername[username] = StoredUser{User: user}
	s.usersByID[user.ID] = user
	return user, nil
}

func (s *repositoryStub) ListUsers() ([]User, error) {
	users := make([]User, 0, len(s.usersByID))
	for _, user := range s.usersByID {
		users = append(users, user)
	}
	return users, nil
}

func (s *repositoryStub) ListStoredUsers() ([]StoredUser, error) {
	users := make([]StoredUser, 0, len(s.usersByUsername))
	for _, user := range s.usersByUsername {
		users = append(users, user)
	}
	return users, nil
}

func (s *repositoryStub) GetUserByUsername(username string) (StoredUser, bool, error) {
	user, ok := s.usersByUsername[username]
	return user, ok, nil
}

func (s *repositoryStub) GetUserByID(id string) (User, bool, error) {
	user, ok := s.usersByID[id]
	return user, ok, nil
}

func (s *repositoryStub) CreateUser(user StoredUser) (User, error) {
	s.usersByUsername[user.Username] = user
	s.usersByID[user.ID] = user.User
	return user.User, nil
}

func (s *repositoryStub) UpdateUserPassword(username, passwordHash, passwordSalt string) (User, error) {
	user := s.usersByUsername[username]
	user.PasswordHash = passwordHash
	user.PasswordSalt = passwordSalt
	user.UpdatedAt = time.Now().UTC()
	s.usersByUsername[username] = user
	s.usersByID[user.ID] = user.User
	return user.User, nil
}

func (s *repositoryStub) SetUserActive(username string, active bool) (User, error) {
	user := s.usersByUsername[username]
	user.IsActive = active
	user.UpdatedAt = time.Now().UTC()
	s.usersByUsername[username] = user
	s.usersByID[user.ID] = user.User
	return user.User, nil
}

func (s *repositoryStub) CreateSession(record SessionRecord) error {
	user := s.usersByID[record.UserID]
	s.sessionsByHash[record.TokenHash] = sessionWithUser{session: record, user: user}
	return nil
}

func (s *repositoryStub) GetSessionByTokenHash(tokenHash string) (SessionRecord, User, bool, error) {
	record, ok := s.sessionsByHash[tokenHash]
	if !ok {
		return SessionRecord{}, User{}, false, nil
	}
	return record.session, record.user, true, nil
}

func (s *repositoryStub) TouchSession(sessionID string, seenAt time.Time) error {
	if s.touchErr != nil {
		return s.touchErr
	}
	for hash, value := range s.sessionsByHash {
		if value.session.ID == sessionID {
			value.session.LastSeenAt = seenAt
			s.sessionsByHash[hash] = value
		}
	}
	return nil
}

func (s *repositoryStub) RevokeSession(sessionID string, revokedAt time.Time) error {
	for hash, value := range s.sessionsByHash {
		if value.session.ID == sessionID {
			value.session.RevokedAt = &revokedAt
			s.sessionsByHash[hash] = value
		}
	}
	return nil
}

func (s *repositoryStub) DeleteExpiredSessions(now time.Time) error {
	s.expiredSweeps++
	for hash, value := range s.sessionsByHash {
		if value.session.ExpiresAt.Before(now) {
			delete(s.sessionsByHash, hash)
		}
	}
	return nil
}

func (s *repositoryStub) TrimActiveSessions(userID string, keepNewest int, now time.Time) error {
	if keepNewest < 0 {
		keepNewest = 0
	}
	type candidate struct {
		hash  string
		entry sessionWithUser
	}
	candidates := []candidate{}
	for hash, value := range s.sessionsByHash {
		if value.session.UserID != userID {
			continue
		}
		if value.session.RevokedAt != nil || !value.session.ExpiresAt.After(now) {
			continue
		}
		candidates = append(candidates, candidate{hash: hash, entry: value})
	}
	for len(candidates) > keepNewest {
		oldestIndex := 0
		for index := range candidates {
			if candidates[index].entry.session.LastSeenAt.Before(candidates[oldestIndex].entry.session.LastSeenAt) {
				oldestIndex = index
			}
		}
		delete(s.sessionsByHash, candidates[oldestIndex].hash)
		candidates = append(candidates[:oldestIndex], candidates[oldestIndex+1:]...)
	}
	return nil
}

func TestResolveRequestAndCapabilities(t *testing.T) {
	repository := newRepositoryStub()
	service, err := NewService("legacy-admin", "viewer-token:viewer:alice,editor-token:editor:bob", repository, time.Hour)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	request := httptest.NewRequest("GET", "/api/v1/session", nil)
	request.Header.Set("Authorization", "Bearer editor-token")
	session := service.SessionForRequest(request)
	if session.Principal.Subject != "bob" || session.Principal.Role != RoleEditor {
		t.Fatalf("unexpected principal: %#v", session.Principal)
	}
	if !session.Capabilities["edit_dashboards"] || !session.Capabilities["edit_metadata"] || session.Capabilities["run_admin_terminal"] {
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
	if adminSession.Principal.UserID == "" {
		t.Fatalf("expected bootstrap admin to resolve to a stored user id")
	}

	anonymousRequest := httptest.NewRequest("GET", "/api/v1/session", nil)
	anonymousSession := service.SessionForRequest(anonymousRequest)
	if anonymousSession.Capabilities["view_platform"] {
		t.Fatalf("expected anonymous session to be denied product access: %#v", anonymousSession.Capabilities)
	}
}

func TestLoginAndLogoutWithNativeSessions(t *testing.T) {
	repository := newRepositoryStub()
	service, err := NewService("bootstrap-token", "", repository, time.Hour)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	user, err := service.CreateUser("analyst", "Analyst User", RoleViewer, "secret-password")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	login, err := service.Login("analyst", "secret-password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if login.Token == "" {
		t.Fatalf("expected session token")
	}
	if login.Session.Principal.UserID != user.ID {
		t.Fatalf("expected session user id %s, got %#v", user.ID, login.Session.Principal)
	}

	request := httptest.NewRequest("GET", "/api/v1/session", nil)
	request.Header.Set("Authorization", "Bearer "+login.Token)
	resolved := service.SessionForRequest(request)
	if resolved.Principal.Subject != "analyst" || resolved.Principal.Role != RoleViewer {
		t.Fatalf("unexpected resolved session %#v", resolved.Principal)
	}

	logoutRequest := httptest.NewRequest("DELETE", "/api/v1/session", nil)
	logoutRequest.Header.Set("Authorization", "Bearer "+login.Token)
	if err := service.Logout(logoutRequest); err != nil {
		t.Fatalf("logout: %v", err)
	}

	postLogout := service.SessionForRequest(request)
	if postLogout.Principal.Role != RoleAnonymous {
		t.Fatalf("expected revoked session to resolve anonymous, got %#v", postLogout.Principal)
	}
}

func TestLoginSupportsLegacyPasswordHashes(t *testing.T) {
	repository := newRepositoryStub()
	service, err := NewService("bootstrap-token", "", repository, time.Hour)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	salt := "legacy-salt"
	user := User{
		ID:          "user_legacy",
		Username:    "legacy-user",
		DisplayName: "Legacy User",
		Role:        RoleViewer,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	repository.usersByUsername[user.Username] = StoredUser{
		User:         user,
		PasswordSalt: salt,
		PasswordHash: deriveLegacyPasswordHash("legacy-password", salt),
	}
	repository.usersByID[user.ID] = user

	login, err := service.Login("legacy-user", "legacy-password")
	if err != nil {
		t.Fatalf("legacy login: %v", err)
	}
	if login.Token == "" {
		t.Fatalf("expected token for legacy login")
	}
}

func TestLoginRateLimitsRepeatedFailuresByClientKey(t *testing.T) {
	repository := newRepositoryStub()
	service, err := NewService("bootstrap-token", "", repository, time.Hour)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	for attempt := 0; attempt < defaultFailedLoginLimit; attempt++ {
		_, err := service.LoginWithKey("10.0.0.1", "unknown", "bad-password")
		if err == nil || err != ErrInvalidCredentials {
			t.Fatalf("expected invalid credentials on attempt %d, got %v", attempt+1, err)
		}
	}
	_, err = service.LoginWithKey("10.0.0.1", "unknown", "bad-password")
	if err == nil || err == ErrInvalidCredentials || err.Error() == "" {
		t.Fatalf("expected rate limit error, got %v", err)
	}
}

func TestLoginTrimsActiveSessionsToMaximum(t *testing.T) {
	repository := newRepositoryStub()
	service, err := NewService("bootstrap-token", "", repository, time.Hour)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CreateUser("operator", "Operator", RoleEditor, "secret-password"); err != nil {
		t.Fatalf("create user: %v", err)
	}

	for attempt := 0; attempt < defaultMaxActiveSessions+2; attempt++ {
		if _, err := service.LoginWithKey("10.0.0.2", "operator", "secret-password"); err != nil {
			t.Fatalf("login attempt %d: %v", attempt+1, err)
		}
	}

	active := 0
	now := time.Now().UTC()
	for _, value := range repository.sessionsByHash {
		if value.user.Username == "operator" && value.session.RevokedAt == nil && value.session.ExpiresAt.After(now) {
			active++
		}
	}
	if active != defaultMaxActiveSessions {
		t.Fatalf("expected %d active sessions, got %d", defaultMaxActiveSessions, active)
	}
}

func TestNewServiceSweepsExpiredSessionsAtStartup(t *testing.T) {
	repository := newRepositoryStub()
	now := time.Now().UTC()
	repository.sessionsByHash["expired"] = sessionWithUser{
		session: SessionRecord{ID: "expired", UserID: "user_bootstrap_admin", TokenHash: "expired", ExpiresAt: now.Add(-time.Minute)},
		user:    User{ID: "user_bootstrap_admin", Username: "bootstrap-admin", Role: RoleAdmin, IsActive: true},
	}

	if _, err := NewService("bootstrap-token", "", repository, time.Hour); err != nil {
		t.Fatalf("new service: %v", err)
	}
	if repository.expiredSweeps == 0 {
		t.Fatalf("expected startup expired-session sweep")
	}
	if _, ok := repository.sessionsByHash["expired"]; ok {
		t.Fatalf("expected expired session to be removed during startup sweep")
	}
}

func deriveLegacyPasswordHash(password, salt string) string {
	seed := []byte(salt + ":" + password)
	sum := sha256.Sum256(seed)
	current := sum[:]
	for range 120000 {
		next := sha256.Sum256(current)
		current = next[:]
	}
	return hex.EncodeToString(current)
}
