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
- Analytical execution: DuckDB behind an adapter boundary
- Local runtime: Docker Compose with ARM64-friendly defaults

## Current Scope

This initial implementation establishes the platform skeleton and first vertical slice around a personal-finance analytics domain. The first slice is designed to prove the architecture end to end: ingestion, orchestration, transformation, metadata registration, quality checks, analytics serving, and dashboard rendering.

## Public Repo Safety

- No real secrets, IPs, tokens, or host-specific credentials should ever be committed here.
- `.env.example` contains placeholders and local-development defaults only.
- The Compose stack binds the API and web UI to loopback by default and does not publish PostgreSQL externally.
- The admin terminal is a platform command surface, not arbitrary shell access.
- If you rotate local tokens or database credentials, keep them in untracked local env files.

## Built-In Operations Surface

The platform now includes first-party operational features owned by this repository:

- in-process telemetry for request metrics and command history
- in-process recent log buffer surfaced through the API
- a system overview API and admin diagnostics page
- a browser-based admin terminal in the management portal
- a `platformctl remote ...` command that connects to the running app from any local terminal

## Local Bootstrap

1. Copy `.env.example` to `.env` and replace placeholder credentials and the admin token.
2. Apply database migrations when PostgreSQL is available:

```bash
cd backend
go run ./cmd/platformctl migrate
```

3. Start PostgreSQL and the platform services with Docker Compose or run the binaries locally.
4. Start both `platform-api` and `platform-worker`; manual runs are queued by the API and executed by the worker.
5. Start `platform-scheduler` if you want scheduled queueing enabled.
6. Open the web UI on `http://127.0.0.1:3000`.
7. Use the Pipelines page `Run now` action or the System page admin terminal command `trigger personal_finance_pipeline`.
8. Use `platformctl remote --token <token> status`, `trigger personal_finance_pipeline`, or `artifacts <run_id>` from any local terminal.

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

## Localhost Safety Defaults

- API and web bindings are loopback-first by default.
- PostgreSQL is not published externally by Compose.
- Example tokens and passwords are placeholders only.
- The smoke script uses a temporary root under `/tmp` instead of mutating the
  normal repo-local `var/` tree.

## Important Constraint

`codex.md` was reviewed before starting implementation. The next operational step before the first build should still be to re-check it in case the guidance evolves.
