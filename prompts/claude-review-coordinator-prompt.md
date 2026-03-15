# Claude Review And Coordination Prompt

You are performing a deep code review and coordination pass for a serious
self-hosted data orchestration and analytics platform intended for a 3-person
on-prem team.

Your job has two parts:

1. Review the platform honestly against the original product requirements.
2. Coordinate the next work across 3 models so every requirement category is
   driven to `100%` in order, with trust, safety, and production-readiness
   prioritized over breadth.

Do not treat this as an MVP review. Treat it as a pre-production review for a
real internal platform that the team wants to use for on-prem data management,
orchestration, metadata, analytics serving, reporting, and operations.

## First Read These

- `README.md`
- `codex.md`
- `plan.md`
- `new-thread-eng-feedback.md`
- `docs/security.md`
- `docs/architecture/runtime-wiring.md`
- `docs/architecture/system-overview.md`
- `docs/architecture/service-boundaries.md`
- `docs/runbooks/operator-manual.md`
- `docs/runbooks/quickstart.md`
- `docs/runbooks/bootstrap.md`
- `docs/runbooks/localhost-e2e.md`
- `docs/runbooks/backups.md`
- `docs/runbooks/benchmarking.md`
- `docs/product/management-console-blueprint.md`
- `docs/product/management-console-integration-map.md`
- `docs/product/web-terminal-blueprint.md`
- `docs/product/operator-followup-blueprint.md`
- `docs/product/operator-evidence-blueprint.md`
- `docs/product/opsview-ui-bridge.md`
- `docs/reference/external-tool-jobs.md`
- `docs/reference/opsview-read-models.md`
- `prompts/api-error-audit.md`
- `prompts/v2-model1-completion.md`
- `prompts/v2-model2-completion.md`
- `prompts/v2-model3-completion.md`

If code inspection is allowed, prioritize these:

- `backend/internal/app/runtime.go`
- `backend/internal/authz/service.go`
- `backend/internal/execution/runner.go`
- `backend/internal/analytics/service.go`
- `backend/internal/metadata/handler.go`
- `backend/internal/orchestration/handler.go`
- `backend/internal/backup/restore.go`
- `backend/internal/opsview/handler.go`
- `web/src/app/App.tsx`
- `web/src/pages/DashboardPage.tsx`
- `web/src/pages/ManagementPage.tsx`
- `web/src/pages/DatasetsPage.tsx`

## Important Current Context

- `new-thread-eng-feedback.md` has been fully completed and should be treated as
  closed.
- Model 1 v2 security hardening has been completed.
- The last identified security bug was a raw pipeline validation text leak in
  the orchestration handler response, and that has now been fixed by returning a
  sanitized validation summary instead.
- The product is no longer a prototype. It is best described as a strong
  self-hosted late-beta / pre-production platform.

## Current Honest Status

### Product lifecycle

- prototype: exceeded
- MVP: exceeded
- advanced internal alpha: exceeded
- real beta / pre-v1 self-hosted product: yes
- enterprise-ready GA: not yet

Best current label:

- `late beta / pre-GA self-hosted platform`

### Readiness estimates

- strong self-hosted local-first v1 candidate: `90-93%`
- fully polished self-hosted product for real internal team use: `78-84%`
- enterprise-ready deployable platform: `60-70%`

### Business readiness estimates

- internal on-prem rollout to the 3-person team: `88-92% ready`
- polished self-hosted v1: `90-93% ready`
- business ready for stable internal production use: `85-90% ready`
- stricter enterprise-grade deployment: `70-78% ready`

## Status Against Original Requirements

### 1. Pipeline orchestration — `82%`

- pipeline definitions: `90%`
- task/job definitions: `82%`
- dependencies: `80%`
- schedules: `80%`
- manual runs: `95%`
- retries: `60%`
- monitoring/run visibility: `82%`
- execution history: `88%`
- auditability: `86%`

Main remaining gaps:
- retries are weaker than the rest of the control plane
- stronger concurrency and idempotency guarantees are still needed
- scheduling semantics can go deeper

### 2. Data ingestion and transformation — `78%`

- CSV/JSON/local files: `88%`
- mock API ingestion: `70%`
- database ingestion: `45%`
- raw layer: `90%`
- staging layer: `78%`
- intermediate layer: `75%`
- mart layer: `88%`
- metrics layer: `88%`
- SQL transforms: `88%`
- Python transforms: `72%`
- freshness tracking: `84%`
- quality tracking: `80%`

Main remaining gaps:
- database connectors are still light
- ingestion/runtime breadth is still narrow
- only one main domain slice is deeply proven

### 3. Integrated metadata and documentation — `88%`

