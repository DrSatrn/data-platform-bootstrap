# Contributing Guide

> **Audience:** Someone who can write scripts, has never shipped production software, and may or may not have access to an AI coding assistant. This guide treats you as a capable adult who just hasn't been exposed to the rituals of professional software projects yet.

---

## Table of Contents

1. [Mental Models Before You Touch Anything](#mental-models-before-you-touch-anything)
2. [Your Local Toolchain](#your-local-toolchain)
3. [The Golden Rule: Verify Before You Push](#the-golden-rule-verify-before-you-push)
4. [Understanding Git Workflow](#understanding-git-workflow)
5. [Making Your First Change (Step-By-Step)](#making-your-first-change-step-by-step)
6. [Where To Make Changes](#where-to-make-changes)
7. [Backend Development (Go)](#backend-development-go)
8. [Frontend Development (React + TypeScript)](#frontend-development-react--typescript)
9. [SQL and Data Changes](#sql-and-data-changes)
10. [YAML Manifest Changes](#yaml-manifest-changes)
11. [Documentation Changes](#documentation-changes)
12. [Writing Tests](#writing-tests)
13. [The Verification Ritual](#the-verification-ritual)
14. [Working WITH AI Assistance](#working-with-ai-assistance)
15. [Working WITHOUT AI Assistance](#working-without-ai-assistance)
16. [Common Mistakes and How To Avoid Them](#common-mistakes-and-how-to-avoid-them)
17. [How To Debug When Things Break](#how-to-debug-when-things-break)
18. [Maintenance Tasks](#maintenance-tasks)
19. [Project Coordination Files](#project-coordination-files)
20. [Growing As An Engineer Through This Project](#growing-as-an-engineer-through-this-project)

---

## Mental Models Before You Touch Anything

### Scripts vs. Services

When you write a Python script, you run it and it finishes. A service is a program that starts and keeps running, waiting for work. This project has three services (API, Worker, Scheduler) that stay alive and communicate through a shared database and file queue. If you change the behavior of one, you need to think about how it affects the others.

### "It Works On My Machine" Is Not Enough

In script-land, if your script runs, it works. In a production codebase:
- Does it still compile?
- Do the existing tests still pass?
- Did you break someone else's feature?
- Does the Docker build still succeed?
- Does the smoke test still pass?

This is why we have verification commands. Always run them.

### Files Have Owners

Not every file is equally safe to edit. Some files are "hot" — they are central wiring that everything depends on. Others are "safe" — they are isolated and additive. The guide-wire.md file in the repo root tracks which is which.

**Hot files** (edit carefully):
- `backend/internal/app/runtime.go` — wires the entire backend together
- `backend/internal/config/config.go` — every service reads this
- `backend/cmd/platformctl/main.go` — the CLI entrypoint
- `infra/compose/docker-compose.yml` — the entire runtime topology

**Safe files** (edit freely):
- Anything under `docs/`
- New test files (`*_test.go`, `*.test.ts`)
- New files in existing packages
- SQL files under `packages/sql/`
- Python tasks under `packages/python/`

---

## Your Local Toolchain

Before contributing, make sure you have these installed:

### Required Tools

| Tool | What It's For | Install Command | Verify |
|------|-------------|----------------|--------|
| **Go 1.24+** | Compiles and runs the backend | `brew install go` | `go version` |
| **Node.js 20+** | Runs the frontend build and dev server | `brew install node` | `node --version` |
| **Python 3** | Runs data enrichment tasks in the pipeline | `brew install python3` | `python3 --version` |
| **Docker** or **OrbStack** | Runs containers for Compose deployments | `brew install orbstack` | `docker --version` |
| **Xcode CLI Tools** | C compiler needed by DuckDB's Go driver (CGO) | `xcode-select --install` | `gcc --version` |
| **Git** | Version control | `brew install git` | `git --version` |

### Optional But Helpful

| Tool | What It's For |
|------|-------------|
| **VS Code** or **Cursor** | Code editor with Go and TypeScript support |
| **Go extension for VS Code** | Autocompletion, linting, test running |
| **GitHub Copilot / Cursor AI** | AI coding assistance (covered in depth below) |

### First-Time Setup

```bash
# Clone the repo
git clone https://github.com/streanor/data-platform.git
cd data-platform

# Set up environment variables
cp .env.example .env

# Install frontend dependencies
cd web && npm install && cd ..

# Verify Go dependencies are resolved
cd backend && go mod download && cd ..

# Run the smoke test to make sure everything works
make smoke
```

If `make smoke` passes, your toolchain is working.

---

## The Golden Rule: Verify Before You Push

Before you commit ANYTHING, run these commands in order:

```bash
# 1. Do the backend tests pass?
cd backend && go test ./...

# 2. Do the manifests validate?
cd backend && go run ./cmd/platformctl validate-manifests

# 3. Does the frontend build?
cd web && npm run build

# 4. Do the frontend tests pass?
cd web && npm test
```

If any of these fail, **do not commit**. Fix the failure first.

For bigger changes, also run:
```bash
# 5. Does the full smoke test pass?
make smoke
```

Think of these as your "safety net." Professional engineers run these dozens of times a day. It's not a chore — it's how you catch mistakes before they become real problems.

---

## Understanding Git Workflow

If you've only used Git to `git add . && git commit && git push`, here's the professional workflow:

### Branching

Never work directly on `main`. Always create a branch:

```bash
# Create a new branch for your work
git checkout -b feature/add-new-quality-check

# Do your work, then commit
git add -A
git commit -m "Add null check for raw_transactions.amount column"

# Push your branch
git push origin feature/add-new-quality-check
```

### Branch Naming Conventions

Use descriptive, lowercase, hyphenated names:
- `feature/add-budget-forecast-mart` — new functionality
- `fix/dashboard-filter-not-applied` — bug fix
- `docs/improve-quickstart-guide` — documentation only
- `test/add-analytics-service-tests` — test additions
- `chore/update-go-dependencies` — maintenance

### Commit Messages

Write commit messages that explain **why**, not just **what**:

```
# Bad
"fixed stuff"
"update file"
"WIP"

# Good
"Fix dashboard preset filters not applying to bar chart widgets"
"Add quality check for null transaction amounts"  
"Update quickstart.md to clarify .env loading behavior"
```

### When To Commit

- Commit after each logically complete change
- Don't bundle unrelated changes into one commit
- It's fine to have many small commits — that's better than one giant one

---

## Making Your First Change (Step-By-Step)

Let's walk through a real example: adding a new data quality check.

### Step 1: Understand what you're changing

A quality check is a SQL query that verifies data integrity. The existing checks live in `packages/sql/quality/`. The pipeline runs them after the transforms complete.

### Step 2: Create a branch

```bash
git checkout -b feature/add-null-amount-quality-check
```

### Step 3: Write the SQL

Create a new file `packages/sql/quality/check_null_amounts.sql`:

```sql
-- This check counts transactions with null or zero amounts,
-- which would indicate data quality issues in the raw source.
select count(*) as null_amount_count
from raw_transactions
where amount is null or amount = 0;
```

### Step 4: Add a quality manifest

Create or update a quality manifest in `packages/manifests/quality/` to register your new check. Look at the existing YAML files for the pattern.

### Step 5: Verify it works

```bash
cd backend && go run ./cmd/platformctl validate-manifests
cd backend && go test ./...
```

### Step 6: Commit and push

```bash
git add -A
git commit -m "Add quality check for null/zero transaction amounts"
git push origin feature/add-null-amount-quality-check
```

---

## Where To Make Changes

### The Easiest Contributions (Start Here)

These require the least knowledge of the system:

1. **Documentation improvements** — Fix typos, clarify instructions, add examples
2. **New SQL quality checks** — Write a SQL query, add a manifest entry
3. **New SQL transforms** — Write a SQL query that creates a new analytical view
4. **Dashboard manifest changes** — Edit `packages/dashboards/finance_overview.yaml`

### Intermediate Contributions

These require understanding Go or TypeScript basics:

5. **New backend tests** — Write `*_test.go` files for existing packages
6. **New frontend tests** — Write `*.test.ts` or `*.test.tsx` files
7. **Frontend UI tweaks** — Modify page components in `web/src/pages/`
8. **New admin terminal commands** — Add handlers in `backend/internal/admin/`

### Advanced Contributions

These require understanding the system architecture:

9. **New backend packages** — Create new directories under `backend/internal/`
10. **Database migrations** — Add new PostgreSQL schema files
11. **Runtime wiring changes** — Modify how services connect to each other
12. **New pipeline job types** — Extend the execution engine

---

## Backend Development (Go)

### How Go Code Is Organized In This Repo

Each directory under `backend/internal/` is a "package." A package is a group of related Go files that:
- Share the same `package` declaration at the top of each file
- Can access each other's private (lowercase) functions
- Expose public (uppercase) functions to other packages

### Reading Go Code If You've Never Seen It

```go
// This is a comment. Package-level comments explain what the package does.
package reporting

// Import block — these are dependencies (like Python's import)
import (
    "encoding/json"    // Standard library: JSON encoding
    "fmt"              // Standard library: string formatting
    "os"               // Standard library: file I/O
)

// A struct is like a Python class (but without methods attached inline)
type Dashboard struct {
    ID   string `json:"id"`    // The `json:"id"` part controls JSON serialization
    Name string `json:"name"`
}

// A method on a struct — the (s *FileStore) part means "this method belongs to FileStore"
func (s *FileStore) ListDashboards() ([]Dashboard, error) {
    // Functions return multiple values. The second return is always an error.
    bytes, err := os.ReadFile(s.path)
    if err != nil {
        return nil, fmt.Errorf("read dashboards: %w", err)  // Wrap errors with context
    }
    // ...
}
```

### Key Go Patterns In This Codebase

1. **Interfaces**: The repo uses interfaces (like `Store`, `RunQueue`) so implementations can be swapped. `FileStore` and `DBStore` both implement the same `Store` interface.
2. **Error wrapping**: Errors are wrapped with `fmt.Errorf("context: %w", err)` so you can trace where things went wrong.
3. **Mutex locks**: Shared state uses `sync.RWMutex` for thread safety. You'll see `s.mu.Lock()` and `defer s.mu.Unlock()` everywhere.

### Running Individual Go Tests

```bash
# Run all tests
cd backend && go test ./...

# Run tests for one specific package
cd backend && go test ./internal/reporting/...

# Run one specific test function
cd backend && go test ./internal/reporting/... -run TestListDashboards

# Run tests with verbose output (see each test name)
cd backend && go test -v ./internal/reporting/...
```

### Adding a New Go Test

Create a file ending in `_test.go` in the same directory as the code you're testing:

```go
// File: backend/internal/reporting/store_test.go
package reporting

import "testing"

func TestValidateDashboard_RequiresID(t *testing.T) {
    dashboard := Dashboard{Name: "Test", Widgets: []DashboardWidget{{ID: "w1", Name: "W", DatasetRef: "test"}}}
    err := validateDashboard(dashboard)
    if err == nil {
        t.Fatal("expected error for missing dashboard ID")
    }
}
```

---

## Frontend Development (React + TypeScript)

### How The Frontend Is Structured

The frontend follows a feature-based organization:
- `pages/` — One file per screen (what the user sees)
- `features/` — Business logic grouped by domain (auth, pipelines, dashboards)
- `components/` — Reusable UI building blocks
- `lib/` — Shared utilities (API client, formatters)

### Running The Frontend Dev Server

```bash
cd web
npm install    # Only needed once or when dependencies change
npm run dev    # Starts Vite dev server on http://localhost:3000
```

The dev server has **hot module replacement (HMR)** — it automatically refreshes the browser when you save a file. This is the fastest feedback loop for UI changes.

### Running Frontend Tests

```bash
cd web
npm test              # Run all tests
npm test -- --watch   # Re-run tests automatically on file changes
```

### Making a Frontend Change

Example: changing the dashboard page title.

1. Open `web/src/pages/DashboardPage.tsx`
2. Find the heading text
3. Change it
4. Save — the browser updates instantly (thanks to Vite HMR)
5. Run `npm run build` to ensure the TypeScript compiler is happy
6. Run `npm test` to ensure you didn't break existing tests

---

## SQL and Data Changes

### Adding a New SQL Transform

1. Write your SQL file in `packages/sql/transforms/`:
   ```sql
   -- File: packages/sql/transforms/monthly_savings_target.sql
   create or replace table mart_monthly_savings_target as
   select
       month,
       income * 0.20 as savings_target,
       income - expenses as actual_savings
   from mart_monthly_cashflow;
   ```

2. Register the transform as a job in the pipeline manifest (`packages/manifests/pipelines/personal_finance_pipeline.yaml`). You'll need to add a new job entry with `type: transform_sql` and a `transform_ref` pointing to your SQL.

3. Register the output as a data asset (`packages/manifests/assets/`). Create a new YAML file describing the columns, owner, and data layer.

4. Validate: `cd backend && go run ./cmd/platformctl validate-manifests`

### Adding a New Metric

Same pattern but use `packages/sql/metrics/` for the SQL and `packages/manifests/metrics/` for the manifest.

---

## YAML Manifest Changes

Manifests are the "declarative truth" of the system. They define:
- What pipelines exist and what jobs they contain
- What data assets exist and their schemas
- What quality checks enforce
- What metrics are computed
- What dashboards display

### The Validation Safety Net

Every manifest change should be validated:

```bash
cd backend && go run ./cmd/platformctl validate-manifests
```

This catches:
- Missing required fields
- Invalid job types
- Broken cross-references between manifests
- Duplicate IDs

If this command passes, your YAML is structurally correct.

---

## Documentation Changes

Docs are the easiest and most impactful contribution. Here's what goes where:

| Directory | Purpose | Style |
|-----------|---------|-------|
| `docs/runbooks/` | Step-by-step operational procedures | Copy-paste safe commands, expected outputs |
| `docs/tutorials/` | Learning walkthroughs | "Follow along and learn" tone |
| `docs/architecture/` | System design explanations | Diagrams, data flow descriptions |
| `docs/decisions/` | Why we made specific choices | Problem → Options → Decision → Consequences |
| `docs/reference/` | API docs, config reference | Exhaustive, lookup-friendly |
| `docs/product/` | Feature requirements and blueprints | What should the product do? |

### Documentation Quality Bar

Every runbook should include:
- The exact command to run
- The expected output (what "success" looks like)
- What to do if it fails
- What role/token is required (if applicable)

---

## Writing Tests

### Why Tests Matter (The Non-Obvious Truth)

Tests aren't about proving your code works today. They're about catching when someone else (or future-you) accidentally breaks it tomorrow. Every test you write is a permanent guardian for that behavior.

### What To Test

- **Happy path**: Does the normal case work?
- **Edge cases**: What happens with empty input? nil values? very large data?
- **Error cases**: Does the system fail gracefully with bad input?

### Test File Naming

- Go: `*_test.go` (same directory as the code)
- TypeScript: `*.test.ts` or `*.test.tsx` (same directory or `tests/`)

---

## The Verification Ritual

Run these in order before every commit. Memorize them.

```bash
# Backend
cd backend && go test ./...
cd backend && go run ./cmd/platformctl validate-manifests

# Frontend
cd web && npm run build
cd web && npm test

# Full stack (for significant changes)
make smoke
```

Create a shell alias if you do this often:

```bash
# Add to your ~/.zshrc
alias platform-check="cd ~/Documents/Playground/data-platform && cd backend && go test ./... && go run ./cmd/platformctl validate-manifests && cd ../web && npm run build && npm test && cd .."
```

---

## Working WITH AI Assistance

This section is for when you have access to AI coding tools (Cursor, Copilot, Codex, etc).

### What AI Is Good At For This Project

1. **Explaining existing code**: "What does `buildRuntimePersistence` in `runtime.go` do?"
2. **Writing boilerplate**: Test files, YAML manifests, SQL queries
3. **Refactoring**: "Rename this function and update all call sites"
4. **Debugging**: "This error shows up when I run the worker. What's causing it?"
5. **Generating documentation**: README sections, code comments
6. **Writing migrations**: "Add a `tags` column to the `dashboards` table"

### What AI Is Bad At For This Project

1. **Understanding the full system holistically** — AI works in windows. It can't hold all 22 backend packages in context simultaneously.
2. **Making architectural decisions** — It will happily build whatever you ask for, even if it's wrong.
3. **Knowing what NOT to change** — It doesn't know about `guide-wire.md` or hot-file rules unless you tell it.
4. **Testing its own output** — It can write code that looks correct but doesn't compile or pass tests.

### The AI Workflow

1. **Always provide context**: Point the AI at `codex.md` first. That file exists specifically to onboard AI agents.
2. **Ask it to read before writing**: "Read `backend/internal/reporting/store.go` and then add a test for the `SaveDashboard` method."
3. **Verify immediately**: After AI generates code, run the verification commands before committing.
4. **Review the diff**: `git diff` shows exactly what changed. Read every line. AI sometimes makes subtle mistakes (wrong package name, incorrect import path).
5. **Use `codex.md` as the prompt seed**: The "Best next session starting point" section tells AI what to work on next.

### Using Codex (OpenAI) or Similar Autonomous Agents

The repo is designed for agentic AI workflows:

```
# The prompt pattern that works with this repo:
"Read codex.md and new-thread-eng-feedback.md in the data-platform repo.
Then execute Workstream 1 from the feedback document, following the 
checklist items in order. After each change, run `go test ./...` 
and `go run ./cmd/platformctl validate-manifests` to verify."
```

Key files for AI onboarding:
- `codex.md` — Full project context, what's built, what's pending, architectural rules
- `new-thread-eng-feedback.md` — Prioritized workstreams with checklists
- `plan.md` — Short-horizon build plan
- `guide-wire.md` — Which files are safe to edit vs. hot

### Checking AI's Work

After any AI session:

```bash
# 1. See what changed
git diff --stat

# 2. Read the actual changes
git diff

# 3. Run the full verification
cd backend && go test ./...
cd backend && go run ./cmd/platformctl validate-manifests
cd web && npm run build && npm test

# 4. Run the smoke test for big changes
make smoke
```

---

## Working WITHOUT AI Assistance

This section is for when you don't have AI tools available. Everything here uses only your editor, terminal, and brain.

### How To Understand Code You Didn't Write

This is the most important skill. Here's the systematic approach:

#### 1. Start from the entry point

Every program starts somewhere. In this repo:
- API starts at `backend/cmd/platform-api/main.go`
- Worker starts at `backend/cmd/platform-worker/main.go`
- Scheduler starts at `backend/cmd/platform-scheduler/main.go`

Open the `main.go` file. It will call something in `backend/internal/app/`. Follow that call.

#### 2. Use grep to trace function calls

```bash
# "Where is this function defined?"
grep -rn "func NewService" backend/internal/

# "Where is this function called?"
grep -rn "NewService(" backend/internal/

# "What files import this package?"
grep -rn '"github.com/streanor/data-platform/backend/internal/reporting"' backend/
```

This is how professional engineers navigate large codebases. It's not glamorous, but it works.

#### 3. Read the package comment

Every package in this repo has a comment at the top of the first file explaining what it does. Read it.

```go
// Package execution runs queued pipeline jobs and materializes local artifacts
// for the first end-to-end finance slice.
package execution
```

#### 4. Read the tests

Tests are executable documentation. They show you exactly how each function is intended to be used.

```bash
# Find all test files for a package
ls backend/internal/reporting/*_test.go
```

### How To Debug Without AI

#### The print statement approach

Go uses `fmt.Println()` or the structured logger:

```go
// Quick and dirty debugging
fmt.Printf("DEBUG: dashboard ID is %s\n", dashboard.ID)

// Using the repo's structured logger (preferred)
r.logger.Info("debug checkpoint", slog.String("dashboard_id", dashboard.ID))
```

#### Reading logs

When running locally:
```bash
# API logs appear in the terminal where you ran the API
cd backend && go run ./cmd/platform-api

# Worker logs appear in its terminal
cd backend && go run ./cmd/platform-worker
```

When running in Docker Compose:
```bash
# See logs for a specific service
docker compose -f infra/compose/docker-compose.yml logs api
docker compose -f infra/compose/docker-compose.yml logs worker

# Follow logs in real-time
docker compose -f infra/compose/docker-compose.yml logs -f api
```

#### The "binary search" debugging technique

If something breaks and you don't know why:
1. Find the last commit that worked: `git log --oneline -20`
2. Check out that commit: `git checkout <commit-hash>`
3. Verify it works: `make smoke`
4. Come back to your branch: `git checkout your-branch`
5. Now you know the breakage happened between those two commits

### Learning Go Without AI

**Free resources:**
- [Go Tour](https://go.dev/tour/) — Interactive browser-based Go tutorial (start here)
- [Go by Example](https://gobyexample.com/) — Quick reference with runnable examples
- [Effective Go](https://go.dev/doc/effective_go) — The official style guide

**The patterns you'll see most in this repo:**
- Interfaces + Structs (polymorphism)
- Error handling with `if err != nil`
- Goroutines and channels (concurrency)
- `sync.Mutex` (thread-safe shared state)
- `defer` (cleanup code that runs when a function exits)

### Learning React/TypeScript Without AI

**Free resources:**
- [React docs](https://react.dev/learn) — The official tutorial
- [TypeScript handbook](https://www.typescriptlang.org/docs/handbook/) — Official guide

**The patterns you'll see most in this repo:**
- Functional components with hooks (`useState`, `useEffect`)
- Custom hooks (`useAuth`, `useDatasets`) that encapsulate data fetching
- Conditional rendering based on auth state
- Fetch API calls to the backend

---

## Common Mistakes and How To Avoid Them

### 1. Editing `.env.example` when you mean `.env`

`.env.example` is tracked in Git and visible to everyone. Your actual secrets go in `.env` (which is gitignored). Never put real tokens in `.env.example`.

### 2. Forgetting to run `npm install` after pulling

If someone else added a frontend dependency, your local `node_modules/` is out of date:
```bash
cd web && npm install
```

### 3. Committing `var/` or `node_modules/`

These are gitignored for good reason. If `git status` shows them, something is wrong with your `.gitignore`.

### 4. Editing `docker-compose.yml` for local hacks

If you need to change ports or env vars temporarily, use `.env.compose` (which is gitignored) instead of editing the tracked Compose file.

### 5. Making changes without understanding the data flow

Before adding a feature, trace the data flow through the existing system. Read `infra-overview.md` first.

---

## How To Debug When Things Break

### "The backend won't compile"

```bash
cd backend && go build ./...
# Read the error message carefully. It tells you the file and line number.
```

### "The frontend won't build"

```bash
cd web && npm run build
# TypeScript errors show the file, line, and what type was expected vs. received.
```

### "The smoke test fails"

```bash
make smoke
# Look at the output for the first FAIL line
# Check the log files it prints (api.log, worker.log, scheduler.log)
```

### "Docker Compose won't start"

```bash
# Check which services are unhealthy
docker compose -f infra/compose/docker-compose.yml ps

# Check logs for the broken service
docker compose -f infra/compose/docker-compose.yml logs api

# Nuclear option: tear everything down and rebuild
make down
docker system prune -f
make bootstrap
```

### "The pipeline runs but produces no data"

1. Check the run status in the Pipelines UI
2. Check `var/data/raw/` — did ingestion produce files?
3. Check `var/data/mart/` — did transforms produce output?
4. Check worker logs for errors

---

## Maintenance Tasks

### Updating Go Dependencies

```bash
cd backend
go get -u ./...          # Update all dependencies
go mod tidy              # Remove unused dependencies
go test ./...            # Make sure nothing broke
```

### Updating Frontend Dependencies

```bash
cd web
npm update               # Update within version ranges
npm audit                # Check for security vulnerabilities
npm run build            # Make sure nothing broke
npm test                 # Run tests
```

### Running Database Migrations

Migrations run automatically on `make bootstrap`. To run them manually:
```bash
cd backend && go run ./cmd/platformctl migrate
```

### Creating a Backup

```bash
make backup
# Output: var/backups/platform-backup-<timestamp>.tar.gz
```

### Verifying a Backup

```bash
make restore-drill
# This extracts to /tmp without overwriting your live data
```

---

## Project Coordination Files

This repo has several files that coordinate work. Here's what each one is for:

| File | Purpose | When To Update |
|------|---------|---------------|
| `codex.md` | Onboarding context for AI agents | After every significant change |
| `plan.md` | Short-horizon implementation plan | When starting a new work block |
| `guide-wire.md` | Tracks safe vs. hot files for parallel work | When adding new files or starting parallel work |
| `new-thread-eng-feedback.md` | Architectural review and prioritized workstreams | After completing a workstream |
| `README.md` | Public-facing project overview | When features or setup steps change |

**Rule**: If you make a significant change, update `codex.md` so the next person (or AI) can pick up where you left off.

---

## Growing As An Engineer Through This Project

This repo is intentionally designed to teach good engineering. Here's a learning roadmap:

### Week 1-2: Read and Understand
- Read `infra-overview.md` end to end
- Run `make smoke` and `make bootstrap`
- Open the UI, trigger a pipeline, explore dashboards
- Read 2-3 backend packages (start with `reporting/` and `analytics/`)

### Week 3-4: Small Documentation Contributions
- Fix a typo in a runbook
- Add a "what to check if this fails" section to an existing doc
- Write a new tutorial based on something you struggled to understand

### Month 2: First Code Contributions
- Add a new SQL quality check
- Add a new test for an existing Go function
- Add a frontend test for a page component

### Month 3: Deeper Engineering
- Add a new SQL transform and wire it into the pipeline
- Add a new admin terminal command
- Modify a frontend page to display new data

### Month 4+: System-Level Work
- Add a new database migration
- Create a new backend package
- Extend the execution engine with a new job type

### The Meta-Skill

The most important thing you'll learn from this project is not Go or React or SQL. It's **how to navigate and reason about a system you didn't design**. That skill transfers to every engineering role you'll ever have.

When you encounter something confusing:
1. Don't panic
2. Read the package comment
3. Read the tests
4. Grep for usage
5. Trace the data flow from entry point to output
6. If you're still stuck, write down what you DO understand and what you DON'T

That process is what professional engineers do every single day. You're not behind — you're practicing.
