// Package authz implements a lightweight role-based access layer for the local
// platform. It keeps the current product safe enough for self-hosted use
// without dragging in a full external identity provider before the core
// platform is ready.
package authz

import (
	"fmt"
	"net/http"
	"strings"
)

// Role identifies the effective permission level for a bearer token.
type Role string

const (
	RoleAnonymous Role = "anonymous"
	RoleViewer    Role = "viewer"
	RoleEditor    Role = "editor"
	RoleAdmin     Role = "admin"
)

// Principal is the resolved identity attached to one request.
type Principal struct {
	Subject string `json:"subject"`
	Role    Role   `json:"role"`
}

// Service resolves bearer tokens and exposes capability checks for handlers.
type Service struct {
	tokens map[string]Principal
}

// Session summarizes the current request identity and product capabilities.
type Session struct {
	Principal    Principal       `json:"principal"`
	Capabilities map[string]bool `json:"capabilities"`
}

// NewService builds a role resolver from configured access tokens plus the
// legacy admin token.
func NewService(adminToken, accessTokens string) (*Service, error) {
	service := &Service{tokens: map[string]Principal{}}
	if strings.TrimSpace(adminToken) != "" {
		service.tokens[strings.TrimSpace(adminToken)] = Principal{
			Subject: "local-admin",
			Role:    RoleAdmin,
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
		service.tokens[parts[0]] = Principal{
			Subject: parts[2],
			Role:    role,
		}
	}
	return service, nil
}

// ResolveRequest converts a bearer token into a principal. Missing or unknown
// tokens degrade to anonymous rather than exploding the request path.
func (s *Service) ResolveRequest(r *http.Request) Principal {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return Principal{Subject: "anonymous", Role: RoleAnonymous}
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
	if principal, ok := s.tokens[token]; ok {
		return principal
	}
	return Principal{Subject: "anonymous", Role: RoleAnonymous}
}

// Allowed reports whether the principal is allowed to perform an action at the
// requested minimum role.
func Allowed(principal Principal, minimum Role) bool {
	return roleRank(principal.Role) >= roleRank(minimum)
}

// SessionForRequest returns the resolved principal plus coarse UI capability
// flags so the frontend can adapt without duplicating role logic.
func (s *Service) SessionForRequest(r *http.Request) Session {
	principal := s.ResolveRequest(r)
	return Session{
		Principal: principal,
		Capabilities: map[string]bool{
			"view_platform":      true,
			"trigger_runs":       Allowed(principal, RoleEditor),
			"edit_dashboards":    Allowed(principal, RoleEditor),
			"run_admin_terminal": Allowed(principal, RoleAdmin),
		},
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
