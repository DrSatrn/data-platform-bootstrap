// This hook gathers platform health and quality summaries for the system page.
import { useEffect, useState } from "react";

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

export function useSystemData() {
  const [health, setHealth] = useState<HealthPayload | null>(null);
  const [quality, setQuality] = useState<QualityPayload | null>(null);
  const [overview, setOverview] = useState<OverviewPayload | null>(null);
  const [logs, setLogs] = useState<LogsPayload | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      fetchJSON<HealthPayload>("/healthz"),
      fetchJSON<QualityPayload>("/api/v1/quality"),
      fetchJSON<OverviewPayload>("/api/v1/system/overview"),
      fetchJSON<LogsPayload>("/api/v1/system/logs")
    ])
      .then(([nextHealth, nextQuality, nextOverview, nextLogs]) => {
        setHealth(nextHealth);
        setQuality(nextQuality);
        setOverview(nextOverview);
        setLogs(nextLogs);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown system error"));
  }, []);

  return { health, quality, overview, logs, error };
}
