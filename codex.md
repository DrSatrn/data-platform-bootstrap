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
	•	The finance slice now includes curated category spend and budget-variance marts plus a category-variance metric, not just the original monthly cashflow and savings-rate outputs.
	•	Scheduler cron evaluation now honors declared pipeline timezones and supports the cron subset needed by the current sample slice, including step fields and day-of-week matching.
	•	Packaged Compose deployment with a built frontend service image, one-shot migrations, health-gated startup, and a repo-owned `compose_smoke.sh` workflow that validates the hosted UI plus the API, worker, scheduler, analytics, quality, artifacts, and CLI paths.
	•	Frontend build passes, backend tests pass, manifest validation passes, compose config resolves, and live localhost API, worker, scheduler, admin terminal, artifact API, CLI, Compose-backed PostgreSQL checks, DuckDB-backed analytics/quality checks, and packaged Compose smoke checks passed.

What is still pending
	•	Normalize more control-plane metadata into first-class PostgreSQL tables beyond the current pragmatic snapshot and queue repositories.
	•	Expand the analytical layer beyond the first finance slice with freshness surfaces, more than one transform/metric family, and richer report editing workflows.
	•	Add a UI path for editing and saving dashboards instead of relying on API-level persistence alone.
	•	Broaden scheduler coverage beyond the currently supported cron subset if future slices need ranges, named weekdays, or more advanced catchup semantics.

Important current architectural direction
	•	Do not reintroduce Prometheus or Grafana as core platform observability dependencies.
	•	Built-in operational surfaces should remain first-party wherever possible.
	•	Public repo safety matters: keep placeholders only in tracked env/config files, avoid real secrets, avoid publishing local-only services broadly, and prefer loopback bindings by default.
	•	Keep docs and top-of-file instructional comments up to date every time a file is modified.

Rolling Workstep Log

Latest completed workstep
	•	Added file-backed saved dashboards seeded from `packages/dashboards`, then rewired the dashboard UI so it hydrates widgets through the reporting API and constrained analytics queries instead of hardcoded page logic.
	•	Expanded the finance slice with budget rules input, new curated marts for category spend and budget-versus-actual, and a new category-variance metric, all materialized through version-controlled DuckDB SQL.
	•	Updated the worker to ingest budget rules, materialize the richer marts, publish the new metric artifacts, and expose those outputs through the artifacts API.
	•	Made scheduler cron evaluation timezone-aware and added tests for timezone and day-of-week matching.
	•	Strengthened both localhost and Compose smoke workflows so they now verify the richer v2 analytics outputs and saved dashboard/reporting surfaces.

Next workstep to execute
	•	Move from API-level dashboard persistence to full UI editing flows, normalize more reporting/control-plane metadata into PostgreSQL, and broaden the product beyond the finance slice with additional domains or connector families.

Session Close Handoff

Use this section at the start of the next fresh-context session. It is the
session-close handoff, not just a rolling summary.

Current state at session end
	•	The platform is in a v2-ready state for the personal-finance slice and is fully runnable both through host-run binaries and the packaged Docker Compose deployment.
	•	The backend supports API, worker, scheduler, admin terminal, artifact inspection, constrained analytics serving, quality status, reporting APIs, and `platformctl`.
	•	The frontend now renders the dashboard from saved dashboard definitions plus constrained analytics queries rather than hardcoded page-specific data loading.
	•	Dashboard definitions are seeded from repo-managed YAML under `packages/dashboards` and persisted locally under the platform data root through the file-backed reporting store.
	•	The analytical layer now includes `mart_monthly_cashflow`, `mart_category_spend`, `mart_budget_vs_actual`, `metrics_savings_rate`, and `metrics_category_variance`.
	•	The worker ingests transactions, account balances, and budget rules, then materializes the richer marts and metrics through version-controlled DuckDB SQL.
	•	The scheduler now honors declared pipeline timezones and supports the cron subset needed by the current slice, including step fields and day-of-week matching.

