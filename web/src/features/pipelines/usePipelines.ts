// This hook loads orchestration state for the pipelines page.
import { useEffect, useState } from "react";

import { fetchJSON } from "../../lib/api";

type Pipeline = {
  id: string;
  name: string;
  description: string;
  owner: string;
  jobs: Array<{ id: string; type: string }>;
};

type PipelinePayload = {
  pipelines: Pipeline[];
  runs: Array<{ id: string; status: string; pipeline_id: string }>;
};

export function usePipelines() {
  const [data, setData] = useState<PipelinePayload | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJSON<PipelinePayload>("/api/v1/pipelines")
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown pipelines error"));
  }, []);

  return { data, error };
}
