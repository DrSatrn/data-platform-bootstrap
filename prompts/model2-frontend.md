# Model 2 Reassigned Prompt — Major Documentation Review & Design Pass

You are working on the `data-platform` repository. Your frontend code contract is complete. You are now reassigned to a comprehensive documentation review and design pass.

Your audience is someone who can write scripts but has never worked as a professional software engineer. The docs must be mentoring-level — clear, exhaustive, assumes nothing.

## Read First

- `codex.md` (project context and architectural direction)
- `README.md` (current entry point)
- `doc.md` (local setup guide)
- `infra-overview.md` (architecture walkthrough)
- `contributing.md` (contributor guide)
- `uat-checklist.md` (testing guide)

Then scan ALL 37 docs under `docs/`:
- `docs/architecture/` (4 files)
- `docs/decisions/` (2 files)
- `docs/product/` (6 files)
- `docs/reference/` (5 files)
- `docs/runbooks/` (12 files)
- `docs/tutorials/` (3 files)

## Your Mission

The documentation across this repo has grown organically across many sessions. It's scattered, inconsistent, sometimes stale, sometimes duplicated, and often written for AI agents rather than human users. Your job is to audit every doc, identify what's broken, and produce a concrete redesign plan + execute the highest-value improvements.

## Phase 1: Documentation Audit

Read every `.md` file in the repo (root + `docs/` tree) and produce a documentation audit report at `docs/documentation-audit.md`.

For EACH document, assess:

| Field | Description |
|-------|-------------|
| **File** | Path |
| **Purpose** | What this doc is trying to do |
| **Audience** | Who this is written for (user? operator? AI agent? developer?) |
| **Quality** | 1-5 rating (1=useless, 5=excellent) |
| **Issues** | Stale content, wrong audience, duplicates another doc, too vague, too dense, missing sections, etc. |
| **Recommendation** | Rewrite / Update / Merge into X / Delete / Keep as-is |

Then add a summary section with:
- Top 5 worst docs that need immediate attention
- Documents that should be merged (duplicates or heavily overlapping)
- Documents that should be deleted (stale, superseded, AI-only)
- Missing documentation that should exist but doesn't
- Proposed new documentation structure (if the current tree is wrong)

## Phase 2: Rewrite the Highest-Impact Docs

After the audit, rewrite the **top 5 worst docs** identified in Phase 1. For each rewrite:

- Write for a human user who is seeing this project for the first time
- Use clear section headers, numbered steps, expected outputs
- Include "What success looks like" after every procedural section
- Include "If something goes wrong" troubleshooting after every risky step
- Remove any AI-agent-specific language ("the next model should...", "fresh-context session", etc.)
- Keep runbooks copy-paste safe — every command should be exact
- Keep architecture docs diagram-rich — use ASCII art or mermaid where it helps

## Phase 3: Design a Documentation Information Architecture

Produce a `docs/README.md` rewrite that serves as a documentation map. It should:
- Explain the documentation structure
- Tell readers which doc to read based on what they're trying to do
- Link to every document with a 1-line description
- Group docs by user journey, not by folder convention

Proposed structure (adapt as you see fit):

```
## I want to...

### Get this running on my machine
→ doc.md (setup guide)
→ docs/runbooks/quickstart.md (quick start)
→ docs/runbooks/bootstrap.md (Docker Compose path)

### Understand how this works
→ infra-overview.md (full architecture walkthrough)
→ docs/architecture/system-overview.md
→ docs/tutorials/trace-one-pipeline.md

### Test it / verify it works
→ uat-checklist.md
→ docs/runbooks/localhost-e2e.md

### Contribute code
→ contributing.md
→ docs/tutorials/making-changes.md

### Operate and maintain it
→ docs/runbooks/operator-manual.md
→ docs/runbooks/backups.md
→ docs/runbooks/benchmarking.md

### Understand design decisions
→ docs/decisions/
→ docs/product/
```

## Phase 4: Spot-Check Specific Problem Areas

These are known documentation problems based on the project review:

1. **`docs/runbooks/quickstart.md` vs `doc.md` vs `docs/runbooks/bootstrap.md`** — Three docs covering "how to start the platform." There should be ONE clear starting point. Consolidate or clearly differentiate.

2. **`docs/runbooks/local-host-run.md` vs `docs/runbooks/localhost-e2e.md`** — Confusingly similar names. Clarify or merge.

3. **`docs/product/*.md`** — These are internal blueprints for features. Most reference "Model 2" or "staged modules." Audit whether any of these are useful as user-facing docs or should be archived as historical decision records.

4. **`docs/reference/management-console-demo-assets.md`** and **`docs/reference/opsview-read-models.md`** — These read like internal API notes. Assess whether they belong in `reference/` or should be archived.

5. **Root-level `.md` sprawl** — The repo root has: `README.md`, `doc.md`, `codex.md`, `plan.md`, `guide-wire.md`, `contributing.md`, `infra-overview.md`, `uat-checklist.md`, `new-thread-eng-feedback.md`, `temp-model1-frontend-wire-plan.md`, `v1-review-coordination-plan.md`. That's 11 markdown files at root. Propose which should stay, which should move under `docs/`, and which should be deleted.

## Allowed Files

- ALL files under `docs/` (create, edit, delete, reorganize)
- `README.md` (update)
- `doc.md` (update or propose merge)
- `infra-overview.md` (update)
- `contributing.md` (update)
- `docs/documentation-audit.md` (NEW — your audit report)

## Forbidden Files

- ❌ `backend/` (all code)
- ❌ `web/src/` (all frontend code)
- ❌ `infra/compose/docker-compose.yml`
- ❌ `Makefile`
- ❌ `packages/`
- ❌ `codex.md` (Model 3 owns this)
- ❌ `plan.md` (Model 3 owns this)
- ❌ `guide-wire.md` (Model 3 owns this)

## Stop Conditions

- If you find a doc that references code behavior you can't verify without running the system, mark it as "UNVERIFIED — needs UAT confirmation" rather than guessing
- If two docs fundamentally contradict each other, flag both and escalate rather than choosing one
- If reorganizing requires changing links in code files (like import paths or references), document the needed changes but don't make them

## How To Report Completion

Create `prompts/model2-completion.md` (overwrite the existing no-work-needed version) with:

```markdown
# Model 2 Completion Report — Documentation Pass

## Phase 1: Audit
- Total docs reviewed: X
- Top 5 worst docs: (list)
- Docs recommended for deletion: (list)
- Docs recommended for merge: (list)
- Missing docs identified: (list)

## Phase 2: Rewrites
- (list each rewritten doc with before/after quality rating)

## Phase 3: Information Architecture
- docs/README.md rewritten: YES/NO

## Phase 4: Problem Areas
- quickstart consolidation: (what you did)
- product docs audit: (what you did)
- root-level sprawl: (proposal)

## Files Changed
- (list every file created, modified, or deleted)
```
