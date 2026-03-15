import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";

const mockAuth = {
  session: {
    capabilities: {
      edit_metadata: true
    }
  }
};

let mockDatasetState: any = {
  data: {
    summary: {
      total_assets: 1,
      documented_columns: 2,
      total_columns: 2,
      assets_missing_docs: 0,
      assets_missing_quality: 0
    },
    assets: [
      {
        id: "mart_monthly_cashflow",
        name: "Monthly Cashflow",
        layer: "mart",
        kind: "table",
        description: "Curated mart",
        owner: "finance",
        source_refs: [],
        quality_check_refs: [],
        documentation_refs: [],
        freshness_status: { state: "fresh", message: "Fresh" },
        coverage: {
          documented_columns: 2,
          total_columns: 2,
          has_documentation: true,
          has_quality_checks: true,
          contains_pii: false
        },
        lineage: {
          upstream: [],
          downstream: []
        },
        columns: [
          { name: "month", type: "string", description: "Month" },
          { name: "net_cashflow", type: "number", description: "Cashflow" }
        ]
      }
    ]
  },
  drilldown: null,
  drilldownError: null,
  drilldownFilters: {
    fromMonth: "",
    toMonth: "",
    category: "",
    groupBy: "",
    drillDimension: "",
    drillValue: "",
    sortBy: "",
    sortDirection: "asc"
  },
  drilldownLoading: false,
  error: null,
  profile: null,
  profileError: null,
  profileLoading: false,
  saveAnnotations: vi.fn(),
  saveError: null,
  savePending: false,
  selectedAsset: {
    id: "mart_monthly_cashflow",
    name: "Monthly Cashflow",
    layer: "mart",
    kind: "table",
    description: "Curated mart",
    owner: "finance",
    source_refs: [],
    quality_check_refs: [],
    documentation_refs: [],
    freshness_status: { state: "fresh", message: "Fresh" },
    coverage: {
      documented_columns: 2,
      total_columns: 2,
      has_documentation: true,
      has_quality_checks: true,
      contains_pii: false
    },
    lineage: {
      upstream: [],
      downstream: []
    },
    columns: [
      { name: "month", type: "string", description: "Month" },
      { name: "net_cashflow", type: "number", description: "Cashflow" }
    ]
  },
  selectedAssetID: "mart_monthly_cashflow",
  setSelectedAssetID: vi.fn(),
  updateDrilldownFilter: vi.fn()
};

vi.mock("../features/auth/useAuth", () => ({
  useAuth: () => mockAuth
}));

vi.mock("../features/datasets/useDatasets", () => ({
  useDatasets: () => mockDatasetState
}));

import { DatasetsPage } from "./DatasetsPage";

describe("DatasetsPage", () => {
  it("renders without crashing", () => {
    const html = renderToStaticMarkup(<DatasetsPage />);
    expect(html).toContain("Catalog Trust Summary");
    expect(html).toContain("Monthly Cashflow");
  });

  it("shows loading state initially", () => {
    mockDatasetState = {
      ...mockDatasetState,
      data: null,
      selectedAsset: null,
      error: null
    };

    const html = renderToStaticMarkup(<DatasetsPage />);
    expect(html).toContain("Loading datasets...");
  });

  it("shows error message on API failure", () => {
    mockDatasetState = {
      ...mockDatasetState,
      error: "Catalog API unavailable"
    };

    const html = renderToStaticMarkup(<DatasetsPage />);
    expect(html).toContain("Datasets error");
    expect(html).toContain("Catalog API unavailable");
  });
});
