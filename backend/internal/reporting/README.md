# Reporting Package

This package owns saved reports, dashboards, and the backend API contract used
by the frontend reporting experience.

The current implementation is local-first rather than purely in-memory:

- dashboard definitions are seeded from repo-managed YAML under
  `packages/dashboards`
- saved dashboards are persisted under the platform data root so edits survive
  restart
- when PostgreSQL is bootstrapped, reporting persistence can be mirrored into
  the database while still keeping the filesystem-backed source of truth
  available as a local-first fallback
- the dashboard UI hydrates widgets through constrained analytics queries rather
  than hardcoded page-specific data fetches
- the UI now supports creating, duplicating, editing, deleting, and reordering
  dashboards and widgets directly in the browser
- widget rendering now covers KPI, table, line-chart, and bar-chart surfaces
  without introducing third-party charting dependencies
