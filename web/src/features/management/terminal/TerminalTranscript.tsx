import type { TerminalSession } from "./sessionModel";

type TerminalTranscriptProps = {
  session: TerminalSession;
};

export function TerminalTranscript({ session }: TerminalTranscriptProps) {
  return (
    <article style={panelStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Live Session Draft</p>
          <h3 style={{ margin: "6px 0 0" }}>{session.title}</h3>
        </div>
        <span style={statusBadgeStyle(session.status)}>{session.status}</span>
      </div>

      <div style={metaRowStyle}>
        <span style={chipStyle}>{session.scope}</span>
        {session.lastExitCode !== undefined ? <span style={chipStyle}>exit {session.lastExitCode}</span> : null}
        {session.recommendedRunbook ? <span style={chipStyle}>runbook ready</span> : null}
      </div>

      <div style={transcriptStyle}>
        {session.entries.map((entry) => (
          <div key={entry.id} style={entryStyle(entry.kind)}>
            <span style={entryPrefixStyle(entry.kind)}>{prefixFor(entry.kind)}</span>
            <div style={{ display: "grid", gap: "4px" }}>
              <code style={entryContentStyle}>{entry.content}</code>
              <span style={timestampStyle}>{entry.at}</span>
            </div>
          </div>
        ))}
      </div>

      <div style={artifactBlockStyle}>
        <strong>Pinned Artifacts</strong>
        {session.pinnedArtifacts.length === 0 ? (
          <p style={mutedStyle}>No promoted artifacts from this session yet.</p>
        ) : (
          <div style={artifactListStyle}>
            {session.pinnedArtifacts.map((artifact) => (
              <span key={artifact} style={artifactChipStyle}>
                {artifact}
              </span>
            ))}
          </div>
        )}
      </div>
    </article>
  );
}

function prefixFor(kind: TerminalSession["entries"][number]["kind"]) {
  switch (kind) {
    case "command":
      return "$";
    case "stdout":
      return ">";
    case "stderr":
      return "!";
    default:
      return "i";
  }
}

const panelStyle = {
  display: "grid",
  gap: "14px",
  padding: "18px",
  borderRadius: "22px",
  background: "rgba(19, 24, 32, 0.96)",
  color: "#f5efe5",
  border: "1px solid rgba(214, 170, 101, 0.18)"
} as const;

const headerStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "12px",
  alignItems: "flex-start"
} as const;

const metaRowStyle = {
  display: "flex",
  gap: "8px",
  flexWrap: "wrap" as const
};

const chipStyle = {
  display: "inline-flex",
  alignItems: "center",
  padding: "5px 9px",
  borderRadius: "999px",
  background: "rgba(214, 170, 101, 0.12)",
  color: "#f4d7a2",
  fontSize: "0.78rem"
} as const;

const transcriptStyle = {
  display: "grid",
  gap: "10px",
  padding: "14px",
  borderRadius: "18px",
  background: "rgba(8, 11, 16, 0.84)",
  minHeight: "220px"
} as const;

const artifactBlockStyle = {
  display: "grid",
  gap: "10px"
} as const;

const artifactListStyle = {
  display: "flex",
  gap: "8px",
  flexWrap: "wrap" as const
} as const;

const artifactChipStyle = {
  display: "inline-flex",
  alignItems: "center",
  padding: "8px 10px",
  borderRadius: "14px",
  background: "rgba(255, 255, 255, 0.08)",
  color: "#efe2c2",
  fontSize: "0.8rem"
} as const;

const entryContentStyle = {
  whiteSpace: "pre-wrap" as const,
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace"
} as const;

const timestampStyle = {
  color: "rgba(244, 230, 199, 0.55)",
  fontSize: "0.75rem"
} as const;

const eyebrowStyle = {
  margin: 0,
  textTransform: "uppercase" as const,
  letterSpacing: "0.12em",
  fontSize: "0.72rem",
  color: "#d6aa65"
} as const;

const mutedStyle = {
  margin: 0,
  color: "rgba(245, 239, 229, 0.68)"
} as const;

function entryStyle(kind: TerminalSession["entries"][number]["kind"]) {
  const accent =
    kind === "stderr"
      ? "rgba(220, 117, 88, 0.42)"
      : kind === "command"
        ? "rgba(214, 170, 101, 0.34)"
        : "rgba(117, 168, 214, 0.26)";

  return {
    display: "grid",
    gridTemplateColumns: "18px minmax(0, 1fr)",
    gap: "10px",
    padding: "10px 12px",
    borderRadius: "14px",
    background: "rgba(255, 255, 255, 0.03)",
    border: `1px solid ${accent}`
  } as const;
}

function entryPrefixStyle(kind: TerminalSession["entries"][number]["kind"]) {
  const color =
    kind === "stderr" ? "#ff9d7a" : kind === "command" ? "#f4d7a2" : "#9ad0ff";
  return {
    color,
    fontWeight: 700,
    lineHeight: 1.5
  } as const;
}

function statusBadgeStyle(status: TerminalSession["status"]) {
  const background =
    status === "failed"
      ? "rgba(184, 61, 33, 0.18)"
      : status === "completed"
        ? "rgba(87, 162, 110, 0.18)"
        : "rgba(214, 170, 101, 0.16)";
  const color = status === "failed" ? "#ffb0a0" : status === "completed" ? "#b5e3bf" : "#f4d7a2";
  return {
    display: "inline-flex",
    alignItems: "center",
    borderRadius: "999px",
    padding: "7px 10px",
    background,
    color,
    textTransform: "uppercase" as const,
    letterSpacing: "0.08em",
    fontSize: "0.74rem"
  } as const;
}
