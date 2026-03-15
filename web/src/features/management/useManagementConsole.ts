// This hook wires the staged management-console modules to the live platform
// APIs so the browser can expose a real operator workflow instead of mock data.
import { useCallback, useEffect, useMemo, useState } from "react";

import { useAuth } from "../auth/useAuth";
import { fetchJSON, postJSON } from "../../lib/api";
import { defaultCommandCatalog, mockRunbooks } from "./mockControlPlane";
import type { AssetAttentionCard, QueueSnapshotCard, RunbookShortcut, ServiceStatusCard, TerminalCommandTemplate } from "./types";
import { createTerminalSession, completeCommand, startCommand, type TerminalSession } from "./terminal/sessionModel";
import { planFollowupActions, prioritizeActions, type FollowupAction } from "./terminal/followupPlanner";
import { buildRunbookDock } from "./runbooks/runbookDock";
import { buildEvidenceBoard } from "./evidence/evidenceBoard";
import type { OpsviewAttentionSummary, OpsviewExternalToolRunSummary } from "./opsview/opsviewBridge";

type HealthPayload = {
  status: string;
  environment: string;
};

type OverviewPayload = {
  run_summary: {
    total_runs: number;
    queued_runs: number;
    running_runs: number;
    succeeded_runs: number;
    failed_runs: number;
  };
  queue_summary: {
    queued: number;
    active: number;
    total: number;
  };
  backup_summary: {
    bundle_count: number;
  };
  scheduler_summary: {
    lag_seconds: number;
    refreshed_at?: string;
    last_error?: string;
  };
  persistence_modes: Record<
    string,
    {
      source_of_truth: string;
      read_path: string;
      write_path: string;
    }
  >;
};

type CatalogPayload = {
  summary: {
    assets_missing_docs: number;
    assets_missing_quality: number;
  };
  assets: Array<{
    id: string;
    layer: string;
    freshness_status: { state: "fresh" | "late" | "stale" | "missing" | "unknown" };
    coverage: {
      has_documentation: boolean;
      has_quality_checks: boolean;
    };
  }>;
};

type PipelinePayload = {
  runs: Array<{
    id: string;
    pipeline_id: string;
    status: "queued" | "running" | "failed" | "succeeded" | "canceled" | "pending";
    trigger: string;
    updated_at: string;
  }>;
};

type ReportsPayload = {
  dashboards: Array<{ id: string; name: string }>;
};

type OpsviewRunSnapshot = {
  run_id: string;
  pipeline_id: string;
  status: string;
  trigger: string;
  updated_at: string;
  external_tool_runs: OpsviewExternalToolRunSummary[];
};

type OpsviewPayload = {
  snapshots: OpsviewRunSnapshot[];
  external_tool_attention: OpsviewAttentionSummary;
  attention_rollup: {
    total_runs: number;
    failed_runs: number;
    running_runs: number;
    succeeded_runs: number;
    runs_with_external_tool_failures: number;
    runs_missing_evidence: number;
    external_tool_job_count: number;
  };
};

type TerminalResponse = {
  command: string;
  success: boolean;
  output: string[];
};

type FollowupDeck = {
  sessionID: string;
  sessionTitle: string;
  actions: FollowupAction[];
};

const runbookCatalog: RunbookShortcut[] = [
  {
    id: "operator-manual",
    label: "Operator Manual",
    path: "/Users/streanor/Documents/Playground/data-platform/docs/runbooks/operator-manual.md",
    reason: "Best central reference while operating the management surface."
  },
  {
    id: "backups",
    label: "Backups And Restore",
    path: "/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md",
    reason: "Use this when recovery or bundle verification becomes part of the workflow."
  },
  {
    id: "external-tools",
    label: "External Tool Jobs",
    path: "/Users/streanor/Documents/Playground/data-platform/docs/reference/external-tool-jobs.md",
    reason: "Reference for dbt-style external tool execution, artifacts, and boundaries."
  }
];

