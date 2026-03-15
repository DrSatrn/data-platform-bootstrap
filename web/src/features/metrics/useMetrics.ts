// This hook powers the semantic metric browser by loading repo-managed metric
// definitions and hydrating preview data through the constrained analytics API.
import { useEffect, useMemo, useState } from "react";

import { fetchJSON } from "../../lib/api";

export type MetricDefinition = {
  id: string;
  name: string;
  description: string;
  owner: string;
  dataset_ref: string;
  time_dimension: string;
  dimensions: string[];
  measures: string[];
  default_visualization: string;
};

export type MetricEntry = {
  definition: MetricDefinition;
  preview: Array<Record<string, string | number>>;
};

type MetricsPayload = {
  metrics: MetricEntry[];
};

export function useMetrics() {
  const [metrics, setMetrics] = useState<MetricEntry[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [selectedMetricID, setSelectedMetricID] = useState<string | null>(null);
  const [filters, setFilters] = useState({ fromMonth: "", toMonth: "", category: "" });
  const [preview, setPreview] = useState<Array<Record<string, string | number>>>([]);

  useEffect(() => {
    fetchJSON<MetricsPayload>("/api/v1/metrics")
      .then((payload) => {
        setMetrics(payload.metrics);
        setSelectedMetricID((current) => current ?? payload.metrics[0]?.definition.id ?? null);
        setPreview(payload.metrics[0]?.preview ?? []);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Unknown metrics error"));
  }, []);

  const selectedMetric = useMemo(
    () => metrics.find((item) => item.definition.id === selectedMetricID) ?? metrics[0] ?? null,
    [metrics, selectedMetricID]
  );

  useEffect(() => {
    if (!selectedMetric) {
      setPreview([]);
      return;
    }
    const query = new URLSearchParams({ metric: selectedMetric.definition.id });
    if (filters.fromMonth) {
      query.set("from_month", filters.fromMonth);
    }
    if (filters.toMonth) {
      query.set("to_month", filters.toMonth);
    }
    if (filters.category) {
      query.set("category", filters.category);
    }
    fetchJSON<{ dashboard: { series: Array<Record<string, string | number>> } }>(`/api/v1/analytics?${query.toString()}`)
      .then((payload) => setPreview(payload.dashboard.series))
      .catch((err) => setError(err instanceof Error ? err.message : "Metric preview error"));
  }, [filters, selectedMetric]);

  function updateFilter(field: "fromMonth" | "toMonth" | "category", value: string) {
    setFilters((current) => ({ ...current, [field]: value }));
  }

  return {
    metrics,
    error,
    selectedMetricID,
    selectedMetric,
    setSelectedMetricID,
    filters,
    updateFilter,
    preview
  };
}
