# Data Platform Local Setup & UAT Guide

This document provides singular, step-by-step instructions for getting the Data Platform running on your laptop. It also covers how to use the GUI for User Acceptance Testing (UAT) and highlights what currently works and what doesn't.

## Part 1: Starting the Platform

The easiest and most reliable way to start the platform is using the pre-configured Docker Compose environment, which packages the UI, Postgres, API, Scheduler, and Worker.

### 1. Configure the Environment
The repository comes with safe local defaults. We will use the provided example configuration to quickly establish static user tokens for our testing.
Run the following commands in the root of the repository:
```bash
cp .env.example .env
cp .env.compose.example .env.compose
```

### 2. Start the Monolith
Use the provided `Makefile` target to build and detach the entire stack:
```bash
make bootstrap
```
_Note: On first run, this may take a minute or two to build the containers. The command will exit once the images are built and spun up in the background._

To verify everything is healthy, you can run:
```bash
docker compose -f infra/compose/docker-compose.yml ps
```
You should see healthy/running states for `api`, `postgres`, `worker`, `scheduler`, and `web`.

### 3. Stopping the Platform
When you are completely finished with your session, you can tear everything down by running:
```bash
make down
```

---

## Part 2: UAT via the GUI

Once the stack is running, the Web UI is available locally.

