# Active Build Plan

## Current Goal

Ship a credible self-hosted v1 of the Data Platform.

## Active Workstreams

### Model 1

- add missing backend test coverage on critical paths
- harden error handling and integration seams
- resolve the `ingestion/` stub story

### Model 2

- add real URL routing
- break down large frontend page files
- add loading/error states and page-level tests

### Model 3

- keep coordination docs current
- add CI/CD workflow coverage
- update project handoff docs
- run and document UAT after Models 1 and 2 are complete

## Immediate Priorities

1. remove stale coordination history from `guide-wire.md`
2. align `plan.md` and `codex.md` to the v1 review
3. establish CI for backend tests, manifest validation, frontend build, and
   frontend tests
4. defer UAT until completion notes from Models 1 and 2 exist

## v1 Release Criteria

Use `v1-review-coordination-plan.md` as the source of truth.

The release is credible when:

- the required v1 items in Section 3 are satisfied
- the highest-priority gaps in Section 4 are either fixed or consciously
  deferred
- the ownership matrix in Section 6 is respected
- the top tasks in Section 8 have either landed or been explicitly deferred
- UAT from `uat-checklist.md` is executed and annotated

## Current Risks

- reviewer-controlled chokepoints (`runtime.go`, `App.tsx`, Compose, Makefile`)
- unverified completion claims from parallel work until completion notes exist
- DuckDB CGO complexity in CI
- frontend polish and coverage gaps remain the most visible v1 risk

## Next Release-Manager Actions

1. keep docs synchronized with actual repo state
2. keep CI green and minimal
3. document UAT deferral or results clearly
4. leave a clean completion report for the merge pass
