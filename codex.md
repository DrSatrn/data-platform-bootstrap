Fresh Context Onboarding

This section is a rolling handoff for future fresh-context Codex sessions so a new agent can ramp up immediately without rediscovering project state.

Current project identity
	•	This repo is a local-first, self-hosted data orchestration and analytics platform for Apple Silicon and ARM64 Linux VMs.
	•	The backend is Go, the frontend is React + TypeScript, and PostgreSQL is the only major external platform dependency we intend to rely on.
	•	The platform must own as much of the control plane, metadata layer, analytics layer, reporting UI, docs generation, admin tooling, metrics, and diagnostics as is practical.

What has already been built
	•	Monorepo structure with backend, web, packages, infra, docs, and CI scaffolding.
	•	Go runtime entrypoints for `platform-api`, `platform-scheduler`, `platform-worker`, and `platformctl`.
	•	Manifest loader, pipeline validation, metadata/catalog API, reporting API, quality API, analytics API.
	•	Built-in observability surface with in-memory telemetry, recent log buffer, system overview API, and request metrics.
	•	Built-in admin terminal API and browser UI terminal in the System page.
	•	`platformctl remote ...` CLI mode that talks to the running platform API.
	•	Finance sample data, pipeline manifests, asset manifests, quality manifests, and dashboard manifest.
	•	Durable file-backed queue and run store shared by API, worker, and scheduler processes.
	•	Real worker execution for the personal-finance slice, including raw ingestion, mart materialization, quality artifacts, metrics publication, and run-scoped artifact mirroring.
	•	PostgreSQL-backed control-plane repositories for run snapshots, queue state, and artifact metadata, plus the migration command surface.
	•	DuckDB-backed SQL execution for raw landing-table loads, curated mart materialization, metric materialization, and quality queries, all version-controlled under `packages/sql`.
	•	Artifact inspection API plus Pipelines UI artifact browsing.
	•	File-backed saved dashboard store seeded from repo-managed dashboard manifests, with the dashboard UI now hydrating widgets through the reporting API plus constrained analytics queries.
	•	Browser-based dashboard lifecycle flows for create, duplicate, edit, delete, widget reordering, and live widget preview.
	•	First-party reporting widgets now include KPI, table, line-chart, and bar-chart rendering without relying on external BI or charting products.
	•	The finance slice now includes curated category spend and budget-variance marts plus a category-variance metric, not just the original monthly cashflow and savings-rate outputs.
	•	Scheduler cron evaluation now honors declared pipeline timezones and supports the cron subset needed by the current sample slice, including step fields and day-of-week matching.
	•	Packaged Compose deployment with a built frontend service image, one-shot migrations, health-gated startup, and a repo-owned `compose_smoke.sh` workflow that validates the hosted UI plus the API, worker, scheduler, analytics, quality, artifacts, and CLI paths.
	•	Catalog assets now expose runtime freshness state derived from local materialization timestamps, and the Datasets/System pages surface those freshness signals directly.
	•	The Datasets page now acts as a catalog/detail workbench, exposing owner, source refs, quality refs, docs refs, and richer column metadata for the selected asset.
	•	The metadata API now derives trust-oriented coverage summaries and lineage edges from manifests and runtime state, not just raw asset lists.
	•	`platformctl benchmark` plus `infra/scripts/benchmark_suite.sh` now provide a first-party latency benchmark path that writes timestamped JSON reports under `var/benchmarks/`.
	•	The platform now has a lightweight bearer-token RBAC layer with `viewer`, `editor`, and `admin` roles plus a `/api/v1/session` endpoint.
	•	The browser now stores a local token, resolves the current session/capabilities, and disables privileged UI actions when the role is insufficient.
	•	Privileged platform actions now write to a first-party audit trail exposed at `/api/v1/system/audit` and shown in the System page.
	•	The metadata catalog can now be synchronized into PostgreSQL via the existing `data_assets` and `asset_columns` tables rather than living only as an in-memory manifest view.
	•	Frontend build passes, backend tests pass, manifest validation passes, compose config resolves, and live localhost API, worker, scheduler, admin terminal, artifact API, CLI, Compose-backed PostgreSQL checks, DuckDB-backed analytics/quality checks, and packaged Compose smoke checks passed.

