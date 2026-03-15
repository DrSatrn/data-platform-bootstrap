# Infrastructure & Architecture Overview

> **Audience:** Someone who can write scripts but has never worked on a large-scale software project. This document will walk you through the entire codebase as if you were sitting next to a senior engineer who was explaining every folder, every system, and every decision.

---

## Table of Contents

1. [The 30-Second Summary](#the-30-second-summary)
2. [How Software Projects Are Organized (Primer)](#how-software-projects-are-organized-primer)
3. [Top-Level Directory Map](#top-level-directory-map)
4. [backend/ — The Brain](#backend--the-brain)
5. [web/ — The Face](#web--the-face)
6. [packages/ — The Content Library](#packages--the-content-library)
7. [infra/ — The Plumbing](#infra--the-plumbing)
8. [docs/ — The Knowledge Base](#docs--the-knowledge-base)
9. [var/ — The Runtime Scratchpad](#var--the-runtime-scratchpad)
10. [How The System Talks To Itself](#how-the-system-talks-to-itself)
11. [Homegrown vs Off-The-Shelf](#homegrown-vs-off-the-shelf)
12. [The Data Lifecycle (End-To-End Walkthrough)](#the-data-lifecycle-end-to-end-walkthrough)
13. [Key Concepts Glossary](#key-concepts-glossary)

---

## The 30-Second Summary

This project is a **self-hosted data platform**. Think of it like building your own miniature version of tools like Airflow + dbt + Metabase + DataHub, all in one repository, running on your laptop.

It does four things:
1. **Orchestrates** data pipelines (schedules them, runs them, tracks success/failure).
2. **Transforms** raw data into useful analytical tables (monthly cashflow, budget variance, etc).
3. **Serves** that transformed data through a web UI with dashboards, charts, and a dataset explorer.
4. **Manages** itself with built-in health checks, audit logs, backups, and an admin terminal.

The backend is written in **Go**, the frontend in **React + TypeScript**, and data transformation happens via **DuckDB** (an embedded analytical database) and **Python** scripts.

---

## How Software Projects Are Organized (Primer)

If you've only written scripts before, here's the mental model shift:

- **A script** is one file that does one thing. You run it, it finishes.
- **A service** is a program that starts up, stays running, and waits for requests (like a web server).
- **A monorepo** is a single Git repository that contains multiple services, libraries, and configuration that all work together.

This project is a **monorepo** containing:
- A Go backend (3 long-running services + 1 CLI tool)
- A React frontend (1 web application)
- SQL files, Python scripts, YAML configuration, Docker definitions, and documentation

When you run `make bootstrap`, Docker Compose starts **6 containers** that work together as one platform.

---

## Top-Level Directory Map

```
data-platform/
├── backend/          # All Go source code (API server, worker, scheduler, CLI)
├── web/              # React + TypeScript frontend application
├── packages/         # Content: SQL queries, Python tasks, YAML manifests, sample data, dashboards
├── infra/            # Infrastructure: Docker, Compose, database migrations, shell scripts
├── docs/             # Architecture docs, runbooks, tutorials, decision records
├── var/              # Runtime data (gitignored): DuckDB files, artifacts, backups
├── Makefile          # Developer command shortcuts (make smoke, make bootstrap, etc.)
├── codex.md          # Instructions for AI coding agents working on this repo
├── README.md         # Project overview and navigation
├── .env.example      # Environment variable template for host-run mode
└── .env.compose.example  # Environment variable template for Docker Compose mode
```

---

## backend/ — The Brain

This is where all the server-side logic lives. It's written in Go, a language designed for building fast, concurrent network services.

### How Go Projects Are Structured

Go projects follow a convention:
- `cmd/` contains **entry points** — the `main()` functions that start each binary.
- `internal/` contains **private packages** — reusable logic that only this project can import.
- `pkg/` contains **public packages** — logic that could theoretically be imported by other projects.
- `go.mod` and `go.sum` are dependency management files (like `package.json` for Node).

### backend/cmd/ — The Four Executables

```
cmd/
├── platform-api/        # The HTTP API server — the "front door" for all requests
├── platform-scheduler/  # The cron-like loop that checks if pipelines are due to run
├── platform-worker/     # The execution engine that actually runs pipeline jobs
└── platformctl/         # A CLI tool for admin tasks (validate manifests, run backups, etc.)
```

**Why four separate binaries?** Even though they share the same codebase, separating them means you can:
- Scale them independently (run 3 workers but only 1 scheduler)
- Restart one without affecting the others
- Debug one process in isolation

In Docker Compose, each of these becomes its own container.

### backend/internal/ — The 22 Internal Packages

This is the core of the platform. Each folder is a self-contained "package" responsible for one concern:

#### Core Runtime
| Package | What It Does | Off-the-shelf alternative |
|---------|-------------|--------------------------|
| `app/` | **Wires everything together at startup.** Reads config, connects to databases, builds the HTTP router, and decides which persistence backends to use. | Spring Boot (Java), Express middleware (Node) |
| `config/` | **Loads environment variables** and provides typed settings to every other package. | Viper (Go), dotenv (Node) |

#### Orchestration & Execution
| Package | What It Does | Off-the-shelf alternative |
|---------|-------------|--------------------------|
| `orchestration/` | **Defines the data model** for Pipelines, Jobs, Runs, and Queues. Manages state transitions (pending → running → succeeded/failed). | Apache Airflow, Dagster, Prefect |
| `execution/` | **Actually runs each job.** Dispatches to the right handler based on job type (ingest, SQL transform, Python task, quality check, metric publish). | Airflow operators, Dagster ops |
| `scheduler/` | **Evaluates cron expressions** on a timer loop and enqueues pipeline runs when they're due. | Airflow scheduler, cron daemon, Celery Beat |
| `manifests/` | **Reads YAML pipeline definitions** from disk and turns them into Go structs. | dbt project loader, Airflow DAG parser |

#### Data Processing
| Package | What It Does | Off-the-shelf alternative |
|---------|-------------|--------------------------|
| `transforms/` | **Wraps DuckDB** to execute SQL transformations. Materializes raw tables, runs transform queries, and queries results. | dbt (SQL transforms), Spark |
| `python/` | **Runs Python scripts as subprocesses** with structured input/output contracts. The Go service stays in control. | Airflow PythonOperator, subprocess calls |
| `ingestion/` | **Handles raw data landing.** Copies sample CSV/JSON files into the data root. | Fivetran, Airbyte, custom ETL scripts |
| `quality/` | **Runs data quality SQL checks** (e.g., "are there duplicate transactions?") and produces pass/fail artifacts. | Great Expectations, dbt tests, Soda |

#### Metadata & Catalog
| Package | What It Does | Off-the-shelf alternative |
|---------|-------------|--------------------------|
| `metadata/` | **Builds a data catalog** from YAML asset manifests. Computes lineage graphs, coverage metrics, and freshness signals. Optionally projects into PostgreSQL. | DataHub, Amundsen, OpenMetadata |

#### Analytics & Reporting
| Package | What It Does | Off-the-shelf alternative |
|---------|-------------|--------------------------|
| `analytics/` | **Serves curated analytical queries** over DuckDB. Intentionally constrained — you can't run arbitrary SQL, only query known datasets with known filters. | Metabase, Superset, Looker |
| `reporting/` | **Manages saved dashboards.** Stores dashboard definitions (widgets, filters, presets) in files and/or PostgreSQL. | Metabase dashboards, Grafana |

#### Security & Governance
| Package | What It Does | Off-the-shelf alternative |
|---------|-------------|--------------------------|
| `authz/` | **Resolves bearer tokens into roles** (viewer, editor, admin) and enforces access control on API endpoints. | Auth0, Keycloak, Casbin |
| `audit/` | **Records privileged actions** (pipeline triggers, dashboard saves, admin commands) into an append-only log. | Audit logging libraries, ELK stack |

#### Operations & Admin
| Package | What It Does | Off-the-shelf alternative |
|---------|-------------|--------------------------|
| `observability/` | **Provides health checks, request metrics, and a recent log buffer.** All built in-process, no external dependencies. | Prometheus + Grafana, Datadog, New Relic |
| `admin/` | **Powers the browser-based admin terminal.** Accepts text commands and returns structured responses. | Custom admin panels, Django admin |
| `backup/` | **Creates portable backup bundles** containing PostgreSQL exports, DuckDB files, manifests, and checksummed manifests. | pg_dump scripts, Velero (Kubernetes) |

#### Storage & Database
| Package | What It Does | Off-the-shelf alternative |
|---------|-------------|--------------------------|
| `storage/` | **Manages run-scoped artifacts** (the files produced by each pipeline run). Can optionally index them in PostgreSQL. | S3, MinIO, artifact stores |
| `db/` | **PostgreSQL integration.** Runs migrations, provides typed repositories for runs, queue state, dashboards, audit events, and metadata. | GORM, sqlx, Prisma |
| `shared/` | **Small shared utilities** used across packages. | stdlib helpers |
| `externaltools/` | **Adapters for optional external tools** (dbt, dlt, PySpark). Runs them as subprocesses and collects their artifacts. | Native dbt CLI, Spark submit |

---

## web/ — The Face

The frontend is a React + TypeScript single-page application (SPA) built with Vite (a modern build tool).

### Directory Structure

```
web/src/
├── app/              # Top-level app shell, routing, and layout
├── pages/            # One file per major UI screen
│   ├── DashboardPage.tsx    # Dashboard viewer, editor, widget builder
│   ├── DatasetsPage.tsx     # Data catalog explorer with schema + lineage
│   ├── MetricsPage.tsx      # Metric browser showing KPIs and trends
│   ├── PipelinesPage.tsx    # Pipeline list, run history, manual triggers
│   └── SystemPage.tsx       # System health, audit logs, admin terminal
├── features/         # Feature-specific logic (hooks, state, API calls)
│   ├── auth/         # Token storage, session resolution, capability checks
│   ├── dashboard/    # Dashboard CRUD, widget state management
│   ├── datasets/     # Catalog data fetching, profiling hooks
│   ├── management/   # Admin terminal integration
│   ├── metrics/      # Metric catalog hooks
│   ├── pipelines/    # Pipeline list, trigger, run polling
│   └── system/       # System overview, audit feed, logs
├── components/       # Reusable UI elements (buttons, cards, charts)
├── lib/              # Shared utilities (API client, formatters)
└── styles/           # Global CSS
```

### How the Frontend Talks to the Backend

The frontend makes HTTP requests to the backend API. In Docker Compose, a small Node.js proxy server (`web/server.mjs`) sits in front of the built React app and forwards `/api/*` requests to the `platform-api` container. This is why you access the UI on port `3000` but the API lives on port `8080`.

### What the Frontend Does NOT Do

- It does not run SQL queries directly
- It does not talk to PostgreSQL or DuckDB
- It does not manage files on disk
- All data comes through the backend API

---

## packages/ — The Content Library

This is where the actual "business logic content" lives — the pipeline definitions, SQL queries, Python scripts, dashboard layouts, and sample data. Think of it as the "what" while the backend is the "how."

### packages/manifests/ — Pipeline & Asset Definitions

```
manifests/
├── pipelines/        # YAML files defining pipeline DAGs (what jobs run in what order)
│   ├── personal_finance_pipeline.yaml     # The main finance pipeline
│   └── personal_finance_dbt_pipeline.yaml # An alternative pipeline using dbt
├── assets/           # YAML files describing each data asset (table/dataset)
│   ├── raw_transactions.yaml              # Raw CSV transactions
│   ├── staging_transactions_enriched.yaml  # Python-enriched staging table
│   ├── mart_monthly_cashflow.yaml         # Final analytical mart
│   └── ... (8 total asset definitions)
├── metrics/          # YAML files defining computed metrics
├── quality/          # YAML files defining data quality checks
└── owners/           # YAML files mapping assets to responsible teams
```

**What's a manifest?** It's a YAML file that **declares** what something is without containing the implementation. The pipeline manifest says "run job A, then job B, then job C" but doesn't contain the actual SQL or Python code. That lives in `packages/sql/` and `packages/python/`.

**Off-the-shelf alternative:** dbt's `schema.yml` + `sources.yml`, Airflow DAG files.

### packages/sql/ — Version-Controlled SQL

```
sql/
├── bootstrap/        # DDL statements to create initial DuckDB tables
├── transforms/       # SQL queries that build analytical tables
│   ├── monthly_cashflow.sql              # Aggregates income vs expenses by month
│   ├── category_spend.sql                # Summarizes spending by category
│   ├── budget_vs_actual.sql              # Compares actual spend to budgets
│   └── intermediate_category_monthly_rollup.sql
├── metrics/          # SQL queries that compute KPI metrics
└── quality/          # SQL queries that check data integrity
```

Every SQL file is version-controlled. The worker reads these files at runtime and executes them against DuckDB. This is similar to how dbt manages SQL transformations, except here the orchestration and execution engine is homegrown.

### packages/python/ — Python Data Tasks

```
python/
└── tasks/            # Python scripts invoked by the Go worker as subprocesses
```

Python scripts handle tasks that are awkward in pure SQL, like enriching raw transaction CSVs with category mappings or profiling dataset schemas. The Go worker calls Python via `subprocess`, passing structured JSON input and reading structured JSON output. The Go process always stays in control.

### packages/dashboards/ — Dashboard Definitions

```
dashboards/
└── finance_overview.yaml    # The default dashboard with widgets, presets, and filters
```

Dashboard layouts are defined as YAML and seeded into the platform on first startup. After that, users can edit them through the browser and the changes persist to the database/filesystem.

### packages/sample_data/ — Seed Data

```
sample_data/
└── personal_finance/
    ├── transactions.csv       # 12 months of fake personal transactions
    ├── account_balances.json  # Bank account balance snapshots
    └── budget_rules.json      # Monthly budget targets per category
```

This is fake data used to prove the platform works end-to-end. The worker "ingests" this data as if it were coming from a real source.

---

## infra/ — The Plumbing

Everything needed to build, run, and operate the platform as containers.

### infra/compose/ — Docker Compose

```
compose/
└── docker-compose.yml    # Defines 6 services: postgres, migrate, api, scheduler, worker, web
```

This single file defines the entire runtime topology. When you run `make bootstrap`:
1. **postgres** starts first (the relational database)
2. **migrate** runs once to create/update database tables, then exits
3. **api** starts the HTTP server (waits for migrate to finish)
4. **scheduler** starts the cron loop
5. **worker** starts the job execution loop
6. **web** starts the frontend proxy server

### infra/docker/ — Dockerfiles

```
docker/
├── backend.Dockerfile     # Multi-stage Go build → slim runtime image
└── web.Dockerfile         # Node.js build → production server image
```

**Multi-stage builds** mean the final container image is small. The first stage compiles the Go code (needs the full Go toolchain), the second stage copies just the compiled binary into a minimal Linux image.

### infra/migrations/ — PostgreSQL Schema

```
migrations/
├── 0001_initial_schema.sql          # Core tables: run_snapshots, artifact_snapshots, dashboards
├── 0002_control_plane_mirror.sql    # Tables for mirroring filesystem state into Postgres
├── 0003_queue_requests.sql          # The durable job queue table
├── 0004_audit_events.sql            # The append-only audit log table
└── 0005_metadata_projection.sql     # Tables for data_assets and asset_columns
```

Migrations run in order and are idempotent. Each file creates new tables or columns. This is a standard pattern for evolving database schemas over time without losing data.

**Off-the-shelf alternative:** golang-migrate, Flyway, Alembic (Python), Prisma Migrate.

### infra/scripts/ — Operational Shell Scripts

```
scripts/
├── localhost_smoke.sh     # Boots the full stack locally, runs assertions, tears down
├── compose_smoke.sh       # Same but against the Docker Compose stack
├── backup_snapshot.sh     # Creates a backup bundle via platformctl
├── benchmark_suite.sh     # Measures API response latencies
├── restore_drill.sh       # Extracts a backup to a tmp directory (safe drill)
└── restore_e2e.sh         # Full end-to-end restore test
```

These are the "make sure everything works" scripts. `localhost_smoke.sh` is particularly important — it starts the API, worker, and scheduler, triggers a pipeline, waits for it to complete, checks the outputs, and tears everything down. If this passes, the platform is working.

---

## docs/ — The Knowledge Base

```
docs/
├── architecture/     # How the system is designed (runtime wiring, data flow diagrams)
├── decisions/        # ADRs (Architecture Decision Records) — why choices were made
├── product/          # Product requirements and feature descriptions
├── reference/        # API reference, configuration reference
├── runbooks/         # Step-by-step operational procedures
│   ├── quickstart.md         # The one canonical "start here" document
│   ├── bootstrap.md          # Docker Compose startup procedure
│   ├── localhost-e2e.md      # Host-run debugging procedure
│   ├── operator-manual.md    # Day-to-day operations guide
│   ├── backups.md            # Backup and recovery procedures
│   └── benchmarking.md       # Performance measurement guide
└── tutorials/        # Learning-oriented walkthroughs
    ├── making-changes.md         # How to modify the platform
    └── trace-one-pipeline.md     # Follow one pipeline from trigger to output
```

---

## var/ — The Runtime Scratchpad

This directory is **gitignored** — it only exists on your machine when the platform is running.

```
var/
├── artifacts/        # Files produced by pipeline runs (per-run subdirectories)
├── backups/          # .tar.gz backup bundles
├── benchmarks/       # JSON benchmark reports
├── data/             # The data lake: raw/, staging/, intermediate/, mart/, metrics/, quality/
│   ├── raw/          # Ingested CSV/JSON files
│   ├── staging/      # Python-enriched staging outputs
│   ├── intermediate/ # Intermediate rollup tables
│   ├── mart/         # Final analytical tables
│   ├── metrics/      # Computed KPI metric files
│   └── quality/      # Data quality check results
└── duckdb/           # The DuckDB database file (platform.duckdb)
```

This is the **medallion architecture** in action — data flows from messy raw files through increasingly refined layers until it reaches clean, query-ready marts.

---

## How The System Talks To Itself

Here's the complete request flow when you click "Run Pipeline" in the browser:

```
┌─────────────┐     HTTP POST      ┌─────────────┐
│   Browser   │ ──────────────────→ │   Web Proxy  │  (port 3000)
│  (React UI) │                     │  (server.mjs) │
└─────────────┘                     └──────┬────────┘
                                           │ forwards /api/* requests
                                           ▼
                                    ┌─────────────┐
                                    │  Platform    │  (port 8080)
                                    │    API       │
                                    └──────┬────────┘
                                           │
                    ┌──────────────────────┤──────────────────────┐
                    │                      │                      │
                    ▼                      ▼                      ▼
             ┌────────────┐       ┌──────────────┐      ┌──────────────┐
             │ PostgreSQL │       │  File Queue  │      │  Audit Log   │
             │ (run state)│       │ (run request)│      │ (who did it) │
             └────────────┘       └──────┬───────┘      └──────────────┘
                                         │
                                         │ worker polls for new jobs
                                         ▼
                                  ┌─────────────┐
                                  │  Platform    │
                                  │   Worker     │
                                  └──────┬───────┘
                                         │
                    ┌────────────────────┤────────────────────┐
                    │                    │                    │
                    ▼                    ▼                    ▼
             ┌────────────┐      ┌────────────┐     ┌────────────┐
             │  Ingest    │      │  DuckDB    │     │  Python    │
             │  (copy     │      │  (run SQL  │     │  (enrichment│
             │   files)   │      │   transforms)    │   scripts) │
             └────────────┘      └────────────┘     └────────────┘
                                         │
                                         ▼
                                  ┌─────────────┐
                                  │  var/data/   │
                                  │  (mart JSON  │
                                  │   artifacts) │
                                  └─────────────┘
```

Meanwhile, on a separate timer:

```
┌──────────────┐    every 15s     ┌──────────────┐
│   Platform   │ ───────────────→ │  Manifest    │
│  Scheduler   │                  │  Loader      │
└──────┬───────┘                  └──────────────┘
       │
       │ if cron matches → enqueue run request
       ▼
┌──────────────┐
│  File/PG     │
│  Queue       │
└──────────────┘
```

And when you view a dashboard:

```
┌─────────────┐    GET /api/v1/analytics    ┌─────────────┐
│   Browser   │ ──────────────────────────→ │  Platform    │
│  (chart     │                             │    API       │
│   widget)   │                             └──────┬───────┘
└─────────────┘                                    │
                                                   ▼
                                            ┌─────────────┐
                                            │  DuckDB     │
                                            │  (query     │
                                            │   mart table)│
                                            └──────┬───────┘
                                                   │
                                                   ▼
                                            ┌─────────────┐
                                            │  JSON rows  │
                                            │  → browser  │
                                            │  → chart    │
                                            └─────────────┘
```

---

## Homegrown vs Off-The-Shelf

This is one of the most interesting aspects of this project. Here's a comprehensive comparison:

### What Was Built From Scratch (And What It Replaces)

| Capability | This Platform | Industry Standard |
|-----------|--------------|------------------|
| **Pipeline orchestration** | Go scheduler + worker + file/PG queue | Apache Airflow, Dagster, Prefect |
| **SQL transformations** | DuckDB engine wrapper + version-controlled SQL | dbt |
| **Data catalog** | YAML manifests + enrichment logic + PG projection | DataHub, Amundsen, OpenMetadata |
| **Dashboard/reporting** | React UI + Go reporting API + constrained analytics | Metabase, Superset, Looker, Grafana |
| **Data quality** | SQL-based checks + quality manifests | Great Expectations, dbt tests, Soda |
| **Observability** | In-process metrics, logs, health checks | Prometheus + Grafana, Datadog |
| **Admin terminal** | Browser-based command shell | Custom admin panels |
| **Backup/recovery** | Go-based tarball creator + verification | pg_dump, Velero, cloud snapshots |
| **RBAC/auth** | Static bearer tokens + role resolver | Auth0, Keycloak, Ory Kratos |
| **Audit logging** | Append-only event store (file + PG) | ELK stack, audit libraries |

### What Was NOT Built From Scratch

| Component | What's Used | Why |
|-----------|------------|-----|
| **Relational database** | PostgreSQL | Battle-tested, widely understood, handles control-plane state well |
| **Analytical database** | DuckDB | Embedded, fast for OLAP, no server needed, great for local-first |
| **Frontend framework** | React + TypeScript | Industry standard for SPAs |
| **Build tooling** | Vite (frontend), Go toolchain (backend) | Fast, well-supported |
| **Container runtime** | Docker / OrbStack | Standard for local dev |
| **Container orchestration** | Docker Compose | Simple, declarative, good for local + small deployments |

### Why Build So Much From Scratch?

1. **Learning value.** Building orchestration teaches you how orchestration works. Wrapping Airflow teaches you how to configure Airflow.
2. **Control.** No framework decides your data model, your persistence strategy, or your UI.
3. **Performance.** No JVM startup, no Python GIL, no 2GB Docker images. The Go binary starts in milliseconds.
4. **Local-first.** Everything runs on one laptop without cloud dependencies.

### Where Off-The-Shelf Might Be Better

1. **Airflow** has a massive ecosystem of pre-built connectors. If you need to ingest from 50 different sources, writing custom ingest jobs gets tedious.
2. **dbt** has a huge community and tooling ecosystem around SQL transformations. The homegrown transform engine works but doesn't have dbt's testing, documentation generation, or package management.
3. **Metabase/Superset** offer drag-and-drop chart building. The homegrown dashboard editor is functional but less polished.
4. **Auth0/Keycloak** handle real multi-user authentication with password management, SSO, MFA, etc. The static token approach is fine for a personal tool but won't scale to teams.

---

## The Data Lifecycle (End-To-End Walkthrough)

Here's what happens when the `personal_finance_pipeline` runs, step by step:

### Step 1: Ingestion (Job Type: `ingest`)
Three raw files are copied from `packages/sample_data/personal_finance/` into `var/data/raw/`:
- `transactions.csv` → `raw_transactions.csv`
- `account_balances.json` → `raw_account_balances.json`  
- `budget_rules.json` → `raw_budget_rules.json`

### Step 2: Raw Table Materialization
The DuckDB engine reads the raw CSV/JSON files and creates in-memory tables:
- `raw_transactions` (from CSV)
- `raw_account_balances` (from JSON)
- `raw_budget_rules` (from JSON)

### Step 3: Python Enrichment (Job Type: `transform_python`)
A Python script reads raw transactions and enriches them (e.g., categorizing transactions, normalizing dates). The output lands as `staging_transactions_enriched.json` in `var/data/staging/`.

### Step 4: SQL Transformations (Job Type: `transform_sql`)
Four SQL files execute in dependency order against DuckDB:

1. `intermediate_category_monthly_rollup.sql` → Groups transactions by category and month
2. `monthly_cashflow.sql` → Computes income vs expenses vs savings rate
3. `category_spend.sql` → Summarizes actual spending by category
4. `budget_vs_actual.sql` → Joins actual spend against budget rules

Each produces a JSON artifact under `var/data/mart/`.

### Step 5: Quality Checks (Job Type: `quality_check`)
SQL queries check data integrity:
- Are there uncategorized transactions?
- Are there duplicate transactions?

Results are written to `var/data/quality/`.

### Step 6: Metric Publication (Job Type: `publish_metric`)
Two metrics are computed from the mart tables:
- `metrics_savings_rate` (savings rate by month)
- `metrics_category_variance` (budget variance by category and month)

Results are written to `var/data/metrics/`.

### Step 7: Artifact Mirroring
Every output from every job step is also copied into `var/artifacts/runs/<run_id>/` so you can inspect exactly what each run produced.

### Step 8: Dashboard Serving
The analytics API queries DuckDB for the mart tables and serves filtered, chart-ready JSON to the React frontend. The dashboard widgets render this as tables, KPIs, line charts, and bar charts.

---

## Key Concepts Glossary

| Term | Meaning |
|------|---------|
| **Manifest** | A YAML file that declares "what" something is (a pipeline, an asset, a metric) without containing the implementation code. |
| **Pipeline** | A named workflow containing ordered jobs with dependencies. |
| **Job** | A single unit of work within a pipeline (ingest a file, run a SQL query, execute a Python script). |
| **Run** | One execution attempt of a pipeline. Has a unique ID and status (pending, running, succeeded, failed). |
| **DAG** | Directed Acyclic Graph — a graph of dependencies where no job can depend on itself (no cycles). |
| **Medallion Architecture** | A data organization pattern: Raw → Staging → Intermediate → Mart → Metrics. Each layer is cleaner. |
| **Control Plane** | The part of the system that manages state (run history, queue, dashboards). Lives in PostgreSQL and/or the filesystem. |
| **Data Plane** | The part of the system that processes data (DuckDB, Python scripts, file I/O). |
| **RBAC** | Role-Based Access Control — users get a role (viewer, editor, admin) that determines what they can do. |
| **Projection** | Copying data from one format into another (e.g., manifest YAML → PostgreSQL rows) so it can be queried differently. |
| **Idempotent** | An operation that produces the same result whether you run it once or ten times. Important for retries. |
| **CGO** | A Go mechanism for calling C code. DuckDB's Go driver uses CGO, which is why you need C build tools (Xcode on Mac). |
| **Multi-stage build** | A Docker technique where one container compiles code and another runs the compiled binary, keeping the final image small. |
| **Seed** | Default data loaded on first startup (e.g., the default finance dashboard). |
