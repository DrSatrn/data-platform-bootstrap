# Retention Package

This package owns repo-driven retention enforcement for local materializations,
run-scoped artifacts, and mirrored control-plane history.

The retention workflow is intentionally CLI-friendly:
- load asset and pipeline manifests
- derive retention windows from asset metadata
- purge stale filesystem materializations and run artifacts
- purge mirrored PostgreSQL run rows when a database executor is available

This keeps the policy logic testable without forcing long-running janitor logic
into the main runtime process.
