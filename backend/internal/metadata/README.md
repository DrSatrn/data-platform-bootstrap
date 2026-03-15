# Metadata Package

This package owns dataset, lineage, ownership, freshness, documentation
coverage, and runtime profiling metadata. It is a core product subsystem, not a
support detail.

Current responsibilities include:

- synchronizing repo-managed assets into the catalog API surface
- projecting asset and column metadata into PostgreSQL when enabled
- deriving freshness and lineage summaries for operators
- generating cached dataset profiles from current local materializations
