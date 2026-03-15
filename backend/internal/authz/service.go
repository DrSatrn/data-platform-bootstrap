// Package authz implements the platform's native identity, session, and role
// enforcement layer. The current design keeps bootstrap simple for self-hosted
// installs while moving normal operator access onto a database-backed user and
// session model.
package authz

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Role identifies the effective permission level for a principal.
type Role string

const (
	RoleAnonymous Role = "anonymous"
	RoleViewer    Role = "viewer"
	RoleEditor    Role = "editor"
	RoleAdmin     Role = "admin"
)

const defaultSessionTTL = 24 * time.Hour

const (
	defaultBcryptCost        = 12
	defaultFailedLoginLimit  = 5
	defaultFailedLoginWindow = time.Minute
	defaultMaxActiveSessions = 5
)

var (
	ErrInvalidCredentials        = errors.New("invalid username or password")
	ErrNativeIdentityUnavailable = errors.New("native identity store is unavailable")
	ErrUsernamePasswordRequired  = errors.New("username and password are required")
	ErrLoginRateLimited          = errors.New("too many failed login attempts")
)

// Principal is the resolved identity attached to one request.
type Principal struct {
	UserID     string `json:"user_id,omitempty"`
	Subject    string `json:"subject"`
	Role       Role   `json:"role"`
	AuthSource string `json:"auth_source,omitempty"`
}

