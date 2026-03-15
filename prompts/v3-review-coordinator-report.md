# Data Platform — Pre-Production Review & v3 Coordinator Plan

**Date:** March 15, 2026
**Context:** Review performed after the completion of v2 (Model 1 security hardening, Model 2 resilience/domains, Model 3 release ops).

## 1. Executive Summary

Following the comprehensive v2 hardening pass, the platform has crossed a critical threshold. It has successfully shed its "prototype/MVP" status and is now a genuinely strong **late-beta / pre-production self-hosted platform**. 

The v2 completions addressed the most vital blockers: bcrypt password hashing, login rate-limiting, session capping, multi-domain proofs (personal finance + inventory), tested resilience against postgres/worker disruptions, deployment guides, and release checklists. This changes the posture from "dangerous for real data" to "functionally robust for a disciplined small team." 

However, against the complete original vision of an enterprise-level orchestration, reporting, and semantics platform, it falls short of being a 100% comprehensive data engineering suite. The remaining distance to `100%` is dominated by breadth extensions (database connectors), production operations safeguards (retries, alerting), and advanced analytics/reporting depth.

## 2. Honest Readiness Assessment

*   **3-Person On-Prem Internal Rollout:** **Ready (92%)**. With the v2 security and operations work, the 3-person team can safely deploy this to an internal LAN, authenticate securely, run diverse domain pipelines, and recover from transient failures. It is safe for real internal operations.
*   **Small-Team Production Usage:** **Almost Ready (88%)**. To be truly stable for hands-off production, the platform still needs native database connectors (too reliant on static CSV/JSON), stronger orchestration retries, and proactive webhook alerting for failures.
*   **Enterprise-Grade Rollout:** **Not Ready (70%)**. The platform's auth natively handles individuals but lacks SAML/OIDC and team-based RBAC. It has no automated data retention/purging policies, no SIEM-compatible audit log exports, and lacks deep column-level lineage tracking.

## 3. Requirement-by-Requirement Review

*Reflecting on the baseline percentages and adjusting for v2 completions + identifying remaining gaps to `100%`.*

| Category | Prev % | Current % (Est.) | Review & Path to 100% |
| :--- | :--- | :--- | :--- |
| **1. Pipeline Orchestration** | 82% | **85%** | **Fair.** v2 generalized execution, but retries and idempotency controls remain weak. *Path to 100%: Declarative exponential backoff retries, explicit idempotency tokens, and advanced timezone-aware cron scheduling.* |
| **2. Data Ingestion/Transform** | 78% | **84%** | **Low.** v2 built the inventory slice and pluggable ingestion, raising this score. *Path to 100%: Implement native Postgres/MySQL database connectors instead of relying purely on file-based ingestion.* |
| **3. Integrated Metadata** | 88% | **88%** | **Fair.** Still lacks deep data-flow visibility. *Path to 100%: Column-level lineage parser, ownership group structures, and data retention/purge policies.* |
| **4. Monitoring, Ops & DBs** | 84% | **87%** | **Fair.** v2 added ops/resilience proofs. *Path to 100%: Proactive webhook alerting, long-term metric retention, and Prometheus `/metrics` endpoints.* |
| **5. Analytics Serving Layer** | 84% | **84%** | **Fair.** Very constrained to prevent chaos, but too constrained for arbitrary dimensional pivots. *Path to 100%: A richer semantic querying layer supporting dynamic dimensional combinations and aggregate rollups.* |
| **6. Reporting / Viz App** | 89% | **89%** | **Fair.** UX is great but closed. *Path to 100%: Shareable report links, CSV/PDF widget export, and polished filtering ergonomics.* |
| **7. Platform as Code** | 91% | **91%** | **Fair.** High marks for manifests, though some state remains mutable. *Path to 100%: Strict immutability enforcement tools and environment promotion logic.* |
| **8. CI/CD** | 76% | **82%** | **Fair.** v2 hardened the GitHub actions and release checklists. *Path to 100%: Automated semantic release tagging, multi-architecture tarball building via Actions, and rollback automation.* |

## 4. Top 10 Highest-Risk Gaps (By Severity)

