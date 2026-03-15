# Backup And Recovery Bundles

This runbook owns the first-party backup and recovery story. It is written as a
procedure, not just a description.

## Current Recovery Reality

What is implemented today:

- bundle creation
- bundle verification
- one-command filesystem restore via `platformctl backup restore`
- automatic PostgreSQL control-plane replay when the configured database is reachable
- isolated restore drills that do not overwrite the live runtime
- a restore E2E drill that boots the API against restored state

What is still intentionally unfinished:

- artifact metadata rows are rebuilt lazily from the filesystem rather than exported/restored as first-class snapshot rows
- in-flight queue claims are restored as `queued`, not resumable `active`, because claim tokens are intentionally not exported in backup bundles
- restore does not attempt to revive a running stack in place; stop the target runtime first

That means recovery is now symmetric and executable, with the remaining edges called out plainly instead of hidden behind manual steps.

## Create A Backup

Working directory:

```sh
cd /Users/streanor/Documents/Playground/data-platform
```

Command:

```sh
make backup
```

Expected success result:

- output includes `backup bundle created:`
- output includes `backup bundle verified:`
- the bundle exists under `var/backups/`

## Verify An Existing Bundle

Command:

```sh
cd backend
go run ./cmd/platformctl backup verify --file ../var/backups/<bundle-name>.tar.gz
```

Expected success result:

- command exits `0`
- output includes `generated_at=`
- output includes counts for runs, queue requests, dashboards, data assets, and bundle files

## List Existing Bundles

Command:

```sh
cd backend
go run ./cmd/platformctl backup list
```

Expected success result:

- one line per known bundle
- each line includes the full path and size in bytes

## Safe Restore Drill

This drill proves the real restore command can rebuild a fresh runtime root
without touching live state.

Command:

```sh
cd /Users/streanor/Documents/Playground/data-platform
make restore-drill
```

Expected success result:

- output includes `restore drill passed`
- output includes:
  - `bundle_path=...`
  - `restore_root=/tmp/...`
  - `restored_data_root=...`
  - `restored_artifact_root=...`

What this checks:

- the bundle verifies before restore
- the restore command can rebuild a fresh data root, artifact root, and DuckDB file
- the restored runtime root contains control-plane runs plus data/artifact directories

## One-Command Restore

Use this only into a stopped target environment.

Command:

```sh
cd backend
go run ./cmd/platformctl backup restore \
  --file ../var/backups/<bundle-name>.tar.gz \
  --yes
```

What the command does:

- verifies the bundle before applying it
- replaces the configured data root with the archived `files/data` tree
- replaces the configured artifact root with the archived `files/artifacts` tree
- replaces the configured DuckDB file with the archived snapshot
- if PostgreSQL is reachable, reapplies migrations and restores:
  - `run_snapshots`
  - `queue_requests`
  - `dashboards`
  - `audit_events`
  - `data_assets`
  - `asset_columns`

Expected success result:

- output includes `backup bundle restored:`
- output includes the target `data_root`, `artifact_root`, and `duckdb_path`
- output includes `postgres_restored=true` in preferred Postgres mode, or a warning if auto mode skipped PostgreSQL because it was unavailable

Optional flags:

- `--postgres-mode auto|required|skip`
  - `auto`: restore PostgreSQL when reachable, otherwise restore only the filesystem and print a warning
  - `required`: fail if PostgreSQL cannot be reached
  - `skip`: restore only the filesystem roots
- `--data-root`, `--artifact-root`, `--duckdb-path`
  - override the target runtime roots
- `--extract-root`
  - keep the extracted bundle workspace instead of using an internal temporary directory

## Restore E2E Drill

This drill proves more than tar extraction. It creates a real bundle from an
isolated smoke runtime, restores that bundle into a second isolated runtime,
and boots the API against the restored state.

Command:

```sh
cd /Users/streanor/Documents/Playground/data-platform
make restore-e2e
```

Expected success result:

- output includes `restore e2e passed`
- output includes:
  - `backup_path=...`
  - `restore_root=/tmp/...`
  - `restored_api_url=http://127.0.0.1:...`
  - `manual_run_id=...`

What this checks:

- a real runtime can generate a valid bundle
- the restore command can rebuild a second runtime root from that bundle
- the restored API can serve reports, catalog, dataset profile, analytics, and artifacts from restored state

## What Comes Back Today

Restored directly from the bundle:

- local data root content, including `raw`, `staging`, `intermediate`, `mart`, `metrics`, `quality`, `profiles`, and control-plane files when present
- local artifacts
- DuckDB file
- PostgreSQL control-plane snapshot tables when PostgreSQL restore is enabled

Restored with deliberate caveats:

- active queue rows come back as `queued` because bundle exports intentionally do not carry worker claim tokens
- artifact bytes are restored directly, while PostgreSQL artifact index rows are rebuilt lazily by the running platform if needed

Not restored yet:

- a richer normalized relational history beyond the current snapshot/projection tables

## Post-Restore Validation

Run these checks after restart:

```sh
curl http://127.0.0.1:8080/healthz
curl -H "Authorization: Bearer viewer-token" http://127.0.0.1:8080/api/v1/catalog
curl -H "Authorization: Bearer viewer-token" http://127.0.0.1:8080/api/v1/reports
curl -H "Authorization: Bearer viewer-token" "http://127.0.0.1:8080/api/v1/catalog/profile?asset_id=mart_budget_vs_actual"
```

Expected success result:

- all commands return HTTP `200`
- catalog returns assets
- reports returns dashboards
- dataset profile returns `row_count`

Then run:

```sh
make smoke
```

Expected success result:

- `localhost smoke test passed`

## If This Fails, Check Next

1. confirm the bundle verified before extraction
2. confirm the target runtime was stopped before copying files
3. confirm `var/data`, `var/artifacts`, and `var/duckdb` were restored into the correct repo root
4. if using Postgres, confirm migrations ran successfully before restart
5. if API is healthy but catalog/profile fail, confirm the restored data files actually exist under `var/data`
6. if `postgres_restored=false`, confirm whether the target environment was intentionally file-only or whether PostgreSQL was unreachable during restore
