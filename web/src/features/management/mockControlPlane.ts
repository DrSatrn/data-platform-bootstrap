import type {
  AssetAttentionCard,
  QueueSnapshotCard,
  RunbookShortcut,
  ServiceStatusCard,
  TerminalCommandTemplate
} from "./types";

export const defaultCommandCatalog: TerminalCommandTemplate[] = [
  {
    id: "status",
    label: "Status",
    command: "status",
    area: "diagnostics",
    minimumRole: "admin",
    summary: "Show environment, uptime, request volume, and command counts."
  },
  {
    id: "pipelines",
    label: "Pipelines",
    command: "pipelines",
    area: "operations",
    minimumRole: "admin",
    summary: "List known pipelines and their ownership."
  },
  {
    id: "runs",
    label: "Runs",
    command: "runs",
    area: "operations",
    minimumRole: "admin",
    summary: "Inspect recent run snapshots from the control plane."
  },
  {
    id: "backup-create",
    label: "Create Backup",
    command: "backup create",
    area: "recovery",
    minimumRole: "admin",
    summary: "Create a first-party recovery bundle for the current environment."
  },
  {
    id: "backup-verify",
    label: "Verify Backup",
    command: "backup verify latest.tar.gz",
    area: "recovery",
    minimumRole: "admin",
    summary: "Verify a bundle before trusting it as a recovery point."
  },
  {
    id: "quality",
    label: "Quality",
    command: "quality",
    area: "governance",
    minimumRole: "admin",
    summary: "List operator-visible quality checks and their current state."
  },
  {
    id: "assets",
    label: "Assets",
    command: "assets",
    area: "catalog",
    minimumRole: "admin",
    summary: "List manifest-backed data assets and ownership metadata."
  },
  {
    id: "dashboards",
    label: "Dashboards",
    command: "dashboards",
    area: "reports",
    minimumRole: "admin",
    summary: "Inspect saved dashboard definitions in the reporting layer."
  }
];

export const mockServiceStatus: ServiceStatusCard[] = [
  { id: "api", label: "API", state: "healthy", detail: "Serving control-plane traffic." },
  { id: "worker", label: "Worker", state: "healthy", detail: "Polling queue and executing runs." },
  { id: "scheduler", label: "Scheduler", state: "degraded", detail: "Active but waiting on next refresh slot." },
  { id: "postgres", label: "Postgres", state: "healthy", detail: "Preferred control-plane backend is reachable." }
];

export const mockQueue: QueueSnapshotCard[] = [
  {
    runID: "run_20260315T120001.000000000",
    pipelineID: "personal_finance_pipeline",
    status: "running",
    trigger: "manual_api",
    requestedAt: "2026-03-15T12:00:01Z"
  },
  {
    runID: "run_20260315T121501.000000000",
    pipelineID: "personal_finance_pipeline",
    status: "queued",
    trigger: "scheduled",
    requestedAt: "2026-03-15T12:15:01Z"
  }
];

export const mockAttentionAssets: AssetAttentionCard[] = [
  {
    assetID: "mart_budget_vs_actual",
    layer: "mart",
    freshnessState: "late",
    hasDocs: true,
    hasQuality: false
  },
  {
    assetID: "raw_account_balances",
    layer: "raw",
    freshnessState: "missing",
    hasDocs: true,
    hasQuality: false
  },
  {
    assetID: "staging_transactions_enriched",
    layer: "staging",
    freshnessState: "fresh",
    hasDocs: false,
    hasQuality: false
  }
];

export const mockRunbooks: RunbookShortcut[] = [
  {
    id: "quickstart",
    label: "Quickstart",
    path: "/docs/runbooks/bootstrap.md",
    reason: "Best first stop for an operator bringing the system up."
  },
  {
    id: "config",
    label: "Config Reality",
    path: "/docs/runbooks/config-reality.md",
    reason: "Explains the actual runtime config behavior and env loading order."
  },
  {
    id: "recovery",
    label: "Recovery Drill",
    path: "/docs/runbooks/recovery-drill.md",
    reason: "Future home for concrete restore validation steps."
  }
];
