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
2. Start PostgreSQL and the platform services with Docker Compose or run the binaries locally.
3. Start both `platform-api` and `platform-worker`; manual runs are queued by the API and executed by the worker.
4. Open the web UI on `http://127.0.0.1:3000`.
5. Use the Pipelines page `Run now` action or the System page admin terminal command `trigger personal_finance_pipeline`.
6. Use `platformctl remote --token <token> status` or `platformctl remote --token <token> trigger personal_finance_pipeline` from any local terminal.

## Important Constraint

`codex.md` was reviewed before starting implementation. The next operational step before the first build should still be to re-check it in case the guidance evolves.
