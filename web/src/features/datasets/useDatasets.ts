// This hook loads dataset catalog metadata for the datasets page.
import { useEffect, useMemo, useState } from "react";

import { fetchJSON } from "../../lib/api";

export type Asset = {
  id: string;
  name: string;
  layer: string;
  kind: string;
  description: string;
  owner: string;
  source_refs: string[];
  quality_check_refs: string[];
  documentation_refs: string[];
  freshness_status: {
    state: string;
    last_updated?: string;
    lag_seconds?: number;
    message: string;
  };
  coverage: {
    documented_columns: number;
    total_columns: number;
    has_documentation: boolean;
    has_quality_checks: boolean;
    contains_pii: boolean;
  };
  lineage: {
    upstream: string[];
    downstream: string[];
  };
  columns: Array<{ name: string; type: string; description: string; is_pii?: boolean }>;
};

type DatasetPayload = {
  assets: Asset[];
  summary: {
    total_assets: number;
    by_layer: Record<string, number>;
    by_freshness: Record<string, number>;
    assets_missing_docs: number;
    assets_missing_quality: number;
    assets_containing_pii: number;
    documented_columns: number;
    total_columns: number;
    lineage_edges: number;
  };
  lineage: Array<{
    from: string;
    to: string;
  }>;
};

export function useDatasets() {
  const [data, setData] = useState<DatasetPayload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedAssetID, setSelectedAssetID] = useState<string | null>(null);

  useEffect(() => {
    fetchJSON<DatasetPayload>("/api/v1/catalog")
      .then((payload) => {
        setData(payload);
        setSelectedAssetID((current) => current ?? payload.assets[0]?.id ?? null);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown datasets error"));
  }, []);

  const selectedAsset = useMemo(
    () => data?.assets.find((asset) => asset.id === selectedAssetID) ?? data?.assets[0] ?? null,
    [data, selectedAssetID]
  );

  return { data, error, selectedAssetID, selectedAsset, setSelectedAssetID };
}
