// This file groups the PostgreSQL-backed control-plane repositories used when
// the normalized database path is available. The runtime still falls back to
// the local filesystem path when PostgreSQL or migrations are unavailable.
package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/streanor/data-platform/backend/internal/reporting"
)

// ControlPlane bundles the PostgreSQL-backed repositories required by the API,
// scheduler, worker, and artifact metadata layer.
type ControlPlane struct {
	Conn        *sql.DB
	RunStore    *RunStore
	RunQueue    *RunQueue
	ArtifactIdx *ArtifactIndex
	Dashboards  reporting.Store
}

// NewControlPlane opens PostgreSQL and verifies the required control-plane
// tables exist before exposing the repositories to the runtime.
func NewControlPlane(ctx context.Context, dsn string) (*ControlPlane, error) {
	conn, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	requiredTables := []string{"run_snapshots", "queue_requests", "artifact_snapshots", "dashboards"}
	for _, tableName := range requiredTables {
		present, err := tableExists(ctx, conn, tableName)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		if !present {
			_ = conn.Close()
			return nil, fmt.Errorf("%s table is missing", tableName)
		}
	}

	return &ControlPlane{
		Conn:        conn,
		RunStore:    NewRunStoreFromConn(conn),
		RunQueue:    NewRunQueueFromConn(conn),
		ArtifactIdx: NewArtifactIndexFromConn(conn),
		Dashboards:  NewDashboardStoreFromConn(conn),
	}, nil
}
