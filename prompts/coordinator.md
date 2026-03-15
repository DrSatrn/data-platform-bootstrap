# Master Coordinator Note

**Date:** March 15, 2026  
**Objective:** Ship a credible self-hostable v1 of the Data Platform.  
**Status:** Post-model review. All 3 initial contracts complete. Remaining work is tactical.

---

## Model Status

| Model | Initial Contract | Follow-Up Status |
|-------|-----------------|-----------------|
| Model 1 | ✅ Complete | 🔧 Fix 4 external-tool test timeouts → `go test ./...` must pass |
| Model 2 | ✅ Complete | ✅ No further work needed |
| Model 3 | ✅ Complete (UAT deferred) | 🔧 Run UAT, update README, clean up repo root |

## What Changed From Initial Contracts

- Model 1 delivered all tasks but revealed 4 **pre-existing** test failures outside its allowed surface
- Model 2 delivered all tasks with clean verification
- Model 3 delivered 5 of 6 tasks; UAT was correctly deferred until Models 1/2 completed

## Remaining v1 Checklist

See `prompts/v1-remaining-work.md` for full details.

- [ ] Model 1: Fix external-tool test timeouts (surgical: 3 files, 1-line changes)
- [ ] Verify: `cd backend && go test ./...` passes globally
- [ ] Model 3: Run full UAT (`make bootstrap` → `uat-checklist.md`)
- [ ] Model 3: Update `README.md`
- [ ] Model 3: Delete `temp-model1-frontend-wire-plan.md`, archive `new-thread-eng-feedback.md`
- [ ] Reviewer: Final `make smoke` verification
- [ ] Reviewer: Tag v1 release

## Chokepoint Files (Still Locked)

- `backend/internal/app/runtime.go`
- `web/src/app/App.tsx` (Model 2 already edited this under approval)
- `infra/compose/docker-compose.yml`
- `Makefile`

## Merge Order

1. Model 1 fix merges first (test-only change, zero conflict risk)
2. Model 3 merges second (docs + UAT annotation, zero code conflict)
3. Reviewer does final verification and tags
