# Benchmarking Runbook

This runbook explains how to execute the repo-owned benchmark suite against a
running platform stack and how to interpret the output.

## Goal

Track the platform's response budgets with a first-party benchmark path instead
of relying only on ad hoc manual checks.

The benchmark command currently measures:

- `/healthz`
- `/api/v1/catalog`
- `/api/v1/analytics?dataset=mart_budget_vs_actual`
- `/api/v1/analytics?metric=metrics_category_variance`
- `/api/v1/reports`
- `/api/v1/system/overview`
- `/api/v1/admin/terminal/execute` with the `status` command

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

## How to use it

- Run after major backend or frontend changes that affect API hydration paths.
- Compare reports over time to spot regressions in catalog, analytics, or admin
  surfaces.
- Use the benchmark output together with the smoke scripts:
  - smoke scripts prove behavior
  - benchmark reports quantify performance

## Follow-on work

The current benchmark suite is intentionally small and deterministic. Future
work should expand it with:

- scheduled-run latency budgets
- artifact lookup latency
- report-save latency
- dashboard hydration timings
- larger-sample load testing against multiple pipeline runs
