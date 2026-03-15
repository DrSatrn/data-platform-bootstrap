# Guide Wire
This file is the live coordination rail for the v1 release push.

## Goal
Ship a credible self-hostable v1 without merge collisions across the three active models.

## Active Models
- Model 1: backend hardening and test coverage
- Model 2: frontend consolidation and polish
- Model 3: platform, infra, docs, CI/CD, release readiness

## Reviewer-Controlled Chokepoints
Do not edit without explicit reviewer approval:
- `backend/internal/app/runtime.go`
- `web/src/app/App.tsx`
- `infra/compose/docker-compose.yml`
- `Makefile`

## Backend Ownership
- Model 1 open:
  - `backend/internal/transforms/`
  - `backend/internal/quality/`
  - `backend/internal/ingestion/`
  - `backend/internal/execution/`
  - `backend/internal/storage/`
  - `backend/internal/analytics/`
  - `backend/test/`
- No-fly zone for all models:
  - `backend/internal/authz/`
  - `backend/internal/db/`
  - `backend/internal/backup/`
  - `backend/internal/orchestration/`
  - `backend/internal/scheduler/`
  - `backend/internal/reporting/`
  - `backend/internal/audit/`
  - `backend/internal/opsview/`

## Frontend Ownership
- Model 2 open:
  - `web/src/pages/` except `ManagementPage.tsx`
  - `web/src/components/`
  - `web/src/features/auth/`
  - `web/src/features/dashboard/`
  - `web/src/styles/`
  - `web/src/lib/`
- Model 3 maintenance only:
  - `web/src/pages/ManagementPage.tsx`
  - `web/src/features/management/`

## Infra And Docs Ownership
- Model 3 open:
  - `.github/`
  - `docs/`
  - `infra/scripts/`
  - `guide-wire.md`
  - `plan.md`
  - `codex.md`
  - `README.md`
  - `doc.md`
  - `contributing.md`
  - `infra-overview.md`
  - `uat-checklist.md`
  - `new-thread-eng-feedback.md`
- No-fly zone:
  - `infra/migrations/`

## Current Priorities
1. Model 1: backend tests and API hardening
2. Model 2: routing, page consolidation, loading/error states, page tests
3. Model 3: rewrite stale coordination docs, create CI, update handoff docs, run UAT when Models 1 and 2 are complete

## Merge Protocol
1. Check `v1-review-coordination-plan.md` Section 6 before editing anything.
2. If a task requires a file you do not own, stop and escalate.
3. Leave a completion note with files changed, verification, untouched files, and escalation items.
4. Merge order is strict: Model 1, then Model 2, then Model 3.

## Release Gates
- `cd backend && go test ./...`
- `cd backend && go run ./cmd/platformctl validate-manifests`
- `cd web && npm run build`
- `cd web && npm test`
- `make smoke`

## Current Status
- The principal v1 review is complete.
- `guide-wire.md` and `plan.md` now track current work, not historical tranche notes.
- UAT is the final verification step and should only run after Model 1 and Model 2 completion is explicitly documented.
