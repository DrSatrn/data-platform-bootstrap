# Data Platform Setup Guide

This file exists because people naturally look for `doc.md` at the repo root.
It is not the best place to learn the project anymore.

Use the docs in this order instead:

1. [README.md](/Users/streanor/Documents/Playground/data-platform/README.md)
   Start here if you want the plain-English project overview.
2. [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
   Use this for your first successful run.
3. [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
   Use this when you want the packaged Docker Compose stack.
4. [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)
   Use this when you want to run the API, worker, scheduler, and web app
   directly on your machine for debugging.
5. [uat-checklist.md](/Users/streanor/Documents/Playground/data-platform/uat-checklist.md)
   Use this after the platform is running and you want to confirm the main user
   journeys work.

## Fastest Path To Success

From the repo root:

```sh
make smoke
```

What success looks like:

- the command exits `0`
- the output includes `localhost smoke test passed`
- you get a temporary localhost stack without hand-starting multiple services

If something goes wrong:

1. open [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
2. follow the troubleshooting section there exactly
3. if you need the packaged stack instead, switch to
   [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)

## Bringing the Service Online and Offline

If you just want to run the platform (without development debugging or testing), use the packaged Docker Compose stack:

**To bring the service online:**
```sh
make bootstrap
```
*This starts the API, worker, scheduler, frontend, and PostgreSQL database. It also automatically runs schema migrations.*

**To bring the service offline:**
```sh
make down
```
*This gracefully stops and removes the Docker containers. Your data is safely preserved in the database volume for the next run.*

## Which Document Should You Use?

- I want the project explained first:
  [README.md](/Users/streanor/Documents/Playground/data-platform/README.md)
- I want one clear first-run path:
  [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
- I want Docker Compose:
  [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
- I want host-run debugging:
  [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)
- I want to verify the product manually:
  [uat-checklist.md](/Users/streanor/Documents/Playground/data-platform/uat-checklist.md)

## Why This File Is Short

The repo previously had several overlapping "start here" docs. That made it
harder, not easier, for a new reader to know which path was authoritative.

This file now acts as a compatibility pointer so there is one clear answer to
"where do I start?" without duplicating setup steps in multiple places.
