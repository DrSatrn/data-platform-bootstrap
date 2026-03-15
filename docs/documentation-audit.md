# Documentation Audit

## Scope

- Audit date: March 15, 2026
- Scope reviewed: 95 tracked, repo-authored markdown files
- Focus areas:
  - root docs
  - all 37 docs under `docs/`
  - package, backend, infra, prompt, and web markdown that affects navigation,
    onboarding, or contributor understanding
- Excluded:
  - vendored dependency docs under `web/node_modules`

Quality ratings below are the audit-start ratings used to decide rewrite
priority. Some of the worst docs were rewritten during this pass; those
rewrites are called out in the recommendations and in the completion report.

## Summary

### Top 5 Worst Docs That Needed Immediate Attention

1. `doc.md`
   Duplicate setup/UAT doc, stale auth model, stale cloud guidance, and no
   clear role in the current documentation set.
2. `uat-checklist.md`
   Written against old token assumptions and not rigorous enough for a real
   operator handoff.
3. `docs/runbooks/local-host-run.md`
   Explicitly described itself as a draft and duplicated
   `docs/runbooks/localhost-e2e.md`.
4. `docs/runbooks/access-matrix.md`
   Said it was additive and unverified instead of acting like a usable operator
   reference.
5. `docs/tutorials/trace-one-pipeline-complete.md`
   Described itself as a replacement draft instead of being a tutorial someone
   could actually follow.

### Documents That Should Be Merged

- `doc.md` into the `README.md` + `docs/runbooks/quickstart.md` path
- `docs/runbooks/local-host-run.md` into `docs/runbooks/localhost-e2e.md`
- `docs/tutorials/trace-one-pipeline.md` and
  `docs/tutorials/trace-one-pipeline-complete.md`
  eventually into one canonical tutorial
- root-level human navigation into `docs/README.md`

### Documents That Should Be Deleted Or Archived

These should not be presented as normal user docs:

- `guide-wire.md`
- `plan.md`
- `new-thread-eng-feedback.md`
- `temp-model1-frontend-wire-plan.md`
- `v1-review-coordination-plan.md`
- `prompts/*.md`

Recommendation:

- keep them if the repo still relies on them for AI coordination
- move them under a clearly internal location later
- do not leave them mixed into the main human-facing docs surface

### Missing Documentation That Should Exist

- a single docs landing page grouped by user journey
  - created in this pass as `docs/README.md`
- a short "what should I read first?" explanation that does not assume the
  reader knows the folder structure
- a clean distinction between:
  - first-run
  - packaged self-host path
  - host-run debug path
- a clearer explanation that product blueprint docs are internal design notes,
  not operator manuals

### Proposed Documentation Structure

Keep the folders, but change how readers enter them:

- `README.md`
  - product overview and "start here" pointer
- `docs/README.md`
  - human-first documentation map
- `docs/runbooks/`
  - setup, operation, troubleshooting, validation
- `docs/architecture/`
  - how the system is built
- `docs/tutorials/`
  - learn by following a complete path
- `docs/reference/`
  - stable contracts and lookups
- `docs/product/`
  - internal design notes only
- internal AI coordination docs
  - move out of root visibility over time

## Audit Tables

### Root Docs

| File | Purpose | Audience | Quality | Issues | Recommendation |
| --- | --- | --- | --- | --- | --- |
| `README.md` | repo overview and entry point | new user, operator, contributor | 4 | dense, but mostly current | Keep and lightly refine |
| `codex.md` | AI fresh-context handoff | AI agent | 2 | not human-facing, root clutter | Keep for AI use, move out of human path later |
| `contributing.md` | contributor onboarding | contributor | 4 | long, but useful and empathetic | Keep |
| `doc.md` | setup guide | new user | 1 | duplicate, stale, mixed concerns | Rewrite as pointer doc |
| `guide-wire.md` | parallel-work coordination | AI/reviewer | 2 | root clutter, not user-facing | Keep internal, move later |
| `infra-overview.md` | repo architecture walkthrough | learner, contributor | 4 | long but high-value | Keep |
| `new-thread-eng-feedback.md` | engineering review contract | AI/reviewer | 1 | not user-facing, root clutter | Archive from main docs path |
| `plan.md` | active build plan | AI/reviewer | 2 | not user-facing, root clutter | Keep internal, move later |
| `temp-model1-frontend-wire-plan.md` | model handoff | AI/reviewer | 1 | temporary by name and purpose | Archive or delete later |
| `uat-checklist.md` | manual product verification | operator, reviewer | 1 | stale and under-explained before rewrite | Rewrite |
| `v1-review-coordination-plan.md` | multi-model release coordination | AI/reviewer | 2 | root clutter, not for normal readers | Keep internal, move later |

