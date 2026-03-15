// This file provides PostgreSQL-backed persistence for native platform users
// and login sessions. It lets the control plane own identity without relying
// on an external auth provider.
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/streanor/data-platform/backend/internal/authz"
)

// IdentityStore persists users and sessions in PostgreSQL.
type IdentityStore struct {
	conn *sql.DB
}

// NewIdentityStoreFromConn constructs an identity store from an existing pool.
func NewIdentityStoreFromConn(conn *sql.DB) *IdentityStore {
	return &IdentityStore{conn: conn}
}

func (s *IdentityStore) EnsureBootstrapUser(username, displayName string) (authz.User, error) {
	now := time.Now().UTC()
	user := authz.User{}
	if err := s.conn.QueryRow(`
		insert into platform_users (
			id, username, display_name, role, password_hash, password_salt,
			is_active, is_bootstrap, created_at, updated_at
		)
		values ($1, $2, $3, 'admin', '', '', true, true, $4, $4)
		on conflict (username) do update set
			display_name = excluded.display_name,
			role = 'admin',
			is_active = true,
			is_bootstrap = true,
			updated_at = excluded.updated_at
		returning id, username, display_name, role, is_active, is_bootstrap, created_at, updated_at
	`, "user_bootstrap_admin", username, displayName, now).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Role,
		&user.IsActive,
		&user.IsBootstrap,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return authz.User{}, fmt.Errorf("ensure bootstrap user: %w", err)
	}
	return user, nil
}

func (s *IdentityStore) ListUsers() ([]authz.User, error) {
	storedUsers, err := s.ListStoredUsers()
	if err != nil {
		return nil, err
	}
	users := make([]authz.User, 0, len(storedUsers))
	for _, user := range storedUsers {
		users = append(users, user.User)
	}
	return users, nil
}

func (s *IdentityStore) ListStoredUsers() ([]authz.StoredUser, error) {
	rows, err := s.conn.Query(`
		select
			id, username, display_name, role, is_active, is_bootstrap, created_at, updated_at,
			password_hash, password_salt
		from platform_users
		order by username
	`)
	if err != nil {
		return nil, fmt.Errorf("query platform users: %w", err)
	}
	defer rows.Close()

	users := []authz.StoredUser{}
	for rows.Next() {
		var user authz.StoredUser
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.DisplayName,
			&user.Role,
			&user.IsActive,
			&user.IsBootstrap,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.PasswordHash,
			&user.PasswordSalt,
		); err != nil {
			return nil, fmt.Errorf("scan platform user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platform users: %w", err)
	}
	return users, nil
}

