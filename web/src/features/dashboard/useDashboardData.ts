// This hook loads the saved dashboard definition and then hydrates each widget
// from the constrained analytics API. Keeping the orchestration here prevents
// the page component from becoming a fetch-and-transform tangle as the
// reporting surface grows.
import { useEffect, useState } from "react";

import { fetchJSON } from "../../lib/api";

type DashboardWidget = {
  id: string;
  name: string;
  type: string;
  description?: string;
  dataset_ref?: string;
  metric_ref?: string;
  value_field?: string;
  limit?: number;
  filters?: {
    from_month?: string;
    to_month?: string;
    category?: string;
  };
};

type DashboardDefinition = {
  id: string;
  name: string;
  description: string;
  widgets: DashboardWidget[];
};

type ReportsPayload = {
  dashboards: DashboardDefinition[];
};

type QueryPayload = {
  query: {
    dataset: string;
    series: Array<Record<string, string | number>>;
  };
};

export type DashboardData = {
  dashboard: DashboardDefinition;
  widgetData: Record<string, QueryPayload["query"]>;
};

export function useDashboardData() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const reports = await fetchJSON<ReportsPayload>("/api/v1/reports");
        const dashboard = reports.dashboards[0];
        if (!dashboard) {
          throw new Error("No dashboards are available yet.");
        }

        const entries = await Promise.all(
          dashboard.widgets.map(async (widget) => {
            const params = new URLSearchParams();
            if (widget.metric_ref) {
              params.set("metric", widget.metric_ref);
            }
            if (widget.dataset_ref) {
              params.set("dataset", widget.dataset_ref);
            }
            if (widget.limit) {
              params.set("limit", String(widget.limit));
            }
            if (widget.filters?.from_month) {
              params.set("from_month", widget.filters.from_month);
            }
            if (widget.filters?.to_month) {
              params.set("to_month", widget.filters.to_month);
            }
            if (widget.filters?.category) {
              params.set("category", widget.filters.category);
            }

            const query = await fetchJSON<QueryPayload>(`/api/v1/analytics?${params.toString()}`);
            return [widget.id, query.query] as const;
          })
        );

        setData({
          dashboard,
          widgetData: Object.fromEntries(entries)
        });
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unknown dashboard error");
      }
    }

    load();
  }, []);

  return { data, error };
}
