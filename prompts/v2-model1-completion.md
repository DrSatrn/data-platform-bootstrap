# v2 Model 1 Completion Report

## Files Changed

- `backend/go.mod`
- `backend/go.sum`
- `backend/internal/authz/service.go`
- `backend/internal/authz/ratelimit.go`
- `backend/internal/authz/service_test.go`
- `backend/internal/authz/handler.go`
- `backend/internal/authz/handler_test.go`
- `backend/internal/db/identity_store.go`
- `backend/internal/shared/errors.go`
- `backend/internal/admin/handler.go`
- `backend/internal/analytics/handler.go`
- `backend/internal/analytics/registry_handler.go`
- `backend/internal/metadata/handler.go`
- `backend/internal/metadata/profile_handler.go`
- `backend/internal/orchestration/handler.go`
- `backend/internal/audit/handler.go`
- `backend/internal/reporting/handler.go`
- `backend/internal/quality/handler.go`
- `backend/internal/storage/handler.go`
- `docs/security.md`
- `docs/runbooks/operator-manual.md`
- `prompts/api-error-audit.md`

## What Is Now Verifiably True

- New and reset passwords now use bcrypt instead of the legacy SHA-256 loop.
- Legacy password hashes still verify, so existing users are not immediately locked out during the transition.
- Failed login attempts are rate limited per client identity and return HTTP `429` after repeated failures.
- Expired sessions are swept during auth-service startup and before successful login session creation.
- Active sessions per user are capped at `5` by default and older sessions are trimmed.
- Session touch failures are logged instead of being silently ignored.
- Handler responses no longer return raw `err.Error()` strings directly to API clients.
- Handler-local role checks now use the same structured `403` envelope shape as the shared role middleware.
- The security model is now documented for operators in `docs/security.md`.

## Verification Commands And Results

- `cd backend && go test -count=1 ./internal/authz/...` — PASS
- `cd backend && go test -count=1 ./...` — PASS
- `cd backend && go run ./cmd/platformctl validate-manifests` — PASS
- `git diff --check` — PASS
- `rg -n 'shared\\.WriteJSON\\([^\\n]*err\\.Error\\(\\)|\"error\": err\\.Error\\(\\)' backend/internal --glob 'handler.go' --glob 'handlers.go' --glob 'profile_handler.go' --glob 'registry_handler.go'` — only remaining hits are audit-detail fields and server logs, not client error payloads

## What Was Explicitly NOT Changed

- `backend/internal/app/runtime.go`
- all frontend files under `web/`
- `infra/`
- `Makefile`
- database schema migrations
- report/dashboard product behavior outside handler sanitization

## Escalation Items

- Host-run defaults still bind `:8080` / `:3000`, which is broader than the safer loopback-first Compose posture. This is now documented in `docs/security.md` but not changed in code during this pass.
- The admin terminal remains an intentionally privileged surface and may still expose detailed backend command output to administrators.
- `analytics.MetricCatalogHandler` still swallows preview-query failures and returns empty previews; this remains a product/API ergonomics issue rather than an information-leak issue.