export function useManagementConsole() {
  const { loading, token, session } = useAuth();
  const [health, setHealth] = useState<HealthPayload | null>(null);
  const [overview, setOverview] = useState<OverviewPayload | null>(null);
  const [catalog, setCatalog] = useState<CatalogPayload | null>(null);
  const [pipelines, setPipelines] = useState<PipelinePayload | null>(null);
  const [reports, setReports] = useState<ReportsPayload | null>(null);
  const [opsview, setOpsview] = useState<OpsviewPayload | null>(null);
  const [sessions, setSessions] = useState<TerminalSession[]>([]);
  const [commandDraft, setCommandDraft] = useState("status");
  const [selectedCommand, setSelectedCommand] = useState<TerminalCommandTemplate | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [pendingCommand, setPendingCommand] = useState(false);

  const load = useCallback(async () => {
    if (loading) {
      return;
    }
    if (!session?.capabilities.view_platform) {
      setError("Viewer role required to access the management console.");
      setHealth(null);
      setOverview(null);
      setCatalog(null);
      setPipelines(null);
      setReports(null);
      setOpsview(null);
      return;
    }
    setRefreshing(true);
    try {
      const [nextHealth, nextOverview, nextCatalog, nextPipelines, nextReports, nextOpsview] = await Promise.all([
        fetchJSON<HealthPayload>("/healthz"),
        fetchJSON<OverviewPayload>("/api/v1/system/overview"),
        fetchJSON<CatalogPayload>("/api/v1/catalog"),
        fetchJSON<PipelinePayload>("/api/v1/pipelines"),
        fetchJSON<ReportsPayload>("/api/v1/reports"),
        fetchJSON<OpsviewPayload>("/api/v1/opsview?limit=12")
      ]);
      setHealth(nextHealth);
      setOverview(nextOverview);
      setCatalog(nextCatalog);
      setPipelines(nextPipelines);
      setReports(nextReports);
      setOpsview(nextOpsview);
      setError(null);
    } catch (nextError) {
      setError(nextError instanceof Error ? nextError.message : "Unknown management-console error");
    } finally {
      setRefreshing(false);
    }
  }, [loading, session]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    if (!session?.capabilities.view_platform) {
      return;
    }
    const interval = window.setInterval(() => {
      void load();
    }, 10000);
    return () => window.clearInterval(interval);
  }, [load, session]);

  const serviceStatus = useMemo<ServiceStatusCard[]>(() => {
    if (!health || !overview) {
      return [];
    }
    return [
      {
        id: "api",
        label: "API",
        state: health.status === "ok" ? "healthy" : "offline",
        detail: `Environment ${health.environment} is ${health.status}.`
      },
      {
        id: "scheduler",
        label: "Scheduler",
        state: overview.scheduler_summary.last_error
          ? "offline"
          : overview.scheduler_summary.lag_seconds > 60
            ? "degraded"
            : "healthy",
        detail: overview.scheduler_summary.last_error
          ? overview.scheduler_summary.last_error
          : `Lag ${Math.round(overview.scheduler_summary.lag_seconds)}s since last heartbeat.`
      },
      {
        id: "queue",
        label: "Queue",
        state: overview.queue_summary.active > 0 || overview.queue_summary.queued > 0 ? "degraded" : "healthy",
        detail: `${overview.queue_summary.queued} queued · ${overview.queue_summary.active} active · ${overview.queue_summary.total} total.`
      },
      {
        id: "control-plane",
        label: "Control Plane",
        state: overview.persistence_modes.runs?.source_of_truth === "postgres" ? "healthy" : "degraded",
        detail: `Runs: ${overview.persistence_modes.runs?.source_of_truth ?? "unknown"} · dashboards: ${overview.persistence_modes.dashboards?.source_of_truth ?? "unknown"}.`
      }
    ];
  }, [health, overview]);

  const queue = useMemo<QueueSnapshotCard[]>(() => {
    return (pipelines?.runs ?? [])
      .filter((run) => run.status === "queued" || run.status === "running" || run.status === "failed")
      .slice(0, 8)
      .map((run) => ({
        runID: run.id,
        pipelineID: run.pipeline_id,
        status: run.status === "failed" ? "failed" : run.status === "running" ? "running" : "queued",
        trigger: run.trigger,
        requestedAt: run.updated_at
      }));
  }, [pipelines]);

  const attentionAssets = useMemo<AssetAttentionCard[]>(() => {
    return (catalog?.assets ?? [])
      .filter((asset) => asset.freshness_status.state !== "fresh" || !asset.coverage.has_documentation || !asset.coverage.has_quality_checks)
      .slice(0, 8)
      .map((asset) => ({
        assetID: asset.id,
        layer: asset.layer,
        freshnessState: asset.freshness_status.state,
        hasDocs: asset.coverage.has_documentation,
        hasQuality: asset.coverage.has_quality_checks
      }));
  }, [catalog]);

  const externalToolSummaries = useMemo(() => (opsview?.snapshots ?? []).flatMap((snapshot) => snapshot.external_tool_runs), [opsview]);

  const followupDecks = useMemo<FollowupDeck[]>(() => {
    return sessions.map((session) => ({
      sessionID: session.id,
      sessionTitle: session.title,
      actions: prioritizeActions(planFollowupActions(session))
    }));
  }, [sessions]);

  const runbookItems = useMemo(() => {
    return buildRunbookDock(
      sessions.map((operatorSession) => ({
        id: operatorSession.id,
        title: operatorSession.title,
        status: operatorSession.status,
        recommendedRunbook: operatorSession.recommendedRunbook
      })),
      followupDecks.flatMap((deck) => deck.actions)
    );
  }, [followupDecks, sessions]);

  const evidenceItems = useMemo(() => {
    return buildEvidenceBoard(
      sessions.map((operatorSession) => ({
        id: operatorSession.id,
        title: operatorSession.title,
        status: operatorSession.status,
        pinnedArtifacts: operatorSession.pinnedArtifacts
      })),
      externalToolSummaries.map((summary) => ({
        jobID: summary.job_id,
        status: summary.status,
        logArtifacts: summary.log_artifacts,
        outputArtifacts: summary.output_artifacts
      }))
    );
  }, [externalToolSummaries, sessions]);

  async function executeCommand(command: string, template?: TerminalCommandTemplate | null) {
    const normalizedCommand = command.trim();
    if (!normalizedCommand) {
      return;
    }
    if (!session?.capabilities.run_admin_terminal) {
      setError("Admin role required to run guided terminal commands.");
      return;
    }

    const chosenTemplate = template ?? defaultCommandCatalog.find((entry) => entry.command === normalizedCommand) ?? null;
    const sessionID = `session_${Date.now()}`;
    const startedAt = new Date().toISOString();
    const baseSession = createTerminalSession({
      id: sessionID,
      title: chosenTemplate?.label ?? normalizedCommand,
      scope: scopeForCommand(chosenTemplate, normalizedCommand),
      currentCommand: normalizedCommand,
      recommendedRunbook: runbookForCommand(chosenTemplate, normalizedCommand)
    });
    const runningSession = startCommand(baseSession, normalizedCommand, startedAt);
    setSessions((current) => [runningSession, ...current].slice(0, 8));
    setPendingCommand(true);
    setError(null);

    try {
      const result = await postJSON<TerminalResponse, { command: string }>(
        "/api/v1/admin/terminal/execute",
        { command: normalizedCommand },
        token.trim() || undefined
      );
      const finishedSession = completeCommand(runningSession, {
        exitCode: result.success ? 0 : 1,
        at: new Date().toISOString(),
        stdout: result.success ? result.output : [],
        stderr: result.success ? [] : result.output,
        artifacts: extractArtifactPaths(result.output)
      });
      setSessions((current) => current.map((entry) => (entry.id === sessionID ? finishedSession : entry)));
      setCommandDraft("");
      await load();
    } catch (nextError) {
      const message = nextError instanceof Error ? nextError.message : "Unknown terminal error";
      const failedSession = completeCommand(runningSession, {
        exitCode: 1,
        at: new Date().toISOString(),
        stderr: [message]
      });
      setSessions((current) => current.map((entry) => (entry.id === sessionID ? failedSession : entry)));
      setError(message);
    } finally {
      setPendingCommand(false);
    }
  }

  return {
    loading,
    refreshing,
    pendingCommand,
    error,
    session,
    commandDraft,
    setCommandDraft,
    selectedCommand,
    setSelectedCommand,
    commands: defaultCommandCatalog,
    runbooks: runbookCatalog.length > 0 ? runbookCatalog : mockRunbooks,
    reports,
    opsview,
    externalToolSummaries,
    health,
    overview,
    serviceStatus,
    queue,
    attentionAssets,
    sessions,
    followupDecks,
    runbookItems,
    evidenceItems,
    recentCommands: sessions.map((operatorSession) => operatorSession.currentCommand).filter(Boolean),
    executeCommand,
    refresh: load
  };
}