### `docs/architecture`

| File | Purpose | Audience | Quality | Issues | Recommendation |
| --- | --- | --- | --- | --- | --- |
| `docs/architecture/README.md` | section index | contributor | 3 | minimal | Keep |
| `docs/architecture/data-model.md` | domain model summary | contributor | 3 | too thin for a true data model doc | Update later |
| `docs/architecture/runtime-wiring.md` | runtime ownership and state map | contributor, operator | 5 | dense but excellent | Keep as-is |
| `docs/architecture/service-boundaries.md` | subsystem boundaries | contributor | 4 | short but clear | Keep |
| `docs/architecture/system-overview.md` | high-level system summary | new contributor | 4 | could use a diagram | Keep, expand later |

### `docs/decisions`

| File | Purpose | Audience | Quality | Issues | Recommendation |
| --- | --- | --- | --- | --- | --- |
| `docs/decisions/0001-modular-monolith.md` | architecture decision record | contributor | 4 | concise but effective | Keep |
| `docs/decisions/README.md` | ADR section intro | contributor | 3 | very minimal | Keep |

### `docs/product`

| File | Purpose | Audience | Quality | Issues | Recommendation |
| --- | --- | --- | --- | --- | --- |
| `docs/product/README.md` | explain product docs | contributor | 2 | did not clearly separate design notes from user docs | Rewrite |
| `docs/product/management-console-blueprint.md` | management surface intent | product, frontend contributor | 3 | still carries staged-build language | Keep as design history, lightly update later |
| `docs/product/management-console-integration-map.md` | management module wiring map | contributor | 2 | references old model handoff language | Archive or rewrite as historical note later |
| `docs/product/operator-evidence-blueprint.md` | evidence UX intent | product, frontend contributor | 2 | stale filenames, staged language | Rewrite or archive |
| `docs/product/operator-followup-blueprint.md` | follow-up UX intent | product, frontend contributor | 2 | staged-build language, not user-facing | Rewrite or archive |
| `docs/product/opsview-ui-bridge.md` | backend/frontend bridge intent | contributor | 3 | still partly historical | Keep, maybe move to reference later |
| `docs/product/web-terminal-blueprint.md` | web terminal design intent | product, frontend contributor | 3 | product note, not user doc | Keep as internal design note |

### `docs/reference`

| File | Purpose | Audience | Quality | Issues | Recommendation |
| --- | --- | --- | --- | --- | --- |
| `docs/reference/README.md` | reference section intro | contributor | 3 | minimal | Keep |
| `docs/reference/external-tool-jobs.md` | external tool manifest contract | contributor, operator | 4 | good, but still narrow | Keep |
| `docs/reference/management-console-demo-assets.md` | demo payload note | frontend contributor | 2 | internal staging note, not broad reference | Archive or move under product/internal later |
| `docs/reference/operator-command-taxonomy.md` | command families | product, contributor | 3 | design-heavy, not yet canonical | Keep as internal reference |
| `docs/reference/opsexport-bundles.md` | export bundle note | backend contributor | 3 | internal and narrow | Keep |
| `docs/reference/opsview-read-models.md` | opsview read-model note | backend/frontend contributor | 3 | internal API note, not broad reference | Keep or move under architecture later |

### `docs/runbooks`

