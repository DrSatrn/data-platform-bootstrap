-- This migration promotes metadata into a database-first runtime entity by
-- adding explicit annotation columns. Manifest-backed fields remain seeded for
-- structure, while operator edits now persist directly in PostgreSQL.

alter table data_assets
  add column if not exists annotation_owner_id text,
  add column if not exists annotation_description text,
  add column if not exists annotation_quality_check_refs jsonb,
  add column if not exists annotation_documentation_refs jsonb,
  add column if not exists annotation_updated_at timestamptz,
  add column if not exists manifest_synced_at timestamptz not null default now();

alter table asset_columns
  add column if not exists annotation_description text,
  add column if not exists annotation_updated_at timestamptz;
