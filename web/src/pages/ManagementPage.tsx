// This page is the real operator-facing management console. It wires the
// staged management modules into live platform APIs so operators can inspect
// posture, run guided commands, and follow evidence without leaving the app.
import { ControlPlaneWorkspace } from "../features/management/console/ControlPlaneWorkspace";
import { EvidenceBoardPanel } from "../features/management/evidence/EvidenceBoardPanel";
import { ExternalToolRunInspector } from "../features/management/externalTools/ExternalToolRunInspector";
import { OpsviewSummaryPanel } from "../features/management/opsview/OpsviewSummaryPanel";
import { RunbookDockPanel } from "../features/management/runbooks/RunbookDockPanel";
import { NextStepBoard } from "../features/management/terminal/NextStepBoard";
import { OperatorSessionDeck } from "../features/management/terminal/OperatorSessionDeck";
import { OperatorWorkbench } from "../features/management/terminal/OperatorWorkbench";
import { useManagementConsole } from "../features/management/useManagementConsole";

export function ManagementPage() {
  const {
    loading,
    refreshing,
    pendingCommand,
    error,
    session,
    commandDraft,
    setCommandDraft,
    selectedCommand,
    setSelectedCommand,
    commands,
    runbooks,
    reports,
    opsview,
    externalToolSummaries,
    overview,
    serviceStatus,
    queue,
    attentionAssets,
    sessions,
    followupDecks,
    runbookItems,
    evidenceItems,
    recentCommands,
    executeCommand,
    refresh
  } = useManagementConsole();

  if (loading && !overview) {
    return <section className="panel">Loading management console...</section>;
  }

  if (!overview || !opsview) {
    return <section className="panel">Management error: {error ?? "Management console data is unavailable."}</section>;
  }

  return (
    <section style={pageStyle}>
      <article style={heroStyle}>
        <div>
          <p style={eyebrowStyle}>Integrated Operator Surface</p>
          <h2 style={{ margin: "8px 0 10px" }}>Management Console</h2>
          <p style={mutedStyle}>
            Live control-plane posture, guided terminal workflows, backend opsview summaries, and operator evidence in one in-app surface.
          </p>
        </div>
        <div style={heroStatsStyle}>
          <StatPill label="Tracked Runs" value={String(opsview.attention_rollup.total_runs)} />
          <StatPill label="External Tool Jobs" value={String(opsview.attention_rollup.external_tool_job_count)} />
          <StatPill label="Dashboards" value={String(reports?.dashboards.length ?? 0)} />
          <button className="mini-button" onClick={() => void refresh()} type="button">
            {refreshing ? "Refreshing..." : "Refresh"}
          </button>
        </div>
      </article>

      {error ? <p className="muted">Management note: {error}</p> : null}

      <div style={gridStyle}>
        <OperatorWorkbench
          commands={commands}
          recentCommands={recentCommands}
          role={session?.principal.role as "viewer" | "editor" | "admin" | undefined}
          runbooks={runbooks}
          onCommandSelect={(command) => {
            setSelectedCommand(command);
            setCommandDraft(command.command);
          }}
        />
        <article style={panelStyle}>
          <div style={rowBetweenStyle}>
            <div>
              <p style={eyebrowStyle}>Guided Terminal</p>
              <h3 style={{ margin: "6px 0 0" }}>Run Platform Command</h3>
            </div>
            <span style={badgeStyle}>{pendingCommand ? "running" : "ready"}</span>
          </div>
          <p style={mutedStyle}>
            The terminal stays platform-oriented. Commands run through the admin API and become session transcripts instead of arbitrary shell access.
          </p>
          {selectedCommand ? (
            <div style={hintCardStyle}>
              <strong>{selectedCommand.label}</strong>
              <span style={mutedStyle}>{selectedCommand.summary}</span>
              <code style={codeStyle}>{selectedCommand.command}</code>
            </div>
          ) : null}
          <div style={commandRowStyle}>
            <input
              className="terminal-input"
              onChange={(event) => setCommandDraft(event.target.value)}
              placeholder="Choose a guided command or type a supported platform command"
              value={commandDraft}
            />
            <button
              className="nav-button active terminal-submit"
              disabled={pendingCommand}
              onClick={() => void executeCommand(commandDraft, selectedCommand)}
              type="button"
            >
              {pendingCommand ? "Running..." : "Run command"}
            </button>
          </div>
          <div style={chipRowStyle}>
            {commands.slice(0, 6).map((command) => (
              <button
                key={command.id}
                className="mini-button"
                onClick={() => {
                  setSelectedCommand(command);
                  setCommandDraft(command.command);
                }}
                type="button"
              >
                {command.label}
              </button>
            ))}
          </div>
        </article>
      </div>

      <ControlPlaneWorkspace attentionAssets={attentionAssets} queue={queue} services={serviceStatus} />

      <div style={twoUpStyle}>
        <OpsviewSummaryPanel attention={opsview.external_tool_attention} summaries={externalToolSummaries} />
        <article style={panelStyle}>
          <div style={rowBetweenStyle}>
            <div>
              <p style={eyebrowStyle}>Run Snapshot Window</p>
              <h3 style={{ margin: "6px 0 0" }}>External Tool Inspection</h3>
            </div>
            <span style={badgeStyle}>{externalToolSummaries.length} jobs</span>
          </div>
          <ExternalToolRunInspector summaries={externalToolSummaries} />
        </article>
      </div>

      <div style={twoUpStyle}>
        <OperatorSessionDeck sessions={sessions} />
        <RunbookDockPanel items={runbookItems} />
      </div>

      <div style={twoUpStyle}>
        <EvidenceBoardPanel items={evidenceItems} />
        <NextStepBoard decks={followupDecks} />
      </div>
    </section>
  );
}

