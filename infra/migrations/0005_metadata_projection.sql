-- This migration promotes the metadata catalog from a read-only manifest view
-- into a persisted control-plane projection. The repo-managed manifests remain
-- the declarative source, but PostgreSQL now stores a synchronized catalog
-- snapshot for durability and future governance workflows.

alter table data_assets
  add column if not exists source_refs jsonb not null default '[]'::jsonb,
  add column if not exists freshness_expected_within text not null default '',
  add column if not exists freshness_warn_after text not null default '',
  add column if not exists quality_check_refs jsonb not null default '[]'::jsonb,
  add column if not exists documentation_refs jsonb not null default '[]'::jsonb;
