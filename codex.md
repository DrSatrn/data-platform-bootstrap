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
	•	Frontend build passes, backend tests pass, manifest validation passes, compose config resolves, and live localhost API/admin terminal checks passed.

What is still pending
	•	True end-to-end pipeline execution path that materializes assets and records run history persistently across API and worker processes.
	•	Persistent run state store instead of the current in-memory orchestration store.
	•	A triggerable pipeline API and UI flow for manual runs.
	•	Worker consumption of queued run requests and real execution of ingest/transform/quality/metric jobs.
	•	Localhost runbook updates for full end-to-end verification after the execution path is wired.

Important current architectural direction
	•	Do not reintroduce Prometheus or Grafana as core platform observability dependencies.
	•	Built-in operational surfaces should remain first-party wherever possible.
	•	Public repo safety matters: keep placeholders only in tracked env/config files, avoid real secrets, avoid publishing local-only services broadly, and prefer loopback bindings by default.
	•	Keep docs and top-of-file instructional comments up to date every time a file is modified.

Rolling Workstep Log

Latest completed workstep
	•	Implemented a durable file-backed run store and file-backed queue shared by API and worker processes.
	•	Implemented real worker execution for the personal-finance pipeline, including raw ingestion, mart materialization, quality artifacts, and metric publication.
	•	Added manual run triggering through the Pipelines UI, admin terminal, and `platformctl remote`.
	•	Verified a real localhost end-to-end run: queue, worker, run history, materialized artifacts, and analytics response all succeeded.

Next workstep to execute
	•	Wire PostgreSQL-backed control-plane repositories, add richer run detail and artifact inspection in the UI, and move the scheduler from catalog refresh into real scheduled queueing.

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
