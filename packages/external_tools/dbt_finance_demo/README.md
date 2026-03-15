# DBT Finance Demo

This is an additive example dbt project used to demonstrate the
`external_tool` pipeline contract.

It is intentionally optional:

- the platform does not install dbt for you
- the example pipeline using this project is paused by default
- the worker will only run it when a pipeline manifest explicitly declares the
  `dbt` external-tool adapter

Primary example manifest:

- [personal_finance_dbt_pipeline.yaml](/Users/streanor/Documents/Playground/data-platform/packages/manifests/pipelines/personal_finance_dbt_pipeline.yaml)
