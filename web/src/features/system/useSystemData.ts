// This hook gathers platform health and quality summaries for the system page.
import { useEffect, useState } from "react";

import { useAuth } from "../auth/useAuth";
import { fetchJSON } from "../../lib/api";

type HealthPayload = {
  status: string;
  environment: string;
};

type QualityPayload = {
  checks: Array<{ id: string; name: string; status: string; message: string }>;
};

type OverviewPayload = {
  environment: string;
  http_addr: string;
  web_addr: string;
  known_pipelines: number;
  known_assets: number;
  run_history: number;
  run_summary: {
    total_runs: number;
    queued_runs: number;
    running_runs: number;
    succeeded_runs: number;
    failed_runs: number;
    completed_last_24_hours: number;
    failed_last_24_hours: number;
    average_duration_seconds: number;
    latest_failure_run_id?: string;
    latest_failure_message?: string;
  };
  queue_summary: {
    queued: number;
    active: number;
    completed: number;
    total: number;
  };
  scheduler_summary: {
    refreshed_at?: string;
    lag_seconds: number;
    pipeline_count: number;
    asset_count: number;
    last_enqueue_at?: string;
    last_error?: string;
  };
  backup_summary: {
    bundle_count: number;
    latest_bundle_path?: string;
    latest_bundle_bytes?: number;
  };
  persistence_modes: Record<
    string,
    {
      source_of_truth: string;
      read_path: string;
      write_path: string;
      mirrors?: string[];
      fallback?: string;
    }
  >;
  telemetry: {
    uptime_seconds: number;
    total_requests: number;
    total_errors: number;
    total_commands: number;
    request_counts: Record<string, number>;
    recent_log_summary: Array<{ time: string; level: string; message: string }>;
  };
};

type LogsPayload = {
  logs: Array<{ time: string; level: string; message: string }>;
};

type AuditPayload = {
  events: Array<{
    time: string;
    actor_user_id?: string;
    actor_subject: string;
    actor_role: string;
    action: string;
    resource: string;
    outcome: string;
    details?: Record<string, string | number | boolean>;
  }>;
};

type CatalogPayload = {
  summary: {
    by_freshness: Record<string, number>;
    assets_missing_docs: number;
    assets_missing_quality: number;
    assets_containing_pii: number;
    lineage_edges: number;
  };
  assets: Array<{
    id: string;
    freshness_status: {
      state: string;
    };
  }>;
};

export function useSystemData() {
  const { loading, session } = useAuth();
  const [health, setHealth] = useState<HealthPayload | null>(null);
  const [quality, setQuality] = useState<QualityPayload | null>(null);
  const [overview, setOverview] = useState<OverviewPayload | null>(null);
  const [logs, setLogs] = useState<LogsPayload | null>(null);
  const [audit, setAudit] = useState<AuditPayload | null>(null);
  const [catalog, setCatalog] = useState<CatalogPayload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  async function load() {
    if (loading) {
      return;
    }
    if (!session?.capabilities.view_platform) {
      setError("Viewer role required to access the system view.");
      setQuality(null);
      setOverview(null);
      setLogs(null);
      setAudit(null);
      setCatalog(null);
      return;
    }

    setRefreshing(true);
    try {
      const [nextHealth, nextQuality, nextOverview, nextLogs, nextAudit, nextCatalog] = await Promise.all([
        fetchJSON<HealthPayload>("/healthz"),
        fetchJSON<QualityPayload>("/api/v1/quality"),
        fetchJSON<OverviewPayload>("/api/v1/system/overview"),
        fetchJSON<LogsPayload>("/api/v1/system/logs"),
        fetchJSON<AuditPayload>("/api/v1/system/audit"),
        fetchJSON<CatalogPayload>("/api/v1/catalog")
      ]);
      setHealth(nextHealth);
      setQuality(nextQuality);
      setOverview(nextOverview);
      setLogs(nextLogs);
      setAudit(nextAudit);
      setCatalog(nextCatalog);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown system error");
    } finally {
      setRefreshing(false);
    }
  }

  useEffect(() => {
    void load();
  }, [loading, session]);

  useEffect(() => {
    if (!session?.capabilities.view_platform) {
      return;
    }
    const interval = window.setInterval(() => {
      void load();
    }, 10000);
    return () => window.clearInterval(interval);
  }, [session]);

  return { health, quality, overview, logs, audit, catalog, error, refreshing, refresh: load };
}
