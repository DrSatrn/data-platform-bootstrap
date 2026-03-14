# Orchestration Package

This package owns pipeline and job definitions, runtime state transitions, dependency evaluation, and operator-facing orchestration APIs. Reliability matters here because weak orchestration semantics quickly undermine trust in the whole platform.

The current implementation now includes:

- a durable file-backed run store for localhost development
- a file-backed run queue shared by the API and worker
- a control service that validates and queues manual pipeline runs

This local-first design keeps the system restart-safe and inspectable while the PostgreSQL-backed orchestration repositories are still pending.
