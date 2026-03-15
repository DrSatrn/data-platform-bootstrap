# Service Boundaries

This document explains how the platform is decomposed into bounded contexts and runtime processes.

## Runtime Processes

- `platform-api`: serves control-plane, metadata, analytics, reporting, and health endpoints.
- `platform-scheduler`: evaluates schedules, dependency readiness, and retry timing.
- `platform-worker`: executes raw landing, external-tool, transform, metric publication, and quality jobs.
- `platform-web`: renders the internal UI for operators and analysts.

## Backend Bounded Contexts

- `orchestration`: pipeline/job definitions, runs, state transitions, dependencies
- `scheduler`: release timing and refresh loops
- `execution`: worker execution contract and artifact handling
- `externaltools`: bounded adapters and runners for optional dbt-style external tools
- `transforms`: SQL and Python transform boundaries
- `metadata`: datasets, columns, lineage, owners, freshness, docs metadata
- `quality`: quality definitions and results
- `analytics`: curated query-serving endpoints
- `reporting`: saved reports and dashboards
- `observability`: logs, health, metrics, diagnostics
- `storage`: local filesystem and future object-store abstraction

## Why This Decomposition

- It preserves clear ownership while keeping the system locally operable.
- It avoids premature microservice sprawl.
- It makes later extraction possible without forcing distributed complexity into v1.
