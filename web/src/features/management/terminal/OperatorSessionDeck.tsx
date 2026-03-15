import { useState } from "react";

import { mockTerminalSessions } from "./mockSessions";
import { sessionNeedsAttention, summarizeSession, type TerminalSession } from "./sessionModel";
import { TerminalTranscript } from "./TerminalTranscript";

type OperatorSessionDeckProps = {
  sessions?: TerminalSession[];
};

export function OperatorSessionDeck({ sessions = mockTerminalSessions }: OperatorSessionDeckProps) {
  const [activeID, setActiveID] = useState(sessions[0]?.id ?? "");
  const activeSession = sessions.find((session) => session.id === activeID) ?? sessions[0];

  if (!activeSession) {
    return null;
  }

  return (
    <section style={shellStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Terminal Sessions Draft</p>
          <h2 style={{ margin: "6px 0 8px" }}>Operator Session Deck</h2>
          <p style={mutedStyle}>
            Unwired staging component for an in-app terminal that can manage restores, dbt runs,
            and platform diagnostics without leaving the control plane.
          </p>
        </div>
      </div>

      <div style={layoutStyle}>
        <div style={sessionListStyle}>
          {sessions.map((session) => (
            <button
              key={session.id}
              onClick={() => setActiveID(session.id)}
              style={sessionCardStyle(session.id === activeSession.id, sessionNeedsAttention(session))}
              type="button"
            >
              <div style={rowBetweenStyle}>
                <strong>{session.title}</strong>
                <span style={statusPillStyle(session.status)}>{session.status}</span>
              </div>
              <span style={scopeStyle}>{session.scope}</span>
              <p style={summaryStyle}>{summarizeSession(session)}</p>
            </button>
          ))}
        </div>

        <TerminalTranscript session={activeSession} />
      </div>
    </section>
  );
}

const shellStyle = {
  display: "grid",
  gap: "18px",
  padding: "24px",
  borderRadius: "24px",
  border: "1px solid rgba(77, 57, 35, 0.18)",
  background:
    "linear-gradient(135deg, rgba(249, 244, 233, 0.98), rgba(237, 228, 212, 0.95))"
} as const;

const headerStyle = {
  display: "grid",
  gap: "8px"
} as const;

const layoutStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(280px, 320px) minmax(0, 1fr)",
  gap: "16px"
} as const;

const sessionListStyle = {
  display: "grid",
  gap: "10px"
} as const;

function sessionCardStyle(active: boolean, attention: boolean) {
  return {
    display: "grid",
    gap: "8px",
    padding: "16px",
    textAlign: "left" as const,
    borderRadius: "18px",
    border: active
      ? "1px solid rgba(188, 121, 45, 0.45)"
      : attention
        ? "1px solid rgba(184, 81, 54, 0.36)"
        : "1px solid rgba(114, 87, 58, 0.14)",
    background: active ? "rgba(255, 248, 237, 0.94)" : "rgba(255, 255, 255, 0.8)",
    cursor: "pointer"
  } as const;
}

const rowBetweenStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "12px",
  alignItems: "flex-start"
} as const;

function statusPillStyle(status: TerminalSession["status"]) {
  const color =
    status === "failed" ? "#8f2b1f" : status === "completed" ? "#29583a" : "#6f4a1f";
  const background =
    status === "failed"
      ? "rgba(184, 81, 54, 0.12)"
      : status === "completed"
        ? "rgba(85, 149, 102, 0.14)"
        : "rgba(212, 151, 73, 0.16)";
  return {
    display: "inline-flex",
    alignItems: "center",
    padding: "5px 8px",
    borderRadius: "999px",
    color,
    background,
    fontSize: "0.74rem",
    textTransform: "uppercase" as const,
    letterSpacing: "0.08em"
  } as const;
}

const scopeStyle = {
  color: "#7e6b57",
  fontSize: "0.84rem",
  textTransform: "uppercase" as const,
  letterSpacing: "0.08em"
} as const;

const summaryStyle = {
  margin: 0,
  color: "#564a3e"
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
  color: "#66584b"
} as const;
