// This hook loads orchestration state for the pipelines page and exposes a
// manual trigger action so the UI can kick off a real worker-backed run.
import { useCallback, useEffect, useState } from "react";

import { useAuth } from "../auth/useAuth";
import { fetchJSON, postJSON } from "../../lib/api";

type Pipeline = {
  id: string;
  name: string;
  description: string;
  owner: string;
  jobs: Array<{ id: string; type: string }>;
};

type PipelinePayload = {
  pipelines: Pipeline[];
  runs: Array<{
    id: string;
    status: string;
    pipeline_id: string;
    trigger: string;
    updated_at: string;
    job_runs: Array<{ job_id: string; status: string }>;
    events: Array<{ time: string; message: string; level: string }>;
  }>;
};

export function usePipelines() {
  const { loading, token, session } = useAuth();
  const [data, setData] = useState<PipelinePayload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [pendingPipelineID, setPendingPipelineID] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(() => {
    if (loading) {
      return;
    }
    if (!session?.capabilities.view_platform) {
      setData(null);
      setError("Viewer role required to access pipelines.");
      return;
    }
    setRefreshing(true);
    fetchJSON<PipelinePayload>("/api/v1/pipelines")
      .then((payload) => {
        setData(payload);
        setError(null);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown pipelines error"))
      .finally(() => setRefreshing(false));
  }, [loading, session]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!session?.capabilities.view_platform) {
      return;
    }
    const interval = window.setInterval(load, 10000);
    return () => window.clearInterval(interval);
  }, [load, session]);

  async function triggerPipeline(pipelineID: string) {
    setPendingPipelineID(pipelineID);
    setError(null);
    try {
      if (!session?.capabilities.trigger_runs) {
        throw new Error("Editor role required to trigger pipelines.");
      }
      await postJSON<{ run: { id: string } }, { pipeline_id: string }>("/api/v1/pipelines", { pipeline_id: pipelineID }, token.trim() || undefined);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown trigger error");
    } finally {
      setPendingPipelineID(null);
    }
  }

  return { data, error, pendingPipelineID, refreshing, triggerPipeline, refresh: load };
}
