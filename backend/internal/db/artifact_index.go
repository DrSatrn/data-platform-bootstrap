// This file stores lightweight artifact metadata in PostgreSQL so the control
// plane can list run outputs without scanning the filesystem every time.
package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/streanor/data-platform/backend/internal/storage"
)

// ArtifactIndex persists run artifact metadata in PostgreSQL.
type ArtifactIndex struct {
	conn *sql.DB
}

// NewArtifactIndexFromConn constructs an artifact metadata repository around
// an existing PostgreSQL connection pool.
func NewArtifactIndexFromConn(conn *sql.DB) *ArtifactIndex {
	return &ArtifactIndex{conn: conn}
}

// RecordArtifact upserts one run artifact metadata row.
func (i *ArtifactIndex) RecordArtifact(artifact storage.Artifact) error {
	_, err := i.conn.ExecContext(context.Background(), `
		insert into artifact_snapshots (run_id, relative_path, content_type, size_bytes, recorded_at)
		values ($1, $2, $3, $4, $5)
		on conflict (run_id, relative_path) do update set
		  content_type = excluded.content_type,
		  size_bytes = excluded.size_bytes,
		  recorded_at = excluded.recorded_at
	`,
		artifact.RunID,
		artifact.RelativePath,
		artifact.ContentType,
		artifact.SizeBytes,
		artifact.ModifiedAt,
	)
	if err != nil {
		return fmt.Errorf("record artifact %s/%s: %w", artifact.RunID, artifact.RelativePath, err)
	}
	return nil
}

// ListRunArtifacts returns artifact metadata rows ordered by relative path.
func (i *ArtifactIndex) ListRunArtifacts(runID string) ([]storage.Artifact, error) {
	rows, err := i.conn.QueryContext(context.Background(), `
		select relative_path, content_type, size_bytes, recorded_at
		from artifact_snapshots
		where run_id = $1
		order by relative_path asc
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("list artifacts for %s: %w", runID, err)
	}
	defer rows.Close()

	artifacts := []storage.Artifact{}
	for rows.Next() {
		var artifact storage.Artifact
		artifact.RunID = runID
		if err := rows.Scan(&artifact.RelativePath, &artifact.ContentType, &artifact.SizeBytes, &artifact.ModifiedAt); err != nil {
			return nil, fmt.Errorf("scan artifact for %s: %w", runID, err)
		}
		artifacts = append(artifacts, artifact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate artifacts for %s: %w", runID, err)
	}
	return artifacts, nil
}
