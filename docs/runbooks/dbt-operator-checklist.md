# DBT Operator Checklist

Use this checklist when an operator needs to trust a dbt-backed
`external_tool` run.

## Before Triggering

- confirm the manifest uses `type: external_tool`
- confirm `external_tool.tool` is `dbt`
- confirm `project_ref` and `config_ref` are repo-relative
- confirm the dbt project contains `dbt_project.yml`
- confirm the configured binary resolves:

```bash
which dbt
echo "$PLATFORM_DBT_BINARY"
```

## After Triggering

- confirm the run includes external-tool lifecycle events
- confirm `stdout.log` and `stderr.log` appear when the process emits output
- confirm declared output artifacts appear under:

```text
external_tools/<job_id>/...
```

- confirm the run does not end with:
  - `tool_unavailable`
  - `execution_failed`
  - `artifact_missing`
  - `invalid_spec`

## Drill Fixtures

Additive fixture assets for operator drills:

- [dbt_non_zero_exit_pipeline.yaml](/Users/streanor/Documents/Playground/data-platform/packages/manifests/fixtures/external_tools/dbt_non_zero_exit_pipeline.yaml)
- [dbt_missing_artifact_pipeline.yaml](/Users/streanor/Documents/Playground/data-platform/packages/manifests/fixtures/external_tools/dbt_missing_artifact_pipeline.yaml)
- [dbt_invalid_repo_ref_pipeline.yaml](/Users/streanor/Documents/Playground/data-platform/packages/manifests/fixtures/external_tools/dbt_invalid_repo_ref_pipeline.yaml)

## Verification

```bash
cd backend && go test ./internal/orchestration ./internal/storage ./internal/execution ./internal/externaltools ./internal/manifests ./internal/config
```
