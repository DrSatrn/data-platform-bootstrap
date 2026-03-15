import { ControlPlaneWorkspace } from "./ControlPlaneWorkspace";
import { previewExternalToolArtifacts, previewExternalToolEvents } from "./mockManagementPreview";
import { buildExternalToolRunSummaries } from "../externalTools/externalToolRunSummary";
import { ExternalToolRunInspector } from "../externalTools/ExternalToolRunInspector";
import { OperatorWorkbench } from "../terminal/OperatorWorkbench";
import { OperatorSessionDeck } from "../terminal/OperatorSessionDeck";
import { NextStepBoard } from "../terminal/NextStepBoard";
import { mockTerminalSessions } from "../terminal/mockSessions";
import { mockFollowupDecks } from "../terminal/mockFollowups";
import { buildRunbookDock } from "../runbooks/runbookDock";
import { RunbookDockPanel } from "../runbooks/RunbookDockPanel";
import { buildEvidenceBoard } from "../evidence/evidenceBoard";
import { EvidenceBoardPanel } from "../evidence/EvidenceBoardPanel";
import { mockOpsviewAttention, mockOpsviewSummaries } from "../opsview/mockOpsview";
import { OpsviewSummaryPanel } from "../opsview/OpsviewSummaryPanel";

export function ManagementConsolePreview() {
  const externalToolSummaries = buildExternalToolRunSummaries(previewExternalToolEvents, previewExternalToolArtifacts);
  const dockItems = buildRunbookDock(
    mockTerminalSessions.map((session) => ({
      id: session.id,
      title: session.title,
      status: session.status,
      recommendedRunbook: session.recommendedRunbook
    })),
    mockFollowupDecks.flatMap((deck) => deck.actions)
  );
  const evidenceItems = buildEvidenceBoard(
    mockTerminalSessions.map((session) => ({
      id: session.id,
      title: session.title,
      status: session.status,
      pinnedArtifacts: session.pinnedArtifacts
    })),
    externalToolSummaries
  );

  return (
    <section style={shellStyle}>
      <div style={heroStyle}>
        <div>
          <p style={eyebrowStyle}>Integration Preview Draft</p>
          <h1 style={{ margin: "8px 0 12px" }}>Management Console Preview</h1>
          <p style={mutedStyle}>
            Unwired composite surface showing how the future in-app terminal and full management
            GUI can work together without leaving the control plane.
          </p>
        </div>
        <div style={pillRowStyle}>
          <span style={pillStyle}>workbench</span>
          <span style={pillStyle}>external tools</span>
          <span style={pillStyle}>sessions</span>
          <span style={pillStyle}>follow-up</span>
        </div>
      </div>

      <div style={stackStyle}>
        <OperatorWorkbench />
        <ControlPlaneWorkspace />
      </div>

      <div style={twoUpStyle}>
        <section style={panelStyle}>
          <ExternalToolRunInspector events={previewExternalToolEvents} artifacts={previewExternalToolArtifacts} />
        </section>
        <OperatorSessionDeck />
      </div>

      <div style={twoUpStyle}>
        <OpsviewSummaryPanel attention={mockOpsviewAttention} summaries={mockOpsviewSummaries} />
        <section style={panelStyle}>
          <RunbookDockPanel items={dockItems} />
        </section>
      </div>

      <div style={twoUpStyle}>
        <section style={panelStyle}>
          <EvidenceBoardPanel items={evidenceItems} />
        </section>
        <NextStepBoard />
      </div>
    </section>
  );
}

const shellStyle = {
  display: "grid",
  gap: "22px",
  padding: "28px",
  borderRadius: "28px",
  background:
    "radial-gradient(circle at top left, rgba(246, 211, 149, 0.28), transparent 28%), linear-gradient(180deg, rgba(250, 246, 239, 0.98), rgba(237, 228, 213, 0.96))",
  border: "1px solid rgba(96, 71, 47, 0.16)"
} as const;

const heroStyle = {
  display: "grid",
  gap: "14px",
  padding: "24px",
  borderRadius: "24px",
  background: "rgba(255, 251, 244, 0.82)",
  border: "1px solid rgba(96, 71, 47, 0.14)"
} as const;

const stackStyle = {
  display: "grid",
  gap: "18px"
} as const;

const twoUpStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) minmax(0, 1fr)",
  gap: "16px",
  alignItems: "start"
} as const;

const panelStyle = {
  display: "grid",
  gap: "12px",
  padding: "20px",
  borderRadius: "24px",
  background: "rgba(255, 255, 255, 0.82)",
  border: "1px solid rgba(96, 71, 47, 0.14)"
} as const;

const pillRowStyle = {
  display: "flex",
  gap: "8px",
  flexWrap: "wrap" as const
} as const;

const pillStyle = {
  display: "inline-flex",
  alignItems: "center",
  padding: "7px 10px",
  borderRadius: "999px",
  background: "rgba(214, 160, 84, 0.14)",
  color: "#73481b",
  fontSize: "0.78rem",
  textTransform: "uppercase" as const,
  letterSpacing: "0.08em"
} as const;

const eyebrowStyle = {
  margin: 0,
  textTransform: "uppercase" as const,
  letterSpacing: "0.12em",
  fontSize: "0.72rem",
  color: "#be7b2d"
} as const;

const mutedStyle = {
  margin: 0,
  color: "#65584b",
  maxWidth: "72ch"
} as const;