function scopeForCommand(template: TerminalCommandTemplate | null, command: string) {
  if (template?.area === "recovery" || command.startsWith("backup")) {
    return "recovery";
  }
  if (template?.area === "reports") {
    return "reports";
  }
  if (template?.area === "catalog" || template?.area === "governance") {
    return "catalog";
  }
  if (template?.area === "operations" || command.startsWith("trigger") || command.startsWith("artifacts")) {
    return "pipelines";
  }
  return "diagnostics";
}

function runbookForCommand(template: TerminalCommandTemplate | null, command: string) {
  if (template?.area === "recovery" || command.startsWith("backup")) {
    return "/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md";
  }
  if (command.startsWith("artifacts") || command.startsWith("trigger") || template?.area === "operations") {
    return "/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md";
  }
  if (template?.area === "catalog" || template?.area === "governance") {
    return "/Users/streanor/Documents/Playground/data-platform/docs/runbooks/access-matrix.md";
  }
  return "/Users/streanor/Documents/Playground/data-platform/docs/runbooks/operator-manual.md";
}

function extractArtifactPaths(lines: string[]) {
  const artifacts = new Set<string>();
  const filePattern = /([A-Za-z0-9_./-]+\.(?:json|csv|log|md|tar\.gz|gz))/g;
  for (const line of lines) {
    const matches = line.match(filePattern) ?? [];
    for (const match of matches) {
      if (match.includes("docs/") || match.includes("/docs/")) {
        continue;
      }
      artifacts.add(match);
    }
  }
  return [...artifacts];
}
