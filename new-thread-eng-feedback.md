# Engineering Feedback

Review basis: snapshot of the repository observed on March 15, 2026.

Important context:
- This repo is actively in flight and being edited concurrently.
- Feedback below should be read as a second-pass engineering review of the current visible state, not a judgment on the final intended implementation.
- I have biased this toward durable architectural and product-engineering concerns over transient polish or in-progress feature gaps.

## What This Repo Is Clearly Going For

This project is aiming to be more than a toy pipeline demo. It is trying to become a local-first, self-hosted internal data platform with first-party ownership of:

- orchestration
- scheduling
- execution
- metadata and lineage
- analytics serving
- saved reporting
- operational tooling
- backup and recovery

The other clear goal is educational clarity. The codebase is deliberately organized to teach platform architecture, not just hide complexity behind frameworks and generated layers.

That intent comes through well. The repo already feels more coherent and product-shaped than most “full-stack data platform” experiments.

## Strong Parts

### 1. The repo has a strong point of view

The project has a real opinion about what belongs inside the product and what should not be outsourced to external tooling. That makes the work feel intentional rather than stitched together.

### 2. Documentation quality is unusually high

The architecture docs, runbooks, product notes, and README create a strong narrative for how the system is supposed to work. A new engineer can reconstruct the platform without reading every file in random order.

### 3. The modular monolith structure is the right call

The backend decomposition is disciplined enough to preserve clear ownership, but not overcomplicated with premature distributed boundaries. For this kind of project, that is the correct tradeoff.

### 4. The repo contains real executable proof

This is not just a documentation-heavy scaffold. The backend tests pass, manifest validation runs, and the localhost smoke flow exercises a meaningful end-to-end path. That matters a lot.

### 5. The project is surprisingly product-aware

The presence of audit, backup, access control, browser editing, admin terminal flows, and reporting management shows good product instincts. The work is not narrowly fixated on “run SQL somehow.”

## Hyper-Critical Docs Review

This needs to be said plainly: the docs are the most important part of a repo like this, and in their current state they are not yet reliable enough to be treated as the primary source of operational truth.

The docs are strong at narrative and architecture framing.

They are weaker at:

- giving one unambiguous setup path
- making prerequisites concrete
- distinguishing Compose-only behavior from local-host-run behavior
- telling the reader which token is actually required for each action
- telling the reader what they should literally expect to see when something works
- avoiding duplication and drift across multiple overlapping runbooks

The result is that the docs often feel polished before they feel trustworthy.

### 1. The docs over-assume and under-specify

The biggest concrete issue is that the docs repeatedly tell the user to copy `.env.example` to `.env` and then run the platform locally, but the backend config loader does not load `.env` files at all. It only reads process environment variables.

That means a cold reader can reasonably do this:

1. copy `.env.example` to `.env`
2. run `go run ./cmd/platform-api`
3. assume the app will use that config

But in the current implementation, that assumption is wrong unless the user is using Docker Compose or is separately exporting those variables into the shell.

This is not a small docs nit. It is a first-run trap.

Worse, `.env.example` is Compose-oriented and points at container paths like `/workspace/...` and `/var/lib/platform/...`, while local host-run defaults are relative paths such as `../packages/...` and `../var/...`. So the docs are currently mixing two configuration models without clearly teaching the difference.

For a repo that wants to be highly followable, this is the single biggest documentation failure.

### 2. There is no single canonical “do this if you want it running” path

The README, `bootstrap.md`, `operator-manual.md`, and `localhost-e2e.md` all partially explain how to get started, but none of them cleanly owns the startup story from beginning to end.

What the reader needs is one canonical page that says:

- use this path if you want the easiest success
- use this path if you want packaged Compose
- use this path only if you are debugging individual processes
- here are the exact commands
- here are the exact prerequisites
- here is what should appear on success
- here is what to do when it does not work

Right now the docs are not terrible individually. The problem is that there are too many overlapping “start here” documents, and they leak responsibility into each other.

That makes the repo look documented while still forcing the reader to synthesize the real workflow themselves.

### 3. The docs are too confident about commands that have hidden constraints

A good example is `platformctl remote ...`.

The docs often present it generically as a remote operator path, but this command goes through the admin terminal endpoint, which currently requires `admin`. That means the docs should not talk about it like a general-purpose remote CLI unless they are explicit that this path is admin-only.

More broadly, the docs often say “use a token” where they should say exactly which role is required:

- `admin` required
- `editor` sufficient
- `viewer` sufficient
- no token needed

