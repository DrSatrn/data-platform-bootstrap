// This hook owns dashboard selection, editing, saving, and widget hydration.
// Centralizing that behavior keeps the dashboard page readable even as the
// reporting surface starts behaving more like a real product than a static
// report.
import { useEffect, useMemo, useState } from "react";

import { deleteJSON, fetchJSON, postJSON } from "../../lib/api";

export type DashboardWidget = {
  id: string;
  name: string;
  type: string;
  description?: string;
  dataset_ref?: string;
  metric_ref?: string;
  value_field?: string;
  x_axis?: string;
  y_axis?: string;
  limit?: number;
  filters?: {
    from_month?: string;
    to_month?: string;
    category?: string;
  };
};

export type DashboardDefinition = {
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
  dashboard: DashboardDefinition | null;
  dashboards: DashboardDefinition[];
  draft: DashboardDefinition | null;
  widgetData: Record<string, QueryPayload["query"]>;
  isEditing: boolean;
  isSaving: boolean;
  error: string | null;
  saveError: string | null;
  selectedDashboardID: string | null;
  selectDashboard: (dashboardID: string) => void;
  startEditing: () => void;
  cancelEditing: () => void;
  updateDraft: (field: "name" | "description", value: string) => void;
  updateWidget: (widgetID: string, field: keyof DashboardWidget, value: string | number) => void;
  updateWidgetFilter: (widgetID: string, field: "from_month" | "to_month" | "category", value: string) => void;
  addWidget: () => void;
  removeWidget: (widgetID: string) => void;
  moveWidget: (widgetID: string, direction: -1 | 1) => void;
  createDashboard: () => void;
  duplicateDashboard: () => void;
  deleteDashboard: () => Promise<void>;
  saveDashboard: () => Promise<void>;
};

const emptyWidget = (): DashboardWidget => ({
  id: `widget_${Date.now()}`,
  name: "New Widget",
  type: "table",
  dataset_ref: "mart_monthly_cashflow",
  x_axis: "month",
  y_axis: "net_cashflow",
  limit: 12,
  filters: {}
});

