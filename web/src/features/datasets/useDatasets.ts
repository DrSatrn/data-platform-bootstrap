// This hook loads dataset catalog metadata for the datasets page.
import { useEffect, useMemo, useState } from "react";

import { useAuth } from "../auth/useAuth";
import { fetchJSON, patchJSON } from "../../lib/api";

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

export type DrilldownQuery = {
  dataset: string;
  series: Array<Record<string, string | number>>;
  available_dimensions?: string[];
  available_measures?: string[];
  group_by?: string;
  drill_dimension?: string;
  drill_value?: string;
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

type RawAsset = Omit<Asset, "source_refs" | "quality_check_refs" | "documentation_refs"> & {
  source_refs?: string[] | null;
  quality_check_refs?: string[] | null;
  documentation_refs?: string[] | null;
};

type RawAssetProfile = Omit<AssetProfile, "columns"> & {
  columns: Array<
    Omit<AssetProfile["columns"][number], "sample_values"> & {
      sample_values?: string[] | null;
    }
  >;
};

type RawDatasetPayload = Omit<DatasetPayload, "assets"> & {
  assets: RawAsset[];
};

type AssetUpdatePayload = {
  asset_id: string;
  owner?: string;
  description?: string;
  quality_check_refs?: string[];
  documentation_refs?: string[];
  column_descriptions?: Array<{ name: string; description?: string }>;
};

export function useDatasets() {
  const { loading, session } = useAuth();
  const [data, setData] = useState<DatasetPayload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedAssetID, setSelectedAssetID] = useState<string | null>(null);
  const [profile, setProfile] = useState<AssetProfile | null>(null);
  const [profileError, setProfileError] = useState<string | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);
  const [drilldown, setDrilldown] = useState<DrilldownQuery | null>(null);
  const [drilldownError, setDrilldownError] = useState<string | null>(null);
  const [drilldownLoading, setDrilldownLoading] = useState(false);
  const [drilldownFilters, setDrilldownFilters] = useState({
    fromMonth: "",
    toMonth: "",
    category: "",
    groupBy: "",
    drillDimension: "",
    drillValue: "",
    sortBy: "",
    sortDirection: "asc"
  });
  const [savePending, setSavePending] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  function loadCatalog() {
    return fetchJSON<RawDatasetPayload>("/api/v1/catalog").then((payload) => {
      const normalized = {
        ...payload,
        assets: payload.assets.map(normalizeAsset)
      };
      setData(normalized);
      setError(null);
      setSelectedAssetID((current) => current ?? normalized.assets[0]?.id ?? null);
      return normalized;
    });
  }

  useEffect(() => {
    if (loading) {
      return;
    }
    if (!session?.capabilities.view_platform) {
      setData(null);
      setError("Viewer role required to access the catalog.");
      return;
    }

    loadCatalog()
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
    fetchJSON<RawAssetProfile>(`/api/v1/catalog/profile?asset_id=${encodeURIComponent(selectedAssetID)}`)
      .then((payload) => setProfile(normalizeProfile(payload)))
      .catch((err) => {
        setProfile(null);
        setProfileError(err instanceof Error ? err.message : "Unknown dataset profile error");
      })
      .finally(() => setProfileLoading(false));
  }, [selectedAssetID, session]);

  useEffect(() => {
    if (!selectedAsset) {
      setDrilldown(null);
      return;
    }
    if (!session?.capabilities.view_platform) {
      setDrilldown(null);
      return;
    }
    if (!isCuratedQueryTarget(selectedAsset.id)) {
      setDrilldown(null);
      setDrilldownError(null);
      return;
    }

    setDrilldownLoading(true);
    setDrilldownError(null);
    const query = new URLSearchParams();
    if (selectedAsset.id.startsWith("metrics_")) {
      query.set("metric", selectedAsset.id);
    } else {
      query.set("dataset", selectedAsset.id);
    }
    if (drilldownFilters.fromMonth) {
      query.set("from_month", drilldownFilters.fromMonth);
    }
    if (drilldownFilters.toMonth) {
      query.set("to_month", drilldownFilters.toMonth);
    }
    if (drilldownFilters.category) {
      query.set("category", drilldownFilters.category);
    }
    if (drilldownFilters.groupBy) {
      query.set("group_by", drilldownFilters.groupBy);
    }
    if (drilldownFilters.drillDimension) {
      query.set("drill_dimension", drilldownFilters.drillDimension);
    }
    if (drilldownFilters.drillValue) {
      query.set("drill_value", drilldownFilters.drillValue);
    }
    if (drilldownFilters.sortBy) {
      query.set("sort_by", drilldownFilters.sortBy);
    }
    if (drilldownFilters.sortDirection) {
      query.set("sort_direction", drilldownFilters.sortDirection);
    }

    fetchJSON<{ query: DrilldownQuery }>(`/api/v1/analytics?${query.toString()}`)
      .then((payload) => setDrilldown(payload.query))
      .catch((err) => {
        setDrilldown(null);
        setDrilldownError(err instanceof Error ? err.message : "Unknown dataset drilldown error");
      })
      .finally(() => setDrilldownLoading(false));
  }, [drilldownFilters, selectedAsset, session]);

  async function saveAnnotations(payload: AssetUpdatePayload) {
    if (!session?.capabilities.edit_metadata) {
      throw new Error("Editor role required to update metadata.");
    }
    setSavePending(true);
    setSaveError(null);
    try {
      const response = await patchJSON<{ asset: Asset }, AssetUpdatePayload>("/api/v1/catalog", payload);
      await loadCatalog();
      setSelectedAssetID(response.asset.id);
      return response.asset;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Unknown metadata update error";
      setSaveError(message);
      throw err;
    } finally {
      setSavePending(false);
    }
  }

  return {
    data,
    drilldown,
    drilldownError,
    drilldownFilters,
    drilldownLoading,
    error,
    profile,
    profileError,
    profileLoading,
    saveAnnotations,
    saveError,
    savePending,
    selectedAssetID,
    selectedAsset,
    setSelectedAssetID,
    updateDrilldownFilter: (
      field: "fromMonth" | "toMonth" | "category" | "groupBy" | "drillDimension" | "drillValue" | "sortBy" | "sortDirection",
      value: string
    ) => setDrilldownFilters((current) => ({ ...current, [field]: value }))
  };
}

function isCuratedQueryTarget(assetID: string) {
  return assetID.startsWith("mart_") || assetID.startsWith("metrics_");
}

function normalizeAsset(asset: RawAsset): Asset {
  return {
    ...asset,
    source_refs: asset.source_refs ?? [],
    quality_check_refs: asset.quality_check_refs ?? [],
    documentation_refs: asset.documentation_refs ?? []
  };
}

function normalizeProfile(profile: RawAssetProfile): AssetProfile {
  return {
    ...profile,
    columns: profile.columns.map((column) => ({
      ...column,
      sample_values: column.sample_values ?? []
    }))
  };
}
