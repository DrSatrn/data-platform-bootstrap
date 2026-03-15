import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";

const mockAuth = {
  session: {
    capabilities: {
      trigger_runs: true
    }
  }
};

let mockPipelineState: any = {
  data: {
    pipelines: [
      {
        id: "finance",
        name: "Finance Pipeline",
        description: "Loads finance marts",
        owner: "platform-team",
        jobs: [{ id: "build_finance_dbt", type: "external_tool" }]
      }
    ],
    runs: []
  },
  error: null,
  pendingPipelineID: null,
  refreshing: false,
  triggerPipeline: vi.fn(),
  refresh: vi.fn()
};

let mockArtifactState: any = {
  artifacts: [],
  error: null
};

vi.mock("../features/auth/useAuth", () => ({
  useAuth: () => mockAuth
}));

vi.mock("../features/pipelines/usePipelines", () => ({
  usePipelines: () => mockPipelineState
}));

vi.mock("../features/pipelines/useRunArtifacts", () => ({
  useRunArtifacts: () => mockArtifactState
}));

import { PipelinesPage } from "./PipelinesPage";

describe("PipelinesPage", () => {
  it("renders without crashing", () => {
    const html = renderToStaticMarkup(<PipelinesPage />);
    expect(html).toContain("Finance Pipeline");
    expect(html).toContain("Recent Runs");
  });

  it("shows loading state initially", () => {
    mockPipelineState = {
      ...mockPipelineState,
      data: null,
      error: null
    };

    const html = renderToStaticMarkup(<PipelinesPage />);
    expect(html).toContain("Loading pipelines...");
  });

  it("shows error message on API failure", () => {
    mockPipelineState = {
      ...mockPipelineState,
      error: "Pipelines service unavailable"
    };

    const html = renderToStaticMarkup(<PipelinesPage />);
    expect(html).toContain("Pipelines error");
    expect(html).toContain("Pipelines service unavailable");
  });
});