The difference matters because the repo is explicitly pitching RBAC and operator safety.

### 4. Some docs read like summaries, not procedures

The runbooks frequently describe what the system does, but they do not always function as stress-proof operational procedures.

What is missing too often:

- copy-paste-safe command blocks
- explicit working directory assumptions
- expected output or success criteria
- exact port expectations
- exact file-path expectations
- exact token expectations
- exact failure diagnosis steps in order

The tone says “operator manual,” but parts still behave more like engineering notes.

### 5. The tutorial story is currently not acceptable

`docs/tutorials/trace-one-pipeline.md` is effectively a stub. For a repo that emphasizes teaching value, that is a serious miss.

The tutorial directory explicitly promises a learner walkthrough of the platform experience, and the main pipeline-trace tutorial currently does not deliver one.

If teaching is a core repo goal, an incomplete tutorial is not a minor docs gap. It is a broken product promise.

### 6. There is too much duplication for a repo changing this quickly

The same setup and runtime ideas are repeated across:

- `README.md`
- `docs/runbooks/bootstrap.md`
- `docs/runbooks/operator-manual.md`
- `docs/runbooks/localhost-e2e.md`

In a fast-moving repo, this is dangerous. Repetition makes the docs feel thorough, but it also creates more surfaces for drift. I would rather see:

- one canonical quickstart
- one canonical operator manual
- one canonical debugging runbook
- everything else link back to those

Right now the docs are over-redundant for the speed of iteration happening in the codebase.

### 7. Some prose is too hand-wavey for someone following cold

There are several places where the docs assume knowledge the reader may not have yet.

Examples of the pattern:

- “adjust local paths if needed”
- “start PostgreSQL and the platform services”
- “run the binaries locally”
- “confirm the worker is polling”
- “the stack is healthy”

These are all reasonable phrases for a teammate who already knows the system. They are not strong enough for true onboarding docs.

The docs should reduce ambiguity, not describe the shape of the ambiguity.

### 8. The docs sometimes talk as if the reader already did something

One example is the operator manual line implying “You already installed the main host toolchains.” That kind of phrasing is risky. Docs should never assume successful prior setup unless the preceding step explicitly verified it.

That kind of sentence makes the docs feel less like an operational reference and more like a guided conversation frozen halfway through.

### 9. Recovery docs are directionally good but still not operational enough

The backup docs are honest about restore not being automated yet, which is good.

But the recovery path still stops one level too early. It tells the user what categories of files to restore, but not how to do it safely in a reproducible way. A real recovery runbook should eventually include:

- an example extraction command
- a concrete restore target layout
- warnings about overwriting existing state
- explicit Postgres expectations
- a post-restore validation checklist with expected responses

Right now it is a solid note. It is not yet a safe disaster-recovery procedure.

### 10. The docs need a harder distinction between “verified” and “intended”

This is probably the most important style correction for the whole docs set.

In a repo evolving this fast, every major instruction should be one of:

- verified and currently exercised
- intended but not fully proven
- in progress / not finished

The docs already do some of this, but not consistently enough. When docs are highly polished, readers assume everything described has been equally battle-tested. That is not currently true, and the docs should make the difference more explicit.

## Documentation Recommendations

If documentation followability is truly the top priority, I would do a dedicated docs hardening pass with these goals:

1. Create one canonical quickstart
- Prefer `make smoke` as the first success path.
- Tell the reader exactly why this is the recommended first path.
- State exactly what it validates and what it does not.

2. Split local-host-run and Compose docs aggressively
- Do not mix them in the same instructions unless the distinction is explicit.
- Call out that `.env.example` is Compose-oriented unless exported manually.

3. Add a “Configuration Reality” section
- Explain which config is read from process env.
- Explain whether `.env` is auto-loaded or not.
- Explain default paths for local runs versus Compose.

4. Add role-specific callouts to every privileged workflow
- Mark commands and UI actions as `admin`, `editor`, `viewer`, or anonymous.

5. Replace duplication with canonical ownership
- `README.md`: short orientation and links.
- one quickstart
- one operator manual
- one recovery runbook
- one tutorial that is actually complete

6. Rewrite docs as checklists where possible
- command
- expected result
- if not, check this next

7. Finish or remove incomplete tutorial promises
- An incomplete tutorial is worse than no tutorial if the docs claim it exists.

## Main Engineering Concerns

### 1. The platform hardening story is not yet as strong as the docs suggest

The repo says it now has lightweight RBAC with `viewer`, `editor`, and `admin`, but in the current snapshot the practical read surface is still much more open than that framing implies.

