// This hook loads orchestration state for the pipelines page and exposes a
// manual trigger action so the UI can kick off a real worker-backed run.
import { useCallback, useEffect, useState } from "react";

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
  const [data, setData] = useState<PipelinePayload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [pendingPipelineID, setPendingPipelineID] = useState<string | null>(null);

  const load = useCallback(() => {
    fetchJSON<PipelinePayload>("/api/v1/pipelines")
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown pipelines error"));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  async function triggerPipeline(pipelineID: string) {
    setPendingPipelineID(pipelineID);
    setError(null);
    try {
      await postJSON<{ run: { id: string } }, { pipeline_id: string }>("/api/v1/pipelines", { pipeline_id: pipelineID });
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown trigger error");
    } finally {
      setPendingPipelineID(null);
    }
  }

  return { data, error, pendingPipelineID, triggerPipeline };
}
