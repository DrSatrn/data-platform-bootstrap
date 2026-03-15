// This file provides PostgreSQL-backed persistence for audit events so
// sensitive operator actions remain queryable after restarts.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streanor/data-platform/backend/internal/audit"
)

// AuditStore persists audit events in PostgreSQL.
type AuditStore struct {
	conn *sql.DB
}

// NewAuditStoreFromConn constructs an audit store from an existing pool.
func NewAuditStoreFromConn(conn *sql.DB) *AuditStore {
	return &AuditStore{conn: conn}
}

func (s *AuditStore) Append(event audit.Event) error {
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	details, err := json.Marshal(event.Details)
	if err != nil {
		return fmt.Errorf("encode audit details: %w", err)
	}
	_, err = s.conn.Exec(`
		insert into audit_events (event_time, actor_subject, actor_role, action, resource, outcome, details)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, event.Time, event.ActorSubject, event.ActorRole, event.Action, event.Resource, event.Outcome, details)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func (s *AuditStore) ListRecent(limit int) ([]audit.Event, error) {
	rows, err := s.conn.Query(`
		select event_time, actor_subject, actor_role, action, resource, outcome, details
		from audit_events
		order by event_time desc
		limit $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	events := []audit.Event{}
	for rows.Next() {
		var (
			event   audit.Event
			details []byte
		)
		if err := rows.Scan(&event.Time, &event.ActorSubject, &event.ActorRole, &event.Action, &event.Resource, &event.Outcome, &details); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		if len(details) > 0 {
			if err := json.Unmarshal(details, &event.Details); err != nil {
				return nil, fmt.Errorf("decode audit details: %w", err)
			}
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}
	return events, nil
}
