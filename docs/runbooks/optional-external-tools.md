# Optional External Pipeline Tools

This runbook documents the additive-first support for bring-your-own pipeline
tools.

## What This Capability Is

The platform control plane can optionally execute a bounded external tool job
inside a pipeline.

In this first slice:

- `dbt` is implemented
- `dlt` and `pyspark` are intentionally reserved but not yet implemented
- the platform still owns scheduling, queueing, run state, audit, and artifact
  indexing
- the external tool adapter only owns command construction, subprocess
  execution, and artifact discovery

## When To Use It

Use an external tool job when:

- you already have a working dbt project you want to bring into the platform
- the platform should orchestrate and audit the run, but not replace the tool
- you want declared tool outputs mirrored into run-scoped artifacts

Do not use it when:

- the transformation is small enough to live as a native SQL or Python job
- you need a remote cluster submission story in this slice
- you expect `dlt` or `pyspark` to work today

## Required Runtime Configuration

The feature is inert unless a pipeline manifest declares a job with
`type: external_tool`.

Optional runtime settings:

- `PLATFORM_EXTERNAL_TOOL_ROOT`
- `PLATFORM_DBT_BINARY`
- `PLATFORM_DLT_BINARY`
- `PLATFORM_PYSPARK_BINARY`
- `PLATFORM_EXTERNAL_TOOL_TIMEOUT_DEFAULT`

Recommended local-first values:

```bash
PLATFORM_EXTERNAL_TOOL_ROOT=..
PLATFORM_DBT_BINARY=dbt
PLATFORM_EXTERNAL_TOOL_TIMEOUT_DEFAULT=5m
```

## Manifest Shape

Example job block:

```yaml
- id: run_finance_dbt
  name: Build Finance Models In DBT
  type: external_tool
  timeout: 5m
  external_tool:
    tool: dbt
    action: build
    project_ref: packages/external_tools/dbt_finance_demo
    config_ref: packages/external_tools/dbt_finance_demo/profiles
    profile: dbt_finance_demo
    target: dev
    selector: monthly_cashflow
    artifacts:
      - path: target/manifest.json
        required: false
      - path: target/run_results.json
        required: true
```

Notes:

- `project_ref` and `config_ref` are repo-relative
- artifact paths are relative to the external project root
- `config_ref` is used as the dbt profiles directory in this slice
- `profile`, `target`, `selector`, `args`, and `vars` are optional dbt
  adapter controls

## Expected Artifacts

When the job succeeds, the control plane should mirror:

- `stdout` log output into a run-scoped log artifact
- `stderr` log output into a run-scoped log artifact
- each declared artifact file into the run artifact index

The default artifact destination pattern is:

```text
external_tools/<job_id>/<declared artifact path>
```

## Failure Modes

This slice should fail clearly for:

- unsupported tool declarations
- unsafe or missing repo-relative paths
- missing external binaries
- non-zero tool exits
- declared artifacts that were not produced

Expected failure classes:

- `invalid_spec`
- `tool_unavailable`
- `execution_failed`
- `artifact_missing`
- `unsupported_tool`

## Example Assets

Additive example assets created in this thread:

- `packages/external_tools/dbt_finance_demo/`
- `packages/manifests/pipelines/personal_finance_dbt_pipeline.yaml`

These assets are intentionally additive and can be wired into canonical docs
later, after the parallel critique/docs thread stabilizes.
