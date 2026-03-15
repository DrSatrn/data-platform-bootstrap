// These tests provide a minimal regression net for the most important operator
// page states: read-only pipeline behavior, dashboard edit mode, and system
// loading or error rendering.
import { renderToStaticMarkup } from "react-dom/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

let mockAuth: any = {
  session: {
    principal: { subject: "alice", role: "viewer" },
    capabilities: {
      view_platform: true,
      trigger_runs: false,
      edit_dashboards: false,
      run_admin_terminal: false
    }
  }
};

let mockPipelineState: any = {
  data: {
    pipelines: [{ id: "personal_finance_pipeline", name: "Finance", description: "desc", owner: "platform-team", jobs: [{ id: "job_a", type: "ingest" }] }],
    runs: []
  },
  error: null,
  pendingPipelineID: null,
  refreshing: false,
  triggerPipeline: vi.fn(),
  refresh: vi.fn()
};

let mockArtifactState: any = { artifacts: [], error: null };

let mockDashboardState: any = {
  dashboard: null,
  dashboards: [],
  draft: null,
  widgetData: {},
  isEditing: false,
  isSaving: false,
  error: null,
  saveError: null,
  selectedDashboardID: null,
  selectDashboard: vi.fn(),
  startEditing: vi.fn(),
  cancelEditing: vi.fn(),
  updateDraft: vi.fn(),
  updateWidget: vi.fn(),
  updateWidgetFilter: vi.fn(),
  addWidget: vi.fn(),
  removeWidget: vi.fn(),
  moveWidget: vi.fn(),
  createDashboard: vi.fn(),
  duplicateDashboard: vi.fn(),
  deleteDashboard: vi.fn(async () => {}),
  saveDashboard: vi.fn(async () => {})
};

let mockSystemState: any = {
  health: { status: "ok", environment: "development" },
  quality: { checks: [] },
  overview: {
    known_pipelines: 1,
    known_assets: 1,
    run_history: 0,
    telemetry: { uptime_seconds: 1, total_requests: 1, total_errors: 0, total_commands: 0, request_counts: {}, recent_log_summary: [] },
    run_summary: { total_runs: 0, queued_runs: 0, running_runs: 0, succeeded_runs: 0, failed_runs: 0, completed_last_24_hours: 0, failed_last_24_hours: 0, average_duration_seconds: 0 },
    queue_summary: { queued: 0, active: 0, completed: 0, total: 0 },
    backup_summary: { bundle_count: 0 },
    persistence_modes: {}
  },
  logs: { logs: [] },
  audit: { events: [] },
  catalog: { summary: { assets_missing_docs: 0, assets_missing_quality: 0, assets_containing_pii: 0, lineage_edges: 0 }, assets: [] },
  error: null,
  refreshing: false,
  refresh: vi.fn(async () => {})
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
vi.mock("../features/dashboard/useDashboardData", () => ({
  useDashboardData: () => mockDashboardState
}));
vi.mock("../features/system/useSystemData", () => ({
  useSystemData: () => mockSystemState
}));
vi.mock("../components/AdminTerminal", () => ({
  AdminTerminal: () => <div>Admin Terminal Stub</div>
}));

import { DashboardPage } from "./DashboardPage";
import { PipelinesPage } from "./PipelinesPage";
import { SystemPage } from "./SystemPage";

describe("operator page states", () => {
  beforeEach(() => {
    mockAuth = {
      session: {
        principal: { subject: "alice", role: "viewer" },
        capabilities: {
          view_platform: true,
          trigger_runs: false,
          edit_dashboards: false,
          run_admin_terminal: false
        }
      }
    };
    mockPipelineState = {
      data: {
        pipelines: [{ id: "personal_finance_pipeline", name: "Finance", description: "desc", owner: "platform-team", jobs: [{ id: "job_a", type: "ingest" }] }],
        runs: []
      },
      error: null,
      pendingPipelineID: null,
      refreshing: false,
      triggerPipeline: vi.fn(),
      refresh: vi.fn()
    };
    mockArtifactState = { artifacts: [], error: null };
    mockDashboardState = {
      dashboard: null,
      dashboards: [],
      draft: null,
      widgetData: {},
      isEditing: false,
      isSaving: false,
      error: null,
      saveError: null,
      selectedDashboardID: null,
      selectDashboard: vi.fn(),
      startEditing: vi.fn(),
      cancelEditing: vi.fn(),
      updateDraft: vi.fn(),
      updateWidget: vi.fn(),
      updateWidgetFilter: vi.fn(),
      addWidget: vi.fn(),
      removeWidget: vi.fn(),
      moveWidget: vi.fn(),
      createDashboard: vi.fn(),
      duplicateDashboard: vi.fn(),
      deleteDashboard: vi.fn(async () => {}),
      saveDashboard: vi.fn(async () => {})
    };
    mockSystemState = {
      health: { status: "ok", environment: "development" },
      quality: { checks: [] },
      overview: {
        known_pipelines: 1,
        known_assets: 1,
        run_history: 0,
        telemetry: { uptime_seconds: 1, total_requests: 1, total_errors: 0, total_commands: 0, request_counts: {}, recent_log_summary: [] },
        run_summary: { total_runs: 0, queued_runs: 0, running_runs: 0, succeeded_runs: 0, failed_runs: 0, completed_last_24_hours: 0, failed_last_24_hours: 0, average_duration_seconds: 0 },
        queue_summary: { queued: 0, active: 0, completed: 0, total: 0 },
        backup_summary: { bundle_count: 0 },
        persistence_modes: {}
      },
      logs: { logs: [] },
      audit: { events: [] },
      catalog: { summary: { assets_missing_docs: 0, assets_missing_quality: 0, assets_containing_pii: 0, lineage_edges: 0 }, assets: [] },
      error: null,
      refreshing: false,
      refresh: vi.fn(async () => {})
    };
  });

  it("shows the read-only pipeline guidance when the session cannot trigger runs", () => {
    const html = renderToStaticMarkup(<PipelinesPage />);
    expect(html).toContain("Editor token required to queue manual runs from the UI.");
    expect(html).toContain("Run now");
  });

  it("renders dashboard edit mode when a draft is active", () => {
    mockAuth.session.capabilities.edit_dashboards = true;
    mockAuth.session.principal.role = "editor";
    mockDashboardState = {
      ...mockDashboardState,
      dashboards: [{ id: "finance_overview", name: "Finance Overview", description: "desc", widgets: [] }],
      dashboard: { id: "finance_overview", name: "Finance Overview", description: "desc", widgets: [] },
      draft: {
        id: "finance_overview",
        name: "Finance Overview",
        description: "desc",
        widgets: [{ id: "savings_rate_kpi", name: "Savings Rate", type: "kpi", metric_ref: "metrics_savings_rate" }]
      },
      isEditing: true
    };
    const html = renderToStaticMarkup(<DashboardPage />);
    expect(html).toContain("Dashboard Editor");
    expect(html).toContain("Save dashboard");
  });

  it("shows a system loading state before the first payload arrives", () => {
    mockSystemState = {
      ...mockSystemState,
      health: null,
      overview: null,
      catalog: null
    };
    const html = renderToStaticMarkup(<SystemPage />);
    expect(html).toContain("Loading system view...");
  });

  it("shows a system error state clearly", () => {
    mockSystemState = {
      ...mockSystemState,
      error: "Viewer role required to access the system view."
    };
    const html = renderToStaticMarkup(<SystemPage />);
    expect(html).toContain("System error");
    expect(html).toContain("Viewer role required to access the system view.");
  });
});