What is still pending
	•	Normalize more control-plane metadata into first-class PostgreSQL tables beyond the current pragmatic snapshot and queue repositories.
	•	Expand the analytical layer beyond the first finance slice with freshness surfaces, more than one transform/metric family, and richer report editing workflows.
	•	Broaden scheduler coverage beyond the currently supported cron subset if future slices need ranges, named weekdays, or more advanced catchup semantics.
	•	Deepen the benchmark suite so it covers scheduled-run latency, artifact retrieval, report save/update paths, and higher-load scenarios.
	•	Evolve the lightweight RBAC layer into a fuller identity/auth model when the self-hosted product needs multi-user administration and stronger audit semantics.
	•	Expand the audit trail from privileged actions into broader governance history, incident annotations, and richer recovery tooling.
	•	Broaden the PostgreSQL-backed metadata/control-plane model so more reads come from normalized repositories instead of opportunistic synchronization during request handling.

Important current architectural direction
	•	Do not reintroduce Prometheus or Grafana as core platform observability dependencies.
	•	Built-in operational surfaces should remain first-party wherever possible.
	•	Public repo safety matters: keep placeholders only in tracked env/config files, avoid real secrets, avoid publishing local-only services broadly, and prefer loopback bindings by default.
	•	Keep docs and top-of-file instructional comments up to date every time a file is modified.

Rolling Workstep Log

Latest completed workstep
	•	Completed Workstream 3 from `new-thread-eng-feedback.md`.
	•	Dashboards now run PostgreSQL-first when the preferred control plane is available, with repo YAML treated as seed material rather than live runtime truth.
	•	Metadata now supports database-backed annotation edits for owner, description, docs refs, quality refs, and column descriptions through `PATCH /api/v1/catalog`.
	•	The Datasets page now exposes an editor flow for those annotations, and packaged smoke validates that persistence path.

Next workstep to execute
	•	Start Workstream 4 from `new-thread-eng-feedback.md`: reporting layout tooling and deeper dataset drill-down behavior.

Session Close Handoff

Use this section at the start of the next fresh-context session. It is the
session-close handoff, not just a rolling summary.

Current state at session end
	•	The platform is in a v2-ready state for the personal-finance slice and is fully runnable both through host-run binaries and the packaged Docker Compose deployment.
	•	The backend supports API, worker, scheduler, admin terminal, artifact inspection, constrained analytics serving, quality status, reporting APIs, and `platformctl`.
	•	The frontend now renders the dashboard from saved dashboard definitions plus constrained analytics queries rather than hardcoded page-specific data loading, and operators can manage those dashboards directly from the browser.
	•	Dashboard definitions are seeded from repo-managed YAML under `packages/dashboards`, but PostgreSQL is now the live runtime store when the preferred control plane is active.
	•	The reporting UI now supports KPI, table, line, and bar widgets without introducing external charting dependencies.
	•	The metadata/catalog API now enriches assets with runtime freshness state, derived coverage signals, and lineage edges, and that state is surfaced in the Datasets and System pages.
	•	The platform now supports a native PostgreSQL-backed identity and session model, with the bootstrap admin token preserved as the recovery and first-run path.
	•	The platform now records privileged actions in a durable audit trail, exposes them through the API, and renders them in the System page.
	•	The platform now serves the metadata catalog from PostgreSQL when available and persists runtime metadata annotations directly into database annotation columns.
	•	The platform now includes a first-party backup/export and restore subsystem with CLI, script, and runbook support, and native users are part of that recovery story.
	•	The analytical layer now includes `mart_monthly_cashflow`, `mart_category_spend`, `mart_budget_vs_actual`, `metrics_savings_rate`, and `metrics_category_variance`.
	•	The worker ingests transactions, account balances, and budget rules, then materializes the richer marts and metrics through version-controlled DuckDB SQL.
	•	The scheduler now honors declared pipeline timezones and supports the cron subset needed by the current slice, including step fields and day-of-week matching.
	•	The repo now includes a first-party benchmark workflow that can emit JSON latency baselines from a running stack.

