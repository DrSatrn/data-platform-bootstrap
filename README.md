# Data Platform

This repository contains a local-first, self-hosted data orchestration and analytics platform built as a serious engineering project. The platform combines orchestration, ingestion, transformations, metadata, observability, analytics serving, and an internal reporting interface into one coherent product designed to run well on an Apple Silicon laptop and ARM64 Linux VM.

The implementation intentionally emphasizes teaching value. Code is organized around clear subsystem boundaries, package-level responsibility, explicit runtime behavior, and heavily documented entrypoints so the project can be studied as much as it can be used.

## Product Goals

- Reliable orchestration with schedules, dependencies, retries, audit history, and understandable failure handling.
- Medallion-style data movement through raw, staging, intermediate, mart, and metrics layers.
- Metadata-first operation with lineage, ownership, freshness, quality, and documentation coverage.
- Curated analytics serving rather than direct raw-table access.
- A custom operational and reporting UI tailored to platform operators and internal analysts.
- Version-controlled manifests, infra, migrations, dashboards, and documentation.
- Built-in metrics, logs, diagnostics, and admin tooling owned by this repo instead of outsourced dashboard products.

## Stack

- Backend and control plane: Go
- Data execution helpers: Python subprocess hooks where needed
- Frontend: React + TypeScript
- Control-plane state: PostgreSQL
- Analytical execution: DuckDB behind a repo-owned SQL execution adapter
- Local runtime: Docker Compose with ARM64-friendly defaults

## Current Scope

This implementation now covers a more polished v2-style finance slice. The
platform proves the architecture end to end with ingestion, orchestration,
transformation, metadata registration, quality checks, analytics serving,
file-backed saved dashboards, and a richer reporting surface around cashflow,
category spend, and budget variance.

When PostgreSQL has been migrated and is reachable, the runtime now prefers a
PostgreSQL-backed control plane for run snapshots, queue state, and artifact
metadata. The filesystem path remains the local-first fallback when PostgreSQL
is unavailable.

## Public Repo Safety

- No real secrets, IPs, tokens, or host-specific credentials should ever be committed here.
- `.env.example` contains placeholders and local-development defaults only.
- The Compose stack binds the API and web UI to loopback by default and does not publish PostgreSQL externally.
- The admin terminal is a platform command surface, not arbitrary shell access.
- If you rotate local tokens or database credentials, keep them in untracked local env files.

## Access Control

The platform now supports lightweight bearer-token RBAC for self-hosted use:

- `admin` can use the admin terminal and all editor actions
- `editor` can trigger runs and modify saved dashboards
- `viewer` is read-only

Configuration:

- `PLATFORM_ADMIN_TOKEN` remains supported and maps to the `admin` role
- `PLATFORM_ACCESS_TOKENS` adds extra tokens in `token:role:subject` format,
  comma-separated

Example:

```bash
PLATFORM_ACCESS_TOKENS=viewer-token:viewer:alice,editor-token:editor:bob
```

The browser UI stores one bearer token locally and uses `/api/v1/session` to
discover capabilities, so write/admin controls disable themselves when the
token is missing or under-privileged.

## Audit Trail

Privileged platform actions now write to a first-party append-only audit trail.
The current audit scope covers:

- manual pipeline triggers
- dashboard saves and deletes
- admin terminal command execution

The audit feed is exposed at `/api/v1/system/audit`, shown in the System page,
and mirrored into PostgreSQL when the DB-backed control plane is available.

## Analytical SQL

Curated SQL now lives under [packages/sql](/Users/streanor/Documents/Playground/data-platform/packages/sql). The worker loads landed raw files into DuckDB, materializes curated tables from those SQL files, and the analytics and quality APIs query the same DuckDB-backed layer when it is available.

The finance slice now includes these curated outputs:

- `mart_monthly_cashflow`
- `mart_category_spend`
- `mart_budget_vs_actual`
- `metrics_savings_rate`
- `metrics_category_variance`

