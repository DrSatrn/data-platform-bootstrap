-- This migration adds a persistent audit trail for privileged platform
-- actions. The table is intentionally append-only so the control plane can
-- retain operator history across restarts and deployments.

create table if not exists audit_events (
  id bigserial primary key,
  event_time timestamptz not null default now(),
  actor_subject text not null,
  actor_role text not null,
  action text not null,
  resource text not null,
  outcome text not null,
  details jsonb not null default '{}'::jsonb
);

create index if not exists idx_audit_events_event_time
  on audit_events (event_time desc);
