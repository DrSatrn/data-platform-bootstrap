import type { FollowupAction } from "./followupPlanner";

type ArtifactFollowupPanelProps = {
  actions: FollowupAction[];
};

export function ArtifactFollowupPanel({ actions }: ArtifactFollowupPanelProps) {
  return (
    <article style={panelStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Artifact Follow-up Draft</p>
          <h3 style={{ margin: "6px 0 0" }}>Post-Run Evidence</h3>
        </div>
      </div>

      <div style={listStyle}>
        {actions.map((action) => (
          <div key={action.id} style={cardStyle}>
            <div style={rowStyle}>
              <strong>{action.title}</strong>
              <span style={priorityBadgeStyle(action.priority)}>{action.priority}</span>
            </div>
            <p style={bodyStyle}>{action.rationale}</p>
            {action.artifactPath ? <code style={codeStyle}>{action.artifactPath}</code> : null}
            {action.suggestedCommand ? <code style={codeStyle}>{action.suggestedCommand}</code> : null}
            {action.runbookPath ? <span style={hintStyle}>Runbook: {action.runbookPath}</span> : null}
          </div>
        ))}
      </div>
    </article>
  );
}

const panelStyle = {
  display: "grid",
  gap: "12px",
  padding: "18px",
  borderRadius: "22px",
  background: "rgba(255, 255, 255, 0.86)",
  border: "1px solid rgba(115, 90, 61, 0.15)"
} as const;

const headerStyle = {
  display: "grid",
  gap: "4px"
} as const;

const listStyle = {
  display: "grid",
  gap: "10px"
} as const;

const cardStyle = {
  display: "grid",
  gap: "8px",
  padding: "14px",
  borderRadius: "16px",
  background: "rgba(250, 246, 239, 0.95)",
  border: "1px solid rgba(115, 90, 61, 0.12)"
} as const;

const rowStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "10px",
  alignItems: "flex-start"
} as const;

const bodyStyle = {
  margin: 0,
  color: "#5d5043"
} as const;

const codeStyle = {
  display: "block",
  padding: "8px 10px",
  borderRadius: "12px",
  background: "rgba(44, 33, 19, 0.06)",
  color: "#46392e",
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
  fontSize: "0.82rem",
  overflowX: "auto" as const
} as const;

const hintStyle = {
  color: "#7b6a5a",
  fontSize: "0.8rem"
} as const;

const eyebrowStyle = {
  margin: 0,
  textTransform: "uppercase" as const,
  letterSpacing: "0.12em",
  fontSize: "0.72rem",
  color: "#be7b2d"
} as const;

function priorityBadgeStyle(priority: FollowupAction["priority"]) {
  const background =
    priority === "now"
      ? "rgba(192, 79, 53, 0.12)"
      : priority === "soon"
        ? "rgba(214, 155, 78, 0.14)"
        : "rgba(92, 117, 141, 0.12)";
  const color = priority === "now" ? "#922f1f" : priority === "soon" ? "#7a521c" : "#44586c";
  return {
    display: "inline-flex",
    alignItems: "center",
    padding: "5px 8px",
    borderRadius: "999px",
    background,
    color,
    textTransform: "uppercase" as const,
    letterSpacing: "0.08em",
    fontSize: "0.72rem"
  } as const;
}
