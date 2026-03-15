# Model 2 — Operational Resilience & Domain Expansion

**Priority:** HIGH — this proves the platform generalizes and recovers from failures.  
**Owner:** Model 2  
**Merge order:** 2nd (after Model 1)  

---

## Mission

Prove the data platform architecture generalizes beyond one domain slice and handles real-world failure scenarios. This work is the second biggest blocker to production because a platform that only proves one data pipeline and has never been failure-tested cannot be trusted with real data operations.

## Context

Read these first:
- `backend/internal/execution/runner.go` — the hardcoded ingestion switch-case
- `backend/internal/analytics/service.go` — the hardcoded dataset query map
- `backend/internal/transforms/engine.go` — the DuckDB execution engine
- `docs/runbooks/benchmarking.md` — current benchmark scope
- `packages/manifests/` — manifest structure for pipelines, assets, quality, metrics

## Tasks (In Priority Order)

### Task 1: Make Ingestion Pluggable

**Current state:** `runner.go` L177-188 uses a switch-case on job IDs to copy specific sample files. Adding a second domain requires modifying this function.

**Required change:**
- Replace the hardcoded switch-case with a registry or manifest-driven pattern
- Each ingest job should declare its source and target in the pipeline manifest
- The runner should look up source/target paths from the job spec, not from code

**Files allowed:**
- `backend/internal/execution/runner.go` — refactor `runIngest`
- `backend/internal/orchestration/models.go` — extend Job spec if needed
- New test files

**Files forbidden:**
- ❌ `backend/internal/app/runtime.go`
- ❌ `backend/internal/authz/` (Model 1's domain)

**Completion signal:** Existing personal finance pipeline works unchanged after refactor. `go test ./internal/execution/...` passes.

### Task 2: Add a Second Domain Slice

**Required change:**
- Add a second sample domain (e.g., operational metrics, sales, inventory — your choice)
- Create: sample data, pipeline manifest, asset manifest, quality manifest, metric manifest, SQL transforms
- Wire it through the same pipeline lifecycle as personal finance
- Add at least one analytics query path for the new domain

**Files allowed:**
- `packages/sample_data/` — new sample data directory
- `packages/manifests/` — new manifests
- `packages/sql/` — new transform + metric SQL
- `backend/internal/analytics/service.go` — extend `QueryDataset` for new mart(s)

**Completion signal:** Both pipelines succeed via `make smoke` or manual trigger. Both domains return analytics data.

### Task 3: Add Failure Recovery Tests

**Required change:**
- Create test scenarios that prove the platform handles:
  1. Postgres becoming unavailable mid-run (worker should complete with filesystem fallback or fail gracefully)
  2. DuckDB file being locked/corrupted (error should be clear, not a panic)
  3. Worker process crash and restart (queued runs should be re-claimable)
  4. Restore under active state (verify `make restore-e2e` after a run)

**Files allowed:**
- `backend/test/` — integration test files
- `infra/scripts/` — failure simulation scripts

**Completion signal:** At least 3 failure scenarios are scripted and documented with pass/fail results.

### Task 4: Expand Benchmark Suite

**Required change:**
- Add to the existing benchmark suite:
  1. Concurrent analytics queries (5 simultaneous `/api/v1/analytics` requests)
  2. Back-to-back pipeline triggers (stress the queue)
  3. Post-restore benchmark (prove restored state performs the same)

**Files allowed:**
- `backend/cmd/platformctl/` — benchmark command extensions
- `infra/scripts/benchmark_suite.sh`

**Completion signal:** `make benchmark` runs extended scenarios without regression.

### Task 5: Create Upgrade/Migration Runbook

**Required change:**
- Document in `docs/runbooks/upgrading.md`:
  1. How to upgrade from one commit to the next without data loss
  2. Migration ordering guarantees
  3. Backup-before-upgrade procedure
  4. Rollback procedure (restore from backup)
  5. What cannot be rolled back (schema migrations)

**Files allowed:**
- `docs/runbooks/upgrading.md` (NEW)
- `docs/runbooks/operator-manual.md` — link to new runbook

**Completion signal:** An operator can follow the runbook to upgrade a running instance.

## Escalation Triggers

- If ingestion refactor requires changes to `runtime.go` → escalate
- If the second domain needs new job types → document the proposal first
- If DuckDB concurrent access is fundamentally broken → document the constraint

## Completion Note Format

After completing, leave `prompts/v2-model2-completion.md` containing:
1. Files changed
2. What is now verifiably true
3. Verification commands and results
4. What was explicitly NOT changed
5. Any escalation items
