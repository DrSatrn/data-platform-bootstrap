# Operator Command Taxonomy

This reference groups platform commands by operator intent so the terminal and
future GUI can share one conceptual model.

## Operations

Primary questions:

- what is running
- what is queued
- what failed
- what can I trigger safely

Examples:

- `status`
- `pipelines`
- `runs`
- `trigger <pipeline_id>`
- `artifacts <run_id>`

## Catalog And Metadata

Primary questions:

- what assets exist
- who owns them
- what is stale or missing

Examples:

- `assets`
- future: `asset <asset_id>`
- future: `freshness`

## Reporting

Primary questions:

- what dashboards exist
- what should be editable
- what definitions look invalid or stale

Examples:

- `dashboards`
- future: `dashboard <dashboard_id>`

## Governance

Primary questions:

- what quality checks are failing
- where are docs or ownership incomplete
- what trust signals need operator attention

Examples:

- `quality`
- `metrics`
- future: `audit recent`

## Recovery

Primary questions:

- do we have a usable backup
- when was it verified
- what should be restored or drilled next

Examples:

- `backups`
- `backup create`
- `backup verify <bundle>`

## Diagnostics

Primary questions:

- is the system healthy
- what do the logs say
- what is the platform doing right now

Examples:

- `status`
- `logs [limit]`

## Why This Matters

A terminal with many commands becomes harder to use if it is only a flat list.

The GUI and the terminal should eventually share:

- command families
- role expectations
- user intent
- runbook linkage

This taxonomy is an additive reference for that future wiring pass.
