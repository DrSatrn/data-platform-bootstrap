# Backup And Recovery Bundles

This runbook owns the first-party backup and recovery story. It is written as a
procedure, not just a description.

## Current Recovery Reality

What is implemented today:

- bundle creation
- bundle verification
- isolated restore drill extraction

What is intentionally not implemented yet:

- destructive one-command restore into a live runtime root
- automatic PostgreSQL row replay from bundle exports

That means recovery is now concrete and drillable, but still operator-driven.

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

This drill proves you can extract and inspect a bundle without overwriting live
state.

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
  - `next_step_copy_data_from=...`
  - `next_step_copy_artifacts_from=...`

What this checks:

- `manifest.json` is extractable
- control-plane exports are present
- extracted data and artifact directories exist

## Manual Restore Procedure

Use this only into a stopped target environment.

1. Stop the platform processes or Compose stack.
2. Verify the bundle:

```sh
cd backend
go run ./cmd/platformctl backup verify --file ../var/backups/<bundle-name>.tar.gz
```

3. Extract the bundle into an isolated working directory:

```sh
mkdir -p /tmp/platform-restore
tar -xzf var/backups/<bundle-name>.tar.gz -C /tmp/platform-restore
```

4. Restore filesystem-backed data into the target runtime root:

```sh
cp -R /tmp/platform-restore/files/data/. var/data/
cp -R /tmp/platform-restore/files/artifacts/. var/artifacts/
cp /tmp/platform-restore/files/duckdb/platform.duckdb var/duckdb/platform.duckdb
```

5. If PostgreSQL was recreated, reapply migrations before starting the stack:

```sh
cd backend
go run ./cmd/platformctl migrate
```

6. Restart the platform.

7. Run the post-restore checks below.

## What Comes Back From Filesystem Restore

Restored directly from the bundle:

- local data root content
- local artifacts
- DuckDB file
- control-plane JSON snapshots in the extracted export directory

Recreated separately:

- PostgreSQL schema via migrations
- PostgreSQL rows are not automatically replayed yet

In preferred Postgres mode, the live DB remains authoritative after restart.
Today’s bundle exports give you inspection and future replay material, not an
automated relational restore path.

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
