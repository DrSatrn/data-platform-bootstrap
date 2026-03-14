// This hook loads run-scoped materialized artifacts for the selected pipeline
// run. It gives operators a direct window into what the worker produced.
import { useEffect, useState } from "react";

import { fetchJSON } from "../../lib/api";

type Artifact = {
  run_id: string;
  relative_path: string;
  size_bytes: number;
  modified_at: string;
  content_type: string;
};

type ArtifactPayload = {
  artifacts: Artifact[];
};

export function useRunArtifacts(runID: string | null) {
  const [artifacts, setArtifacts] = useState<Artifact[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!runID) {
      setArtifacts([]);
      return;
    }

    fetchJSON<ArtifactPayload>(`/api/v1/artifacts?run_id=${encodeURIComponent(runID)}`)
      .then((payload) => setArtifacts(payload.artifacts))
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown artifact error"));
  }, [runID]);

  return { artifacts, error };
}
