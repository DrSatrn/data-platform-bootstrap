// This hook loads dataset catalog metadata for the datasets page.
import { useEffect, useState } from "react";

import { fetchJSON } from "../../lib/api";

type Asset = {
  id: string;
  name: string;
  layer: string;
  description: string;
  owner: string;
  columns: Array<{ name: string; type: string }>;
};

type DatasetPayload = {
  assets: Asset[];
};

export function useDatasets() {
  const [data, setData] = useState<DatasetPayload | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJSON<DatasetPayload>("/api/v1/catalog")
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown datasets error"));
  }, []);

  return { data, error };
}