At a minimum, I would revisit whether these should require a bearer token with at least `viewer`:

- analytics endpoints
- catalog endpoints
- run history
- artifacts
- logs
- audit trail

Right now the product language sounds like “role-based platform access,” while the implementation still behaves closer to “anonymous read, restricted write/admin.” That is not necessarily wrong for local-first development, but it should be an explicit product choice, not an accidental middle state.

This is the highest-value issue to tighten because it affects trust in the whole hardening narrative.

### 2. The database-backed control plane still feels half-committed

The repo is clearly moving toward PostgreSQL as the preferred control-plane backend, but some of the implementation still feels like a mirror strategy layered on top of the original filesystem model.

That creates a few risks:

- ambiguity about which store is truly authoritative
- subtle divergence between the file and Postgres views of the world
- design choices optimized for “keep both alive” instead of “make one mode clean”

This is a natural transitional state, but the transition should not stay fuzzy for long. Once the project starts advertising a preferred database-backed control plane, consistency semantics need to be very clear.

### 3. Some read paths are doing too much work

The metadata/catalog path currently behaves more like “sync on read” than “query a stable projected state.” That is acceptable in an early slice, but it is not a good long-term pattern.

A read request should not be the trigger for heavy projection synchronization unless that choice is extremely deliberate and documented. Over time, this will make the system harder to reason about operationally and harder to scale even at modest local/self-hosted usage.

I would prefer one of these models:

- sync metadata during startup and scheduler refresh
- sync metadata on explicit control-plane mutation
- provide an explicit `sync metadata` control-plane action

### 4. Validation is too shallow relative to project ambition

The repo has good smoke coverage for the happy path, but the static validation contract is still fairly narrow.

For a platform repo that wants to be manifest-first and product-grade, I would expect validation to eventually cover:

- pipeline DAG correctness
- supported job types
- existence and validity of SQL refs
- existence and validity of Python task refs
- asset manifest consistency
- quality manifest consistency
- metric manifest consistency
- dashboard manifest consistency
- cross-manifest reference integrity

Right now the repo has enough structure that stronger validation would pay off immediately.

### 5. Frontend confidence is lagging backend confidence

The backend has meaningful tests and smoke coverage. The frontend currently has no actual test files, and `npm test` exits because none exist.

That does not mean the web app is poor quality, but it does mean UI behavior is relying heavily on manual confidence and backend stability. Given how much of the product value now lives in dashboards, datasets, metrics, and system views, I would start adding a minimal frontend safety net soon.

Even a small set would help:

- auth/session capability rendering
- dashboard editor state transitions
- pipeline trigger UX
- system page loading/error states

### 6. Real-time operator ergonomics are still thin

The UI is already useful, but much of it is still “load once and render” rather than “behave like an operational surface.”

The places where this is most noticeable:

- recent runs
- logs
- audit feed
- system health

For a platform product, these views should eventually feel alive. Polling, refresh controls, or lightweight live updates would make the tool feel much more trustworthy during actual operations.

## Architectural Advice

### 1. Keep resisting microservice sprawl

The current modular monolith approach is the right backbone for this repo. Do not break this into separate deployables prematurely unless there is a very specific operational need.

The current code organization is already expressive enough to support future extraction if it becomes necessary.

### 2. Be stricter about “source of truth” decisions

A recurring theme in the repo is duality:

- filesystem plus Postgres
- manifest truth plus projected truth
- local-first plus more normalized control plane

That is workable in transition, but each subsystem should eventually answer these questions clearly:

- what is authoritative
- what is derived
- when is synchronization allowed
- what happens when stores diverge

The stronger those answers get, the more mature the platform will feel.

### 3. Separate “teaching clarity” from “runtime coupling”

The project’s educational posture is a strength. The risk is that explanatory structure can sometimes preserve transitional runtime patterns longer than necessary.

In other words:

- keep the docs detailed
- keep the package boundaries clear
- but do not keep awkward execution or persistence behavior just because it is easier to explain

If a subsystem needs a cleaner internal contract, prefer that and document it well.

## Product Advice

### 1. The repo is strongest when it behaves like a platform, not a finance demo

The finance slice is doing its job well as the proving ground, but the long-term value is the platform behavior around it:

- orchestration discipline
- control-plane safety
- metadata trust
- operational UX
- recovery readiness

Future work should continue shifting emphasis from “more finance features” to “more platform rigor.”

### 2. Operator trust should be a first-class design principle

If I reduce the whole roadmap to one theme, it would be this:

Build confidence that the platform is telling the truth.

That means:

- auth behavior matches docs
- audit behavior is complete and explainable
- freshness signals are trustworthy
- validation catches broken manifests before runtime
- backups are restorable, not just creatable
- UI surfaces reflect current state, not stale snapshots

That is the difference between an impressive engineering project and a believable platform.

## Suggested Next Priorities

If I were sequencing the next high-leverage pass, I would do this:

1. Access model pass
- Decide whether anonymous read access is intentional.
- If not, require `viewer` for all non-health product endpoints.
- Make the docs and UI capability model match the actual enforcement model.

2. Control-plane consistency pass
- Tighten Postgres/file semantics.
- Resolve any stale or ambiguous mirrored-state behavior.
- Clarify which paths are fallback-only and which are preferred-mode behavior.

3. Validation pass
- Expand `platformctl validate-manifests` into a real contract checker.
- Catch broken refs and unsupported manifest states before runtime.

4. Frontend confidence pass
- Add a small but meaningful UI test suite.
- Add refresh behavior to operational views.

5. Recovery trust pass
- Keep investing in restore realism, not only backup generation.
- Make disaster-recovery expectations as concrete as smoke-test expectations.

## Bottom Line

This is a strong repo.

It already has:

- a clear thesis
- solid repo structure
- unusually good documentation
- real end-to-end proof
- promising product instincts

The main thing holding it back is not lack of effort or lack of architecture. It is that some of the newer “platform hardening” claims are slightly ahead of the current enforcement, validation, and consistency guarantees.

That is a normal in-flight state. It just means the next best move is probably not more feature breadth. It is tightening the contract between what the platform says it is and what the runtime can reliably guarantee.

## What I Verified During Review

I reviewed:

- repo docs and architecture notes
- backend runtime wiring
- orchestration, execution, scheduler, metadata, auth, reporting, and storage paths
- frontend app structure and key hooks/pages
- manifests, SQL, and Python task paths
- smoke/test tooling

I also ran:

- `cd backend && go test ./...`
- `cd backend && go run ./cmd/platformctl validate-manifests`
- localhost smoke flow through the repo script

Current test note:

- backend tests passed in this snapshot
- frontend `npm test` currently reports no test files

## Async Handoff Framework For The Next Model

This section is written as an execution contract for the next model working on the repo.

The goal is not to “consider” these issues. The goal is to close them in a disciplined order and avoid returning to feature expansion before the platform and docs are more trustworthy.

## Operating Rule

Do not move on to new feature breadth until the current tranche of trust, docs, and control-plane concerns is explicitly closed or consciously deferred with a written reason.

If something is partially fixed, do not mark the area complete. Partial closure is how this repo drifts into polished-but-ambiguous behavior.

## Priority Order

Work in this order unless there is a strong reason not to:

1. documentation followability
2. auth and access-model clarity
3. control-plane consistency
4. validation depth
5. frontend confidence and operator ergonomics
6. recovery realism
7. only then resume broader feature work

## Global Gating Rules

Before any area is considered complete:

- the implementation behavior must match the docs
- the docs must match the implementation behavior
- the happy path must be runnable without hidden assumptions
- there must be an explicit verification step
- any unfinished edge must be called out plainly in docs

If any of those are false, the area is not done.

## Workstream 1: Documentation Hardening

Do not leave this workstream until all of the following are true:

- there is one canonical quickstart
- there is one canonical operator path
- local host-run and Compose instructions are clearly separated
- `.env` behavior is explicitly explained
- token/role requirements are called out per workflow
- tutorials are either complete or removed from “start here” flows
- each runbook contains expected results, not just commands
- duplication across README/runbooks is reduced enough that drift risk is materially lower

Checklist:

- [ ] Create or rewrite one canonical quickstart with the easiest verified first-run path.
- [ ] Explicitly state whether `.env` is auto-loaded or not.
- [ ] Explain the difference between Compose-oriented env values and local host-run defaults.
- [ ] Mark every privileged workflow with required role: `admin`, `editor`, `viewer`, or anonymous.
- [ ] Add “expected success result” blocks to startup, run, backup, and verification docs.
- [ ] Add “if this fails, check this next” blocks to the main runbooks.
- [ ] Finish `trace-one-pipeline.md` or stop presenting it as an existing tutorial.
- [ ] Remove or reduce duplicated startup instructions across overlapping docs.

Definition of done:

- A new reader can get to one successful smoke run and one successful packaged boot without guessing.
- A new reader can tell which token they need for each action without reading code.
- There are no major instructions that depend on hidden shell-export behavior.

