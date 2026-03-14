# Data Model

This document summarizes the platform's first-pass domain model.

## Orchestration Entities

- `Pipeline`
- `Job`
- `Schedule`
- `PipelineRun`
- `JobRun`
- `RunEvent`

These entities support explicit state transitions, retry policies, and auditability.

## Metadata Entities

- `DataAsset`
- `Column`
- `Owner`
- `Freshness`

These entities describe the curated and raw data surfaces that operators and analysts rely on.

## Reporting Entities

- `Dashboard`
- curated metrics and chart-ready analytics payloads

The first slice intentionally keeps reporting persistence small while the overall architecture remains open for expansion.
