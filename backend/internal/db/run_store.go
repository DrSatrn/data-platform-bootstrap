// This file mirrors pipeline run state into PostgreSQL snapshot tables. The
// file-backed store remains the primary local execution store, while this
// mirror gives the platform a realistic durable repository immediately.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// RunStore mirrors run state into PostgreSQL.
type RunStore struct {
	conn *sql.DB
}

// NewRunStore constructs a mirrored run store if the required tables exist.
func NewRunStore(ctx context.Context, dsn string) (*RunStore, error) {
	conn, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	var present bool
	if err := conn.QueryRowContext(ctx, `select to_regclass('public.run_snapshots') is not null`).Scan(&present); err != nil {
		return nil, fmt.Errorf("check run_snapshots table: %w", err)
	}
	if !present {
		return nil, fmt.Errorf("run_snapshots table is missing")
	}

	return &RunStore{conn: conn}, nil
}

// SavePipelineRun mirrors a run snapshot into the PostgreSQL snapshot table.
func (s *RunStore) SavePipelineRun(run orchestration.PipelineRun) error {
	ctx := context.Background()
	payload, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("encode run snapshot %s: %w", run.ID, err)
	}
	if _, err := s.conn.ExecContext(ctx, `
		insert into run_snapshots (run_id, pipeline_id, status, trigger_source, payload, updated_at)
		values ($1, $2, $3, $4, $5::jsonb, $6)
		on conflict (run_id) do update set
		  status = excluded.status,
		  trigger_source = excluded.trigger_source,
		  payload = excluded.payload,
		  updated_at = excluded.updated_at
	`,
		run.ID,
		run.PipelineID,
		string(run.Status),
		run.Trigger,
		string(payload),
		run.UpdatedAt,
	); err != nil {
		return fmt.Errorf("upsert run snapshot %s: %w", run.ID, err)
	}
	return nil
}

// ListPipelineRuns is intentionally unsupported for the mirror store because
// the file-backed primary store remains the read source during this phase.
func (s *RunStore) ListPipelineRuns() ([]orchestration.PipelineRun, error) {
	return nil, fmt.Errorf("postgres mirror store is write-focused and should not be used as the primary read store")
}

// GetPipelineRun is intentionally unsupported for the mirror store.
func (s *RunStore) GetPipelineRun(id string) (orchestration.PipelineRun, bool, error) {
	return orchestration.PipelineRun{}, false, fmt.Errorf("postgres mirror store is write-focused and should not be used as the primary read store")
}

// Close releases the underlying connection pool.
func (s *RunStore) Close() error {
	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}
