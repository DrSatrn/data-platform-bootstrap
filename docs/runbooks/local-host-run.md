# Local Host-Run Guide

This file exists because its name is easy to search for, but it is no longer
the primary host-run procedure.

Use [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)
for the real step-by-step host-run path.

## When To Use This File

Use this file only to decide which startup path you want:

- I want the fastest first success:
  [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
- I want Docker Compose:
  [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
- I want to run the API, worker, scheduler, and web app directly on my machine:
  [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)

## Why This File Changed

The repo previously had two host-run docs with very similar names:

- `local-host-run.md`
- `localhost-e2e.md`

That is confusing for a new reader. `localhost-e2e.md` now owns the real
host-run procedure. This file is a signpost so older links and habits do not
send you into a dead end.
