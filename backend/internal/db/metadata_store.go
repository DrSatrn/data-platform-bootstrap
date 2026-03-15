// This file persists the metadata catalog in PostgreSQL. Manifest assets seed
// the structural catalog shape, while operator-driven annotations are stored
// directly in the database and treated as the runtime source of truth.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

// SeedAssets upserts the current manifest-backed asset structure without
// replacing operator-managed annotations.
func (s *MetadataStore) SeedAssets(assets []metadata.DataAsset) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin metadata seed: %w", err)
	}
	defer tx.Rollback()

	seenAssets := make([]string, 0, len(assets))
	for _, asset := range assets {
		seenAssets = append(seenAssets, asset.ID)
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
				quality_check_refs, documentation_refs, manifest_synced_at
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
			on conflict (id) do update set
				name = excluded.name,
				layer = excluded.layer,
				owner_id = excluded.owner_id,
				kind = excluded.kind,
				description = excluded.description,
				source_refs = excluded.source_refs,
				freshness_expected_within = excluded.freshness_expected_within,
				freshness_warn_after = excluded.freshness_warn_after,
				quality_check_refs = excluded.quality_check_refs,
				documentation_refs = excluded.documentation_refs,
				manifest_synced_at = now()
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
			return fmt.Errorf("upsert data asset %s: %w", asset.ID, err)
		}

		seenColumns := make([]string, 0, len(asset.Columns))
		for _, column := range asset.Columns {
			seenColumns = append(seenColumns, column.Name)
			if _, err := tx.Exec(`
				insert into asset_columns (asset_id, column_name, column_type, description, is_pii)
				values ($1, $2, $3, $4, $5)
				on conflict (asset_id, column_name) do update set
					column_type = excluded.column_type,
					description = excluded.description,
					is_pii = excluded.is_pii
			`, asset.ID, column.Name, column.Type, column.Description, column.IsPII); err != nil {
				return fmt.Errorf("upsert asset column %s.%s: %w", asset.ID, column.Name, err)
			}
		}

		if err := deleteMissingColumns(tx, asset.ID, seenColumns); err != nil {
			return fmt.Errorf("delete removed columns for %s: %w", asset.ID, err)
		}
	}

	if err := deleteMissingAssets(tx, seenAssets); err != nil {
		return fmt.Errorf("delete removed assets: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit metadata seed: %w", err)
	}
	return nil
}

// UpdateAnnotations persists operator-managed metadata directly into the
// database. This is the mutable runtime path used by the UI.
func (s *MetadataStore) UpdateAnnotations(patch metadata.AssetAnnotationsPatch) error {
	if strings.TrimSpace(patch.AssetID) == "" {
		return fmt.Errorf("asset_id is required")
	}

	tx, err := s.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin metadata annotation update: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		update data_assets
		set
			annotation_owner_id = coalesce($2, annotation_owner_id),
			annotation_description = coalesce($3, annotation_description),
			annotation_quality_check_refs = coalesce($4, annotation_quality_check_refs),
			annotation_documentation_refs = coalesce($5, annotation_documentation_refs),
			annotation_updated_at = $6
		where id = $1
	`,
		patch.AssetID,
		nullableString(patch.Owner),
		nullableString(patch.Description),
		marshalOptionalStrings(patch.QualityCheckRefs),
		marshalOptionalStrings(patch.DocumentationRefs),
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update metadata asset %s: %w", patch.AssetID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read metadata asset update count for %s: %w", patch.AssetID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("metadata asset %s was not found", patch.AssetID)
	}

	for _, columnPatch := range patch.ColumnDescriptions {
		if strings.TrimSpace(columnPatch.Name) == "" {
			return fmt.Errorf("column description updates require a column name")
		}
		columnResult, err := tx.Exec(`
			update asset_columns
			set
				annotation_description = coalesce($3, annotation_description),
				annotation_updated_at = $4
			where asset_id = $1 and column_name = $2
		`, patch.AssetID, columnPatch.Name, nullableString(columnPatch.Description), time.Now().UTC())
		if err != nil {
			return fmt.Errorf("update metadata column %s.%s: %w", patch.AssetID, columnPatch.Name, err)
		}
		columnRows, err := columnResult.RowsAffected()
		if err != nil {
			return fmt.Errorf("read metadata column update count for %s.%s: %w", patch.AssetID, columnPatch.Name, err)
		}
		if columnRows == 0 {
			return fmt.Errorf("metadata column %s.%s was not found", patch.AssetID, columnPatch.Name)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit metadata annotation update: %w", err)
	}
	return nil
}

func deleteMissingAssets(tx *sql.Tx, keep []string) error {
	if len(keep) == 0 {
		if _, err := tx.Exec(`delete from asset_columns`); err != nil {
			return err
		}
		_, err := tx.Exec(`delete from data_assets`)
		return err
	}

	query, args := notInClause("delete from asset_columns where asset_id not in (%s)", 1, keep)
	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}
	query, args = notInClause("delete from data_assets where id not in (%s)", 1, keep)
	_, err := tx.Exec(query, args...)
	return err
}

func deleteMissingColumns(tx *sql.Tx, assetID string, keep []string) error {
	if len(keep) == 0 {
		_, err := tx.Exec(`delete from asset_columns where asset_id = $1`, assetID)
		return err
	}
	query, args := notInClause("delete from asset_columns where asset_id = $1 and column_name not in (%s)", 2, keep)
	args = append([]any{assetID}, args...)
	_, err := tx.Exec(query, args...)
	return err
}

func notInClause(format string, startAt int, values []string) (string, []any) {
	placeholders := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for index, value := range values {
		placeholders = append(placeholders, fmt.Sprintf("$%d", startAt+index))
		args = append(args, value)
	}
	return fmt.Sprintf(format, strings.Join(placeholders, ", ")), args
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func marshalOptionalStrings(values *[]string) any {
	if values == nil {
		return nil
	}
	bytes, err := json.Marshal(*values)
	if err != nil {
		return []byte("[]")
	}
	return bytes
}

// ListAssets returns the persisted catalog projection with column metadata.
func (s *MetadataStore) ListAssets() ([]metadata.DataAsset, error) {
	rows, err := s.conn.Query(`
		select
			id,
			name,
			layer,
			coalesce(annotation_owner_id, owner_id) as effective_owner_id,
			kind,
			coalesce(annotation_description, description) as effective_description,
			source_refs,
			freshness_expected_within,
			freshness_warn_after,
			coalesce(annotation_quality_check_refs, quality_check_refs) as effective_quality_refs,
			coalesce(annotation_documentation_refs, documentation_refs) as effective_doc_refs
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
			select
				column_name,
				column_type,
				coalesce(annotation_description, description) as effective_description,
				is_pii
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
