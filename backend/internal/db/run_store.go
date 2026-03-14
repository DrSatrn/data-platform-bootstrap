// This file mirrors pipeline run state into PostgreSQL snapshot tables. The
// file-backed store remains the primary local execution store, while this
// mirror gives the platform a realistic durable repository immediately.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// RunStore mirrors run state into PostgreSQL.
type RunStore struct {
	conn *sql.DB
}

// NewRunStoreFromConn constructs a run store around an existing PostgreSQL
// connection pool.
func NewRunStoreFromConn(conn *sql.DB) *RunStore {
	return &RunStore{conn: conn}
}

// NewRunStore constructs a mirrored run store if the required tables exist.
func NewRunStore(ctx context.Context, dsn string) (*RunStore, error) {
	controlPlane, err := NewControlPlane(ctx, dsn)
	if err != nil {
		return nil, err
	}
	present, err := tableExists(ctx, controlPlane.Conn, "run_snapshots")
	if err != nil {
		return nil, err
	}
	if !present {
		return nil, fmt.Errorf("run_snapshots table is missing")
	}
	return controlPlane.RunStore, nil
}

// SavePipelineRun persists a run snapshot into the PostgreSQL snapshot table.
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

// ListPipelineRuns returns the latest run snapshots ordered by update time.
func (s *RunStore) ListPipelineRuns() ([]orchestration.PipelineRun, error) {
	rows, err := s.conn.QueryContext(context.Background(), `
		select payload
		from run_snapshots
		order by updated_at desc
	`)
	if err != nil {
		return nil, fmt.Errorf("list run snapshots: %w", err)
	}
	defer rows.Close()

	runs := []orchestration.PipelineRun{}
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan run snapshot: %w", err)
		}
		var run orchestration.PipelineRun
		if err := json.Unmarshal(payload, &run); err != nil {
			return nil, fmt.Errorf("decode run snapshot: %w", err)
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run snapshots: %w", err)
	}
	return runs, nil
}

// GetPipelineRun loads one run snapshot by identifier.
func (s *RunStore) GetPipelineRun(id string) (orchestration.PipelineRun, bool, error) {
	var payload []byte
	if err := s.conn.QueryRowContext(context.Background(), `
		select payload
		from run_snapshots
		where run_id = $1
	`, id).Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return orchestration.PipelineRun{}, false, nil
		}
		return orchestration.PipelineRun{}, false, fmt.Errorf("get run snapshot %s: %w", id, err)
	}

	var run orchestration.PipelineRun
	if err := json.Unmarshal(payload, &run); err != nil {
		return orchestration.PipelineRun{}, false, fmt.Errorf("decode run snapshot %s: %w", id, err)
	}
	return run, true, nil
}

// Close releases the underlying connection pool.
func (s *RunStore) Close() error {
	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}
