# Runtime Wiring Manual

This document is the practical wiring map for the running system. Read it when
you need to understand which services exist, how they communicate, where state
lives, which ports matter, and which code paths own each responsibility.

## Why This Document Exists

The project now has enough moving parts that a high-level architecture summary
is no longer sufficient. This manual is intended to answer operational and
development questions such as:

- which process is responsible for what
- what storage each process reads and writes
- how a pipeline run flows through the system
- which runtime defaults are safe for localhost
- where to add new behavior without breaking the current boundaries

## Runtime Process Map

### `platform-api`

Purpose:

- serves the control-plane HTTP API
- exposes metadata, analytics, reporting, quality, health, audit, logs, and
  admin-terminal, session, and user-management endpoints
- accepts manual pipeline trigger requests

Listens on:

- `PLATFORM_HTTP_ADDR`
- default local binding: `:8080`
- Compose host binding: `127.0.0.1:8080:8080`

Key code:

- [runtime.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/app/runtime.go)
- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/orchestration/handler.go)
- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/metadata/handler.go)
- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/analytics/handler.go)
- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/reporting/handler.go)

Reads:

- manifests
- dashboard definitions
- local materialized data
- control-plane state from PostgreSQL when available, otherwise filesystem
- native users and sessions from PostgreSQL when available, otherwise the
  bootstrap admin token only

Writes:

- queued run requests
- run snapshots
- audit events
- platform users and session rows when the native identity store is available
- dashboard changes written directly to PostgreSQL when the preferred control
  plane is available, otherwise to the fallback filesystem store
- widget layout coordinates and spans are part of the saved dashboard contract,
  not frontend-only local state
- metadata seeds projected from manifests at startup and on scheduler refresh,
  plus runtime metadata annotation edits written directly to PostgreSQL when
  the preferred control plane is available
- system overview summaries derived from runs, queue state, and backup inventory
- cached dataset profile snapshots under `data/profiles/` when the catalog UI
  requests runtime profiling

### `platform-scheduler`

Purpose:

- wakes up on a tick
- loads pipeline manifests
- determines which scheduled pipelines are due
- enqueues scheduled runs into the control-plane queue

Key code:

- [service.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/scheduler/service.go)
- [runtime.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/app/runtime.go)

Reads:

- pipeline manifests
- run history
- queue state

Writes:

- queued scheduled runs
- scheduler-visible run history events through the orchestration service

### `platform-worker`

Purpose:

- polls the queue
- claims the next runnable run request
- executes jobs in dependency order
- materializes raw, mart, metric, and quality artifacts
- launches bounded Python subprocess tasks where manifests declare them

Key code:

- [worker.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/execution/worker.go)
- [runner.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/execution/runner.go)
- [engine.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/transforms/engine.go)

Reads:

- queued run requests
- pipeline manifests
- sample or landed data
- version-controlled SQL
- repo-managed Python task scripts

Writes:

- run status transitions
- run events
- materialized local data files
- run-scoped artifact copies
- artifact metadata index rows when PostgreSQL is enabled

### `platform-web`

Purpose:

- renders the operator and analyst UI
- talks to the API through HTTP
- stores a bearer token in browser local storage

Local dev mode:

- Vite dev server

Packaged mode:

- repo-owned static server with API proxy

Key code:

- [App.tsx](/Users/streanor/Documents/Playground/data-platform/web/src/app/App.tsx)
- [server.mjs](/Users/streanor/Documents/Playground/data-platform/web/server.mjs)

### `postgres`

Purpose:

- preferred normalized control-plane backend
- durable store for runs, queue state, artifact metadata, dashboards, audit,
  and projected metadata

Important note:

- PostgreSQL is preferred, not mandatory
- if it is unavailable, the platform falls back to local-first filesystem
  stores for core behavior

### DuckDB

Purpose:

- analytical execution engine
- materializes raw landing tables, marts, metrics, and quality query results

Important note:

- DuckDB is embedded and accessed through the Go SQL driver
- it is not deployed as a separate service

## Port Map

- API: `127.0.0.1:8080`
- Web: `127.0.0.1:3000`
- Postgres: internal Compose network only by default

Safe default:

- no non-loopback API or web binding in the tracked Compose file
- Postgres is intentionally not published externally

## State Map

### Filesystem state

By default, runtime state is rooted under `var/` in the repo for normal local
use, or under `/tmp` for smoke runs.

Important paths:

- `var/data/control_plane`
- `var/data/control_plane/scheduler_status.json`
- `var/data/raw`
- `var/data/mart`
- `var/data/metrics`
- `var/data/quality`
- `var/artifacts/runs/<run_id>`
- `var/duckdb/platform.duckdb`
- `var/backups`
- `var/benchmarks`

### PostgreSQL state

Current normalized tables include:

- `run_snapshots`
- `queue_requests`
- `artifact_snapshots`
- `dashboards`
- `audit_events`
- `data_assets`
- `asset_columns`
- `platform_users`
- `platform_sessions`

These are the current bridge between the local-first runtime and a more
enterprise-style control plane.

## Authoritative State By Subsystem

This is the runtime contract the docs and System page should agree on.

### When PostgreSQL is enabled

