# Benchmarking Runbook

This runbook explains how to execute the repo-owned benchmark suite against a
running platform stack and how to interpret the output.

## Goal

Track the platform's response budgets with a first-party benchmark path instead
of relying only on ad hoc manual checks. The benchmark suite is now a release
gate, not just a passive timing report.

The benchmark command currently measures:

- `/healthz`
- `/api/v1/catalog`
- `/api/v1/analytics?dataset=mart_budget_vs_actual`
- `/api/v1/analytics?metric=metrics_category_variance`
- `/api/v1/reports`
- `/api/v1/system/overview`
- `/api/v1/system/audit`
- `/api/v1/admin/terminal/execute` with the `status` command
- a concurrent manual-trigger load scenario against `personal_finance_pipeline`
- queue visibility under that load
- scheduler heartbeat freshness from `/api/v1/system/overview`

## Prerequisites

- A running platform stack on localhost or another explicitly configured local
  endpoint
- A valid admin token if you want the admin-terminal benchmark included
- Go available on the host for the `platformctl` command

## Fastest path

If the platform is already running on the default localhost URL:

```bash
make benchmark
```

That writes a timestamped JSON report under `var/benchmarks/`.

## Custom target

To benchmark another local stack or adjust iteration count:

```bash
PLATFORM_BENCHMARK_URL=http://127.0.0.1:18085 \
PLATFORM_ADMIN_TOKEN=local-dev-admin-token \
PLATFORM_BENCHMARK_ITERATIONS=10 \
PLATFORM_BENCHMARK_LOAD_TRIGGERS=6 \
PLATFORM_BENCHMARK_LOAD_CONCURRENCY=3 \
PLATFORM_BENCHMARK_QUEUE_VISIBLE_THRESHOLD_MS=7000 \
PLATFORM_BENCHMARK_SCHEDULER_LAG_THRESHOLD_SECONDS=120 \
make benchmark
```

## Direct CLI usage

You can also call the CLI directly:

```bash
cd backend
go run ./cmd/platformctl benchmark \
  --server http://127.0.0.1:8080 \
  --token local-dev-admin-token \
  --iterations 10 \
  --pipeline personal_finance_pipeline \
  --load-triggers 6 \
  --load-concurrency 3 \
  --queue-visible-threshold-ms 7000 \
  --scheduler-lag-threshold-seconds 120 \
  --out ../var/benchmarks/manual-benchmark.json
```

## Output shape

The benchmark report includes:

- server URL
- generation time
- iteration count
- one result per target with:
  - successes and failures
  - average latency
  - p50 latency
  - p95 latency
  - min and max latency
  - last HTTP status
  - last observed error
- one load-scenario block with:
  - requested versus accepted triggers
  - queue depth before and after the load burst
  - maximum queued, active, total, and inflight request counts observed
  - queue visibility latency
  - run IDs and trigger errors
- one scheduler summary block with:
  - latest heartbeat time
  - scheduler lag in seconds
  - latest pipeline and asset counts
- assertion results that explicitly pass or fail the benchmark gate

## How to use it

- Run after major backend or frontend changes that affect API hydration paths.
- Compare reports over time to spot regressions in catalog, analytics, or admin
  surfaces.
- Treat a failed benchmark command as a real regression signal. It now fails if:
  - any target records zero successes
  - the load scenario cannot surface accepted queue requests within the budget
  - the scheduler heartbeat is stale beyond the configured threshold
- Use the benchmark output together with the smoke scripts:
  - smoke scripts prove behavior
  - benchmark reports quantify performance
