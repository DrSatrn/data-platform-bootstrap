# Scheduler Package

This package owns the scheduling loop and dependency release mechanics. The scheduler should stay predictable, restart-safe, and easy to reason about because subtle timing logic often becomes a source of accidental complexity.

The current implementation:

- refreshes manifests and catalog state
- persists scheduler enqueue state under the local data root
- enqueues due runs for the subset of cron syntax currently used by the sample slice

This keeps scheduled execution working locally without introducing an external queue or scheduler dependency.