func (s *IdentityStore) GetUserByUsername(username string) (authz.StoredUser, bool, error) {
	var user authz.StoredUser
	if err := s.conn.QueryRow(`
		select
			id, username, display_name, role, is_active, is_bootstrap, created_at, updated_at,
			password_hash, password_salt
		from platform_users
		where username = $1
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Role,
		&user.IsActive,
		&user.IsBootstrap,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.PasswordHash,
		&user.PasswordSalt,
	); err != nil {
		if err == sql.ErrNoRows {
			return authz.StoredUser{}, false, nil
		}
		return authz.StoredUser{}, false, fmt.Errorf("get user %s: %w", username, err)
	}
	return user, true, nil
}

func (s *IdentityStore) GetUserByID(id string) (authz.User, bool, error) {
	var user authz.User
	if err := s.conn.QueryRow(`
		select id, username, display_name, role, is_active, is_bootstrap, created_at, updated_at
		from platform_users
		where id = $1
	`, id).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Role,
		&user.IsActive,
		&user.IsBootstrap,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return authz.User{}, false, nil
		}
		return authz.User{}, false, fmt.Errorf("get user by id %s: %w", id, err)
	}
	return user, true, nil
}

func (s *IdentityStore) CreateUser(user authz.StoredUser) (authz.User, error) {
	created := authz.User{}
	if err := s.conn.QueryRow(`
		insert into platform_users (
			id, username, display_name, role, password_hash, password_salt,
			is_active, is_bootstrap, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, false, $8, $9)
		returning id, username, display_name, role, is_active, is_bootstrap, created_at, updated_at
	`,
		user.ID,
		user.Username,
		user.DisplayName,
		string(user.Role),
		user.PasswordHash,
		user.PasswordSalt,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&created.ID,
		&created.Username,
		&created.DisplayName,
		&created.Role,
		&created.IsActive,
		&created.IsBootstrap,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return authz.User{}, fmt.Errorf("create user %s: %w", user.Username, err)
	}
	return created, nil
}

func (s *IdentityStore) UpdateUserPassword(username, passwordHash, passwordSalt string) (authz.User, error) {
	updated := authz.User{}
	if err := s.conn.QueryRow(`
		update platform_users
		set password_hash = $2,
		    password_salt = $3,
		    updated_at = now()
		where username = $1
		returning id, username, display_name, role, is_active, is_bootstrap, created_at, updated_at
	`, username, passwordHash, passwordSalt).Scan(
		&updated.ID,
		&updated.Username,
		&updated.DisplayName,
		&updated.Role,
		&updated.IsActive,
		&updated.IsBootstrap,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return authz.User{}, fmt.Errorf("user %s does not exist", username)
		}
		return authz.User{}, fmt.Errorf("update password for %s: %w", username, err)
	}
	return updated, nil
}

func (s *IdentityStore) SetUserActive(username string, active bool) (authz.User, error) {
	updated := authz.User{}
	if err := s.conn.QueryRow(`
		update platform_users
		set is_active = $2,
		    updated_at = now()
		where username = $1
		returning id, username, display_name, role, is_active, is_bootstrap, created_at, updated_at
	`, username, active).Scan(
		&updated.ID,
		&updated.Username,
		&updated.DisplayName,
		&updated.Role,
		&updated.IsActive,
		&updated.IsBootstrap,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return authz.User{}, fmt.Errorf("user %s does not exist", username)
		}
		return authz.User{}, fmt.Errorf("set active=%t for %s: %w", active, username, err)
	}
	return updated, nil
}

func (s *IdentityStore) CreateSession(record authz.SessionRecord) error {
	_, err := s.conn.Exec(`
		insert into platform_sessions (id, user_id, token_hash, created_at, last_seen_at, expires_at, revoked_at)
		values ($1, $2, $3, $4, $5, $6, $7)
	`,
		record.ID,
		record.UserID,
		record.TokenHash,
		record.CreatedAt,
		record.LastSeenAt,
		record.ExpiresAt,
		record.RevokedAt,
	)
	if err != nil {
		return fmt.Errorf("create session for user %s: %w", record.UserID, err)
	}
	return nil
}

func (s *IdentityStore) GetSessionByTokenHash(tokenHash string) (authz.SessionRecord, authz.User, bool, error) {
	var (
		session authz.SessionRecord
		user    authz.User
	)
	if err := s.conn.QueryRow(`
		select
			s.id, s.user_id, s.token_hash, s.created_at, s.last_seen_at, s.expires_at, s.revoked_at,
			u.id, u.username, u.display_name, u.role, u.is_active, u.is_bootstrap, u.created_at, u.updated_at
		from platform_sessions s
		join platform_users u on u.id = s.user_id
		where s.token_hash = $1
	`, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Role,
		&user.IsActive,
		&user.IsBootstrap,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return authz.SessionRecord{}, authz.User{}, false, nil
		}
		return authz.SessionRecord{}, authz.User{}, false, fmt.Errorf("get session by token hash: %w", err)
	}
	return session, user, true, nil
}

func (s *IdentityStore) TouchSession(sessionID string, seenAt time.Time) error {
	_, err := s.conn.Exec(`update platform_sessions set last_seen_at = $2 where id = $1`, sessionID, seenAt)
	if err != nil {
		return fmt.Errorf("touch session %s: %w", sessionID, err)
	}
	return nil
}

func (s *IdentityStore) RevokeSession(sessionID string, revokedAt time.Time) error {
	_, err := s.conn.Exec(`update platform_sessions set revoked_at = $2 where id = $1`, sessionID, revokedAt)
	if err != nil {
		return fmt.Errorf("revoke session %s: %w", sessionID, err)
	}
	return nil
}
