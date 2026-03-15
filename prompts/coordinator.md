# Master Coordinator Note

**Date:** March 15, 2026  
**Objective:** Ship a credible self-hostable v1 of the Data Platform.  
**Strategy:** Layer-based parallelism — 3 models working simultaneously on backend, frontend, and platform/infra.

---

## Active Models

| Model | Mission | Merge Order |
|-------|---------|-------------|
| Model 1 | Backend hardening & test coverage | 1st (safest, additive-only) |
| Model 2 | Frontend consolidation & polish | 2nd |
| Model 3 | Platform, infra, docs & release readiness | 3rd (last) |

## Chokepoint Files (Reviewer-Controlled)

These files are **locked**. No model may edit them without explicit reviewer approval:

- `backend/internal/app/runtime.go`
- `web/src/app/App.tsx` (Model 2 may propose changes after routing approval)
- `infra/compose/docker-compose.yml`
- `Makefile`

## No-Fly Zones (Do Not Touch for v1)

These packages are working and stable. No model should modify them:

- `backend/internal/authz/`
- `backend/internal/db/`
- `backend/internal/backup/`
- `backend/internal/orchestration/`
- `backend/internal/scheduler/`
- `backend/internal/reporting/`
- `backend/internal/audit/`
- `backend/internal/opsview/`
- `infra/migrations/`

## Coordination Rules

1. **Before editing any file**, check the ownership matrix in `v1-review-coordination-plan.md` Section 6.
2. **If you need a file you don't own**, stop and document the request. Do not edit it.
3. **After completing a task**, leave a completion note containing:
   - Files changed (list)
   - Verification commands run and results
   - What was explicitly NOT changed
   - Any escalation items
4. **Merge order is strict:** Model 1 → Model 2 → Model 3.

## Verification Commands (All Models Must Know)

```bash
# Backend
cd backend && go test ./...
cd backend && go run ./cmd/platformctl validate-manifests

# Frontend
cd web && npm run build
cd web && npm test

# Full stack
make smoke
```

## Current Status Tracking

Update this section as models complete tasks:

- [ ] Model 1: transforms engine tests
- [ ] Model 1: quality service tests expanded
- [ ] Model 1: ingestion stub resolved
- [ ] Model 2: URL routing added
- [ ] Model 2: DashboardPage component extraction
- [ ] Model 2: Core page tests added
- [ ] Model 2: Loading/error states added
- [ ] Model 3: guide-wire.md and plan.md updated
- [ ] Model 3: CI/CD workflow created
- [ ] Model 3: Full UAT run documented
