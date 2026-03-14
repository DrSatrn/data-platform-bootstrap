# Reporting Package

This package owns saved reports, dashboards, and the backend API contract used
by the frontend reporting experience.

The current implementation is local-first rather than purely in-memory:

- dashboard definitions are seeded from repo-managed YAML under
  `packages/dashboards`
- saved dashboards are persisted under the platform data root so edits survive
  restart
- the dashboard UI hydrates widgets through constrained analytics queries rather
  than hardcoded page-specific data fetches
