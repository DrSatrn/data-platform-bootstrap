// These tests cover the shell-level auth rendering so the browser keeps making
// the current access model obvious to operators as RBAC evolves.
import { renderToStaticMarkup } from "react-dom/server";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

let mockAuth = {
  token: "viewer-token",
  setToken: vi.fn(),
  clearToken: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  session: {
    principal: { subject: "alice", role: "viewer", auth_source: "session" },
    capabilities: {
      view_platform: true,
      trigger_runs: false,
      edit_metadata: false,
      edit_dashboards: false,
      run_admin_terminal: false,
      manage_users: false
    }
  },
  loading: false,
  error: null
};

vi.mock("../features/auth/useAuth", () => ({
  authStorageKey: "data-platform-auth-token",
  useAuth: () => mockAuth
}));

vi.mock("../pages/DashboardPage", () => ({ DashboardPage: () => <div>Dashboard Stub</div> }));
vi.mock("../pages/ManagementPage", () => ({ ManagementPage: () => <div>Management Stub</div> }));
vi.mock("../pages/MetricsPage", () => ({ MetricsPage: () => <div>Metrics Stub</div> }));
vi.mock("../pages/PipelinesPage", () => ({ PipelinesPage: () => <div>Pipelines Stub</div> }));
vi.mock("../pages/DatasetsPage", () => ({ DatasetsPage: () => <div>Datasets Stub</div> }));
vi.mock("../pages/SystemPage", () => ({ SystemPage: () => <div>System Stub</div> }));

import { App } from "./App";

describe("App", () => {
  it("renders the resolved session and access guidance", () => {
    const html = renderToStaticMarkup(
      <MemoryRouter initialEntries={["/dashboard"]}>
        <App />
      </MemoryRouter>
    );
    expect(html).toContain("alice");
    expect(html).toContain("viewer");
    expect(html).toContain("Native sessions are the normal path");
  });

  it("renders the page for a direct route", () => {
    const html = renderToStaticMarkup(
      <MemoryRouter initialEntries={["/pipelines"]}>
        <App />
      </MemoryRouter>
    );

    expect(html).toContain("Pipelines Stub");
    expect(html).toContain("/pipelines");
  });
});
