# Quality Package

This package owns data quality definitions and runtime result surfaces. Quality checks should stay explicit and operator-visible because silent data trust failures are costly.

Quality status now prefers DuckDB-backed SQL queries defined under
`packages/sql/quality`. That keeps the checks version-controlled, reviewable,
and shared between the worker and API surfaces.
