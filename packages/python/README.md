# Python Tasks

This directory contains bounded Python tasks launched by the Go worker.

These tasks are deliberately narrow. They are not an alternate control plane.

Use Python here for:

- data enrichment helpers
- profiling and quality helpers
- connector-specific logic
- other small data-runtime behaviors that are easier to express in Python

Each task reads a JSON request from `PLATFORM_TASK_REQUEST_PATH` and writes a
JSON result to `PLATFORM_TASK_RESULT_PATH`.