// User describes one database-backed platform identity.
type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Role        Role      `json:"role"`
	IsActive    bool      `json:"is_active"`
	IsBootstrap bool      `json:"is_bootstrap"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StoredUser extends the visible user shape with password material needed by
// the repository-backed authentication flow.
type StoredUser struct {
	User
	PasswordHash string
	PasswordSalt string
}

// SessionRecord captures one persisted login session.
type SessionRecord struct {
	ID         string
	UserID     string
	TokenHash  string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

// Repository defines the durable identity and session operations needed by the
// auth service. PostgreSQL-backed implementations satisfy this today.
type Repository interface {
	EnsureBootstrapUser(username, displayName string) (User, error)
	ListUsers() ([]User, error)
	ListStoredUsers() ([]StoredUser, error)
	GetUserByUsername(username string) (StoredUser, bool, error)
	GetUserByID(id string) (User, bool, error)
	CreateUser(user StoredUser) (User, error)
	UpdateUserPassword(username, passwordHash, passwordSalt string) (User, error)
	SetUserActive(username string, active bool) (User, error)
	CreateSession(record SessionRecord) error
	GetSessionByTokenHash(tokenHash string) (SessionRecord, User, bool, error)
	TouchSession(sessionID string, seenAt time.Time) error
	RevokeSession(sessionID string, revokedAt time.Time) error
	DeleteExpiredSessions(now time.Time) error
	TrimActiveSessions(userID string, keepNewest int, now time.Time) error
}

// Session summarizes the current request identity and product capabilities.
type Session struct {
	Principal    Principal       `json:"principal"`
	Capabilities map[string]bool `json:"capabilities"`
}

// LoginResult returns the created session token plus the resolved session
// payload so browser clients can update immediately after login.
type LoginResult struct {
	Token   string  `json:"token"`
	Session Session `json:"session"`
}

// Service resolves bootstrap tokens, legacy static tokens, and database-backed
// sessions. Legacy access tokens are still accepted as a compatibility bridge,
// but the normal path is now username/password login plus session creation.
type Service struct {
	bootstrapToken string
	bootstrapUser  User
	legacyTokens   map[string]Principal
	repository     Repository
	sessionTTL     time.Duration
	bcryptCost     int
	loginLimiter   *loginRateLimiter
	maxSessions    int
}

// NewService constructs the identity/session service.
func NewService(adminToken, accessTokens string, repository Repository, sessionTTL time.Duration) (*Service, error) {
	if sessionTTL <= 0 {
		sessionTTL = defaultSessionTTL
	}

	service := &Service{
		bootstrapToken: strings.TrimSpace(adminToken),
		legacyTokens:   map[string]Principal{},
		repository:     repository,
		sessionTTL:     sessionTTL,
		bcryptCost:     configuredInt("PLATFORM_BCRYPT_COST", defaultBcryptCost),
		loginLimiter:   newLoginRateLimiter(defaultFailedLoginLimit, defaultFailedLoginWindow),
		maxSessions:    configuredInt("PLATFORM_MAX_ACTIVE_SESSIONS", defaultMaxActiveSessions),
	}

	if service.bootstrapToken != "" && repository != nil {
		bootstrapUser, err := repository.EnsureBootstrapUser("bootstrap-admin", "Bootstrap Admin")
		if err != nil {
			return nil, fmt.Errorf("ensure bootstrap user: %w", err)
		}
		service.bootstrapUser = bootstrapUser
	}
	if repository != nil {
		if err := repository.DeleteExpiredSessions(time.Now().UTC()); err != nil {
			slog.Default().Warn("failed to delete expired sessions during auth startup", slog.String("error", err.Error()))
		}
	}

	if strings.TrimSpace(accessTokens) == "" {
		return service, nil
	}
	for _, rawEntry := range strings.Split(accessTokens, ",") {
		entry := strings.TrimSpace(rawEntry)
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid access token entry %q; expected token:role:subject", entry)
		}
		role, err := parseRole(parts[1])
		if err != nil {
			return nil, err
		}
		service.legacyTokens[parts[0]] = Principal{
			Subject:    parts[2],
			Role:       role,
			AuthSource: "legacy_token",
		}
	}
	return service, nil
}

// ResolveRequest converts the current bearer token into a principal. Unknown or
// expired tokens degrade to anonymous rather than exploding the request path.
func (s *Service) ResolveRequest(r *http.Request) Principal {
	token := bearerToken(r)
	if token == "" {
		return anonymousPrincipal()
	}
	if s.bootstrapToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.bootstrapToken)) == 1 {
		return s.bootstrapPrincipal()
	}
	if principal, ok := s.legacyTokens[token]; ok {
		return principal
	}
	if s.repository == nil {
		return anonymousPrincipal()
	}

	tokenHash := hashToken(token)
	session, user, ok, err := s.repository.GetSessionByTokenHash(tokenHash)
	if err != nil || !ok {
		return anonymousPrincipal()
	}
	if !user.IsActive {
		return anonymousPrincipal()
	}
	now := time.Now().UTC()
	if session.RevokedAt != nil || !session.ExpiresAt.After(now) {
		return anonymousPrincipal()
	}
	if err := s.repository.TouchSession(session.ID, now); err != nil {
		slog.Default().Warn("failed to touch session", slog.String("session_id", session.ID), slog.String("error", err.Error()))
	}
	return Principal{
		UserID:     user.ID,
		Subject:    user.Username,
		Role:       user.Role,
		AuthSource: "session",
	}
}

// SessionForRequest returns the resolved principal plus coarse UI capability
// flags so the frontend can adapt without duplicating role logic.
func (s *Service) SessionForRequest(r *http.Request) Session {
	principal := s.ResolveRequest(r)
	return Session{
		Principal: principal,
		Capabilities: map[string]bool{
			"view_platform":      Allowed(principal, RoleViewer),
			"trigger_runs":       Allowed(principal, RoleEditor),
			"edit_metadata":      Allowed(principal, RoleEditor),
			"edit_dashboards":    Allowed(principal, RoleEditor),
			"run_admin_terminal": Allowed(principal, RoleAdmin),
			"manage_users":       Allowed(principal, RoleAdmin),
		},
	}
}

// Login authenticates a username/password against the native identity store
// and returns a bearer session token.
func (s *Service) Login(username, password string) (LoginResult, error) {
	return s.LoginWithKey("local", username, password)
}

// LoginWithKey authenticates a username/password and applies rate limiting
// keyed by the caller's network identity.
func (s *Service) LoginWithKey(clientKey, username, password string) (LoginResult, error) {
	if s.repository == nil {
		return LoginResult{}, fmt.Errorf("%w; use the bootstrap admin token", ErrNativeIdentityUnavailable)
	}
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return LoginResult{}, ErrUsernamePasswordRequired
	}
	now := time.Now().UTC()
	if blocked, retryAfter := s.loginLimiter.Allow(clientKey, now); !blocked {
		return LoginResult{}, fmt.Errorf("%w; retry after %s", ErrLoginRateLimited, retryAfter.Round(time.Second))
	}
	storedUser, ok, err := s.repository.GetUserByUsername(username)
	if err != nil {
		return LoginResult{}, err
	}
	if !ok || !storedUser.IsActive {
		s.loginLimiter.RecordFailure(clientKey, now)
		return LoginResult{}, ErrInvalidCredentials
	}
	if !verifyPassword(password, storedUser.PasswordSalt, storedUser.PasswordHash) {
		s.loginLimiter.RecordFailure(clientKey, now)
		return LoginResult{}, ErrInvalidCredentials
	}
	s.loginLimiter.Reset(clientKey)
	if err := s.repository.DeleteExpiredSessions(now); err != nil {
		slog.Default().Warn("failed to delete expired sessions before login", slog.String("error", err.Error()))
	}
	if s.maxSessions > 0 {
		if err := s.repository.TrimActiveSessions(storedUser.ID, s.maxSessions-1, now); err != nil {
			return LoginResult{}, err
		}
	}

	token, err := randomToken()
	if err != nil {
		return LoginResult{}, err
	}
	record := SessionRecord{
		ID:         "session_" + uuid.NewString(),
		UserID:     storedUser.ID,
		TokenHash:  hashToken(token),
		CreatedAt:  now,
		LastSeenAt: now,
		ExpiresAt:  now.Add(s.sessionTTL),
	}
	if err := s.repository.CreateSession(record); err != nil {
		return LoginResult{}, err
	}

	session := Session{
		Principal: Principal{
			UserID:     storedUser.ID,
			Subject:    storedUser.Username,
			Role:       storedUser.Role,
			AuthSource: "session",
		},
		Capabilities: capabilitiesForRole(storedUser.Role),
	}
	return LoginResult{Token: token, Session: session}, nil
}

// Logout revokes the current database-backed session token. Bootstrap and
// legacy tokens simply return success so the browser can clear local state.
func (s *Service) Logout(r *http.Request) error {
	token := bearerToken(r)
	if token == "" {
		return nil
	}
	if s.bootstrapToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.bootstrapToken)) == 1 {
		return nil
	}
	if _, ok := s.legacyTokens[token]; ok {
		return nil
	}
	if s.repository == nil {
		return nil
	}
	session, _, ok, err := s.repository.GetSessionByTokenHash(hashToken(token))
	if err != nil || !ok {
		return err
	}
	return s.repository.RevokeSession(session.ID, time.Now().UTC())
}

// ListUsers returns the visible platform identities.
func (s *Service) ListUsers() ([]User, error) {
	if s.repository == nil {
		return []User{}, nil
	}
	return s.repository.ListUsers()
}

// CreateUser creates a new native platform user.
func (s *Service) CreateUser(username, displayName string, role Role, password string) (User, error) {
	if s.repository == nil {
		return User{}, ErrNativeIdentityUnavailable
	}
	if strings.TrimSpace(username) == "" || strings.TrimSpace(displayName) == "" || password == "" {
		return User{}, fmt.Errorf("username, display name, role, and password are required")
	}
	normalizedRole, err := parseRole(string(role))
	if err != nil {
		return User{}, err
	}
	passwordSalt, passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	return s.repository.CreateUser(StoredUser{
		User: User{
			ID:          "user_" + uuid.NewString(),
			Username:    strings.TrimSpace(username),
			DisplayName: strings.TrimSpace(displayName),
			Role:        normalizedRole,
			IsActive:    true,
			IsBootstrap: false,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		PasswordHash: passwordHash,
		PasswordSalt: passwordSalt,
	})
}

// ResetPassword rotates one user's password.
func (s *Service) ResetPassword(username, password string) (User, error) {
	if s.repository == nil {
		return User{}, ErrNativeIdentityUnavailable
	}
	if strings.TrimSpace(username) == "" || password == "" {
		return User{}, ErrUsernamePasswordRequired
	}
	passwordSalt, passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	return s.repository.UpdateUserPassword(strings.TrimSpace(username), passwordHash, passwordSalt)
}

// SetUserActive flips one user's active state.
func (s *Service) SetUserActive(username string, active bool) (User, error) {
	if s.repository == nil {
		return User{}, ErrNativeIdentityUnavailable
	}
	if strings.TrimSpace(username) == "" {
		return User{}, fmt.Errorf("username is required")
	}
	return s.repository.SetUserActive(strings.TrimSpace(username), active)
}

// Allowed reports whether the principal is allowed to perform an action at the
// requested minimum role.
func Allowed(principal Principal, minimum Role) bool {
	return roleRank(principal.Role) >= roleRank(minimum)
}

func capabilitiesForRole(role Role) map[string]bool {
	return map[string]bool{
		"view_platform":      roleRank(role) >= roleRank(RoleViewer),
		"trigger_runs":       roleRank(role) >= roleRank(RoleEditor),
		"edit_metadata":      roleRank(role) >= roleRank(RoleEditor),
		"edit_dashboards":    roleRank(role) >= roleRank(RoleEditor),
		"run_admin_terminal": roleRank(role) >= roleRank(RoleAdmin),
		"manage_users":       roleRank(role) >= roleRank(RoleAdmin),
	}
}

func parseRole(value string) (Role, error) {
	role := Role(strings.TrimSpace(value))
	switch role {
	case RoleViewer, RoleEditor, RoleAdmin:
		return role, nil
	default:
		return "", fmt.Errorf("unsupported role %q", value)
	}
}

func roleRank(role Role) int {
	switch role {
	case RoleViewer:
		return 1
	case RoleEditor:
		return 2
	case RoleAdmin:
		return 3
	default:
		return 0
	}
}

func (s *Service) bootstrapPrincipal() Principal {
	if s.bootstrapUser.ID != "" {
		return Principal{
			UserID:     s.bootstrapUser.ID,
			Subject:    s.bootstrapUser.Username,
			Role:       RoleAdmin,
			AuthSource: "bootstrap_token",
		}
	}
	return Principal{
		Subject:    "bootstrap-admin",
		Role:       RoleAdmin,
		AuthSource: "bootstrap_token",
	}
}

func anonymousPrincipal() Principal {
	return Principal{Subject: "anonymous", Role: RoleAnonymous, AuthSource: "anonymous"}
}

func bearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return "session_" + hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func hashPassword(password string) (salt string, hashed string, err error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), configuredInt("PLATFORM_BCRYPT_COST", defaultBcryptCost))
	if err != nil {
		return "", "", fmt.Errorf("hash password with bcrypt: %w", err)
	}
	return "", string(hashedBytes), nil
}

func verifyPassword(password, salt, expected string) bool {
	if expected == "" {
		return false
	}
	if isBcryptHash(expected) {
		return bcrypt.CompareHashAndPassword([]byte(expected), []byte(password)) == nil
	}
	if salt == "" {
		return false
	}
	actual := derivePasswordHash(password, salt)
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func derivePasswordHash(password, salt string) string {
	seed := []byte(salt + ":" + password)
	sum := sha256.Sum256(seed)
	current := sum[:]
	for range 120000 {
		next := sha256.Sum256(current)
		current = next[:]
	}
	return hex.EncodeToString(current)
}

func configuredInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func isBcryptHash(value string) bool {
	return strings.HasPrefix(value, "$2a$") || strings.HasPrefix(value, "$2b$") || strings.HasPrefix(value, "$2y$")
}
