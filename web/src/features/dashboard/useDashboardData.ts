// This hook gathers the curated dashboard data used by the homepage. Keeping
// the data access here prevents the page component from becoming a fetch-and-
// transform tangle as the dashboard grows.
import { useEffect, useState } from "react";

import { fetchJSON } from "../../lib/api";

export type DashboardPayload = {
  dashboard: {
    dataset: string;
    series: Array<Record<string, string | number>>;
  };
};

export function useDashboardData() {
  const [data, setData] = useState<DashboardPayload | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJSON<DashboardPayload>("/api/v1/analytics")
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown dashboard error"));
  }, []);

  return { data, error };
}
