# Upgrading Runbook

This runbook explains how to move from one commit to the next without losing
data, artifacts, or the ability to recover.

## Core rule

Never upgrade a running instance without taking and verifying a backup first.

## Safe upgrade sequence

1. Confirm the current stack is healthy.
2. Run `make backup`.
3. Verify the backup output reports both `created` and `verified`.
4. Stop the running stack or local processes.
5. Update the repo to the target commit.
6. Run migrations before starting the full stack:

```sh
cd backend
go run ./cmd/platformctl migrate
```

7. Start the stack again.
8. Run `make smoke` or `make compose-smoke`.
9. Run `make benchmark` if this is a release-like change.

## Migration ordering guarantees

- filesystem-backed data roots are preserved across upgrades unless you
  explicitly delete them
- DuckDB, data artifacts, and control-plane files are reused in place
- PostgreSQL schema changes are applied through the repo migration set before
  normal API traffic resumes
- the repo-owned backup format captures the pre-upgrade control-plane and
  filesystem state so rollback can restore that snapshot

## Backup-before-upgrade procedure

```sh
make backup
```

Operator expectations:

- the command exits `0`
- the bundle lands under `var/backups/`
- `backup bundle verified:` appears in the output

If verification fails, stop and do not proceed with the upgrade.

## Rollback procedure

If the upgrade fails after the new commit is deployed:

1. Stop the target runtime.
2. Restore the last verified bundle:

```sh
cd backend
go run ./cmd/platformctl backup restore --file ../var/backups/<bundle-name>.tar.gz --yes
```

3. If PostgreSQL rollback is required and available, use the default restore
   mode or `--postgres-mode required`.
4. Restart the stack from the last known-good commit.
5. Run `make restore-e2e` or at minimum verify:
   - `GET /healthz`
   - catalog
   - analytics
   - artifacts for a known run

## What cannot be rolled back cleanly

Be explicit about current limits:

- irreversible schema migrations may require restoring PostgreSQL from backup
  rather than merely checking out an older commit
- active login sessions are intentionally cleared on restore
- queue claims are not restored as active resumable work when PostgreSQL
  control-plane state is rehydrated
- any data written after the backup point is not recoverable from that bundle

## Recommended post-upgrade checks

- `go run ./cmd/platformctl validate-manifests`
- `make smoke`
- `make benchmark`
- `make restore-drill`
