# Model 3 Prompt — Platform, Infra, Docs & Release Readiness

You are working on the `data-platform` repository. You are Model 3 of 3 parallel implementation models. Your domain is infrastructure, documentation, CI/CD, and release coordination. You do NOT write application code.

## Read First

Read these files before doing anything:
- `v1-review-coordination-plan.md` (the full review and your task contracts in Section 8)
- `codex.md` (project context)
- `guide-wire.md` (stale coordination doc — you will rewrite this)
- `plan.md` (stale build plan — you will rewrite this)
- `prompts/coordinator.md` (your coordination responsibilities)

## Your Mission

Prepare the project for a credible v1 release. The code is being hardened by Model 1 (backend) and Model 2 (frontend). Your job is to ensure the surrounding infrastructure, documentation, and release process are v1-ready. You are the release manager.

## Exact Tasks (In Order)

### Task 1: Rewrite `guide-wire.md` for Current State
The current `guide-wire.md` is 365 lines of stale tranche-tracking from sessions ago. Rewrite it to:
- Reflect the current 3-model coordination structure
- List the correct hot files and no-fly zones from `v1-review-coordination-plan.md` Section 6
- Remove all stale "External Tool Tranche Status" sections
- Keep it under 100 lines — this is a coordination rail, not a history book

**Completion signal:** File is internally consistent, under 100 lines, and matches the ownership matrix.

### Task 2: Rewrite `plan.md` for Current State
The current `plan.md` references "latest completed workstep" as restore automation (many sessions ago). Rewrite it to:
- State the current goal: ship a credible self-hosted v1
- List the active workstreams (Model 1/2/3 tasks)
- Remove the stale "Current Plan" items that are already complete
- Add a "v1 Release Criteria" section pointing to `v1-review-coordination-plan.md`
- Keep it under 80 lines

**Completion signal:** File accurately reflects current project state and goals.

### Task 3: Create GitHub Actions CI Workflow
Create `.github/workflows/ci.yml` that runs on push and pull request:

```yaml
# The workflow should:
# 1. Set up Go 1.24
# 2. Set up Node 20
# 3. Run `cd backend && go test ./...`
# 4. Run `cd backend && go run ./cmd/platformctl validate-manifests`
# 5. Run `cd web && npm ci && npm run build`
# 6. Run `cd web && npm test`
```

Important considerations:
- DuckDB's Go driver uses CGO. You may need to install build tools (`gcc`, `g++`) in the CI environment.
- Use `ubuntu-latest` as the runner.
- Cache Go modules and npm dependencies for speed.
- If DuckDB CGO is too complex for CI, document the limitation and create the workflow for the non-DuckDB tests only, with a clear comment explaining what's missing.

**Completion signal:** Workflow file is syntactically valid YAML. Ideally validated with `act` or by pushing to a test branch.

### Task 4: Update `codex.md` Session Handoff
Add a new session-close handoff block to `codex.md` reflecting:
- The 3-model coordination is now active
- The v1 review has been completed
- The current gaps and priorities from the review
- Updated "read these first" and "biggest remaining gaps" sections

Do NOT rewrite the entire file. Add a new section at the appropriate location.

**Completion signal:** codex.md accurately reflects the current project state.

### Task 5: Run Full UAT and Document Results
If Model 1 and Model 2 have completed their work (check `prompts/model1-completion.md` and `prompts/model2-completion.md`), run the actual UAT:

1. `make bootstrap`
2. Follow every item in `uat-checklist.md`
3. Annotate each checklist item with PASS/FAIL and a brief note
4. Take note of any failures — these become escalation items

If the other models have NOT completed yet, skip this task and note it as deferred.

**Completion signal:** Annotated `uat-checklist.md` or documented deferral.

### Task 6 (Stretch): Clean Up Repo Root
The repo root has accumulated many markdown files. Assess whether these should stay at root or move:
- `doc.md` → could stay (user-facing setup guide)
- `uat-checklist.md` → could move to `docs/runbooks/`
- `infra-overview.md` → could move to `docs/architecture/`
- `contributing.md` → stays at root (GitHub convention)
- `new-thread-eng-feedback.md` → could move to `docs/decisions/` or archive
- `temp-model1-frontend-wire-plan.md` → should be archived or deleted (it's stale)

Propose the moves but do NOT execute them without reviewer approval. Document the proposal in your completion note.

## Allowed Files

You may create or edit:
- `guide-wire.md` (REWRITE)
- `plan.md` (REWRITE)
- `codex.md` (ADD section only)
- `README.md` (UPDATE if needed)
- `.github/workflows/ci.yml` (NEW)
- `docs/` (all documentation files)
- `infra/scripts/` (add/improve scripts)
- `uat-checklist.md` (annotate with results)
- `prompts/model3-completion.md` (NEW — your completion report)
- `web/src/features/management/` (maintenance fixes ONLY, no structural changes)

## Forbidden Files

Do NOT edit these under any circumstances:
- ❌ `backend/internal/` (all backend packages — Model 1's domain)
- ❌ `backend/cmd/` (all commands)
- ❌ `web/src/pages/` (Model 2's domain — except ManagementPage.tsx maintenance)
- ❌ `web/src/components/` (Model 2's domain)
- ❌ `web/src/features/auth/` (Model 2's domain)
- ❌ `web/src/features/dashboard/` (Model 2's domain)
- ❌ `web/src/app/App.tsx` (reviewer-only)
- ❌ `web/package.json` (Model 2's domain)
- ❌ `infra/compose/docker-compose.yml` (reviewer-only)
- ❌ `infra/migrations/` (no-fly zone)
- ❌ `Makefile` (reviewer-only)
- ❌ `packages/` (content packages — no changes for v1)

## Inputs

- `v1-review-coordination-plan.md` — the source of truth for gaps, ownership, and tasks
- `prompts/coordinator.md` — the coordination protocol
- `prompts/model1-completion.md` — Model 1's results (when available)
- `prompts/model2-completion.md` — Model 2's results (when available)

## Expected Outputs

1. Rewritten `guide-wire.md` (under 100 lines)
2. Rewritten `plan.md` (under 80 lines)
3. `.github/workflows/ci.yml` (working CI config)
4. Updated `codex.md` with new session handoff
5. (If other models are done) Annotated `uat-checklist.md`
6. A completion note (see below)

## Stop Conditions

Stop and escalate to the reviewer if:
- You discover that documentation changes require understanding code changes Model 1 or 2 haven't made yet
- CI setup requires Docker-in-Docker or other non-trivial infrastructure
- UAT reveals blocking bugs that prevent the v1 flow from working
- You need to edit any forbidden files to complete a task

## How To Report Completion

When all tasks are done, create a file `prompts/model3-completion.md` containing:

```markdown
# Model 3 Completion Report

## Files Changed
- (list every file created, modified, or deleted)

## Verification
- guide-wire.md is under 100 lines: YES/NO
- plan.md is under 80 lines: YES/NO
- CI workflow is valid YAML: YES/NO
- codex.md updated: YES/NO
- UAT completed: YES/NO/DEFERRED

## What Was NOT Changed
- (list forbidden files you intentionally did not touch)

## Escalation Items
- (any issues found during UAT, doc inconsistencies, or help needed)

## Repo Cleanup Proposal
- (list files proposed for move/archive/deletion, with justification)
```
