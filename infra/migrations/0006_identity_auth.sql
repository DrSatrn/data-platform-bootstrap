-- This migration introduces a native identity and session layer for the
-- platform. It keeps self-hosted bootstrap simple while moving day-to-day
-- access control away from static environment-provided bearer tokens.

create table if not exists platform_users (
  id text primary key,
  username text not null unique,
  display_name text not null,
  role text not null,
  password_hash text not null default '',
  password_salt text not null default '',
  is_active boolean not null default true,
  is_bootstrap boolean not null default false,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_platform_users_active
  on platform_users (is_active, username);

create table if not exists platform_sessions (
  id text primary key,
  user_id text not null references platform_users (id) on delete cascade,
  token_hash text not null unique,
  created_at timestamptz not null default now(),
  last_seen_at timestamptz not null default now(),
  expires_at timestamptz not null,
  revoked_at timestamptz
);

create index if not exists idx_platform_sessions_lookup
  on platform_sessions (token_hash);

create index if not exists idx_platform_sessions_active
  on platform_sessions (expires_at, revoked_at);

alter table audit_events
  add column if not exists actor_user_id text not null default '';