function StatPill({ label, value }: { label: string; value: string }) {
  return (
    <div style={statPillStyle}>
      <span style={mutedStyle}>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

const pageStyle = {
  display: "grid",
  gap: "18px"
} as const;

const heroStyle = {
  display: "grid",
  gap: "16px",
  padding: "24px",
  borderRadius: "24px",
  border: "1px solid rgba(101, 74, 47, 0.16)",
  background: "linear-gradient(135deg, rgba(245, 203, 135, 0.24), rgba(255, 246, 231, 0.94))"
} as const;

const heroStatsStyle = {
  display: "flex",
  gap: "10px",
  flexWrap: "wrap" as const,
  alignItems: "center"
} as const;

const gridStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1.2fr) minmax(320px, 0.8fr)",
  gap: "16px",
  alignItems: "start"
} as const;

const twoUpStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
  gap: "16px",
  alignItems: "start"
} as const;

const panelStyle = {
  display: "grid",
  gap: "12px",
  padding: "18px",
  borderRadius: "22px",
  background: "rgba(255, 255, 255, 0.86)",
  border: "1px solid rgba(108, 82, 56, 0.14)"
} as const;

const rowBetweenStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "12px",
  alignItems: "flex-start"
} as const;

const commandRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  gap: "10px"
} as const;

const chipRowStyle = {
  display: "flex",
  gap: "8px",
  flexWrap: "wrap" as const
} as const;

const statPillStyle = {
  display: "grid",
  gap: "4px",
  padding: "10px 12px",
  borderRadius: "14px",
  background: "rgba(255, 255, 255, 0.75)",
  border: "1px solid rgba(101, 74, 47, 0.12)"
} as const;

const hintCardStyle = {
  display: "grid",
  gap: "6px",
  padding: "14px",
  borderRadius: "16px",
  background: "rgba(249, 245, 236, 0.96)",
  border: "1px solid rgba(108, 82, 56, 0.12)"
} as const;

const codeStyle = {
  display: "block",
  padding: "8px 10px",
  borderRadius: "12px",
  background: "rgba(44, 33, 19, 0.05)",
  color: "#4b3f33",
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
  fontSize: "0.82rem",
  overflowX: "auto" as const
} as const;

const badgeStyle = {
  display: "inline-flex",
  alignItems: "center",
  borderRadius: "999px",
  background: "rgba(44, 33, 19, 0.08)",
  color: "#4d4034",
  padding: "4px 8px",
  fontSize: "0.78rem"
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
  color: "#65594d"
} as const;
