import { describe, expect, it } from "vitest";

import { buildDashboardQueryState, buildWidgetAnalyticsParams, parseDashboardQueryState } from "./queryState";

describe("dashboard query state helpers", () => {
  it("parses and rebuilds shareable dashboard params", () => {
    const params = new URLSearchParams(
      "dashboard=finance_overview&preset=current_quarter&from_month=2026-01&to_month=2026-03&category=Groceries"
    );

    const parsed = parseDashboardQueryState(params);

    expect(parsed).toEqual({
      dashboardID: "finance_overview",
      presetID: "current_quarter",
      filters: {
        from_month: "2026-01",
        to_month: "2026-03",
        category: "Groceries"
      }
    });
    expect(buildDashboardQueryState(parsed).toString()).toBe(params.toString());
  });

  it("builds widget analytics params including multi-dimension group by", () => {
    const params = buildWidgetAnalyticsParams(
      {
        id: "widget_1",
        name: "Variance by Month and Category",
        type: "table",
        dataset_ref: "mart_budget_vs_actual",
        group_by: ["month", "category"],
        limit: 20
      },
      {
        from_month: "2026-01",
        category: "Groceries"
      }
    );

    expect(params.get("dataset")).toBe("mart_budget_vs_actual");
    expect(params.get("group_by")).toBe("month,category");
    expect(params.get("from_month")).toBe("2026-01");
    expect(params.get("category")).toBe("Groceries");
  });
});
