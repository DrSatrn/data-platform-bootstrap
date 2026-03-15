# External Tool Troubleshooting

This runbook is the additive operator guide for diagnosing `external_tool`
pipeline jobs.

It is intentionally narrow:

- current implemented adapter: `dbt`
- current reserved but unsupported adapters: `dlt`, `pyspark`
- control-plane ownership remains with the platform worker

## Where To Look First

For a run `run_123` and job `build_finance_dbt`, inspect:

- pipeline run events through `/api/v1/pipelines`
- run artifacts through `/api/v1/artifacts?run_id=run_123`
- log artifacts:
  - `external_tools/build_finance_dbt/logs/stdout.log`
  - `external_tools/build_finance_dbt/logs/stderr.log`
- declared output artifacts:
  - `external_tools/build_finance_dbt/target/run_results.json`
  - `external_tools/build_finance_dbt/target/manifest.json`

These paths are mirrored under the local artifact root:

```text
var/artifacts/runs/<run_id>/external_tools/<job_id>/...
```

## Failure Modes

### Missing Binary

Symptoms:

- run event includes `failure_class=tool_unavailable`
- job error mentions the tool is unavailable
- no declared dbt output artifacts are present
- log artifacts may be absent if the process never started

Checks:

```bash
which dbt
echo "$PLATFORM_DBT_BINARY"
cd backend && go test ./internal/execution ./internal/externaltools -run 'MissingBinary'
```

Typical cause:

- `PLATFORM_DBT_BINARY` points to a missing executable
- `dbt` is not installed on the operator host

### Non-Zero Exit

Symptoms:

- run event includes `failure_class=execution_failed`
- stdout/stderr log artifacts exist
- worker error contains the external tool exit code

Checks:

```bash
cd backend && go test ./internal/execution ./internal/externaltools -run 'NonZero|Exit'
```

Typical cause:

- dbt model failure
- bad profile/target configuration
- database or adapter runtime error

### Missing Required Artifact

Symptoms:

- run event includes `failure_class=artifact_missing`
- stdout/stderr may exist
- process may have exited successfully, but a declared output file was not
  produced

Checks:

```bash
cd backend && go test ./internal/execution ./internal/externaltools -run 'Artifact'
```

Typical cause:

- manifest declares an artifact path the tool does not actually emit
- the dbt selection changed and the expected target output was not written

### Invalid Repo-Relative Refs

Symptoms:

- pipeline validation fails or execution returns `failure_class=invalid_spec`
- error message mentions `repo-relative`

Checks:

```bash
cd backend && go test ./internal/orchestration ./internal/execution -run 'RepoRelative|Invalid'
```

Typical cause:

- `project_ref`, `config_ref`, or artifact paths are absolute
- `..` escapes the repo root

## Current Verification Commands

```bash
cd backend && go test ./internal/orchestration ./internal/storage ./internal/execution ./internal/externaltools ./internal/manifests ./internal/config
```

Targeted frontend visibility staging:

```bash
cd web && npm test -- src/features/management/externalTools/externalToolRunSummary.test.ts src/features/management/externalTools/ExternalToolRunInspector.test.tsx
```

## Fixture Assets

Failure-oriented additive fixtures live under:

- `packages/manifests/fixtures/external_tools/`
- `packages/external_tools/fixtures/`

These are drafts for drills and testing only. They are not canonical product
manifests.
