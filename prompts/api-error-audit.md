# API Error Response Audit

This audit reviews the current backend HTTP handlers and the route registration
 in `backend/internal/app/runtime.go`. The goal is consistency, not redesign.
No production handlers were changed in this work block.

## Summary

- Handlers audited: `11`
- Endpoint/method contracts reviewed: `18`
- Plain-text error responses found: `0`
- Handlers missing JSON `Content-Type` on error: `0`
- Inconsistencies found: `5`
- Critical issues found: `0`

## Post-Hardening Update

The highest-risk findings from the original audit have been addressed:

- handler responses no longer expose raw `err.Error()` strings to clients
- handler-local role checks now use the same structured `403` envelope shape as
  the shared role middleware

The remaining issues are still worth fixing, but they are now mostly about
status-code discipline and API ergonomics rather than direct information
leakage.

## Key Findings

### Moderate

1. `storage.Handler` returns `400` for artifact read/list failures that may be not-found or internal storage errors.
2. `reporting.Handler` uses `400` for save failures but `500` for delete failures from the same store layer.
3. `metadata.CatalogHandler` maps all annotation-store failures to `400`, even if the underlying store error is operational.
4. `authz.UserHandler` maps most service failures to `400`, even when the cause may be persistence/backend failure.
5. `analytics.MetricCatalogHandler` silently swallows preview query errors and returns empty previews.

## Endpoint Review

| Endpoint | Method | Success Status | Error Status | Error Format | Consistent? |
|----------|--------|---------------|-------------|-------------|-------------|
| `/healthz` | `GET` | `200` | none | n/a | Yes |
| `/api/v1/session` | `GET` | `200` | none | n/a | Yes |
| `/api/v1/session` | `POST` | `200` | `400`, `401` | JSON `{error}`; raw service/login error text on auth failure | Mostly |
| `/api/v1/session` | `DELETE` | `200` | `500`, `405` | JSON `{error}`; raw logout error text | Mostly |
| `/api/v1/admin/users` | `GET` | `200` | `403`, `500`, `405` | JSON `{error}`; `403` is generic admin-only error | No |
| `/api/v1/admin/users` | `POST` | `201` | `400`, `403`, `405` | JSON `{error}`; raw role/service errors bubble through | No |
| `/api/v1/admin/users` | `PATCH` | `200` | `400`, `403`, `405` | JSON `{error}`; raw service errors bubble through | No |
| `/api/v1/pipelines` | `GET` | `200` | `403`, `500` | JSON `{error}` from wrapper/handler | Mostly |
| `/api/v1/pipelines` | `POST` | `202` | `400`, `403` | JSON `{error}`; trigger failures return raw service errors | No |
| `/api/v1/catalog` | `GET` | `200` | `403`, `500` | JSON `{error}` from wrapper/handler | Mostly |
| `/api/v1/catalog` | `PATCH` | `200` | `400`, `403`, `404`, `500`, `503` | JSON `{error}`; store errors returned verbatim | No |
| `/api/v1/catalog/profile` | `GET` | `200` | `400`, `403`, `404`, `500` | JSON `{error}`; profiling errors returned verbatim | Mostly |
| `/api/v1/quality` | `GET` | `200` | `403`, `500` | JSON `{error}`; quality service error returned verbatim | Mostly |
| `/api/v1/analytics` | `GET` | `200` | `400`, `403`, `500` | JSON `{error}`; raw validation/service error text | Mostly |
| `/api/v1/metrics` | `GET` | `200` | `403`, `500` | JSON `{error}`; loader errors verbatim, preview errors swallowed | No |
| `/api/v1/opsview` | `GET` | `200` | `403`, `500`, `405` | JSON `{error}` with stable strings | Yes |
| `/api/v1/reports` | `GET` | `200` | `403`, `500` | JSON `{error}`; list errors verbatim | Mostly |
| `/api/v1/reports` | `POST` | `201` | `400`, `403` | JSON `{error}`; save errors verbatim | No |
| `/api/v1/reports` | `DELETE` | `200` | `400`, `403`, `500`, `405` | JSON `{error}`; delete errors verbatim | No |
| `/api/v1/artifacts` | `GET` | `200` | `400`, `403`, `500`, `405` | JSON `{error}` for failures; success may return raw artifact bytes | No |
| `/api/v1/system/overview` | `GET` | `200` | `403`, `500` | JSON `{error}` with stable strings | Yes |
| `/api/v1/system/logs` | `GET` | `200` | `403` | JSON `{error}` from wrapper | Yes |
| `/api/v1/system/audit` | `GET` | `200` | `403`, `500`, `405` | JSON `{error}`; audit store errors verbatim | Mostly |
| `/api/v1/admin/terminal/execute` | `POST` | `200` | `400`, `403`, `405` | JSON `admin.Result` on both success and command failure; auth error uses `{error}` only | No |

