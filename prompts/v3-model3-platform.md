# Model 3 v3 — Operations, Alerting & CI/CD

**Priority:** HIGH — Closes out the final ops and infrastructure blockers for an enterprise-ready pipeline.  
**Owner:** Model 3  
**Merge order:** 3rd  

---

## Mission

Drive Monitoring (Category 4), Platform-as-Code (Category 7), and CI/CD (Category 8) to 100%. Enable proactive pipeline health awareness, standard runtime telemetry, and fully automated release workflows.

## Tasks (In Priority Order)

### Task 1: Webhook Alerting
- **Current state:** Failures require UI polling to discover. 
- **Required change:** Build a backend alerting module that can be configured via environment variables to POST JSON payloads to designated URLs (e.g., Slack/Teams) when a pipeline run fails or an asset exceeds its warning SLA.
- **Completion signal:** E2E unit test confirms a mock HTTP server receives the correct webhook payload when a mocked run transitions to failed.

### Task 2: `/metrics` Prometheus Endpoint
- **Current state:** Opsview is built for the custom React GUI.
- **Required change:** Add a new handler adhering strictly to the Prometheus exposition format. Expose Golang memory stats, current number of active workers, queue depth, and basic HTTP latency histograms. Map this to an `/api/v1/system/metrics` endpoint.
- **Completion signal:** `curl localhost:8080/api/v1/system/metrics` returns plaintext Prometheus-formatted telemetry.

### Task 3: Automated GitHub Releases
- **Current state:** Actions run tests but don't publish release artifacts.
- **Required change:** Update `.github/workflows/ci.yml` to trigger cross-compilation (linux/amd64, linux/arm64, darwin/arm64) on new tags. Build a tarball containing the `platformctl` binary and necessary `docs/runbooks`, and upload it as a GitHub Release asset automatically.
- **Completion signal:** Local verification of the `ci.yml` matrix structure shows it correctly calls the `Makefile` compilation targets.

## Escalation Triggers
- If adding telemetry metrics introduces severe lock contention inside the runtime state manager, consider implementing a dedicated fast-path lockless counter.
