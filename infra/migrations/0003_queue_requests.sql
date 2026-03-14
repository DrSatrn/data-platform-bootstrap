-- This migration promotes PostgreSQL from an optional mirror to a practical
-- primary control-plane store for queued run requests. Rows remain available
-- for diagnostics after completion rather than disappearing immediately.

create table if not exists queue_requests (
  run_id text primary key,
  pipeline_id text not null,
  trigger_source text not null,
  requested_at timestamptz not null,
  status text not null,
  claim_token text not null default '',
  claimed_at timestamptz,
  completed_at timestamptz
);

create index if not exists idx_queue_requests_claimable
  on queue_requests (status, requested_at);
