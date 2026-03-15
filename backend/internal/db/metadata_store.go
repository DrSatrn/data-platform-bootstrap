// This file persists the synchronized metadata catalog in PostgreSQL. The
// store keeps the platform's dataset registry durable while preserving the
// repo-managed manifests as the declarative source of truth.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/streanor/data-platform/backend/internal/metadata"
)

// MetadataStore persists catalog assets and columns in PostgreSQL.
type MetadataStore struct {
	conn *sql.DB
}

// NewMetadataStoreFromConn constructs a metadata store from an existing pool.
func NewMetadataStoreFromConn(conn *sql.DB) *MetadataStore {
	return &MetadataStore{conn: conn}
}

// SyncAssets upserts the full current manifest-backed asset projection.
func (s *MetadataStore) SyncAssets(assets []metadata.DataAsset) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin metadata sync: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`delete from asset_columns`); err != nil {
		return fmt.Errorf("clear asset columns: %w", err)
	}
	if _, err := tx.Exec(`delete from data_assets`); err != nil {
		return fmt.Errorf("clear data assets: %w", err)
	}

	for _, asset := range assets {
		sourceRefs, err := json.Marshal(asset.SourceRefs)
		if err != nil {
			return fmt.Errorf("encode source refs for %s: %w", asset.ID, err)
		}
		qualityRefs, err := json.Marshal(asset.QualityCheckRefs)
		if err != nil {
			return fmt.Errorf("encode quality refs for %s: %w", asset.ID, err)
		}
		docRefs, err := json.Marshal(asset.DocumentationRefs)
		if err != nil {
			return fmt.Errorf("encode doc refs for %s: %w", asset.ID, err)
		}
		if _, err := tx.Exec(`
			insert into data_assets (
				id, name, layer, owner_id, kind, description,
				source_refs, freshness_expected_within, freshness_warn_after,
				quality_check_refs, documentation_refs
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			asset.ID,
			asset.Name,
			asset.Layer,
			asset.Owner,
			asset.Kind,
			asset.Description,
			sourceRefs,
			asset.Freshness.ExpectedWithin,
			asset.Freshness.WarnAfter,
			qualityRefs,
			docRefs,
		); err != nil {
			return fmt.Errorf("insert data asset %s: %w", asset.ID, err)
		}

		for _, column := range asset.Columns {
			if _, err := tx.Exec(`
				insert into asset_columns (asset_id, column_name, column_type, description, is_pii)
				values ($1, $2, $3, $4, $5)
			`, asset.ID, column.Name, column.Type, column.Description, column.IsPII); err != nil {
				return fmt.Errorf("insert asset column %s.%s: %w", asset.ID, column.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit metadata sync: %w", err)
	}
	return nil
}

// ListAssets returns the persisted catalog projection with column metadata.
func (s *MetadataStore) ListAssets() ([]metadata.DataAsset, error) {
	rows, err := s.conn.Query(`
		select
			id, name, layer, owner_id, kind, description,
			source_refs, freshness_expected_within, freshness_warn_after,
			quality_check_refs, documentation_refs
		from data_assets
		order by layer, name
	`)
	if err != nil {
		return nil, fmt.Errorf("query data assets: %w", err)
	}
	defer rows.Close()

	assets := []metadata.DataAsset{}
	for rows.Next() {
		var (
			asset       metadata.DataAsset
			sourceRefs  []byte
			qualityRefs []byte
			docRefs     []byte
		)
		if err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.Layer,
			&asset.Owner,
			&asset.Kind,
			&asset.Description,
			&sourceRefs,
			&asset.Freshness.ExpectedWithin,
			&asset.Freshness.WarnAfter,
			&qualityRefs,
			&docRefs,
		); err != nil {
			return nil, fmt.Errorf("scan data asset: %w", err)
		}
		if err := json.Unmarshal(sourceRefs, &asset.SourceRefs); err != nil {
			return nil, fmt.Errorf("decode source refs for %s: %w", asset.ID, err)
		}
		if err := json.Unmarshal(qualityRefs, &asset.QualityCheckRefs); err != nil {
			return nil, fmt.Errorf("decode quality refs for %s: %w", asset.ID, err)
		}
		if err := json.Unmarshal(docRefs, &asset.DocumentationRefs); err != nil {
			return nil, fmt.Errorf("decode doc refs for %s: %w", asset.ID, err)
		}

		columnRows, err := s.conn.Query(`
			select column_name, column_type, description, is_pii
			from asset_columns
			where asset_id = $1
			order by column_name
		`, asset.ID)
		if err != nil {
			return nil, fmt.Errorf("query columns for %s: %w", asset.ID, err)
		}
		columns := []metadata.Column{}
		for columnRows.Next() {
			var column metadata.Column
			if err := columnRows.Scan(&column.Name, &column.Type, &column.Description, &column.IsPII); err != nil {
				columnRows.Close()
				return nil, fmt.Errorf("scan column for %s: %w", asset.ID, err)
			}
			columns = append(columns, column)
		}
		if err := columnRows.Err(); err != nil {
			columnRows.Close()
			return nil, fmt.Errorf("iterate columns for %s: %w", asset.ID, err)
		}
		columnRows.Close()
		asset.Columns = columns
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate data assets: %w", err)
	}
	return assets, nil
}
