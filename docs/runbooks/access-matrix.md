# Access Matrix

This document is an additive reference for role expectations. It is intended as
guide-wire material until the main docs are updated to present one canonical
auth story.

## Roles

| Role | Intended use |
| --- | --- |
| `anonymous` | unauthenticated browsing when allowed by current implementation |
| `viewer` | read-only product access |
| `editor` | pipeline triggers and dashboard modifications |
| `admin` | full editor capabilities plus admin terminal access |

## Workflow Matrix

| Workflow | Minimum intended role |
| --- | --- |
| view dashboard page | `viewer` |
| view datasets page | `viewer` |
| view metrics page | `viewer` |
| view system page | `viewer` |
| trigger pipeline from UI | `editor` |
| save dashboard | `editor` |
| delete dashboard | `editor` |
| use admin terminal in browser | `admin` |
| use `platformctl remote ...` against admin terminal | `admin` |
| use smoke scripts with write/admin actions | `admin` |

## Important Note

This matrix is a desired operator-facing contract, not a guarantee that every
endpoint currently enforces these minimums exactly.

Before wiring this into canonical docs:

- verify endpoint enforcement
- verify UI capability handling
- verify whether anonymous read access is still intentionally allowed

## Verification Pointers

Use these implementation files when reconciling docs and enforcement:

- `backend/internal/authz/service.go`
- `backend/internal/admin/handler.go`
- `backend/internal/orchestration/handler.go`
- `backend/internal/reporting/handler.go`
