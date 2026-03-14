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
	•	Optional PostgreSQL run snapshot mirror and migration command surface.
	•	Artifact inspection API plus Pipelines UI artifact browsing.
	•	Frontend build passes, backend tests pass, manifest validation passes, compose config resolves, and live localhost API, worker, scheduler, admin terminal, artifact API, and CLI checks passed.

What is still pending
	•	Make PostgreSQL a true primary control-plane repository instead of a write-only optional mirror.
	•	Persist artifact metadata to PostgreSQL alongside run snapshots.
	•	Replace the sample-data analytics execution path with DuckDB-backed transforms and curated query execution.
	•	Deepen scheduler coverage beyond the current cron subset and start honoring manifest timezone semantics explicitly.
	•	Add stronger smoke automation for Compose-backed Postgres migration coverage.

Important current architectural direction
	•	Do not reintroduce Prometheus or Grafana as core platform observability dependencies.
	•	Built-in operational surfaces should remain first-party wherever possible.
	•	Public repo safety matters: keep placeholders only in tracked env/config files, avoid real secrets, avoid publishing local-only services broadly, and prefer loopback bindings by default.
	•	Keep docs and top-of-file instructional comments up to date every time a file is modified.

Rolling Workstep Log

Latest completed workstep
	•	Added an optional PostgreSQL run snapshot mirror, migration command wiring, and run-scoped artifact inspection surfaces in the API, admin terminal, CLI, and Pipelines UI.
	•	Verified a real localhost end-to-end run on loopback with API, worker, and scheduler running against isolated `/tmp` runtime roots.
	•	Verified both scheduled and manual runs reaching `succeeded`, run artifacts being listed and downloaded over the API, admin-terminal artifact listing succeeding, and `platformctl remote "artifacts <run_id>"` succeeding against the live API.
	•	Added a repo-owned localhost smoke script, added a fail-fast port guard so it does not silently attach to an already-running API, wired `make smoke` to that path, and validated `PLATFORM_SMOKE_PORT=18083 make smoke` successfully.

Next workstep to execute
	•	Promote PostgreSQL from optional mirror to primary control-plane persistence for runs, queue state, and artifact metadata while keeping the local-first filesystem path as a safe fallback for offline development.

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