- data source catalog: `85%`
- dataset registry: `90%`
- table/column docs: `84%`
- lineage: `78%`
- owners: `82%`
- refresh schedules: `76%`
- freshness: `88%`
- quality checks in metadata: `84%`
- documentation coverage: `92%`
- UI exposure: `88%`

Main remaining gaps:
- lineage depth
- ownership/governance richness
- more normalized DB-first metadata behavior

### 4. Monitoring, observability, and dashboards — `84%`

- service health: `92%`
- logs: `82%`
- metrics: `78%`
- pipeline health: `82%`
- dataset freshness: `88%`
- error rates/failure surfaces: `78%`
- quality status: `84%`
- diagnostics surfaces: `90%`
- operational dashboards: `84%`
- performance visibility: `78%`

Main remaining gaps:
- longer-term retention/history
- alerting/escalation depth
- stronger proof under load and failure

### 5. Analytics serving layer — `84%`

- curated datasets: `88%`
- metric endpoints: `86%`
- filters: `78%`
- aggregations: `84%`
- chart-ready responses: `88%`
- semantic/curated access control: `84%`
- separation from raw tables: `90%`

Main remaining gaps:
- richer semantic layer
- better dimensional exploration
- stronger analytics contract depth

### 6. Reporting and visualization app — `89%`

- KPI cards: `88%`
- dataset explorer: `86%`
- filters: `82%`
- saved reports/dashboards: `88%`
- charting: `86%`
- metadata-aware browsing: `86%`
- operational dashboards: `90%`
- overall BI feel: `88%`

Main remaining gaps:
- richer sharing/export flows
- more polished report-building ergonomics
- deeper in-app artifact preview

### 7. Version-controlled platform as code — `91%`

- manifests in repo: `95%`
- infra/config in repo: `88%`
- dashboards in repo: `90%`
- docs in repo: `95%`
- tests in repo: `90%`
- schemas/SQL/tasks in repo: `88%`
- reproducibility: `86%`

Main remaining gaps:
- some mutable runtime state remains intentionally DB-backed
- release discipline around change promotion can improve

### 8. CI/CD — `76%`

- local test flows: `90%`
- manifest validation: `92%`
- backend test coverage: `88%`
- frontend test/build: `82%`
- image builds: `76%`
- migrations: `84%`
- smoke tests: `92%`
- GitHub workflow maturity: `65%`
- reproducible deployment from source: `82%`

Main remaining gaps:
- release automation maturity
- stronger rollout/upgrade/rollback automation
- better environment matrix coverage

## Cross-Cutting Non-Functional Goals

- fast: `82%`
- reliable: `78%`
- modern: `90%`
- self-built where it matters: `91%`
- deeply instructive: `95%`

## Highest-Priority Remaining Product Gaps

These are the main reasons the platform is not yet at enterprise-ready GA:

1. orchestration retries and idempotency maturity
2. database connector and ingestion breadth
3. deeper lineage and governance richness
4. longer-term observability, alerting, and incident depth
5. richer semantic analytics exploration
6. stronger CI/CD release automation and rollback discipline
7. broader resilience and soak certification
8. fuller enterprise IAM/compliance depth beyond the current native auth model

## Your Review Goals

Please do all of the following:

1. assess whether the current repo state is truly ready for a 3-person on-prem
   internal rollout
2. identify which requirement categories are genuinely not yet at `100%`
3. challenge the percentages above if the code/docs do not justify them
4. identify the shortest path to getting every original requirement category to
   `100%`
5. coordinate work across 3 models so those requirement categories are closed
   in order of business impact

## Coordination Rules For The 3 Models

Split follow-up work across 3 models, but keep boundaries clean:

- Model 1:
  security, auth, API consistency, backend hardening, operational trust
- Model 2:
  frontend polish, reporting UX, operator workflows, in-app usability
- Model 3:
  infra, CI/CD, release automation, resilience, deployment and rollout safety

If another split is better, say so explicitly and explain why.

## Non-Negotiable Direction

- Do not widen feature breadth just for surface area.
- Drive the original requirement categories to `100%` one by one.
- Prioritize trust, followability, production safety, and operational realism.
- Prefer concrete closeout work over vague future-roadmap advice.
- Be explicit about what is proven versus inferred.
- Call out anything that would make rollout risky for a small on-prem team.

## Required Output

Produce:

1. executive summary
2. honest readiness assessment for:
   - 3-person on-prem internal rollout
   - small-team production usage
   - enterprise-grade rollout
3. requirement-by-requirement review against the original platform vision,
   including whether the stated percentages are fair
4. top 10 highest-risk gaps, ordered by severity
5. top 10 highest-leverage next actions
6. a coordinated 3-model work plan to drive each requirement category to `100%`
7. a final recommendation on whether to:
   - keep building before rehearsals
   - begin real deployment rehearsals now
   - or do a short stabilization pass first

Optimize for rigor, honesty, and usefulness.
