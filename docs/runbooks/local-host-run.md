# Local Host-Run Path

This runbook is a draft canonical path for developers who want to run the
platform directly on the host instead of through Docker Compose.

It is intentionally narrow:

- backend processes run from `backend/`
- web dev server runs from `web/`
- PostgreSQL is optional for first-pass local work

## Recommended First Use

If you only want the fastest verified success, stop here and run:

```bash
make smoke
```

Use the rest of this document only when you specifically want host-run
processes for debugging or iteration.

## Preconditions

- Go installed
- Node and npm installed
- Python 3 installed
- host C/C++ toolchain available for DuckDB CGO builds

Optional:

- PostgreSQL running locally if you want the preferred database-backed control
  plane mode

## Working Directory Assumptions

Run backend commands from:

```bash
cd backend
```

Run web commands from:

```bash
cd web
```

These working directories matter because backend defaults are repo-relative.

## Configuration Options

You have two safe choices:

### Option A: rely on built-in local defaults

This is the lowest-friction host-run path when you are running from `backend/`.

### Option B: create a local env file

Create a repo-root `.env` or `.env.local` if you want explicit tokens or custom
paths. Current backend config loads env files if present, but existing shell
environment variables still win.

See:

- `docs/runbooks/config-reality.md`

## Start The API

```bash
cd backend
PLATFORM_ADMIN_TOKEN=local-dev-admin-token go run ./cmd/platform-api
```

Expected result:

- process starts without exiting
- API responds on `http://127.0.0.1:8080/healthz`

## Start The Worker

In a second terminal:

```bash
cd backend
PLATFORM_ADMIN_TOKEN=local-dev-admin-token go run ./cmd/platform-worker
```

Expected result:

- worker starts without exiting
- log output indicates polling has started

## Start The Scheduler

In a third terminal:

```bash
cd backend
PLATFORM_ADMIN_TOKEN=local-dev-admin-token go run ./cmd/platform-scheduler
```

Expected result:

- scheduler starts without exiting
- refresh logs appear periodically

## Start The Web Dev Server

In a fourth terminal:

```bash
cd web
npm install
npm run dev
```

Expected result:

- Vite serves the UI on `http://127.0.0.1:3000`

## Verify Health

```bash
curl http://127.0.0.1:8080/healthz
```

Expected result:

- JSON health payload with a healthy status

## Trigger A Pipeline

Browser path:

- open `http://127.0.0.1:3000`
- paste an `editor` or `admin` token into the sidebar token field
- go to `Pipelines`
- click `Run now`

CLI path:

```bash
cd backend
PLATFORM_API_BASE_URL=http://127.0.0.1:8080 \
PLATFORM_ADMIN_TOKEN=local-dev-admin-token \
go run ./cmd/platformctl remote trigger personal_finance_pipeline
```

## Expected Artifacts

After a successful run, expect files under the repo-local `var/` tree such as:

- `var/data/raw/raw_transactions.csv`
- `var/data/mart/mart_monthly_cashflow.json`
- `var/data/metrics/metrics_savings_rate.json`
- `var/artifacts/runs/<run_id>/...`

## If This Fails, Check This Next

- API fails to start:
  verify Go toolchain and DuckDB CGO prerequisites
- worker never processes runs:
  verify API and worker are using the same effective data root
- browser cannot trigger runs:
  verify you are using an `editor` or `admin` token
- CLI remote command fails:
  verify you are using an `admin` token
- data paths look wrong:
  inspect `docs/runbooks/config-reality.md` and confirm your working directory
