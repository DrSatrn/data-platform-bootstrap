-- This migration adds pragmatic snapshot tables for the local control-plane
-- mirror. They let the platform persist durable run and artifact snapshots in
-- PostgreSQL without waiting for the full normalized repository layer.

create table if not exists run_snapshots (
  run_id text primary key,
  pipeline_id text not null,
  status text not null,
  trigger_source text not null,
  payload jsonb not null,
  updated_at timestamptz not null default now()
);

create table if not exists artifact_snapshots (
  run_id text not null,
  relative_path text not null,
  content_type text not null,
  size_bytes bigint not null,
  recorded_at timestamptz not null default now(),
  primary key (run_id, relative_path)
);
