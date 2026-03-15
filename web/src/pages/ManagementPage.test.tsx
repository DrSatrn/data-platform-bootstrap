// These tests keep the integrated management surface honest as staged modules
// transition into live app wiring.
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";

const mockState = {
  loading: false,
  refreshing: false,
  pendingCommand: false,
  error: null,
  session: { principal: { role: "admin" } },
  commandDraft: "status",
  setCommandDraft: vi.fn(),
  selectedCommand: { id: "status", label: "Status", command: "status", area: "diagnostics", minimumRole: "admin", summary: "Show environment." },
  setSelectedCommand: vi.fn(),
  commands: [{ id: "status", label: "Status", command: "status", area: "diagnostics", minimumRole: "admin", summary: "Show environment." }],
  runbooks: [{ id: "operator-manual", label: "Operator Manual", path: "/docs/runbooks/operator-manual.md", reason: "Best central reference." }],
  reports: { dashboards: [{ id: "finance_overview", name: "Finance Overview" }] },
  opsview: {
    snapshots: [],
    external_tool_attention: { total_jobs: 0, failed_jobs: 0, running_jobs: 0, succeeded_jobs: 0, jobs_missing_logs: 0, jobs_missing_outputs: 0 },
    attention_rollup: { total_runs: 1, failed_runs: 0, running_runs: 0, succeeded_runs: 1, runs_with_external_tool_failures: 0, runs_missing_evidence: 0, external_tool_job_count: 0 }
  },
  externalToolSummaries: [],
  overview: { queue_summary: { queued: 0, active: 0, total: 0 } },
  serviceStatus: [{ id: "api", label: "API", state: "healthy", detail: "ok" }],
  queue: [],
  attentionAssets: [],
  sessions: [],
  followupDecks: [],
  runbookItems: [],
  evidenceItems: [],
  recentCommands: [],
  executeCommand: vi.fn(),
  refresh: vi.fn()
};

vi.mock("../features/management/useManagementConsole", () => ({
  useManagementConsole: () => mockState
}));

import { ManagementPage } from "./ManagementPage";

describe("ManagementPage", () => {
  it("renders the integrated operator surface with live wiring copy", () => {
    const html = renderToStaticMarkup(<ManagementPage />);
    expect(html).toContain("Management Console");
    expect(html).toContain("Guided Terminal");
    expect(html).toContain("External Tool Inspection");
    expect(html).toContain("Run Platform Command");
  });
});