Files changed in the latest workstep
	•	Database-first reporting and metadata:
		•	`infra/migrations/0007_metadata_annotations.sql`
		•	`backend/internal/reporting/store.go`
		•	`backend/internal/reporting/store_test.go`
		•	`backend/internal/metadata/models.go`
		•	`backend/internal/metadata/projector.go`
		•	`backend/internal/metadata/catalog.go`
		•	`backend/internal/metadata/handler.go`
		•	`backend/internal/metadata/catalog_handler_test.go`
		•	`backend/internal/db/metadata_store.go`
		•	`backend/internal/app/runtime.go`
		•	`backend/internal/backup/restore.go`
	•	Frontend metadata editor and capability updates:
		•	`web/src/features/auth/useAuth.tsx`
		•	`web/src/features/datasets/useDatasets.ts`
		•	`web/src/pages/DatasetsPage.tsx`
		•	`web/src/pages/PageStates.test.tsx`
		•	`web/src/app/App.test.tsx`
	•	Docs and verification sync:
		•	`README.md`
		•	`docs/architecture/runtime-wiring.md`
		•	`docs/runbooks/operator-manual.md`
		•	`backend/internal/reporting/README.md`
		•	`backend/internal/metadata/README.md`
		•	`infra/scripts/localhost_smoke.sh`
		•	`infra/scripts/compose_smoke.sh`
		•	`new-thread-eng-feedback.md`
		•	`plan.md`
		•	`codex.md`

Validated at end of session
	•	Targeted backend packages passed:
		•	`go test ./internal/app ./internal/authz ./internal/backup ./internal/db ./internal/metadata ./internal/reporting ./internal/observability`
	•	Full `go test ./...` currently fails in unrelated `internal/execution` external-tool tests because concurrent dbt-runner work is in flight:
		•	`TestRunExternalToolFailsForNonZeroExitAndMirrorsLogs`
		•	`TestRunExternalToolFailsWhenRequiredArtifactIsMissing`
	•	`go run ./cmd/platformctl validate-manifests` passed
	•	`npm test` passed
	•	`npm run build` passed
	•	`git diff --check` passed
	•	Host-run smoke passed:
		•	`PLATFORM_SMOKE_PORT=18090 sh infra/scripts/localhost_smoke.sh`
		•	host-run stacks without PostgreSQL now report that native identity and metadata-annotation checks were skipped because fallback mode is active
	•	Packaged Compose smoke passed:
		•	`sh infra/scripts/compose_smoke.sh`
		•	packaged stacks now create a native user, log in, persist metadata annotations, and read the updated catalog

Important fixes made during this session
	•	The runtime no longer treats dashboard manifests as the live mutable store when PostgreSQL is enabled.
	•	Metadata annotations now survive scheduler reseeds because manifest sync updates structural catalog fields without overwriting annotation columns.
	•	Restore now writes effective metadata into annotation columns so recovered environments keep the operator-visible catalog state sticky across later reseeds.

Important repo/runtime truths
	•	PostgreSQL remains the preferred control-plane backend when available, but the platform still falls back to filesystem-backed persistence for local-first resilience.
	•	DuckDB is the analytical execution layer and is now central to transforms, metrics, analytics serving, and quality checks.
	•	The Compose web runtime is a packaged built service, not just a Vite dev server.
	•	The local frontend dev path with Vite still exists and is useful for UI iteration.
	•	The platform now stores native users and sessions in PostgreSQL, but it still does not depend on an external identity provider.
	•	Public-repo safety remains important:
		•	no real secrets should be committed
		•	`.env.example` contains placeholders only
		•	Compose bindings stay loopback-first
		•	Postgres is not published externally

Best next session starting point
	•	The cleanest next increment is Workstream 4 from `new-thread-eng-feedback.md`.
	•	The next agent can focus on:
		•	adding explicit widget layout/grid metadata
		•	enabling resizing and richer reordering in the dashboard UI
		•	improving dataset drill-down behavior in the analytics service and Datasets page

