# Model 1 v3 — Backend Orchestration, Ingestion & Governance

**Priority:** HIGH — Closes out the final production blockers around run safety and ingestion.  
**Owner:** Model 1  
**Merge order:** 1st  

---

## Mission

Drive Pipeline Orchestration (Category 1), Data Ingestion (Category 2), and Metadata Governance (Category 3) to 100%. Make the execution runner safe from transient failures, add native database ingestion capabilities, and introduce data retention mechanisms to prevent unbounded growth.

## Tasks (In Priority Order)

### Task 1: Exponential Backoff Retries & Idempotency
- **Current state:** Retries exist but lack sophisticated backoff or idempotency controls.
- **Required change:** Update `backend/internal/execution/runner.go` to support exponential backoff for failed jobs. Jobs must securely use idempotency keys (e.g. `run_id + job_id + attempt_count`) so transient API/DB disruptions don't result in duplicate side effects.
- **Completion signal:** `go test ./internal/execution/...` verifies backoff timing and idempotency.

### Task 2: Native Database Ingestion
- **Current state:** Ingestion relies on static CSV/JSON files. 
- **Required change:** Add native Postgres and MySQL ingestion capabilities. The runner should dynamically connect, query, and dump remote database tables into the raw file lake based on pipeline manifests.
- **Completion signal:** E2E test proves database rows can be ingested into a `.csv` or `.parquet` file in the data root.

### Task 3: Data Retention & Purge Policies
- **Current state:** Run logs and staging artifacts grow indefinitely.
- **Required change:** Create a background GC job (or CLI command triggered via cron) that reads retention policies from asset manifests and physically deletes stale artifacts/run history from both the DB and filesystem.
- **Completion signal:** `make benchmark` or a dedicated shell script proves that expired runs are physically purged.

## Escalation Triggers
- If adding database drivers significantly inflates the binary size beyond acceptable limits, consider making them optional plugins.
