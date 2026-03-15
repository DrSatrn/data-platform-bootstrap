import type { CSSProperties } from "react";

import type { DashboardWidget } from "../../features/dashboard/useDashboardData";

export const datasetOptions = [
  "mart_monthly_cashflow",
  "mart_category_spend",
  "mart_budget_vs_actual"
];

export const metricOptions = ["metrics_savings_rate", "metrics_category_variance"];

export const axisOptions = [
  "month",
  "category",
  "total_income",
  "total_expense",
  "net_cashflow",
  "actual_spend",
  "budget_amount",
  "variance_amount"
];

export function deriveColumns(series: Array<Record<string, string | number>>) {
  const first = series[0];
  return first ? Object.keys(first) : [];
}

export function firstMetricField(row: Record<string, string | number>) {
  return Object.keys(row).find((key) => key !== "month" && key !== "category") ?? "value";
}

export function formatValue(value: string | number | undefined, field: string) {
  if (typeof value === "number" && field.includes("rate")) {
    return `${Math.round(value * 100)}%`;
  }
  if (
    typeof value === "number" &&
    (field.includes("income") ||
      field.includes("expense") ||
      field.includes("spend") ||
      field.includes("budget") ||
      field.includes("variance"))
  ) {
    return new Intl.NumberFormat("en-AU", {
      style: "currency",
      currency: "AUD",
      maximumFractionDigits: 0
    }).format(value);
  }
  return String(value ?? "");
}

export function layoutOf(widget: DashboardWidget) {
  return widget.layout ?? { x: 0, y: 0, w: widget.type === "kpi" ? 3 : 6, h: widget.type === "kpi" ? 1 : 2 };
}

export function widgetPlacementStyle(widget: DashboardWidget) {
  const layout = layoutOf(widget);
  return {
    ["--widget-col-start" as const]: String(layout.x + 1),
    ["--widget-col-span" as const]: String(Math.max(1, layout.w)),
    ["--widget-row-start" as const]: String(layout.y + 1),
    ["--widget-row-span" as const]: String(Math.max(1, layout.h))
  } as CSSProperties;
}

export function sortWidgets(widgets: DashboardWidget[]) {
  return [...widgets].sort((left, right) => {
    const leftLayout = layoutOf(left);
    const rightLayout = layoutOf(right);
    if (leftLayout.y !== rightLayout.y) {
      return leftLayout.y - rightLayout.y;
    }
    if (leftLayout.x !== rightLayout.x) {
      return leftLayout.x - rightLayout.x;
    }
    return left.name.localeCompare(right.name);
  });
}
