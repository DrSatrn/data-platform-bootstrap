# External Tool Jobs

`external_tool` is an additive job type for optional repo-local pipeline tools.
The platform worker still owns scheduling, run state, retries, event emission,
and run-scoped artifact capture. The external tool only performs one bounded
execution step.

## Current Scope

- implemented adapter: `dbt`
- reserved but rejected for now: `dlt`, `pyspark`
- artifact mirroring uses the existing run artifact namespace under
  `var/artifacts/runs/<run_id>/external_tools/<job_id>/...`

## Manifest Shape

```yaml
jobs:
  - id: run_finance_dbt
    type: external_tool
    external_tool:
      tool: dbt
      action: build
      project_ref: packages/external_tools/dbt_finance_demo
      config_ref: packages/external_tools/dbt_finance_demo/profiles
      profile: dbt_finance_demo
      target: dev
      selector: monthly_cashflow
      artifacts:
        - path: target/run_results.json
          required: true
        - path: target/manifest.json
          required: false
```

Field notes:

- `tool`: adapter name
- `action`: bounded tool action such as `build`
- `project_ref`: repo-relative project root
- `config_ref`: repo-relative config root or config file
- `artifacts`: declared files to mirror into run artifacts after success

## Example Assets

- example pipeline manifest:
  [personal_finance_dbt_pipeline.yaml](/Users/streanor/Documents/Playground/data-platform/packages/manifests/pipelines/personal_finance_dbt_pipeline.yaml)
- example dbt project:
  [dbt_project.yml](/Users/streanor/Documents/Playground/data-platform/packages/external_tools/dbt_finance_demo/dbt_project.yml)
- example dbt profile:
  [profiles.yml](/Users/streanor/Documents/Playground/data-platform/packages/external_tools/dbt_finance_demo/profiles/profiles.yml)

## Important Boundaries

- The worker invokes the tool as a subprocess and captures declared artifacts.
- The tool does not register schedules, queue its own jobs, or replace the
  platform control plane.
- Adapter support is intentionally narrow in the first slice so later tool
  additions can reuse the same contract.
