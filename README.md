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

## Stack

- Backend and control plane: Go
- Data execution helpers: Python subprocess hooks where needed
- Frontend: React + TypeScript
- Control-plane state: PostgreSQL
- Analytical execution: DuckDB behind an adapter boundary
- Local runtime: Docker Compose with ARM64-friendly defaults

## Current Scope

This initial implementation establishes the platform skeleton and first vertical slice around a personal-finance analytics domain. The first slice is designed to prove the architecture end to end: ingestion, orchestration, transformation, metadata registration, quality checks, analytics serving, and dashboard rendering.

## Important Constraint

`codex.md` was reviewed before starting implementation. The next operational step before the first build should still be to re-check it in case the guidance evolves.
