# Analytics Package

This package owns curated query-serving behavior. It should remain tightly constrained so reporting can never accidentally couple itself to raw internal tables or uncontrolled SQL surfaces.

The runtime now prefers DuckDB-backed curated datasets and metric tables. It
falls back to materialized JSON artifacts or sample data only when the
analytical database is not yet available, which keeps localhost startup and
recovery paths resilient without weakening the serving contract.