### 1. Access the App
Open your browser and navigate to:
**[http://127.0.0.1:3000](http://127.0.0.1:3000)**

### 2. Authentication (Tokens)
The system currently uses statically defined bearer tokens for role-based access control (RBAC). When you visit the site, you will be an "anonymous" user and see very little. 

To login, use the token input (or authorization header workflow in the UI) and enter one of the following depending on what you want to test:
- **Viewer Role:** `viewer-token` (Can browse catalogs, dashboards, and runs)
- **Editor Role:** `editor-token` (Can trigger pipelines and create/edit dashboards)
- **Admin Role:** `local-dev-admin-token` (Full access, including the Admin Terminal)

_Recommendation: For full UAT, log in with `local-dev-admin-token`._

### 3. Key Workflows to Test
1. **Trigger a Pipeline:**
   - Navigate to the **Pipelines** page.
   - Click to manually run the `personal_finance_pipeline`.
   - Watch the `worker` process the run in real-time. This pipeline seeds SQLite/DuckDB databases, runs data quality checks, and generates metadata.
2. **Explore the Catalog:**
   - Navigate to **Datasets**. 
   - Click into the newly materialized datasets (e.g., `mart_monthly_cashflow` or `mart_category_spend`) to see the Python-profiled statistics and lineage graphs.
3. **Dashboards & Reporting:**
   - Navigate to **Reports/Dashboards**.
   - You should see a default "Finance Overview" dashboard. 
   - Create a new dashboard, add widgets (Table, Bar Chart, KPI), and bind them to the marts mentioned above. Save it.
4. **Admin Terminal & System Overview:**
   - Navigate to the **System** page.
   - Look at the "Source of Truth" card to verify Postgres is active.
   - Look at the Audit Logs to see the footprint of your manual pipeline trigger and dashboard saves.
   - Use the built-in terminal to run `backups` or `benchmark`.

---

## Part 3: What to Expect (Capabilities & Caveats)

### ✅ What IS Expected to Work
- **Orchestration & Execution**: The worker robustly executes the medallion architecture (raw -> staging -> mart -> metrics) using Python and DuckDB SQL. Pipeline history and status reporting are fully operational.
- **Reporting Engine**: You can visually build charts, apply dashboard-level and widget-level filters (e.g. `from_month`, `category`), save layout presets, and persist these to the database.
- **Metadata Generation**: Datasets are dynamically profiled. Schema, row counts, descriptions, nullable states, and upstream/downstream lineage graphs are fully generated and explorable.
- **Backup Generation**: You can run `platformctl backup create` (or via the UI Admin terminal) to package your Postgres database, SQLite/DuckDB instances, and JSON metadata into a portable `.tar.gz`.

### ⚠️ Caveats (What is NOT finished / expected to break)
- **Restore Automation**: While taking a backup works flawlessly, _restoring_ from that backup is currently a highly manual extraction drill. Do not expect to find a one-click `restore` button that magically revives the application state yet.
- **Dynamic Identity Management**: Currently, users cannot "sign up", "change passwords", or be added by admins in the UI. All access control relies strictly on strings defined in your `.env` files.
- **Advanced Dashboard Layouts**: While you can add widgets, dragging, dropping, resizing on a strict grid, and assigning complex layout schemas are not fully built out. 
- **True Metadata Normalization**: Deep database persistence for UI-driven changes to metadata (tagging columns, overriding documentation in the UI instead of YAML) is still heavily tied to "sync-on-read" memory/file representations.
- **Complex Cron Scheduling**: The scheduler natively supports basic crons, but does not support complex ranges, intricate catch-up logic, or complex event-driven triggers beyond its current scaffolded implementation.

---

## Part 4: Proxmox Homelab Deployment

Running this on a dedicated server (like a Proxmox VM) requires modifying the networking bindings, as the repo is extremely defensive and binds to `127.0.0.1` by default to prevent accidental public exposure.

### 1. Update Network Bindings
To access the Web UI from another computer on your local LAN, you must edit `infra/compose/docker-compose.yml`.

Find the `api` and `web` port definitions:
```yaml
# In infra/compose/docker-compose.yml

  api:
    # ...
    ports:
      - "127.0.0.1:8080:8080" # Change this to "8080:8080"
  
  web:
    # ...
    ports:
      - "127.0.0.1:3000:3000" # Change this to "3000:3000"
```
Removing the `127.0.0.1:` forces Docker to bind to `0.0.0.0`, exposing the ports to your entire local network.

### 2. Update Environment Variables
Because the Web UI proxies requests to the backend, it needs to know the correct URL. On a local Macbook, `http://127.0.0.1:8080` works fine. On a network, you should modify `.env.compose`:

```bash
# Append to .env.compose
PLATFORM_API_BASE_URL=http://<YOUR_PROXMOX_VM_IP>:8080
```
Update `<YOUR_PROXMOX_VM_IP>` to the static IP of your Proxmox VM (e.g., `192.168.1.100`). Keep your token definitions secure.

### 3. Execution
Run `make bootstrap` from the VM. You can now open a browser from your Macbook, navigate to `http://<YOUR_PROXMOX_VM_IP>:3000`, and access the platform.

---

## Part 5: SaaS / Cloud Hosting Strategy

How would this look if moved to AWS, GCP, or a hosted service?

### The "Lift and Shift" (Easiest)
Because the application is packaged as a `docker-compose.yml`, the easiest Cloud deployment is spinning up an EC2 instance (or generic cloud VM), installing Docker, and running the exact same Proxmox homelab setup mentioned above. You would place a load balancer or an Nginx reverse proxy (with Let's Encrypt SSL/TLS) in front of port 3000.

### The "Cloud Native" Evolution (Recommended for true SaaS)
To scale this properly as a SaaS platform, you would break the monolith constraints slightly to utilize managed services:

1. **Managed PostgreSQL:** Drop the `postgres` docker container. Point the `PLATFORM_POSTGRES_DSN` directly at AWS RDS or GCP Cloud SQL. This delegates backups, high-availability, and scaling of the control plane to the cloud provider.
2. **Container Orchestration:** Instead of `docker-compose`, deploy the `platform-api`, `platform-scheduler`, `platform-worker`, and `platform-web` containers onto ECS (AWS), GKE (Kubernetes), or a managed serverless container service like AWS Fargate.
3. **Artifact Storage (S3):** Currently, the platform uses local disk (files and DuckDB) via `PLATFORM_ARTIFACT_ROOT` and `PLATFORM_DATA_ROOT`. To work on distributed containers, the platform's `storage.Service` interface would need to be extended to support `s3://` or `gcs://` URIs instead of purely relying on local block storage, ensuring any worker container can read pipeline outputs.
4. **Identity Provider (IdP):** Instead of static RBAC tokens in `.env.compose`, you would integrate OIDC/OAuth2/SAML with Auth0, Okta, or AWS Cognito, validating JWTs at the `platform-api` layer.