1.  **Limited Ingestion Connectors:** Only supporting flat files limits usefulness for live transactional data. (Category 2)
2.  **Weak Orchestration Retries:** Lacking exponential backoff and idempotency leaves the system vulnerable to transient API/DB blips. (Category 1)
3.  **No Proactive Alerting:** Operators must poll the UI to discover pipeline or freshness failures. (Category 4)
4.  **No Automated Data Retention:** Run artifacts and staging data will grow unbounded without purge policies. (Category 3)
5.  **Closed Reporting Ecosystem:** Lack of export or sharing restricts data dissemination. (Category 6)
6.  **Missing Column-Level Lineage:** Trust is diminished when complex SQL obscures dependencies. (Category 3)
7.  **Constrained Semantic Layer:** Hardcoded querying restricts ad-hoc dimensional exploration. (Category 5)
8.  **Manual Release Automation:** Artifact building and tagging relies on human checklists. (Category 8)
9.  **No Outer-Loop Monitoring:** Lack of standard `/metrics` integration with tools like Prometheus. (Category 4)
10. **Basic Scheduling:** Missing timezone-aware advanced cron. (Category 1)

## 5. Top 10 Highest-Leverage Next Actions

1.  **Build Native DB Connectors:** Add Postgres and MySQL ingestion job types.
2.  **Implement Exponential Backoff:** Add configurable retry loops with jitter to `runner.go`.
3.  **Add Webhook Alerting:** Fire payloads to Slack/Teams on run failures and freshness drops.
4.  **Implement Data Purge Policies:** Create a GC job to enforce artifact and database log retention.
5.  **Build CSV Export:** Add a backend endpoint to export any curated widget dataset as CSV.
6.  **Extend Semantic Layer:** Refactor `analytics/service.go` to support dynamic multi-dimension GROUP BY rollups.
7.  **Automate GitHub Releases:** Expand CI to compile binaries and attach release tarballs automatically.
8.  **Add `/metrics` Endpoint:** Export Go runtime, queue depth, and runner metrics for Prometheus.
9.  **Enhance Pipeline Cron:** Migrate to a robust cron parser supporting timezones and complex expressions.
10. **Implement Shareable Links:** Add frontend UUID routing to share specific dashboard parameter states.

## 6. Coordinated 3-Model Work Plan (Path to 100%)

### Model 1: Backend Orchestration, Ingestion & Governance
**Focus:** Driving Pipeline Orchestration (1), Data Ingestion (2), and Metadata (3) to 100%.
*   Add native Postgres/MySQL database ingestion job types.
*   Implement explicit idempotency tokens and exponential backoff retries in the execution runner.
*   Add advanced timezone-aware cron scheduling limits.
*   Implement a data retention garbage collector to purge old artifacts and run logs based on asset metadata.

### Model 2: Analytics, Reporting & Product UX
**Focus:** Driving Analytics Serving (5) and Reporting App (6) to 100%.
*   Refactor the analytics semantic layer to allow dynamic dimensional combinations and rollups.
*   Implement backend handlers and frontend UI for CSV data export from dashboard widgets.
*   Develop robust URL parameter sharing for dashboards to persist filter states across shareable links.
*   Enhance report builder ergonomics with deeper in-app preview mechanisms.

### Model 3: Operations, Alerting & CI/CD
**Focus:** Driving Monitoring (4), Platform-as-Code (7), and CI/CD (8) to 100%.
*   Build a webhook alerting notification module triggered by pipeline failures and asset staleness.
*   Expose a standard Prometheus `/metrics` endpoint wrapping telemetry and queue depth.
*   Automate release tagging and multi-architecture binary compilation via GitHub Actions.
*   Solidify deployment automation scripts for flawless environment promotion and rollback.

## 7. Final Recommendation

**Begin real deployment rehearsals now.**

Do not wait for 100% completion of the enterprise features to start using the platform. The v2 completions have provided the necessary security (bcrypt, rate limiting), architectural resilience, and deployment guidelines to make it safe for a small, trusted team to operate on a private network.

The team should deploy the v2 platform to the target on-prem hardware and begin ingesting their actual data logic. The pain points discovered in this "real metadata, real workloads" rehearsal will perfectly validate (and reprioritize if necessary) the remaining v3 work coordinated above. Models 1, 2, and 3 can execute the v3 burn-down in parallel with the real-world deployment.
