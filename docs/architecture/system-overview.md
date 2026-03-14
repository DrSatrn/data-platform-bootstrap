# System Overview

The platform is split into a small number of runtime processes that share a common codebase and database-backed control plane:

- `platform-api` exposes orchestration, metadata, analytics, reporting, and health endpoints.
- `platform-scheduler` evaluates schedules, dependency readiness, and retry timing.
- `platform-worker` executes jobs and emits structured operational outcomes.
- `platform-web` provides the operator and analyst interface.
- `postgres` stores durable control-plane, metadata, and reporting state.

DuckDB is accessed through an adapter boundary for analytical execution so the project stays laptop-friendly and avoids coupling orchestration logic to analytical query mechanics.
