# Platform Manifests

This directory contains the declarative platform configuration that defines pipelines, assets, metrics, quality checks, and ownership. The runtime should project these manifests into operational state rather than inventing a parallel configuration model.

Analytical SQL now lives alongside these manifests under `packages/sql` so the
platform can keep declarative metadata and executable SQL versioned together
without overloading one directory with mixed concerns.
