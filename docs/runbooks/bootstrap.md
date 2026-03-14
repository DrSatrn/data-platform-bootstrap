# Bootstrap Runbook

This runbook will guide local setup once build commands are wired in.

## Intended Flow

1. Review `codex.md` for repo-specific guidance.
2. Copy `.env.example` to `.env` and adjust local paths if needed.
3. Build the backend and web runtimes.
4. Start the Compose stack or run `platform-api`, `platform-worker`, and the web app locally.
5. Confirm the API health endpoint responds and the worker is polling.
6. Queue a manual pipeline run from the Pipelines page or with `platformctl remote --token <token> trigger personal_finance_pipeline`.
7. Verify that run artifacts appear under the local `var/` directory and that run status moves from `queued` to `running` to `succeeded`.
8. Open the System page and verify the built-in metrics, recent logs, and admin terminal are responding.
9. Inspect the finance dashboard and dataset views to confirm materialized outputs are being served.
