import type { DashboardWidget } from "./useDashboardData";

export type DashboardViewFilters = {
  from_month?: string;
  to_month?: string;
  category?: string;
};

export type DashboardQueryState = {
  dashboardID: string | null;
  presetID: string | null;
  filters: DashboardViewFilters;
};

export function parseDashboardQueryState(searchParams: URLSearchParams): DashboardQueryState {
  return {
    dashboardID: searchParams.get("dashboard"),
    presetID: searchParams.get("preset"),
    filters: {
      from_month: emptyToUndefined(searchParams.get("from_month")),
      to_month: emptyToUndefined(searchParams.get("to_month")),
      category: emptyToUndefined(searchParams.get("category"))
    }
  };
}

export function buildDashboardQueryState(state: DashboardQueryState): URLSearchParams {
  const params = new URLSearchParams();
  if (state.dashboardID) {
    params.set("dashboard", state.dashboardID);
  }
  if (state.presetID) {
    params.set("preset", state.presetID);
  }
  if (state.filters.from_month) {
    params.set("from_month", state.filters.from_month);
  }
  if (state.filters.to_month) {
    params.set("to_month", state.filters.to_month);
  }
  if (state.filters.category) {
    params.set("category", state.filters.category);
  }
  return params;
}

export function buildWidgetAnalyticsParams(
  widget: DashboardWidget,
  filters: DashboardViewFilters
): URLSearchParams {
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
  if (filters.from_month) {
    params.set("from_month", filters.from_month);
  }
  if (filters.to_month) {
    params.set("to_month", filters.to_month);
  }
  if (filters.category) {
    params.set("category", filters.category);
  }
  const groupBy = normalizeGroupBy(widget.group_by);
  if (groupBy.length > 0) {
    params.set("group_by", groupBy.join(","));
  }
  return params;
}

export function normalizeGroupBy(groupBy?: string[]): string[] {
  if (!groupBy) {
    return [];
  }
  const seen = new Set<string>();
  const normalized: string[] = [];
  for (const value of groupBy) {
    const item = value.trim();
    if (!item || seen.has(item)) {
      continue;
    }
    seen.add(item);
    normalized.push(item);
  }
  return normalized;
}

function emptyToUndefined(value: string | null) {
  const next = value?.trim();
  return next ? next : undefined;
}
