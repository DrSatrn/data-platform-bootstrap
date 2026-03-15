# Reporting Package

This package owns saved reports, dashboards, and the backend API contract used
by the frontend reporting experience.

The current implementation is database-first when PostgreSQL is available and
local-first only as a fallback:

- dashboard definitions are seeded from repo-managed YAML under
  `packages/dashboards`
- when PostgreSQL is bootstrapped, the runtime reads and writes dashboards
  directly from the `dashboards` table
- the filesystem store remains the fallback runtime path only when PostgreSQL
  is unavailable
- the dashboard UI hydrates widgets through constrained analytics queries rather
  than hardcoded page-specific data fetches
- the UI now supports creating, duplicating, editing, deleting, and reordering
  dashboards and widgets directly in the browser
- widget rendering now covers KPI, table, line-chart, and bar-chart surfaces
  without introducing third-party charting dependencies
