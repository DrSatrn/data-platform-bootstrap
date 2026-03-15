# Python Runtime

This package defines the bounded Python task contract used by the Go worker.

The design goal is to keep Go in charge of the control plane while letting
Python own small, high-leverage data-runtime helpers where the ecosystem is
useful.

Current rules:

- Go remains the source of truth for orchestration, queueing, state machines,
  retries, audit, and API ownership.
- Python tasks are launched as subprocesses by the Go worker.
- The contract is file-based JSON so task requests and results are easy to
  inspect and debug.
- Python tasks may write platform-owned files under the configured data root,
  but they must report those outputs explicitly back to Go.

Current runtime-owned Python entrypoints:

- `tasks/enrich_transactions.py` for staging enrichment during worker runs
- `tasks/profile_asset.py` for on-demand dataset profiling requested by the
  metadata API and Datasets page