## File-by-File Notes

### `backend/internal/admin/handler.go`
- Errors are JSON.
- Uses `400` for command failures because the terminal endpoint returns `admin.Result`.
- Forbidden response shape differs from `authz.RequireRole`.

### `backend/internal/analytics/handler.go`
- Good: limit validation is explicit and returns `400`.
- Concern: raw service errors are returned directly to clients.
- Client/server classification is string-based via `isClientAnalyticsError`, which is fragile.

### `backend/internal/analytics/registry_handler.go`
- Good: loader failure returns JSON `500`.
- Concern: metric preview query failures are swallowed silently; callers get empty preview arrays without explanation.

### `backend/internal/authz/handler.go`
- Good: all failures are JSON.
- Concern: login/logout/user-management failures return raw service error strings.
- Concern: admin-user handler does its own `403` shape instead of the shared role-gate format.

### `backend/internal/metadata/handler.go`
- Good: method handling is explicit, `Allow` header is set.
- Concern: annotation store errors are always surfaced as `400` with raw error text.
- Concern: `503` for “postgres-backed control plane required” is reasonable, but distinct from the generic auth wrapper shape.

### `backend/internal/metadata/profile_handler.go`
- Good: distinguishes `400`, `404`, and `500`.
- Concern: raw profiler/service error strings are returned directly.

### `backend/internal/observability/handlers.go`
- Good: consistent JSON, stable human-written error messages, no raw internal errors exposed in overview loader failures.
- Good: partial queue/backup failures are degraded to zero summaries instead of breaking the endpoint.

### `backend/internal/orchestration/handler.go`
- Good: read-path failures use stable messages instead of raw store errors.
- Concern: trigger failures return raw control-service errors.
- Concern: write-path auth failure shape differs from the shared role middleware payload.

### `backend/internal/quality/handler.go`
- Good: JSON-only.
- Concern: raw quality-service error strings are returned directly.

### `backend/internal/reporting/handler.go`
- Good: explicit role checks and audit writes.
- Concern: save and delete classify similar store-layer failures differently (`400` vs `500`).
- Concern: store validation/persistence errors are returned verbatim.

### `backend/internal/storage/handler.go`
- Good: JSON on failure, artifact bytes on successful content read are intentional.
- Concern: read/list failures are not distinguished between not-found, invalid path, and internal storage problems.

### `backend/internal/opsview/handler.go`
- Good: stable JSON errors and no raw backend error leakage.
- Good: method guard is explicit and consistent.

### `backend/internal/app/runtime.go`
- Viewer-gated routes consistently use `authz.RequireRole(...)`.
- `admin/terminal/execute` is not wrapped and relies on its own internal admin check, which is one source of the forbidden-shape inconsistency.

## Recommended Follow-Up Fix Order

1. Standardize one shared JSON error envelope for all handlers.
2. Stop returning raw Go/service/store error strings directly to clients.
3. Normalize authorization failures so all protected endpoints return the same `403` shape.
4. Revisit status-code mapping for storage, reporting, metadata patch, and user admin flows.
5. Decide whether preview-query failures in `/api/v1/metrics` should be visible to operators instead of being silently dropped.
