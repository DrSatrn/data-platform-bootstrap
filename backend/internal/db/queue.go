// This file implements a PostgreSQL-backed queue for pipeline run requests. It
// is intentionally pragmatic: queued rows are claimed transactionally and
// completed rows remain recorded with status metadata for diagnostics.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

// RunQueue persists queued run requests in PostgreSQL.
type RunQueue struct {
	conn *sql.DB
}

// NewRunQueueFromConn constructs a PostgreSQL-backed run queue around an
// existing connection pool.
func NewRunQueueFromConn(conn *sql.DB) *RunQueue {
	return &RunQueue{conn: conn}
}

// Enqueue records a queued run request for worker pickup.
func (q *RunQueue) Enqueue(request orchestration.RunRequest) error {
	_, err := q.conn.ExecContext(context.Background(), `
		insert into queue_requests (run_id, pipeline_id, trigger_source, requested_at, status, claim_token)
		values ($1, $2, $3, $4, 'queued', '')
		on conflict (run_id) do update set
		  pipeline_id = excluded.pipeline_id,
		  trigger_source = excluded.trigger_source,
		  requested_at = excluded.requested_at,
		  status = 'queued',
		  claim_token = '',
		  claimed_at = null,
		  completed_at = null
	`,
		request.RunID,
		request.PipelineID,
		request.Trigger,
		request.RequestedAt,
	)
	if err != nil {
		return fmt.Errorf("enqueue run %s: %w", request.RunID, err)
	}
	return nil
}

// ClaimNext returns the next queue row, preferring previously active rows so a
// restarted worker can recover abandoned work.
func (q *RunQueue) ClaimNext() (*orchestration.ClaimedRequest, error) {
	tx, err := q.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("begin queue claim transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	claimToken := fmt.Sprintf("claim_%d", time.Now().UTC().UnixNano())
	row := tx.QueryRowContext(context.Background(), `
		with next as (
			select run_id
			from queue_requests
			where status in ('active', 'queued')
			order by case when status = 'active' then 0 else 1 end, requested_at asc
			limit 1
			for update skip locked
		)
		update queue_requests as q
		set status = 'active',
		    claimed_at = now(),
		    claim_token = $1
		from next
		where q.run_id = next.run_id
		returning q.run_id, q.pipeline_id, q.trigger_source, q.requested_at
	`, claimToken)

	var request orchestration.RunRequest
	if err := row.Scan(&request.RunID, &request.PipelineID, &request.Trigger, &request.RequestedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("claim next run request: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit queue claim transaction: %w", err)
	}
	return &orchestration.ClaimedRequest{
		Request: request,
		Receipt: claimToken,
	}, nil
}

// Complete marks a claimed request as completed so it leaves the active queue
// while remaining inspectable in PostgreSQL.
func (q *RunQueue) Complete(claimed *orchestration.ClaimedRequest) error {
	if claimed == nil {
		return nil
	}
	result, err := q.conn.ExecContext(context.Background(), `
		update queue_requests
		set status = 'completed',
		    completed_at = now()
		where run_id = $1 and claim_token = $2
	`, claimed.Request.RunID, claimed.Receipt)
	if err != nil {
		return fmt.Errorf("complete queue request %s: %w", claimed.Request.RunID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check completed rows for %s: %w", claimed.Request.RunID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("queue request %s was not updated; claim token may be stale", claimed.Request.RunID)
	}
	return nil
}

// ListRequests returns a point-in-time snapshot of queue rows for backup and
// operational export workflows.
func (q *RunQueue) ListRequests() ([]orchestration.QueueSnapshot, error) {
	rows, err := q.conn.QueryContext(context.Background(), `
		select run_id, pipeline_id, trigger_source, status, requested_at, claimed_at, completed_at
		from queue_requests
		order by requested_at asc
	`)
	if err != nil {
		return nil, fmt.Errorf("list queue requests: %w", err)
	}
	defer rows.Close()

	requests := []orchestration.QueueSnapshot{}
	for rows.Next() {
		var snapshot orchestration.QueueSnapshot
		if err := rows.Scan(
			&snapshot.RunID,
			&snapshot.PipelineID,
			&snapshot.Trigger,
			&snapshot.Status,
			&snapshot.RequestedAt,
			&snapshot.ClaimedAt,
			&snapshot.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan queue request: %w", err)
		}
		requests = append(requests, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate queue requests: %w", err)
	}
	return requests, nil
}