- runs:
  source of truth is PostgreSQL `run_snapshots`
  `updated_at` is touched on each run save by the preferred store
  filesystem remains a write mirror and fallback bootstrap path
- queue:
  source of truth is PostgreSQL `queue_requests`
  `claimed_at` and `completed_at` are written by the queue repository at claim
  and completion time
  filesystem queue is fallback only
- scheduler heartbeat:
  source of truth is the scheduler heartbeat file under
  `var/data/control_plane/scheduler_status.json`
  `/api/v1/system/overview` projects that heartbeat into `scheduler_summary`
  so operators and the benchmark suite can assert scheduler freshness
- artifacts:
  source of truth for bytes is the filesystem artifact root
  PostgreSQL stores artifact metadata/index rows only
- dashboards:
  source of truth is PostgreSQL `dashboards`
  repo-managed dashboard YAML is initial seed material only
  filesystem dashboard storage is fallback-only when PostgreSQL is unavailable
- audit:
  source of truth is PostgreSQL `audit_events`
  filesystem audit log is a mirror and fallback
- metadata:
  source of truth is PostgreSQL `data_assets` and `asset_columns`
  manifests seed structural catalog fields on startup and scheduler refresh
  runtime owner/description/docs/quality/column-description edits persist
  directly to PostgreSQL annotation columns
  manifests are only used directly when PostgreSQL is unavailable
- identity:
  source of truth is PostgreSQL `platform_users` and `platform_sessions`
  `PLATFORM_ADMIN_TOKEN` remains a bootstrap override and recovery path

### When PostgreSQL is not enabled

- runs, queue, dashboards, audit, and artifacts all operate from filesystem-backed stores
- metadata reads directly from repo-managed manifests
- identity falls back to bootstrap-admin-token-only access because the native
  session store is unavailable without PostgreSQL
- the product remains runnable, but the preferred normalized control-plane path is not active

## Runtime Metadata Editing Contract

When PostgreSQL is available, metadata edits happen through the API and write
directly into annotation columns on `data_assets` and `asset_columns`.

Today the mutable operator-managed fields are:

- asset owner
- asset description
- documentation refs
- quality check refs
- column descriptions

Those overrides survive scheduler refreshes and manifest reseeds because the
manifest sync updates structural fields only.

## Reporting Layout And Drill-Down Contract

- dashboard widgets now carry explicit `layout.x`, `layout.y`, `layout.w`, and
  `layout.h` values
- the browser editor mutates those values directly when operators move or
  resize widgets
- the analytics API now supports grouped and drilled dataset reads through
  `group_by`, `drill_dimension`, `drill_value`, `sort_by`, and
  `sort_direction`
- the Datasets page consumes that constrained contract to provide curated
  drill-down without bypassing the analytics layer

## Layer Map

The current finance slice now exercises these layers end to end:

- `raw`
- `staging`
- `intermediate`
- `mart`
- `metrics`

## Request And Execution Flow

### Manual run flow

1. User clicks `Run now` in the UI, uses `platformctl remote trigger ...`, or
   runs `trigger <pipeline>` in the admin terminal.
2. `platform-api` validates the pipeline exists.
3. API creates a run snapshot with `queued` state.
4. API enqueues a run request into PostgreSQL or the local queue.
5. `platform-worker` claims the request.
6. Worker loads the pipeline manifest.
7. Worker executes jobs in dependency order.
8. Worker writes artifacts and updates run state after each job.
9. UI reloads run history and artifact lists from the API.

### Scheduled run flow

1. `platform-scheduler` wakes up on `PLATFORM_SCHEDULER_TICK`.
2. Scheduler loads pipeline manifests.
3. Scheduler evaluates each pipeline’s schedule and timezone.
4. If a run is due, it enqueues the run through the orchestration service.
5. Worker processes the run exactly like a manual run.

## Code Ownership Map

- control plane and queueing: `backend/internal/orchestration`
- schedule release logic: `backend/internal/scheduler`
- actual job execution: `backend/internal/execution`
- analytical SQL execution: `backend/internal/transforms`
- metadata and catalog logic: `backend/internal/metadata`
- quality rules and status API: `backend/internal/quality`
- reporting and dashboards: `backend/internal/reporting`
- authz and session model: `backend/internal/authz`
- audit trail: `backend/internal/audit`
- built-in operator commands: `backend/internal/admin`
- storage and artifact serving: `backend/internal/storage`

## Current Python Position

The control plane is Go-first by design. Python should be added only for bounded
execution tasks where the ecosystem is materially better, such as:

- future connectors
- profiling helpers
- transformation helpers not worth expressing in Go
- quality helpers
- docs/code generation utilities

Do not move orchestration, queueing, scheduling, or API ownership into Python.

Current implemented Python usage:

- `packages/python/tasks/enrich_transactions.py`
- `packages/python/tasks/profile_asset.py`

That task enriches landed transaction data into the staging layer and reports
its outputs back to Go through the JSON request/result contract. The profile
utility inspects current materialized assets and feeds the Datasets page with
row counts, observed types, ranges, and sample values.

## Current Queue Position

There is no Redis in the stack right now.

Queue behavior is currently:

- PostgreSQL-backed queue is the preferred implementation
- filesystem queue is the fallback local-first implementation

This is intentional. Redis would add another dependency without clearly solving
today’s highest-priority product gaps.
