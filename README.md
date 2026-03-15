# Data Platform

This repository contains a local-first, self-hosted data orchestration and analytics platform built as a serious engineering project. The platform combines orchestration, ingestion, transformations, metadata, observability, analytics serving, and an internal reporting interface into one coherent product designed to run well on an Apple Silicon laptop and ARM64 Linux VM.

The implementation intentionally emphasizes teaching value. Code is organized around clear subsystem boundaries, package-level responsibility, explicit runtime behavior, and heavily documented entrypoints so the project can be studied as much as it can be used.

## Documentation Map

If you are new to the project, use this reading order:

1. [README.md](/Users/streanor/Documents/Playground/data-platform/README.md)
2. [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
3. [runtime-wiring.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/runtime-wiring.md)
4. [operator-manual.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/operator-manual.md)
5. [making-changes.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/making-changes.md)

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
is unavailable. The live source-of-truth breakdown is exposed in
`GET /api/v1/system/overview` and rendered in the System page as the
`Source Of Truth` card.

## Configuration Reality

- Host-run Go binaries now auto-load `.env` and `.env.local` from the repo root
  or the current working directory.
- [.env.example](/Users/streanor/Documents/Playground/data-platform/.env.example)
  is the host-run example and uses local filesystem paths.
- Compose does not use `.env.example` as its canonical config source.
- The tracked Compose file has safe defaults already, and optional overrides
  belong in `.env.compose`.

## Public Repo Safety

- No real secrets, IPs, tokens, or host-specific credentials should ever be committed here.
- `.env.example` contains placeholders and local-development defaults only.
- The Compose stack binds the API and web UI to loopback by default and does not publish PostgreSQL externally.
- The admin terminal is a platform command surface, not arbitrary shell access.
- If you rotate local tokens or database credentials, keep them in untracked local env files.

## Access Control

The platform now supports lightweight bearer-token RBAC for self-hosted use:

- anonymous can access only `GET /healthz` and `GET /api/v1/session`
- `viewer` can access read-only product APIs and pages
- `editor` can trigger runs and modify saved dashboards
- `admin` can use the admin terminal and all editor actions

Configuration:

- `PLATFORM_ADMIN_TOKEN` remains supported and maps to the `admin` role
- `PLATFORM_ACCESS_TOKENS` adds extra tokens in `token:role:subject` format,
  comma-separated

Example:

```bash
PLATFORM_ACCESS_TOKENS=viewer-token:viewer:alice,editor-token:editor:bob
```

The browser UI stores one bearer token locally and uses `/api/v1/session` to
discover capabilities. Product pages now require at least `viewer`, so an
anonymous browser session will see only health/session access until a token is
provided.

## Audit Trail

Privileged platform actions now write to a first-party append-only audit trail.
The current audit scope covers:

- manual pipeline triggers
- dashboard saves and deletes
- admin terminal command execution

The audit feed is exposed at `/api/v1/system/audit`, shown in the System page,
and mirrored into PostgreSQL when the DB-backed control plane is available.

## Backups

The platform now includes a first-party backup/export path built in-repo:

- `platformctl backup create`
- `platformctl backup verify`
- `platformctl backup list`
- admin terminal commands: `backups`, `backup create`, `backup verify <bundle>`

Each backup bundle is a portable `.tar.gz` archive containing:

- control-plane JSON exports for runs, queue state, dashboards, audit events,
  metadata assets, and sanitized config
- materialized local data and run artifacts when present
- the DuckDB file
- repo-managed manifest and dashboard snapshots
- a checksummed `manifest.json`

Use:

```bash
make backup
```

Use:

```bash
make restore-drill
```

when you want a safe extraction drill without overwriting live state. See
[backups.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md)
for the full recovery procedure.

## Analytical SQL

Curated SQL now lives under [packages/sql](/Users/streanor/Documents/Playground/data-platform/packages/sql). The worker loads landed raw files into DuckDB, materializes curated tables from those SQL files, and the analytics and quality APIs query the same DuckDB-backed layer when it is available.

The finance slice now includes these curated outputs:

- `staging_transactions_enriched`
- `intermediate_category_monthly_rollup`
- `mart_monthly_cashflow`
- `mart_category_spend`
- `mart_budget_vs_actual`
- `metrics_savings_rate`
- `metrics_category_variance`

Because the DuckDB Go driver uses CGO, host builds need working Apple Silicon C
tooling. On macOS that usually means Xcode Command Line Tools are installed.

The worker now also supports bounded Python subprocess tasks behind the Go
control plane. The current finance slice uses Python to enrich raw
transactions into the staging layer before the intermediate and mart layers are
built, and the metadata catalog now uses a second bounded Python utility to
profile materialized assets for the Datasets page.

## Built-In Operations Surface

The platform now includes first-party operational features owned by this repository:

- in-process telemetry for request metrics and command history
- in-process recent log buffer surfaced through the API
- a system overview API and admin diagnostics page
- run throughput, failure, queue, and backup inventory summaries surfaced in the System page
- a browser-based admin terminal in the management portal
- a `platformctl remote ...` command that connects to the running app from any local terminal
- saved dashboards seeded from repo-managed dashboard manifests and persisted through the reporting API
- browser-based dashboard management with create, duplicate, edit, delete, reorder, report owner/audience metadata, dashboard-wide default filters, and saved preset flows
- first-party line and bar chart widgets rendered without external BI or charting products
- a semantic metrics browser page backed by repo-managed metric manifests
- Python-backed dataset profile cards in the Datasets page so operators can
  inspect row counts, observed types, null counts, ranges, and sample values

## Start Here

Use the canonical quickstart instead of guessing between the overlapping
runbooks:

- [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)

The recommended first success path is:

```bash
make smoke
```

That verifies the core platform behavior without requiring you to hand-start
services. After that, use:

- [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
  for the packaged Compose path
- [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)
  for the host-run debug path

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