export function useDashboardData(): DashboardData {
  const [dashboards, setDashboards] = useState<DashboardDefinition[]>([]);
  const [selectedDashboardID, setSelectedDashboardID] = useState<string | null>(null);
  const [widgetData, setWidgetData] = useState<Record<string, QueryPayload["query"]>>({});
  const [draft, setDraft] = useState<DashboardDefinition | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);

  const dashboard = useMemo(
    () => dashboards.find((item) => item.id === selectedDashboardID) ?? dashboards[0] ?? null,
    [dashboards, selectedDashboardID]
  );
  const previewDashboard = isEditing && draft ? draft : dashboard;

  useEffect(() => {
    void loadDashboards();
  }, []);

  useEffect(() => {
    if (!previewDashboard) {
      setWidgetData({});
      return;
    }
    void hydrateDashboard(previewDashboard);
  }, [previewDashboard]);

  async function loadDashboards(preferredID?: string) {
    try {
      const reports = await fetchJSON<ReportsPayload>("/api/v1/reports");
      setDashboards(reports.dashboards);
      setSelectedDashboardID((current) => preferredID ?? current ?? reports.dashboards[0]?.id ?? null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown dashboard error");
    }
  }

  async function hydrateDashboard(activeDashboard: DashboardDefinition) {
    try {
      const entries = await Promise.all(
        activeDashboard.widgets.map(async (widget) => {
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
      setWidgetData(Object.fromEntries(entries));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown widget hydration error");
    }
  }

  function selectDashboard(dashboardID: string) {
    setSelectedDashboardID(dashboardID);
    setIsEditing(false);
    setDraft(null);
    setSaveError(null);
  }

  function startEditing() {
    if (!dashboard) {
      return;
    }
    setDraft(JSON.parse(JSON.stringify(dashboard)) as DashboardDefinition);
    setIsEditing(true);
    setSaveError(null);
  }

  function cancelEditing() {
    setDraft(null);
    setIsEditing(false);
    setSaveError(null);
  }

  function createDashboard() {
    const nextDashboard: DashboardDefinition = {
      id: `dashboard_${Date.now()}`,
      name: "New Dashboard",
      description: "Describe the purpose of this dashboard.",
      widgets: [emptyWidget()]
    };
    setDraft(nextDashboard);
    setIsEditing(true);
    setSaveError(null);
  }

  function duplicateDashboard() {
    if (!dashboard) {
      return;
    }
    const clonedDashboard: DashboardDefinition = {
      ...JSON.parse(JSON.stringify(dashboard)),
      id: `${dashboard.id}_copy_${Date.now()}`,
      name: `${dashboard.name} Copy`,
      widgets: dashboard.widgets.map((widget) => ({
        ...JSON.parse(JSON.stringify(widget)),
        id: `${widget.id}_copy_${Date.now()}`
      }))
    };
    setDraft(clonedDashboard);
    setIsEditing(true);
    setSaveError(null);
  }

  function updateDraft(field: "name" | "description", value: string) {
    setDraft((current) => (current ? { ...current, [field]: value } : current));
  }

  function updateWidget(widgetID: string, field: keyof DashboardWidget, value: string | number) {
    setDraft((current) => {
      if (!current) {
        return current;
      }
      return {
        ...current,
        widgets: current.widgets.map((widget) =>
          widget.id === widgetID
            ? {
                ...widget,
                [field]: value,
                // Keep the widget query source explicit so analytics requests do
                // not silently prefer a metric when the operator intended a dataset.
                ...(field === "dataset_ref" && value ? { metric_ref: "" } : {}),
                ...(field === "metric_ref" && value ? { dataset_ref: "" } : {})
              }
            : widget
        )
      };
    });
  }

  function updateWidgetFilter(widgetID: string, field: "from_month" | "to_month" | "category", value: string) {
    setDraft((current) => {
      if (!current) {
        return current;
      }
      return {
        ...current,
        widgets: current.widgets.map((widget) =>
          widget.id === widgetID
            ? {
                ...widget,
                filters: {
                  ...(widget.filters ?? {}),
                  [field]: value
                }
              }
            : widget
        )
      };
    });
  }

  function addWidget() {
    setDraft((current) =>
      current
        ? {
            ...current,
            widgets: [...current.widgets, emptyWidget()]
          }
        : current
    );
  }

  function removeWidget(widgetID: string) {
    setDraft((current) =>
      current
        ? {
            ...current,
            widgets: current.widgets.filter((widget) => widget.id !== widgetID)
          }
        : current
    );
  }

  function moveWidget(widgetID: string, direction: -1 | 1) {
    setDraft((current) => {
      if (!current) {
        return current;
      }
      const widgets = [...current.widgets];
      const index = widgets.findIndex((widget) => widget.id === widgetID);
      const targetIndex = index + direction;
      if (index < 0 || targetIndex < 0 || targetIndex >= widgets.length) {
        return current;
      }
      const [widget] = widgets.splice(index, 1);
      widgets.splice(targetIndex, 0, widget);
      return {
        ...current,
        widgets
      };
    });
  }

  async function saveDashboard() {
    if (!draft) {
      return;
    }

    setIsSaving(true);
    setSaveError(null);
    try {
      await postJSON<{ dashboard: DashboardDefinition }, DashboardDefinition>("/api/v1/reports", draft);
      await loadDashboards(draft.id);
      setDraft(null);
      setIsEditing(false);
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : "Unknown dashboard save error");
    } finally {
      setIsSaving(false);
    }
  }

  async function deleteDashboard() {
    if (!dashboard) {
      return;
    }

    setIsSaving(true);
    setSaveError(null);
    try {
      await deleteJSON<{ deleted: string }>(`/api/v1/reports?id=${encodeURIComponent(dashboard.id)}`);
      const remainingDashboards = dashboards.filter((item) => item.id !== dashboard.id);
      setSelectedDashboardID(remainingDashboards[0]?.id ?? null);
      setDraft(null);
      setIsEditing(false);
      await loadDashboards(remainingDashboards[0]?.id);
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : "Unknown dashboard delete error");
    } finally {
      setIsSaving(false);
    }
  }

  return {
    dashboard,
    dashboards,
    draft,
    widgetData,
    isEditing,
    isSaving,
    error,
    saveError,
    selectedDashboardID,
    selectDashboard,
    startEditing,
    cancelEditing,
    updateDraft,
    updateWidget,
    updateWidgetFilter,
    addWidget,
    removeWidget,
    moveWidget,
    createDashboard,
    duplicateDashboard,
    deleteDashboard,
    saveDashboard
  };
}
