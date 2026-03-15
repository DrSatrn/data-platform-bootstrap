// This file exposes the session and user-management endpoints for the native
// identity layer. The session endpoint handles login/logout, while the admin
// user endpoint gives bootstrap administrators a first-party way to manage
// accounts.
package authz

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/shared"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userCreateRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Password    string `json:"password"`
}

type userUpdateRequest struct {
	Username string `json:"username"`
	Action   string `json:"action"`
	Password string `json:"password,omitempty"`
	Active   *bool  `json:"active,omitempty"`
}

// SessionHandler serves login, logout, and current-session inspection.
type SessionHandler struct {
	service *Service
	audit   audit.Store
}

// NewSessionHandler constructs the session endpoint.
func NewSessionHandler(service *Service, auditStore audit.Store) http.Handler {
	return &SessionHandler{service: service, audit: auditStore}
}

func (h *SessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		shared.WriteJSON(w, http.StatusOK, h.service.SessionForRequest(r))
	case http.MethodPost:
		var payload loginRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid login payload"})
			return
		}
		result, err := h.service.LoginWithKey(loginRateLimitKey(r), payload.Username, payload.Password)
		if err != nil {
			_ = h.audit.Append(audit.Event{
				ActorSubject: payload.Username,
				ActorRole:    string(RoleAnonymous),
				Action:       "session_login",
				Resource:     payload.Username,
				Outcome:      "failure",
				Details:      map[string]any{"reason": sanitizedAuthMessage(err)},
			})
			switch {
			case errors.Is(err, ErrLoginRateLimited):
				shared.WriteError(w, http.StatusTooManyRequests, "too many failed login attempts; try again later", err)
			case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrUsernamePasswordRequired):
				shared.WriteError(w, http.StatusUnauthorized, sanitizedAuthMessage(err), err)
			case errors.Is(err, ErrNativeIdentityUnavailable):
				shared.WriteError(w, http.StatusServiceUnavailable, "native identity is unavailable; use the bootstrap admin token", err)
			default:
				shared.WriteError(w, http.StatusInternalServerError, "login failed", err)
			}
			return
		}
		_ = h.audit.Append(audit.Event{
			ActorUserID:  result.Session.Principal.UserID,
			ActorSubject: result.Session.Principal.Subject,
			ActorRole:    string(result.Session.Principal.Role),
			Action:       "session_login",
			Resource:     result.Session.Principal.Subject,
			Outcome:      "success",
			Details:      map[string]any{"auth_source": result.Session.Principal.AuthSource},
		})
		shared.WriteJSON(w, http.StatusOK, result)
	case http.MethodDelete:
		principal := h.service.ResolveRequest(r)
		if err := h.service.Logout(r); err != nil {
			shared.WriteError(w, http.StatusInternalServerError, "logout failed", err)
			return
		}
		_ = h.audit.Append(audit.Event{
			ActorUserID:  principal.UserID,
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "session_logout",
			Resource:     principal.Subject,
			Outcome:      "success",
			Details:      map[string]any{"auth_source": principal.AuthSource},
		})
		shared.WriteJSON(w, http.StatusOK, map[string]any{"status": "logged_out"})
	default:
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

// UserHandler exposes admin-only user management.
type UserHandler struct {
	service *Service
	audit   audit.Store
}

// NewUserHandler constructs the admin user-management endpoint.
func NewUserHandler(service *Service, auditStore audit.Store) http.Handler {
	return &UserHandler{service: service, audit: auditStore}
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	principal := h.service.ResolveRequest(r)
	if !Allowed(principal, RoleAdmin) {
		shared.WriteRoleError(w, string(RoleAdmin), string(principal.Role))
		return
	}

	switch r.Method {
	case http.MethodGet:
		users, err := h.service.ListUsers()
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, "failed to load users", err)
			return
		}
		shared.WriteJSON(w, http.StatusOK, map[string]any{"users": users})
	case http.MethodPost:
		var payload userCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid user payload"})
			return
		}
		role, err := parseRole(payload.Role)
		if err != nil {
			shared.WriteError(w, http.StatusBadRequest, "invalid user role", err)
			return
		}
		user, err := h.service.CreateUser(payload.Username, payload.DisplayName, role, payload.Password)
		if err != nil {
			shared.WriteError(w, statusForUserMutation(err), sanitizedUserMutationMessage(err), err)
			return
		}
		_ = h.audit.Append(audit.Event{
			ActorUserID:  principal.UserID,
			ActorSubject: principal.Subject,
			ActorRole:    string(principal.Role),
			Action:       "create_user",
			Resource:     user.Username,
			Outcome:      "success",
		})
		shared.WriteJSON(w, http.StatusCreated, map[string]any{"user": user})
	case http.MethodPatch:
		var payload userUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid user update payload"})
			return
		}
		switch payload.Action {
		case "reset_password":
			user, err := h.service.ResetPassword(payload.Username, payload.Password)
			if err != nil {
				shared.WriteError(w, statusForUserMutation(err), sanitizedUserMutationMessage(err), err)
				return
			}
			_ = h.audit.Append(audit.Event{
				ActorUserID:  principal.UserID,
				ActorSubject: principal.Subject,
				ActorRole:    string(principal.Role),
				Action:       "reset_user_password",
				Resource:     user.Username,
				Outcome:      "success",
			})
			shared.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
		case "set_active":
			if payload.Active == nil {
				shared.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "active must be provided"})
				return
			}
			user, err := h.service.SetUserActive(payload.Username, *payload.Active)
			if err != nil {
				shared.WriteError(w, statusForUserMutation(err), sanitizedUserMutationMessage(err), err)
				return
			}
			_ = h.audit.Append(audit.Event{
				ActorUserID:  principal.UserID,
				ActorSubject: principal.Subject,
				ActorRole:    string(principal.Role),
				Action:       "set_user_active",
				Resource:     user.Username,
				Outcome:      "success",
				Details:      map[string]any{"active": *payload.Active},
			})
			shared.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
		default:
			shared.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "unsupported user update action"})
		}
	default:
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func sanitizedAuthMessage(err error) string {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return "invalid username or password"
	case errors.Is(err, ErrUsernamePasswordRequired):
		return "username and password are required"
	case errors.Is(err, ErrLoginRateLimited):
		return "too many failed login attempts; try again later"
	case errors.Is(err, ErrNativeIdentityUnavailable):
		return "native identity is unavailable; use the bootstrap admin token"
	default:
		return "authentication failed"
	}
}

func statusForUserMutation(err error) int {
	if errors.Is(err, ErrNativeIdentityUnavailable) {
		return http.StatusServiceUnavailable
	}
	return http.StatusBadRequest
}

func sanitizedUserMutationMessage(err error) string {
	switch {
	case errors.Is(err, ErrNativeIdentityUnavailable):
		return "native identity store is unavailable"
	case errors.Is(err, ErrUsernamePasswordRequired):
		return "username and password are required"
	default:
		return "user update request is invalid"
	}
}
