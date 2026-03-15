// This hook owns dashboard selection, editing, saving, and widget hydration.
// Centralizing that behavior keeps the dashboard page readable even as the
// reporting surface starts behaving more like a real product than a static
// report.
import { useEffect, useMemo, useState } from "react";

import { useAuth } from "../auth/useAuth";
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
  owner?: string;
  tags?: string[];
  shared_role?: string;
  default_filters?: {
    from_month?: string;
    to_month?: string;
    category?: string;
  };
  presets?: DashboardPreset[];
  widgets: DashboardWidget[];
};

export type DashboardPreset = {
  id: string;
  name: string;
  description?: string;
  filters?: {
    from_month?: string;
    to_month?: string;
    category?: string;
  };
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
  selectedPresetID: string | null;
  selectDashboard: (dashboardID: string) => void;
  selectPreset: (presetID: string) => void;
  startEditing: () => void;
  cancelEditing: () => void;
  updateDraft: (field: "name" | "description" | "owner" | "shared_role" | "tags", value: string) => void;
  updateDashboardFilter: (field: "from_month" | "to_month" | "category", value: string) => void;
  addPreset: () => void;
  removePreset: (presetID: string) => void;
  updatePreset: (presetID: string, field: "name" | "description", value: string) => void;
  updatePresetFilter: (presetID: string, field: "from_month" | "to_month" | "category", value: string) => void;
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
  const { loading, token, session } = useAuth();
  const [dashboards, setDashboards] = useState<DashboardDefinition[]>([]);
  const [selectedDashboardID, setSelectedDashboardID] = useState<string | null>(null);
  const [selectedPresetID, setSelectedPresetID] = useState<string | null>(null);
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
  const selectedPreset = useMemo(
    () => previewDashboard?.presets?.find((preset) => preset.id === selectedPresetID) ?? null,
    [previewDashboard, selectedPresetID]
  );

  useEffect(() => {
    if (loading) {
      return;
    }
    if (!session?.capabilities.view_platform) {
      setDashboards([]);
      setSelectedDashboardID(null);
      setError("Viewer role required to access dashboards.");
      return;
    }
    void loadDashboards();
  }, [loading, session]);

  useEffect(() => {
    if (!previewDashboard) {
      setWidgetData({});
      return;
    }
    if (!session?.capabilities.view_platform) {
      setWidgetData({});
      return;
    }
    void hydrateDashboard(previewDashboard);
  }, [previewDashboard, selectedPreset, session]);

  async function loadDashboards(preferredID?: string) {
    try {
      const reports = await fetchJSON<ReportsPayload>("/api/v1/reports");
      setDashboards(reports.dashboards);
      const nextDashboardID = preferredID ?? selectedDashboardID ?? reports.dashboards[0]?.id ?? null;
      setSelectedDashboardID(nextDashboardID);
      const nextDashboard = reports.dashboards.find((item) => item.id === nextDashboardID) ?? reports.dashboards[0] ?? null;
      setSelectedPresetID(nextDashboard?.presets?.[0]?.id ?? null);
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
          const mergedFilters = {
            ...(activeDashboard.default_filters ?? {}),
            ...(selectedPreset?.filters ?? {}),
            ...(widget.filters ?? {})
          };
          if (mergedFilters.from_month) {
            params.set("from_month", mergedFilters.from_month);
          }
          if (mergedFilters.to_month) {
            params.set("to_month", mergedFilters.to_month);
          }
          if (mergedFilters.category) {
            params.set("category", mergedFilters.category);
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
    const nextDashboard = dashboards.find((item) => item.id === dashboardID) ?? null;
    setSelectedPresetID(nextDashboard?.presets?.[0]?.id ?? null);
    setIsEditing(false);
    setDraft(null);
    setSaveError(null);
  }

  function selectPreset(presetID: string) {
    setSelectedPresetID(presetID || null);
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
      owner: session?.principal.subject ?? "platform-team",
      tags: ["custom"],
      shared_role: "viewer",
      default_filters: {},
      presets: [],
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
      owner: dashboard.owner,
      tags: [...(dashboard.tags ?? [])],
      shared_role: dashboard.shared_role,
      default_filters: { ...(dashboard.default_filters ?? {}) },
      presets: (dashboard.presets ?? []).map((preset) => ({
        ...JSON.parse(JSON.stringify(preset)),
        id: `${preset.id}_copy_${Date.now()}`
      })),
      widgets: dashboard.widgets.map((widget) => ({
        ...JSON.parse(JSON.stringify(widget)),
        id: `${widget.id}_copy_${Date.now()}`
      }))
    };
    setDraft(clonedDashboard);
    setIsEditing(true);
    setSaveError(null);
  }

  function updateDraft(field: "name" | "description" | "owner" | "shared_role" | "tags", value: string) {
    setDraft((current) => {
      if (!current) {
        return current;
      }
      if (field === "tags") {
        return {
          ...current,
          tags: value
            .split(",")
            .map((item) => item.trim())
            .filter(Boolean)
        };
      }
      return { ...current, [field]: value };
    });
  }

  function updateDashboardFilter(field: "from_month" | "to_month" | "category", value: string) {
    setDraft((current) =>
      current
        ? {
            ...current,
            default_filters: {
              ...(current.default_filters ?? {}),
              [field]: value
            }
          }
        : current
    );
  }

  function addPreset() {
    setDraft((current) =>
      current
        ? {
            ...current,
            presets: [
              ...(current.presets ?? []),
              {
                id: `preset_${Date.now()}`,
                name: "New Preset",
                description: "Describe when this preset should be used.",
                filters: {}
              }
            ]
          }
        : current
    );
  }

  function removePreset(presetID: string) {
    setDraft((current) =>
      current
        ? {
            ...current,
            presets: (current.presets ?? []).filter((preset) => preset.id !== presetID)
          }
        : current
    );
  }

  function updatePreset(presetID: string, field: "name" | "description", value: string) {
    setDraft((current) =>
      current
        ? {
            ...current,
            presets: (current.presets ?? []).map((preset) =>
              preset.id === presetID
                ? {
                    ...preset,
                    [field]: value
                  }
                : preset
            )
          }
        : current
    );
  }

  function updatePresetFilter(presetID: string, field: "from_month" | "to_month" | "category", value: string) {
    setDraft((current) =>
      current
        ? {
            ...current,
            presets: (current.presets ?? []).map((preset) =>
              preset.id === presetID
                ? {
                    ...preset,
                    filters: {
                      ...(preset.filters ?? {}),
                      [field]: value
                    }
                  }
                : preset
            )
          }
        : current
    );
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
      if (!session?.capabilities.edit_dashboards) {
        throw new Error("Editor role required to save dashboards.");
      }
      await postJSON<{ dashboard: DashboardDefinition }, DashboardDefinition>(
        "/api/v1/reports",
        draft,
        token.trim() || undefined
      );
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
      if (!session?.capabilities.edit_dashboards) {
        throw new Error("Editor role required to delete dashboards.");
      }
      await deleteJSON<{ deleted: string }>(
        `/api/v1/reports?id=${encodeURIComponent(dashboard.id)}`,
        token.trim() || undefined
      );
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
    selectedPresetID,
    selectDashboard,
    selectPreset,
    startEditing,
    cancelEditing,
    updateDraft,
    updateDashboardFilter,
    addPreset,
    removePreset,
    updatePreset,
    updatePresetFilter,
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
