-- This migration defines the first control-plane schema for the platform. The
-- tables are grouped by bounded context so the data model remains readable and
-- future migrations can evolve each subsystem intentionally.

create table if not exists pipelines (
  id text primary key,
  name text not null,
  description text not null,
  owner_id text not null,
  schedule_cron text not null,
  schedule_timezone text not null,
  created_at timestamptz not null default now()
);

create table if not exists jobs (
  id text primary key,
  pipeline_id text not null references pipelines (id),
  name text not null,
  job_type text not null,
  retries integer not null default 0,
  command text not null default '',
  transform_ref text not null default ''
);

create table if not exists pipeline_runs (
  id text primary key,
  pipeline_id text not null references pipelines (id),
  status text not null,
  trigger_source text not null,
  error_message text not null default '',
  started_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  finished_at timestamptz
);

create table if not exists job_runs (
  id text primary key,
  pipeline_run_id text not null references pipeline_runs (id),
  job_id text not null references jobs (id),
  status text not null,
  attempts integer not null default 0,
  error_message text not null default '',
  started_at timestamptz,
  finished_at timestamptz
);

create table if not exists run_events (
  id bigserial primary key,
  pipeline_run_id text not null references pipeline_runs (id),
  event_time timestamptz not null default now(),
  level text not null,
  message text not null,
  fields jsonb not null default '{}'::jsonb
);

create table if not exists data_assets (
  id text primary key,
  name text not null,
  layer text not null,
  owner_id text not null,
  kind text not null,
  description text not null
);

create table if not exists asset_columns (
  asset_id text not null references data_assets (id),
  column_name text not null,
  column_type text not null,
  description text not null,
  is_pii boolean not null default false,
  primary key (asset_id, column_name)
);

create table if not exists owners (
  id text primary key,
  display_name text not null,
  email text not null,
  team text not null
);

create table if not exists dashboards (
  id text primary key,
  name text not null,
  description text not null,
  definition jsonb not null default '{}'::jsonb
);
