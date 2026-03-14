# Bootstrap Runbook

This runbook explains the intended local bootstrap path for day-to-day
development and for the verified localhost smoke workflow.

## Intended Flow

1. Review `codex.md` for repo-specific guidance.
2. Copy `.env.example` to `.env` and adjust local paths if needed.
3. Apply database migrations if PostgreSQL is running with `go run ./cmd/platformctl migrate`.
4. Build the backend and web runtimes.
5. Start the Compose stack or run `platform-api`, `platform-worker`, `platform-scheduler`, and the web app locally.
6. Confirm the API health endpoint responds and the worker is polling.
7. Queue a manual pipeline run from the Pipelines page or with `platformctl remote --token <token> trigger personal_finance_pipeline`.
8. Verify that run artifacts appear under the local `var/` directory and that run status moves from `queued` to `running` to `succeeded`.
9. Open the System page and verify the built-in metrics, recent logs, and admin terminal are responding.
10. Inspect the finance dashboard and dataset views to confirm materialized outputs are being served.

## Fastest Verified Path

Use the repo-owned smoke workflow when you want a quick end-to-end confidence
check without mutating your normal repo-local data directory:

```bash
make smoke
```

That workflow starts API, worker, and scheduler against an isolated `/tmp`
runtime root, verifies scheduled and manual runs, checks run artifacts, calls
the admin terminal API, and proves the `platformctl remote` CLI against the
live localhost API.
