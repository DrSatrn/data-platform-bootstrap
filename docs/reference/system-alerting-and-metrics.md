# System Alerting And Metrics

This reference doc describes the additive ops surfaces introduced for the v3
Model 3 lane.

## Webhook Alerting

Two environment variables control outbound alerting:

- `PLATFORM_ALERT_RUN_FAILURE_WEBHOOK_URLS`
  Comma-separated URLs that receive JSON when a pipeline run transitions to
  `failed`.
- `PLATFORM_ALERT_ASSET_WARNING_WEBHOOK_URLS`
  Comma-separated URLs that receive JSON when the scheduler observes an asset
  enter the `stale` freshness state.
- `PLATFORM_ALERT_WEBHOOK_TIMEOUT`
  Optional request timeout. Default: `5s`.

Alert payloads are JSON and include:

- `event_type`
- `environment`
- `severity`
- `occurred_at`
- `summary`
- `run` or `asset`

Current event types:

- `pipeline_run_failed`
- `asset_warning_sla_breached`

## Prometheus Endpoint

The platform now exposes:

- `/api/v1/system/metrics`

This endpoint uses the Prometheus text exposition format and currently includes:

- Go memory statistics
- goroutine count
- active worker count, derived from active queue claims
- queue depth, derived from queued requests
- HTTP request totals, error totals, and request latency histogram buckets

The endpoint is viewer-role protected like the other system read surfaces.

## Release Assets

Tagged pushes matching `v*` now trigger multi-architecture `platformctl`
release packaging in GitHub Actions for:

- `linux/amd64`
- `linux/arm64`
- `darwin/arm64`

Each tarball currently contains:

- `bin/platformctl`
- `docs/runbooks/`
