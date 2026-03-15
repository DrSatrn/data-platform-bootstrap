# Ingestion Package

This package owns bounded ingestion helpers that move source data into the
platform's raw layer.

Current responsibilities:
- file-copy ingestion from repo-managed sample data
- native database query export for PostgreSQL and MySQL sources

Non-goals:
- broad connector frameworks
- long-running sync daemons
- schema management

The execution runner remains the control plane. This package exists so the
physical ingest mechanics can evolve without pushing more source-specific logic
into `execution/runner.go`.
