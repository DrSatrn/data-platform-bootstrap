# Access Matrix

Use this document when you want a quick answer to "which role should be able to
do this?"

Important note:

- this matrix describes the intended operator-facing contract
- if the running product disagrees, trust the live UI plus `/api/v1/session`
- where behavior is not fully verified end to end, that row is marked
  `UNVERIFIED`

## Roles

| Role | Intended use |
| --- | --- |
| `anonymous` | health checks and first sign-in only |
| `viewer` | read-only product access |
| `editor` | operational work that changes platform state but does not manage users |
| `admin` | full operator access including admin tooling and user management |

## Capability Matrix

| Workflow | Anonymous | Viewer | Editor | Admin | Verification status |
| --- | --- | --- | --- | --- | --- |
| `GET /healthz` | Yes | Yes | Yes | Yes | verified by design |
| view current session via `/api/v1/session` | Yes | Yes | Yes | Yes | verified by design |
| sign in through `/api/v1/session` | Yes | Yes | Yes | Yes | verified by design |
| view Dashboard page | No | Yes | Yes | Yes | verified in current UI |
| view Pipelines page | No | Yes | Yes | Yes | verified in current UI |
| view Datasets page | No | Yes | Yes | Yes | verified in current UI |
| view Metrics page | No | Yes | Yes | Yes | verified in current UI |
| view System page | No | Yes | Yes | Yes | verified in current UI |
| trigger a manual pipeline run | No | No | Yes | Yes | verified in current UI |
| create, save, duplicate, or delete dashboards | No | No | Yes | Yes | verified in current UI |
| update metadata annotations | No | No | Yes | Yes | partially verified |
| use the admin terminal | No | No | No | Yes | verified in current UI |
| create or manage users | No | No | No | Yes | verified in current UI |
| run `platformctl remote ...` | No | No | No | Yes | verified by current docs and UI contract |
| use external-tool operator flows | No | Yes | Yes | Yes | UNVERIFIED end to end |

## Practical Guidance

- use the bootstrap admin token only for first-run setup and recovery
- for normal day-to-day work, create native users and sign in through the UI or
  `/api/v1/session`
- if PostgreSQL is unavailable, native users and sessions are unavailable too,
  so the runtime may fall back to bootstrap-token-only behavior

## What Success Looks Like

You should be able to answer these questions quickly:

- can this person view the product at all?
- can this person trigger runs?
- can this person modify dashboards or metadata?
- can this person use the admin terminal?
- should this person be using a normal session or the bootstrap token?

## If Something Goes Wrong

If observed behavior does not match this matrix:

1. check the current session payload from `/api/v1/session`
2. confirm whether the runtime is using native users or bootstrap-token-only mode
3. compare the live UI behavior with
   [operator-manual.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/operator-manual.md)
4. if the docs and product truly disagree, treat the doc as stale and update it
