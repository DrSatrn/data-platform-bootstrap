# Transforms Package

This package will own SQL and Python transform execution contracts. SQL should be the default for transparent, performant analytical transformations, while Python remains available for carefully justified cases.

The current implementation provides a DuckDB-backed SQL engine that loads
version-controlled SQL from `packages/sql`, materializes curated tables, and
returns result rows to the worker, analytics API, and quality API.
