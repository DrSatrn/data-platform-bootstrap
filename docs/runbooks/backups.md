# Backup And Recovery Bundles

This runbook explains the first-party backup/export path built into the
platform. The goal is to give self-hosted operators a real recovery workflow
without depending on external backup products.

## What The Bundle Contains

`platformctl backup create` writes a `.tar.gz` archive that includes:

- control-plane JSON exports for runs, queue state, dashboards, audit events,
  metadata assets, and sanitized config
- local materialized data under the platform data root when present
- run-scoped artifact files
- the DuckDB database file
- repo-managed manifest and dashboard snapshots
- a machine-readable `manifest.json` containing file inventory and checksums

The backup manifest intentionally excludes secrets such as bearer tokens and
PostgreSQL credentials.

## Create A Backup

From the repo root:

```sh
make backup
```

Or directly:

```sh
cd backend
go run ./cmd/platformctl backup create --out ../var/backups/manual-backup.tar.gz
```

## Verify A Backup

Always verify the bundle after creation:

```sh
cd backend
go run ./cmd/platformctl backup verify --file ../var/backups/manual-backup.tar.gz
```

Verification checks:

- required export files exist
- `manifest.json` is present
- archived file sizes match the manifest
- archived file checksums match the manifest

## Discover Existing Bundles

```sh
cd backend
go run ./cmd/platformctl backup list
```

The admin terminal also supports:

- `backups`
- `backup create`
- `backup verify <bundle-name-or-path>`

## Recovery Guidance

This subsystem is currently a backup/export and verification path, not yet a
one-command destructive restore workflow.

Today’s safe recovery path is:

1. Verify the bundle.
2. Extract it into an isolated working directory.
3. Restore the relevant local roots:
   - `files/data/control_plane`
   - `files/data/raw`
   - `files/data/mart`
   - `files/data/metrics`
   - `files/data/quality`
   - `files/artifacts`
   - `files/duckdb/platform.duckdb`
4. Reapply migrations if PostgreSQL was recreated.
5. Restart the platform and exercise:
   - `/healthz`
   - `/api/v1/catalog`
   - `/api/v1/reports`
   - `/api/v1/system/audit`
6. Run the smoke workflow to prove the recovered stack is healthy.

## Planned Next Step

The next evolution is a restore automation command that can replay bundle state
into a clean target environment with stronger safeguards.
