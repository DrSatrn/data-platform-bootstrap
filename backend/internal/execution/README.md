# Execution Package

This package owns job execution contracts, worker-side task handling, artifact capture, and retry-aware result reporting. The execution layer should remain explicit because it is where partial failures become operationally visible.

The current implementation executes the personal-finance pipeline end to end by:

- copying sample source files into the local raw layer
- loading landed raw files into DuckDB
- building a curated monthly cashflow mart from version-controlled SQL
- computing quality artifacts from DuckDB-backed queries
- publishing a materialized savings-rate metric table and artifact

Artifacts are written under the configured local data root so the API, worker, and UI can all observe the same local-first state.
