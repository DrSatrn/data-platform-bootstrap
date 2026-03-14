// This hook gathers platform health and quality summaries for the system page.
import { useEffect, useState } from "react";

import { fetchJSON } from "../../lib/api";

type HealthPayload = {
  status: string;
  environment: string;
  http_addr: string;
  web_addr: string;
};

type QualityPayload = {
  checks: Array<{ id: string; name: string; status: string; message: string }>;
};

export function useSystemData() {
  const [health, setHealth] = useState<HealthPayload | null>(null);
  const [quality, setQuality] = useState<QualityPayload | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      fetchJSON<HealthPayload>("/healthz"),
      fetchJSON<QualityPayload>("/api/v1/quality")
    ])
      .then(([nextHealth, nextQuality]) => {
        setHealth(nextHealth);
        setQuality(nextQuality);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown system error"));
  }, []);

  return { health, quality, error };
}
