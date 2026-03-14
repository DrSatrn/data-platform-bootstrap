# SQL Package

This directory stores version-controlled SQL owned by the platform. Keeping
analytical transforms, metric materializations, and quality queries in git
makes the execution path inspectable, reviewable, and reproducible.

Use these subdirectories intentionally:

- `bootstrap/` for raw landing-table loads into DuckDB
- `transforms/` for curated dataset materializations
- `metrics/` for metric table materializations
- `quality/` for explicit operator-visible quality queries
