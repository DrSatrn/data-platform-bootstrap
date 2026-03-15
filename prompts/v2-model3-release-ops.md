# Model 3 — Release Readiness, Docs & Ops Infrastructure

**Priority:** HIGH — this makes the platform deployable and trustworthy.  
**Owner:** Model 3  
**Merge order:** 3rd (last, after Models 1 and 2)  

---

## Mission

Make the platform credibly deployable for a 3-person on-prem data team by fixing stale documentation, creating a deployment guide, hardening CI, and establishing release discipline. Your work is the bridge between "working repo" and "trustworthy internal product."

## Context

Read these first:
- `docs/runbooks/operator-manual.md` — has stale content (says restore not implemented)
- `docs/architecture/service-boundaries.md` — lists deleted `ingestion` package
- `codex.md` — session handoff (compare with actual repo state)
- `v1-review-coordination-plan.md` — previous review baseline
- `.github/workflows/ci.yml` — current CI config
- `prompts/v1-remaining-work.md` — outstanding items from previous cycle

## Tasks (In Priority Order)

### Task 1: Fix All Stale Documentation

**Current state:** Multiple docs contradict the actual repo:
- `operator-manual.md` L368-375: says "restore automation is not implemented yet" — it IS
- `operator-manual.md` L374: says "restore automation is still not built on top of the backup bundle format" — it IS
- `service-boundaries.md` L17: lists `ingestion` as a bounded context — the package was deleted
- `codex.md` L39-47: "What is still pending" lists items that have been completed

**Required change:**
- Audit every docs file against actual repo state
- Fix all contradictions
- Remove references to deleted packages
- Update "what is still pending" sections to reflect reality

**Files allowed:**
- All files under `docs/`
- `codex.md`
- `README.md`
- `plan.md`

**Completion signal:** No doc file contradicts the current codebase.

### Task 2: Create On-Prem Deployment Guide

**Required change:**
- Create `docs/runbooks/deployment.md` covering:
  1. Hardware requirements (RAM, disk, CPU for 3-person team)
  2. Network configuration (LAN binding, firewall rules)
  3. TLS termination options (reverse proxy setup)
  4. Backup schedule recommendations
  5. Monitoring recommendations (what to check daily/weekly)
  6. User provisioning for a 3-person team
  7. Data directory sizing and growth expectations
  8. OrbStack-specific notes for Mac deployment

**Files allowed:**
- `docs/runbooks/deployment.md` (NEW)
- `docs/runbooks/operator-manual.md` — link to new guide

**Completion signal:** A new operator can deploy to their on-prem machine by following this guide alone.

### Task 3: Fix External-Tool Test Failures

**Current state:** 4 tests fail with `context deadline exceeded` due to 2-second timeouts.

**Required change:**
- `backend/internal/execution/external_tool_test.go` L47: `Timeout: "2s"` → `Timeout: "30s"`
- `backend/internal/execution/external_tool_failures_test.go` L128: `Timeout: "2s"` → `Timeout: "30s"`
- `backend/internal/execution/external_tool_operator_inspection_test.go` L47: `Timeout: "2s"` → `Timeout: "30s"`

**Files allowed:**
- The 3 test files listed above

**Completion signal:** `cd backend && go test ./...` passes globally.

### Task 4: Harden CI Pipeline

**Current state:** `.github/workflows/ci.yml` exists but DuckDB CGO makes it non-trivial. No evidence it has run successfully.

**Required change:**
- Ensure the CI workflow:
  1. Installs CGO prerequisites (gcc, build-essential)
  2. Runs `go test ./...` successfully
  3. Runs `npm run build` and `npm test`
  4. Runs `go run ./cmd/platformctl validate-manifests`
  5. Caches Go modules and npm packages
- Test locally with `act` if possible

**Files allowed:**
- `.github/workflows/ci.yml`
- `Makefile` (only with reviewer approval)

**Completion signal:** CI workflow is syntactically valid and has been tested locally or on GitHub.

### Task 5: Create Release Checklist

**Required change:**
- Create `docs/runbooks/release-checklist.md` with:
  1. Pre-release verification steps
  2. Changelog format
  3. Version tagging strategy
  4. Migration compatibility check
  5. Backup-before-release requirement
  6. Smoke/benchmark pass requirements
  7. Post-release validation steps
  8. Rollback procedure

**Files allowed:**
- `docs/runbooks/release-checklist.md` (NEW)

### Task 6: Run Full UAT

**Required change:**
- Execute `make bootstrap`
- Follow the UAT checklist end-to-end
- Annotate every item with PASS/FAIL + evidence
- Document any bugs found and escalate to the appropriate model

**Files allowed:**
- `uat-checklist.md` — annotate with results
- `docs/` — update based on findings

**Completion signal:** Every UAT item has a documented result.

**Dependency:** This task can only start after Models 1 and 2 report completion.

## Escalation Triggers

- If CI requires changes to `Makefile` → request reviewer approval
- If UAT reveals code bugs → create an escalation note referencing the specific model
- If stale docs reveal actual product gaps → document them, don't fix the code

## Completion Note Format

After completing, leave `prompts/v2-model3-completion.md` containing:
1. Files changed
2. What is now verifiably true
3. Verification commands and results
4. What was explicitly NOT changed
5. Any escalation items