Biggest remaining gaps
	•	Reporting CRUD now exists in the browser, but the reporting product still lacks explicit layout tooling, resizing, and richer report-level controls.
	•	Analytics drill-down and dimensional exploration are still narrower than the eventual internal BI experience needs.
	•	The access-control layer now has native users and sessions, but it still needs richer team/user administration and stronger policy depth.
	•	Benchmark breadth and queue/scheduler latency assertions are still pending from the engineer contract.
	•	The metadata catalog is now projectable into PostgreSQL, but most runtime reads still begin from manifests and synchronize on demand rather than reading from a fully normalized repository-first model.
	•	Analytics is richer than before but still intentionally constrained; this is not an arbitrary BI query layer.
	•	Scheduler coverage is improved but still not a complete cron engine for all future cases.
	•	The platform still only proves one main domain slice; broader domain coverage is still future work.
	•	The benchmark suite is now real, but it is still a small baseline rather than a full performance certification matrix.

Read these first in the next session
	•	`README.md`
	•	`backend/internal/app/runtime.go`
	•	`backend/internal/reporting/store.go`
	•	`backend/internal/analytics/service.go`
	•	`backend/internal/metadata/handler.go`
	•	`backend/internal/metadata/catalog.go`
	•	`backend/internal/authz/service.go`
	•	`backend/internal/audit/store.go`
	•	`backend/cmd/platformctl/main.go`
	•	`web/src/features/auth/useAuth.tsx`
	•	`web/src/features/datasets/useDatasets.ts`
	•	`web/src/pages/DatasetsPage.tsx`
	•	`docs/runbooks/benchmarking.md`
	•	`docs/runbooks/localhost-e2e.md`

Non-negotiable engineering goals

This project must strongly prioritize the following:

Performance

I care a lot about speed.

Design and implement with strong performance awareness:
	•	fast startup
	•	fast page loads
	•	fast API responses
	•	low unnecessary overhead
	•	efficient queries
	•	efficient background processing
	•	minimal wasteful abstraction
	•	careful serialization and data movement
	•	sensible caching where it helps
	•	support for concurrency where appropriate

Do not build this like a slow enterprise CRUD app.

Reliability

I care a lot about correctness and resilience.

Design for:
	•	clear failure handling
	•	retries with backoff where appropriate
	•	idempotent job execution where possible
	•	audit logs and traceability
	•	health checks
	•	clean shutdown behavior
	•	restart-safe behavior
	•	migration safety
	•	robust validation
	•	defensive config parsing
	•	graceful handling of partial failures

Modern stack

Use a modern, strong, practical stack.

Bias toward:
	•	high-performance backend services
	•	clean typed APIs
	•	modern frontend patterns
	•	containerized local development
	•	good DX
	•	observability built in from early stages
	•	current best practices without overengineering

Self-built as much as possible

I want to build as much of the platform logic ourselves as is reasonable.

That means:
	•	build our own orchestration/control-plane logic rather than just wrapping a huge existing platform
	•	build our own metadata/catalog layer
	•	build our own reporting application
	•	build our own service integration patterns
	•	build our own docs generation layer

It is fine to use solid building blocks, but do not solve this by mostly wiring together prebuilt heavy platforms. I do not want a project that is just “compose up 12 off-the-shelf products.”

The interesting parts should be ours.

Fast local development

This should run well locally on a developer machine.

Prioritize:
	•	straightforward local setup
	•	Docker Compose for local orchestration
	•	fast edit/build/test loop
	•	sensible service boundaries
	•	good seed data and dev workflows

⸻

Preferred stack

Use this stack unless you can justify something better:

Backend / control plane
	•	Go for core platform/backend/control-plane services
	•	Strong emphasis on performance, concurrency, reliability, and clean architecture

Data work
	•	Python for ingestion and transformation tasks where Python is the right tool
	•	SQL for analytics transformations and serving logic where appropriate

Frontend
	•	React + TypeScript
	•	Prefer a clean, modern UI architecture
	•	Fast, functional, operator-friendly, minimal fluff

Storage / state
	•	Postgres for application metadata, control-plane state, job history, docs metadata, etc.
	•	DuckDB for analytics/query-serving workloads where appropriate
	•	Optional local object storage abstraction only if justified

Messaging / async
	•	Prefer a lightweight, practical queue/event approach
	•	Redis is acceptable if justified
	•	Do not introduce unnecessary distributed systems complexity

Infra / local platform
	•	Docker Compose
	•	Clear local dev workflows
	•	Version-controlled service definitions