Because the DuckDB Go driver uses CGO, host builds need working Apple Silicon C
tooling. On macOS that usually means Xcode Command Line Tools are installed.

## Built-In Operations Surface

The platform now includes first-party operational features owned by this repository:

- in-process telemetry for request metrics and command history
- in-process recent log buffer surfaced through the API
- a system overview API and admin diagnostics page
- a browser-based admin terminal in the management portal
- a `platformctl remote ...` command that connects to the running app from any local terminal
- saved dashboards seeded from repo-managed dashboard manifests and persisted through the reporting API
- browser-based dashboard management with create, duplicate, edit, delete, reorder, and live widget preview flows
- first-party line and bar chart widgets rendered without external BI or charting products

## Local Bootstrap

1. Copy `.env.example` to `.env` and replace placeholder credentials and tokens.
2. Start PostgreSQL and the platform services with Docker Compose or run the binaries locally.
3. Start both `platform-api` and `platform-worker`; manual runs are queued by the API and executed by the worker.
4. Start `platform-scheduler` if you want scheduled queueing enabled.
5. Open the web UI on `http://127.0.0.1:3000`.
6. Use the Pipelines page `Run now` action or the System page admin terminal command `trigger personal_finance_pipeline`.
7. Use `platformctl remote --token <token> status`, `trigger personal_finance_pipeline`, or `artifacts <run_id>` from any local terminal.
8. Use the Datasets page to inspect freshness states and the Dashboard page to manage saved dashboards directly from the browser.
9. Use the Datasets page catalog/detail split view to inspect ownership, source refs, quality refs, docs refs, and column-level metadata for a selected asset.

## Compose Bootstrap

The Compose path is now a validated local runtime on Apple Silicon with Go 1.24
service images built from the repo Dockerfiles. The stack includes a one-shot
migration service, a packaged web service instead of a Vite dev server, and
health-gated startup ordering:

```bash
make bootstrap
```

Or, without `make`:

```bash
docker compose -f infra/compose/docker-compose.yml up -d --build
```

After the stack is healthy, the platform should be available on:

- `http://127.0.0.1:8080` for the API
- `http://127.0.0.1:3000` for the packaged web UI
- `platformctl remote ...` against `http://127.0.0.1:8080`

## Verified Localhost Smoke Path

The repo now includes a first-party localhost smoke script that starts an
isolated API, worker, and scheduler stack on loopback, drives a scheduled run
plus a manual run, verifies run-scoped artifacts, exercises the admin terminal
API, and proves the `platformctl remote` CLI path.

```bash
make smoke
```

By default the smoke run uses `http://127.0.0.1:18080` and a temporary runtime
root under `/tmp`. It keeps that runtime root after success so you can inspect
logs and artifacts. Set `PLATFORM_SMOKE_KEEP=0` if you want automatic cleanup.
If `127.0.0.1:18080` is already in use, rerun with
`PLATFORM_SMOKE_PORT=<unused-port> make smoke`.

## Verified Compose Smoke Path

The repo also includes a packaged-deployment smoke workflow that boots Docker
Compose, waits for migrations and health, validates the hosted web UI, and
drives a real pipeline run through the API, worker, scheduler, analytics,
quality, artifacts, and CLI layers:

```bash
make compose-smoke
```

## Benchmarking

The repo now includes a first-party benchmark command and wrapper script so we
can start tracking response budgets as the product grows:

```bash
make benchmark
```

That writes a timestamped JSON report under `var/benchmarks/` and currently
covers health, catalog, analytics, reports, system overview, and the admin
terminal status command.

## Localhost Safety Defaults

- API and web bindings are loopback-first by default.
- PostgreSQL is not published externally by Compose.
- Example tokens and passwords are placeholders only.
- The smoke script uses a temporary root under `/tmp` instead of mutating the
  normal repo-local `var/` tree.

## Important Constraint

`codex.md` was reviewed before starting implementation. The next operational step before the first build should still be to re-check it in case the guidance evolves.
