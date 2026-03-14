# Scheduler Package

This package owns the scheduling loop and dependency release mechanics. The scheduler should stay predictable, restart-safe, and easy to reason about because subtle timing logic often becomes a source of accidental complexity.

The current implementation:

- refreshes manifests and catalog state
- persists scheduler enqueue state under the local data root
- evaluates due runs in the pipeline's declared timezone
- supports the cron subset needed by the sample slice, including step fields and
  day-of-week matching

This keeps scheduled execution working locally without introducing an external queue or scheduler dependency.
