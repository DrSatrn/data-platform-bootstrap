# Resilience Drills

This runbook owns the repo-first resilience checks that prove the platform can
recover from common local failure scenarios without guesswork.

## What this covers

The current drill set proves:

- queued work can be reclaimed after a worker restart
- a corrupt DuckDB file returns a clear error instead of panicking
- the restore E2E path still works against a runtime that already executed a run

## Fastest path

Run the full drill bundle:

```sh
sh infra/scripts/resilience_drill.sh
```

Expected pass/fail signals:

- `go test ./test ...` exits `0`
- `restore e2e passed` appears in the output
- the script ends with `resilience drill passed`

If any step exits non-zero, treat it as a real regression in operational
resilience rather than a flaky optional check.

## Individual scenarios

Queue reclaim after restart:

```sh
cd backend
go test ./test -run TestWorkerRestartReclaimsActiveQueueRequest
```

DuckDB corruption handling:

```sh
cd backend
go test ./test -run TestCorruptDuckDBReturnsClearError
```

Restore after a real run:

```sh
make restore-e2e
```

## Interpreting failures

- If queue reclaim fails, inspect
  [queue.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/orchestration/queue.go)
  and
  [worker.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/execution/worker.go).
- If the DuckDB corruption drill fails unclearly, inspect
  [engine.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/transforms/engine.go).
- If restore E2E fails, inspect
  [restore_e2e.sh](/Users/streanor/Documents/Playground/data-platform/infra/scripts/restore_e2e.sh)
  and
  [backups.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md).