| File | Purpose | Audience | Quality | Issues | Recommendation |
| --- | --- | --- | --- | --- | --- |
| `docs/runbooks/README.md` | runbook section intro | operator | 3 | minimal and missing clear map | Keep |
| `docs/runbooks/access-matrix.md` | role matrix | operator | 1 | previously additive/unverified instead of usable | Rewrite |
| `docs/runbooks/backups.md` | backup and restore procedure | operator | 5 | strong and procedural | Keep |
| `docs/runbooks/benchmarking.md` | benchmark procedure | operator, contributor | 4 | good, could use example interpretation | Keep |
| `docs/runbooks/bootstrap.md` | Docker Compose startup | new user, operator | 4 | good but still one of several start docs | Keep |
| `docs/runbooks/config-reality.md` | explain env loading | contributor, operator | 3 | good truth doc but not a great first-read doc | Keep |
| `docs/runbooks/dbt-operator-checklist.md` | dbt quick checklist | operator | 3 | terse and tool-specific | Keep |
| `docs/runbooks/external-tool-troubleshooting.md` | troubleshoot external tools | operator | 4 | useful, but narrow | Keep |
| `docs/runbooks/local-host-run.md` | host-run path | contributor | 1 | draft duplicate of `localhost-e2e.md` | Rewrite as pointer or merge |
| `docs/runbooks/localhost-e2e.md` | host-run debug path | contributor, operator | 4 | solid but name is close to the duplicate doc | Keep |
| `docs/runbooks/operator-manual.md` | main operator handbook | operator | 4 | long, but valuable | Keep |
| `docs/runbooks/optional-external-tools.md` | optional dbt-style tools | operator, contributor | 4 | current and clear | Keep |
| `docs/runbooks/quickstart.md` | canonical first-run doc | new user | 4 | best start doc, but needs stronger docs-map support | Keep |

### `docs/tutorials`

| File | Purpose | Audience | Quality | Issues | Recommendation |
| --- | --- | --- | --- | --- | --- |
| `docs/tutorials/README.md` | tutorial section intro | learner | 3 | minimal | Keep |
| `docs/tutorials/making-changes.md` | change-by-change tutorial | contributor | 4 | dense, but useful | Keep |
| `docs/tutorials/trace-one-pipeline-complete.md` | detailed pipeline walkthrough | learner | 2 | was explicitly a draft and not canonical | Rewrite |
| `docs/tutorials/trace-one-pipeline.md` | shorter pipeline walkthrough | learner | 4 | useful but overlaps the longer version | Merge later |

### Backend, Infra, Package, Web, And Prompt Docs

These docs are mostly internal package overviews and coordination artifacts.
They still matter, but they are not the main source of user confusion. The
ratings below reflect that many are acceptable as internal docs even when they
would be poor public-facing docs.

