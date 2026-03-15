// This hook loads dataset catalog metadata for the datasets page.
import { useEffect, useMemo, useState } from "react";

import { useAuth } from "../auth/useAuth";
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

export type AssetProfile = {
  asset_id: string;
  path: string;
  format: string;
  row_count: number;
  file_bytes: number;
  generated_at: string;
  observed_at?: string;
  profile_state: string;
  columns: Array<{
    name: string;
    observed_type: string;
    null_count: number;
    unique_count: number;
    sample_values: string[];
    min_value?: string;
    max_value?: string;
  }>;
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
  const { loading, session } = useAuth();
  const [data, setData] = useState<DatasetPayload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedAssetID, setSelectedAssetID] = useState<string | null>(null);
  const [profile, setProfile] = useState<AssetProfile | null>(null);
  const [profileError, setProfileError] = useState<string | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);

  useEffect(() => {
    if (loading) {
      return;
    }
    if (!session?.capabilities.view_platform) {
      setData(null);
      setError("Viewer role required to access the catalog.");
      return;
    }

    fetchJSON<DatasetPayload>("/api/v1/catalog")
      .then((payload) => {
        setData(payload);
        setError(null);
        setSelectedAssetID((current) => current ?? payload.assets[0]?.id ?? null);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown datasets error"));
  }, [loading, session]);

  const selectedAsset = useMemo(
    () => data?.assets.find((asset) => asset.id === selectedAssetID) ?? data?.assets[0] ?? null,
    [data, selectedAssetID]
  );

  useEffect(() => {
    if (!selectedAssetID) {
      setProfile(null);
      return;
    }
    if (!session?.capabilities.view_platform) {
      setProfile(null);
      return;
    }

    setProfileLoading(true);
    setProfileError(null);
    fetchJSON<AssetProfile>(`/api/v1/catalog/profile?asset_id=${encodeURIComponent(selectedAssetID)}`)
      .then((payload) => setProfile(payload))
      .catch((err) => {
        setProfile(null);
        setProfileError(err instanceof Error ? err.message : "Unknown dataset profile error");
      })
      .finally(() => setProfileLoading(false));
  }, [selectedAssetID, session]);

  return { data, error, profile, profileError, profileLoading, selectedAssetID, selectedAsset, setSelectedAssetID };
}
