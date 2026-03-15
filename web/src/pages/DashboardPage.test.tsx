import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";

const mockAuth = {
  session: {
    principal: { subject: "alice", role: "editor" },
    capabilities: {
      edit_dashboards: true
    }
  }
};

let mockDashboardState: any = {
  dashboard: {
    id: "finance",
    name: "Finance Overview",
    description: "Saved dashboard",
    owner: "platform-team",
    tags: ["finance"],
    shared_role: "viewer",
    presets: [],
    widgets: [
      { id: "widget_1", name: "Cashflow", type: "kpi", value_field: "net_cashflow", layout: { x: 0, y: 0, w: 3, h: 1 } }
    ]
  },
  dashboards: [],
  draft: null,
  widgetData: {
    widget_1: {
      series: [{ month: "2026-03", net_cashflow: 1200 }]
    }
  },
  isEditing: false,
  isSaving: false,
  error: null,
  saveError: null,
  selectedDashboardID: "finance",
  selectedPresetID: null,
  selectDashboard: vi.fn(),
  selectPreset: vi.fn(),
  startEditing: vi.fn(),
  cancelEditing: vi.fn(),
  updateDraft: vi.fn(),
  updateDashboardFilter: vi.fn(),
  addPreset: vi.fn(),
  removePreset: vi.fn(),
  updatePreset: vi.fn(),
  updatePresetFilter: vi.fn(),
  updateWidget: vi.fn(),
  updateWidgetFilter: vi.fn(),
  addWidget: vi.fn(),
  removeWidget: vi.fn(),
  moveWidget: vi.fn(),
  nudgeWidget: vi.fn(),
  resizeWidget: vi.fn(),
  createDashboard: vi.fn(),
  duplicateDashboard: vi.fn(),
  deleteDashboard: vi.fn(),
  saveDashboard: vi.fn()
};

vi.mock("../features/auth/useAuth", () => ({
  useAuth: () => mockAuth
}));

vi.mock("../features/dashboard/useDashboardData", () => ({
  useDashboardData: () => mockDashboardState
}));

import { DashboardPage } from "./DashboardPage";

describe("DashboardPage", () => {
  it("renders the dashboard content", () => {
    const html = renderToStaticMarkup(<DashboardPage />);
    expect(html).toContain("Finance Overview");
    expect(html).toContain("Cashflow");
  });

  it("shows a loading state before dashboards are ready", () => {
    mockDashboardState = {
      ...mockDashboardState,
      dashboard: null,
      dashboards: [],
      error: null
    };

    const html = renderToStaticMarkup(<DashboardPage />);
    expect(html).toContain("Loading dashboards...");
  });

  it("shows an error state when dashboard loading fails", () => {
    mockDashboardState = {
      ...mockDashboardState,
      error: "Dashboard service unavailable"
    };

    const html = renderToStaticMarkup(<DashboardPage />);
    expect(html).toContain("Dashboard error");
    expect(html).toContain("Dashboard service unavailable");
  });
});