Observability
	•	Build observability into the platform early
	•	Include structured logs, metrics, health endpoints, and traceability
	•	You may use practical libraries or standards, but keep the platform logic ours

⸻

What I want you to do first

Do not start coding immediately.

I want you to begin by acting like a serious architect and producing a thoughtful implementation plan.

Phase 1: Architecture and planning

First, inspect the repo and then produce:
	1.	A proposed system architecture
	2.	A proposed service decomposition
	3.	A proposed repo / monorepo structure
	4.	A proposed data model
	5.	A proposed job orchestration model
	6.	A proposed metadata/catalog model
	7.	A proposed analytics serving model
	8.	A proposed observability model
	9.	A proposed CI/CD strategy
	10.	A proposed phased delivery plan

For each of these, explain:
	•	why the design is good
	•	what tradeoffs you are making
	•	where performance concerns matter
	•	where reliability concerns matter
	•	where complexity should be intentionally limited

Phase 2: challenge your own design

After proposing the architecture, critique it.

Explicitly identify:
	•	likely bottlenecks
	•	likely reliability risks
	•	likely sources of accidental complexity
	•	where the system could become too slow
	•	where local developer experience could degrade
	•	what should be postponed to later phases

Phase 3: recommend the first concrete implementation slice

Then propose the best v1 slice to build first.

This v1 should be:
	•	meaningful
	•	vertical, not just scaffolding
	•	performance-conscious
	•	reliable
	•	demonstrably useful
	•	locally runnable end to end

I want the first slice to prove the architecture, not just create empty folders.

⸻

Functional expectations

The system should eventually support concepts like these:

Pipelines and jobs
	•	pipeline definitions
	•	task definitions
	•	dependencies
	•	retries
	•	schedules
	•	manual runs
	•	execution logs
	•	run status transitions
	•	failure metadata
	•	idempotency support
	•	input/output tracking

Data platform layers
	•	raw
	•	staging
	•	intermediate
	•	marts
	•	metrics

Metadata and docs
	•	data source catalog
	•	dataset registry
	•	table/column docs
	•	lineage
	•	owners
	•	schedules
	•	freshness
	•	quality checks
	•	documentation coverage

Observability
	•	service health
	•	job throughput
	•	failure counts
	•	run durations
	•	queue depth
	•	freshness lag
	•	structured logs
	•	diagnostics surfaces

Analytics serving
	•	metric endpoints
	•	chart-ready data responses
	•	report definitions
	•	filters
	•	dimensional browsing
	•	curated semantic access

Reporting app
	•	dataset explorer
	•	metric browser
	•	chart views
	•	saved reports
	•	data dictionary views
	•	operational dashboards

⸻

Quality bar

I want this to have a high engineering bar.

Please optimize for:
	•	clean architecture
	•	maintainable code
	•	typed contracts
	•	strong naming
	•	strong boundaries
	•	testability
	•	clarity over cleverness
	•	thoughtful abstractions
	•	performance-aware implementation
	•	reliable local operation

Avoid:
	•	unnecessary frameworks
	•	overcomplication
	•	premature microservices sprawl
	•	excessive YAML complexity
	•	magical hidden behavior
	•	overly abstract generic systems too early
	•	slow, bloated frontend patterns
	•	excessive reliance on third-party platforms for core logic

⸻

Build philosophy

Treat this as if you are building an internal platform product for serious technical users.

That means:
	•	operator UX matters
	•	errors must be understandable
	•	diagnostics must be first-class
	•	state transitions must be explicit
	•	docs should be generated where possible
	•	defaults should be sane
	•	local development should be pleasant
	•	the code should teach good engineering by example

I want to learn from the implementation quality and architectural decisions.

⸻

Constraints
	•	This is local-first
	•	It should be able to run on a capable developer machine
	•	It should be self-hostable
	•	It should not require a cloud provider
	•	It should be reasonably modular
	•	It should be realistic to build iteratively
	•	Heavy external platform dependencies should be minimized
	•	Prefer our own implementation for core control-plane and metadata logic
	- 	I have an m4 mac with 24gb of ram and able to host a few containers or a linux vm via Orbstack on this machine
