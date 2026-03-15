import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";

let mockMetricsState: any = {
  metrics: [
    {
      definition: {
        id: "metrics_savings_rate",
        name: "Savings Rate",
        description: "Monthly savings rate",
        owner: "finance",
        dataset_ref: "mart_monthly_cashflow",
        time_dimension: "month",
        dimensions: ["month"],
        measures: ["savings_rate"],
        default_visualization: "line"
      },
      preview: [{ month: "2026-03", savings_rate: 0.34 }]
    }
  ],
  error: null,
  selectedMetricID: "metrics_savings_rate",
  selectedMetric: {
    definition: {
      id: "metrics_savings_rate",
      name: "Savings Rate",
      description: "Monthly savings rate",
      owner: "finance",
      dataset_ref: "mart_monthly_cashflow",
      time_dimension: "month",
      dimensions: ["month"],
      measures: ["savings_rate"],
      default_visualization: "line"
    }
  },
  setSelectedMetricID: vi.fn(),
  filters: { fromMonth: "", toMonth: "", category: "" },
  updateFilter: vi.fn(),
  preview: [{ month: "2026-03", savings_rate: 0.34 }]
};

vi.mock("../features/metrics/useMetrics", () => ({
  useMetrics: () => mockMetricsState
}));

import { MetricsPage } from "./MetricsPage";

describe("MetricsPage", () => {
  it("renders without crashing", () => {
    const html = renderToStaticMarkup(<MetricsPage />);
    expect(html).toContain("Metrics Registry");
    expect(html).toContain("Savings Rate");
  });

  it("shows loading state initially", () => {
    mockMetricsState = {
      ...mockMetricsState,
      metrics: [],
      selectedMetric: null,
      selectedMetricID: null,
      error: null,
      preview: []
    };

    const html = renderToStaticMarkup(<MetricsPage />);
    expect(html).toContain("Loading metrics...");
  });

  it("shows error message on API failure", () => {
    mockMetricsState = {
      ...mockMetricsState,
      error: "Metrics API unavailable"
    };

    const html = renderToStaticMarkup(<MetricsPage />);
    expect(html).toContain("Metrics error");
    expect(html).toContain("Metrics API unavailable");
  });
});
