# Analytics Package

This package owns curated query-serving behavior. It should remain tightly constrained so reporting can never accidentally couple itself to raw internal tables or uncontrolled SQL surfaces.

The current implementation computes the finance dashboard from repo-managed sample data. That keeps the first analytics slice genuinely self-built while we prepare the DuckDB-backed execution path.