| File | Purpose | Audience | Quality | Issues | Recommendation |
| --- | --- | --- | --- | --- | --- |
| `backend/README.md` | backend overview | contributor | 4 | good internal entry point | Keep |
| `backend/internal/admin/README.md` | admin package overview | contributor | 4 | concise | Keep |
| `backend/internal/analytics/README.md` | analytics package overview | contributor | 4 | concise | Keep |
| `backend/internal/app/README.md` | app wiring package overview | contributor | 4 | concise | Keep |
| `backend/internal/audit/README.md` | audit package overview | contributor | 4 | concise | Keep |
| `backend/internal/authz/README.md` | auth package overview | contributor | 4 | concise | Keep |
| `backend/internal/backup/README.md` | backup package overview | contributor | 4 | concise | Keep |
| `backend/internal/config/README.md` | config package overview | contributor | 4 | concise | Keep |
| `backend/internal/db/README.md` | db package overview | contributor | 4 | concise | Keep |
| `backend/internal/execution/README.md` | execution package overview | contributor | 4 | concise | Keep |
| `backend/internal/externaltools/README.md` | external tools package overview | contributor | 4 | concise | Keep |
| `backend/internal/manifests/README.md` | manifests package overview | contributor | 4 | concise | Keep |
| `backend/internal/metadata/README.md` | metadata package overview | contributor | 4 | concise | Keep |
| `backend/internal/observability/README.md` | observability package overview | contributor | 4 | concise | Keep |
| `backend/internal/orchestration/README.md` | orchestration package overview | contributor | 4 | concise | Keep |
| `backend/internal/python/README.md` | python runtime note | contributor | 4 | concise | Keep |
| `backend/internal/quality/README.md` | quality package overview | contributor | 4 | concise | Keep |
| `backend/internal/reporting/README.md` | reporting package overview | contributor | 4 | concise | Keep |
| `backend/internal/scheduler/README.md` | scheduler package overview | contributor | 4 | concise | Keep |
| `backend/internal/shared/README.md` | shared package overview | contributor | 4 | concise | Keep |
| `backend/internal/storage/README.md` | storage package overview | contributor | 3 | still speaks a bit prospectively | Update later |
| `backend/internal/transforms/README.md` | transforms package overview | contributor | 3 | still partly prospective | Update later |
| `backend/test/README.md` | backend integration test note | contributor | 3 | placeholder-like | Keep, expand later |
| `infra/README.md` | infra overview | contributor | 4 | concise | Keep |
| `infra/scripts/README.md` | scripts overview | contributor | 4 | concise | Keep |
| `packages/README.md` | packages overview | contributor | 4 | useful | Keep |
| `packages/demo/management_console/README.md` | demo asset note | frontend contributor | 3 | internal staging note | Keep internal |
| `packages/docs_templates/README.md` | docs template note | contributor | 3 | placeholder-like | Keep |
| `packages/external_tools/dbt_finance_demo/README.md` | dbt demo note | contributor | 4 | clear | Keep |
| `packages/external_tools/fixtures/README.md` | fixture note | contributor | 4 | clear | Keep |
| `packages/manifests/README.md` | manifest overview | contributor | 4 | useful | Keep |
| `packages/manifests/fixtures/external_tools/README.md` | external tool fixture note | contributor | 4 | clear | Keep |
| `packages/python/README.md` | python task overview | contributor | 4 | clear | Keep |
| `packages/sample_data/personal_finance/README.md` | sample data note | learner, contributor | 4 | clear | Keep |
| `packages/schemas/README.md` | schema directory note | contributor | 3 | placeholder-like | Keep |
| `packages/sql/README.md` | SQL directory overview | contributor | 4 | useful | Keep |
| `prompts/coordinator.md` | AI coordinator prompt | AI agent | 2 | not human-facing | Keep internal |
| `prompts/model1-backend.md` | AI model contract | AI agent | 2 | not human-facing | Keep internal |
| `prompts/model1-completion.md` | AI completion note | AI agent | 2 | not human-facing | Keep internal |
| `prompts/model2-completion.md` | AI completion note | AI agent | 2 | not human-facing | Keep internal |
| `prompts/model2-frontend.md` | AI model contract | AI agent | 2 | not human-facing | Keep internal |
| `prompts/model3-completion.md` | AI completion note | AI agent | 2 | not human-facing | Keep internal |
| `prompts/model3-platform.md` | AI model contract | AI agent | 2 | not human-facing | Keep internal |
| `prompts/v1-remaining-work.md` | AI/reviewer coordination | AI agent, reviewer | 2 | not human-facing | Keep internal |
| `web/README.md` | web app overview | contributor | 4 | good internal entry point | Keep |
| `web/src/features/management/README.md` | feature note | contributor | 3 | internal and specific | Keep |
| `web/tests/README.md` | web test note | contributor | 3 | placeholder-like | Keep |

## Root-Level Sprawl Proposal

### Keep At Root

- `README.md`
- `doc.md`
- `contributing.md`
- `infra-overview.md`
- `uat-checklist.md`

### Move Out Of Root Later

- `guide-wire.md`
- `plan.md`
- `new-thread-eng-feedback.md`
- `temp-model1-frontend-wire-plan.md`
- `v1-review-coordination-plan.md`
- `codex.md` if AI tooling still needs it but humans do not

### Why

The root should answer:

- what is this project
- how do I start it
- how do I contribute
- how do I verify it

It should not also be the attic for temporary model contracts and coordination
artifacts.
