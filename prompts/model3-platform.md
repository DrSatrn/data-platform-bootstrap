# Model 3 Updated Prompt — UAT, README, and Cleanup

You are resuming Model 3 work on the `data-platform` repository. Your initial contract is complete. This is the follow-up pass now that Models 1 and 2 have completed.

## Read First

- `prompts/v1-remaining-work.md` (full remaining gap analysis)
- `prompts/model1-completion.md` (Model 1 results)
- `prompts/model2-completion.md` (Model 2 results)
- `uat-checklist.md` (the checklist you need to execute and annotate)

## Mission

Close out the v1 release preparation with 3 tasks: run UAT, update README, and clean up the repo root.

## Prerequisite

Before starting, confirm that `go test ./...` is passing (Model 1 should have fixed the external-tool test timeouts). If it's not passing, escalate and stop.

## Task 1: Run Full UAT

1. Run `make bootstrap` to stand up the full Docker Compose stack
2. Follow EVERY item in `uat-checklist.md`
3. Annotate each checklist item with ✅ PASS or ❌ FAIL and a brief note
4. If any items fail, document the failure clearly but do NOT attempt code fixes
5. Run `make down` when complete

**Completion signal:** Annotated `uat-checklist.md` with results.

## Task 2: Update `README.md`

Update the README to reflect the post-model state:
- URL routing now exists (all pages are bookmarkable)
- Frontend has proper loading/error states and an error boundary
- Dashboard page has been decomposed into reusable components
- Backend test coverage has been expanded (transforms, quality, integration skeleton)
- CI/CD workflow exists in `.github/workflows/ci.yml`
- The `ingestion/` stub has been removed

Keep changes minimal. Update facts, don't rewrite the entire file.

**Completion signal:** README accurately reflects the current repo state.

## Task 3: Repo Root Cleanup

Execute the cleanup proposal from your previous completion report:
- Delete `temp-model1-frontend-wire-plan.md` (stale, replaced by prompts/)
- Move `new-thread-eng-feedback.md` to `docs/decisions/` (closed contract)
- Leave all other files at root for now

Do NOT move `uat-checklist.md` or `infra-overview.md` until after the v1 tag.

**Completion signal:** Stale files removed or archived.

## Allowed Files

- `uat-checklist.md` (annotate)
- `README.md` (update)
- `temp-model1-frontend-wire-plan.md` (delete)
- `new-thread-eng-feedback.md` (move to `docs/decisions/`)
- `docs/` (all docs)
- `prompts/model3-completion.md` (update with addendum)

## Forbidden Files

- ❌ `backend/` (all code)
- ❌ `web/src/` (all frontend code)
- ❌ `infra/compose/docker-compose.yml`
- ❌ `Makefile`
- ❌ `packages/`

## Stop Conditions

- If `make bootstrap` fails, document the error and stop
- If UAT reveals critical failures, document them and stop
- If any change requires editing source code, escalate

## How To Report Completion

Update `prompts/model3-completion.md` with an addendum:

```markdown
## Follow-Up: UAT, README, and Cleanup

### Files Changed
- (list)

### UAT Results
- Total items tested: X
- Passed: X
- Failed: X
- Key failures: (list if any)

### Verification
- `make bootstrap` — PASS/FAIL
- `make down` — PASS/FAIL
```