Files changed in the final workstep
	•	Reporting persistence and API:
		•	`backend/internal/reporting/store.go`
		•	`backend/internal/reporting/handler.go`
		•	`backend/internal/reporting/store_test.go`
	•	Runtime and config wiring:
		•	`backend/internal/app/runtime.go`
		•	`backend/internal/config/config.go`
		•	`.env.example`
		•	`infra/compose/docker-compose.yml`
		•	`infra/docker/backend.Dockerfile`
	•	Analytics and execution:
		•	`backend/internal/analytics/service.go`
		•	`backend/internal/analytics/handler.go`
		•	`backend/internal/analytics/service_test.go`
		•	`backend/internal/execution/runner.go`
		•	`backend/internal/transforms/engine.go`
	•	Scheduler:
		•	`backend/internal/scheduler/service.go`
		•	`backend/internal/scheduler/service_test.go`
	•	Frontend dashboard surface:
		•	`web/src/features/dashboard/useDashboardData.ts`
		•	`web/src/pages/DashboardPage.tsx`
	•	New manifests, SQL, and sample data:
		•	`packages/sample_data/personal_finance/budget_rules.json`
		•	`packages/sql/bootstrap/raw_budget_rules.sql`
		•	`packages/sql/transforms/category_spend.sql`
		•	`packages/sql/transforms/budget_vs_actual.sql`
		•	`packages/sql/metrics/metrics_category_variance.sql`
		•	`packages/manifests/assets/raw_budget_rules.yaml`
		•	`packages/manifests/assets/mart_category_spend.yaml`
		•	`packages/manifests/assets/mart_budget_vs_actual.yaml`
		•	`packages/manifests/metrics/category_variance.yaml`
		•	`packages/manifests/pipelines/personal_finance_pipeline.yaml`
		•	`packages/dashboards/finance_overview.yaml`
	•	Smoke and docs:
		•	`infra/scripts/localhost_smoke.sh`
		•	`infra/scripts/compose_smoke.sh`
		•	`README.md`
		•	`docs/runbooks/bootstrap.md`
		•	`docs/runbooks/localhost-e2e.md`
		•	package READMEs for analytics, reporting, scheduler, manifests, and sample data

Validated at end of session
	•	`go test ./...` passed
	•	`go run ./cmd/platformctl validate-manifests` passed
	•	`npm run build` passed
	•	`git diff --check` passed
	•	Host-run smoke passed:
		•	`PLATFORM_SMOKE_PORT=18084 sh infra/scripts/localhost_smoke.sh`
	•	Packaged Compose smoke passed:
		•	`sh infra/scripts/compose_smoke.sh`
	•	The smoke runs now verify the richer v2 outputs, including:
		•	saved dashboard definitions from `/api/v1/reports`
		•	`mart_budget_vs_actual`
		•	`metrics_category_variance`
		•	run-scoped artifacts for the new marts and metrics

Important fixes made during this session
	•	A compile-time runtime wiring issue was fixed by making the dashboard store use the `reporting.Store` interface before selecting file-backed or in-memory implementations.
	•	A scheduler test expectation was corrected after verifying the timezone-aware day-of-week behavior was correct and the assertion was wrong.
	•	Both smoke scripts were fixed so they no longer infer success by naive run-status matching across the full pipeline list; they now wait for the expected run-scoped artifact to exist, which is a much safer completion signal.

Important repo/runtime truths
	•	PostgreSQL remains the preferred control-plane backend when available, but the platform still falls back to filesystem-backed persistence for local-first resilience.
	•	DuckDB is the analytical execution layer and is now central to transforms, metrics, analytics serving, and quality checks.
	•	The Compose web runtime is a packaged built service, not just a Vite dev server.
	•	The local frontend dev path with Vite still exists and is useful for UI iteration.
	•	Public-repo safety remains important:
		•	no real secrets should be committed
		•	`.env.example` contains placeholders only
		•	Compose bindings stay loopback-first
		•	Postgres is not published externally

Best next session starting point
	•	The cleanest next increment is UI-level dashboard editing and save flows.
	•	The backend/reporting boundary is now ready for that work:
		•	`/api/v1/reports` can list and save dashboards
		•	the file-backed reporting store persists saved dashboards already
		•	the dashboard page already hydrates widgets from saved definitions
	•	That means the next agent can focus on:
		•	adding dashboard-editing UI
		•	saving changes through the reporting API
		•	deciding whether to keep file-backed reporting persistence or promote it into PostgreSQL

Biggest remaining gaps
	•	There is still no UI for editing dashboards; persistence exists, but editing is API-level only.
	•	Reporting persistence is file-backed, not yet PostgreSQL-backed.
	•	Analytics is richer than before but still intentionally constrained; this is not an arbitrary BI query layer.
	•	Scheduler coverage is improved but still not a complete cron engine for all future cases.
	•	The platform still only proves one main domain slice; broader domain coverage is still future work.

Read these first in the next session
	•	`README.md`
	•	`backend/internal/app/runtime.go`
	•	`backend/internal/reporting/store.go`
	•	`backend/internal/analytics/service.go`
	•	`backend/internal/execution/runner.go`
	•	`backend/internal/scheduler/service.go`
	•	`packages/manifests/pipelines/personal_finance_pipeline.yaml`
	•	`packages/dashboards/finance_overview.yaml`
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