## Workstream 2: Auth And Access Model

Do not leave this workstream until the product story and implementation story are the same.

Checklist:

- [ ] Decide whether anonymous read access is intentional.
- [ ] If anonymous read access is intentional, document it explicitly.
- [ ] If it is not intentional, require at least `viewer` for non-health read endpoints.
- [ ] Ensure UI capability messaging matches real backend enforcement.
- [ ] Ensure admin-only surfaces are consistently described as admin-only.
- [ ] Re-check audit visibility, logs visibility, artifacts visibility, and analytics visibility under the chosen policy.

Definition of done:

- A user reading the docs can predict endpoint and UI access behavior correctly.
- Role names in docs are operationally meaningful, not aspirational.

## Workstream 3: Control-Plane Consistency

This repo needs a cleaner answer to “what is authoritative when Postgres is present.”

Checklist:

- [ ] Decide the authoritative source of truth per subsystem.
- [ ] Document primary versus fallback behavior for runs, queue, artifacts, dashboards, audit, and metadata.
- [ ] Remove or tighten any mirrored-state behavior that can produce stale or divergent reads.
- [ ] Review timestamp/update semantics in the preferred Postgres path.
- [ ] Review any read-time synchronization behavior and move it to a cleaner lifecycle if possible.

Definition of done:

- A maintainer can explain where state comes from in each runtime mode without hedging.
- Preferred-mode behavior is cleaner than fallback-mode behavior, not more confusing.

## Workstream 4: Validation Expansion

The repo needs a stronger static contract before runtime.

Checklist:

- [ ] Extend manifest validation beyond DAG shape.
- [ ] Validate supported job types and required fields.
- [ ] Validate SQL refs exist.
- [ ] Validate Python task refs exist.
- [ ] Validate dashboard manifests.
- [ ] Validate metric manifests.
- [ ] Validate quality manifests.
- [ ] Validate cross-manifest references.
- [ ] Fail loudly on unsupported or orphaned definitions.

Definition of done:

- Obvious configuration errors are caught before a smoke run or manual operation.
- `platformctl validate-manifests` feels like a real contract checker, not a partial sanity check.

## Workstream 5: Frontend Confidence And Ops UX

Checklist:

- [ ] Add a minimal frontend test suite.
- [ ] Cover auth/session capability rendering.
- [ ] Cover dashboard editing state transitions.
- [ ] Cover pipeline trigger UX.
- [ ] Cover critical loading/error states for system views.
- [ ] Add refresh or polling behavior for operationally live pages.

Definition of done:

- The frontend has at least a basic regression net.
- The main operator views do not feel static or stale during active use.

## Workstream 6: Recovery Realism

Checklist:

- [ ] Tighten the backup docs into an operationally safer recovery guide.
- [ ] Add explicit extraction and restore examples.
- [ ] Add post-restore validation steps with expected outcomes.
- [ ] Clarify what is restored into filesystem state versus recreated in Postgres.
- [ ] If possible, add a more concrete restore drill before calling the backup path mature.

Definition of done:

- Recovery guidance is runnable, not merely descriptive.
- Backup confidence comes from actual restore thinking, not bundle creation alone.

## “Do Not Move On Until” Checklist

The next model should not return to normal feature expansion until these are all true:

- [ ] Docs first-run path is unambiguous.
- [ ] `.env` and runtime config behavior are explained correctly.
- [ ] Access model is explicit and enforced consistently.
- [ ] Preferred control-plane semantics are clear.
- [ ] Validation meaningfully covers repo-managed definitions.
- [ ] Frontend has at least a basic test floor.
- [ ] Recovery docs are concrete enough to follow under pressure.

## Suggested Working Method

For each workstream:

1. restate the current problem in one paragraph
2. make the implementation change
3. update the docs immediately after
4. run the smallest meaningful verification
5. record what is now true and what is still deferred

Do not batch all docs updates until the end. This repo is moving too fast for that.

## Required Closeout Format For Each Completed Area

When the next model finishes a workstream, it should leave behind a short note in its own summary covering:

- what changed
- what is now verifiably true
- what remains intentionally unfinished
- what command or workflow was used to verify it

That makes the work inspectable for the next pass instead of forcing future reviewers to infer closure from diffs.

## Final Instruction To The Next Model

Optimize for trustworthiness over surface area.

The repo already has enough ambition and enough feature shape. The highest-value work now is making the system easier to believe, easier to operate, and easier to follow without guessing.
