# Quickstart

This is the canonical first-run document for the repo. Use this file first if
you want the platform running without guessing.

## Choose Your Path

Use exactly one of these paths:

1. `make smoke`
   Recommended first success path. Fastest verified result. Uses an isolated
   `/tmp` runtime and does not depend on you hand-starting services.
2. `make bootstrap`
   Recommended packaged self-host path. Starts the Compose stack on loopback.
3. Host-run binaries
   Use only if you are debugging individual processes or changing local
   runtime wiring.

If you are new to the repo, start with `make smoke`.

## Configuration Reality

### Host-run binaries

- The Go backend now auto-loads `.env` and `.env.local`.
- It looks in the current working directory and its parent directory.
- The tracked [.env.example](/Users/streanor/Documents/Playground/data-platform/.env.example)
  is the host-run example.
- Copy it to `.env` when you want `go run` and built binaries to pick up local
  configuration automatically.

### Compose

- Compose does not depend on `.env.example`.
- The tracked Compose file already contains safe local defaults.
- If you want to override Compose-only values such as tokens or the Postgres
  password, copy
  [.env.compose.example](/Users/streanor/Documents/Playground/data-platform/.env.compose.example)
  to `.env.compose`.
- The Make targets `make up`, `make down`, and `make bootstrap` automatically
  pass `--env-file .env.compose` when that file exists.

## Role Requirements

Use this table instead of guessing:

- Anonymous: `GET /healthz`, `GET /api/v1/session`
- Viewer: read product surfaces such as pipelines, catalog, analytics, quality,
  reports, artifacts, system overview, logs, and audit
- Editor: manual pipeline triggers and dashboard create/update/delete
- Admin: admin terminal and `platformctl remote ...`

The normal path is:

- use `PLATFORM_ADMIN_TOKEN` only to bootstrap or recover the environment
- create native users once PostgreSQL is available
- sign in through `/api/v1/session` or the browser login form
- use the returned session token for day-to-day access

If PostgreSQL is unavailable, host-run mode falls back to bootstrap-token-only
auth and will skip the native session smoke checks.

## First Success Path: Smoke

Working directory:

```sh
cd /Users/streanor/Documents/Playground/data-platform
```

Command:

```sh
make smoke
```

Expected success result:

- command exits `0`
- output includes `localhost smoke test passed`
- output prints:
  - `api_url=http://127.0.0.1:<port>`
  - `manual_run_id=...`
  - `backup_path=...`
  - `runtime_root=/tmp/...`

What this verifies:

- API, worker, and scheduler boot
- scheduled and manual run paths work
- run artifacts are created
- admin terminal responds
- CLI path works
- backup create and verify both work

If this fails, check next:

1. inspect the printed `logs_root`
2. open `api.log`, `worker.log`, and `scheduler.log`
3. if the port is already taken, rerun with `PLATFORM_SMOKE_PORT=<unused-port> make smoke`

## Packaged Self-Host Path: Compose

Working directory:

```sh
cd /Users/streanor/Documents/Playground/data-platform
```

Optional config override:

```sh
cp .env.compose.example .env.compose
```

Command:

```sh
make bootstrap
```

Expected success result:

- `docker compose ps` shows healthy `api` and running `worker`, `scheduler`, and `web`
- web is reachable at `http://127.0.0.1:3000`
- API health is reachable at `http://127.0.0.1:8080/healthz`

Quick verification:

```sh
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/api/v1/session
```

Bootstrap a viewer user in packaged mode:

```sh
curl -X POST \
  -H "Authorization: Bearer local-dev-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"username":"viewer-demo","display_name":"Viewer Demo","role":"viewer","password":"viewer-password"}' \
  http://127.0.0.1:8080/api/v1/admin/users
```

Then sign in:

```sh
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"viewer-demo","password":"viewer-password"}' \
  http://127.0.0.1:8080/api/v1/session
```

After you sign in with a viewer-or-higher token, open the System page and
confirm the `Source Of Truth` card matches the runtime you intended to start.

If this fails, check next:

1. `docker compose -f infra/compose/docker-compose.yml ps`
2. `docker compose -f infra/compose/docker-compose.yml logs api`
3. `docker compose -f infra/compose/docker-compose.yml logs worker`
4. confirm `127.0.0.1:8080` and `127.0.0.1:3000` are free

## Host-Run Debug Path

Use this only when you need to run processes individually.

Setup:

```sh
cd /Users/streanor/Documents/Playground/data-platform
cp .env.example .env
```

Then:

```sh
cd backend
go run ./cmd/platform-api
```

In separate terminals:

```sh
cd /Users/streanor/Documents/Playground/data-platform/backend
go run ./cmd/platform-worker
```

```sh
cd /Users/streanor/Documents/Playground/data-platform/backend
go run ./cmd/platform-scheduler
```

```sh
cd /Users/streanor/Documents/Playground/data-platform/web
npm run dev
```

Expected success result:

- API responds on `http://127.0.0.1:8080/healthz`
- web responds on `http://127.0.0.1:3000`
- logs show the worker polling and the scheduler ticking

## Next Documents

After this quickstart:

1. [operator-manual.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/operator-manual.md)
2. [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
3. [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)
4. [trace-one-pipeline.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/trace-one-pipeline.md)
