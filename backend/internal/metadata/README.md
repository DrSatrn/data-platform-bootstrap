# Metadata Package

This package owns dataset, lineage, ownership, freshness, documentation
coverage, and runtime profiling metadata. It is a core product subsystem, not a
support detail.

Current responsibilities include:

- seeding repo-managed assets into PostgreSQL when enabled
- serving the catalog from PostgreSQL when the preferred control plane is available
- persisting runtime metadata annotations such as owner, description, docs refs,
  quality refs, and column descriptions directly into PostgreSQL
- deriving freshness and lineage summaries for operators
- generating cached dataset profiles from current local materializations
