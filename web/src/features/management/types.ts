export type OperatorRole = "viewer" | "editor" | "admin";

export type CommandArea =
  | "operations"
  | "catalog"
  | "reports"
  | "governance"
  | "recovery"
  | "diagnostics";

export type TerminalCommandTemplate = {
  id: string;
  label: string;
  command: string;
  area: CommandArea;
  minimumRole: OperatorRole;
  summary: string;
  destructive?: boolean;
};

export type ServiceStatusCard = {
  id: string;
  label: string;
  state: "healthy" | "degraded" | "offline";
  detail: string;
};

export type QueueSnapshotCard = {
  runID: string;
  pipelineID: string;
  status: "queued" | "running" | "retrying" | "failed";
  trigger: string;
  requestedAt: string;
};

export type AssetAttentionCard = {
  assetID: string;
  layer: string;
  freshnessState: "fresh" | "late" | "stale" | "missing" | "unknown";
  hasDocs: boolean;
  hasQuality: boolean;
};

export type RunbookShortcut = {
  id: string;
  label: string;
  path: string;
  reason: string;
};
