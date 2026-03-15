import type { RunbookDockItem } from "./runbookDock";

type RunbookDockPanelProps = {
  items: RunbookDockItem[];
};

export function RunbookDockPanel({ items }: RunbookDockPanelProps) {
  return (
    <article style={panelStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Runbook Dock Draft</p>
          <h3 style={{ margin: "6px 0 0" }}>Operator Runbooks</h3>
        </div>
      </div>

      <div style={listStyle}>
        {items.map((item) => (
          <div key={item.id} style={cardStyle}>
            <div style={rowStyle}>
              <strong>{item.label}</strong>
              <span style={badgeStyle(item.urgency)}>{item.urgency}</span>
            </div>
            <code style={pathStyle}>{item.path}</code>
            <p style={reasonStyle}>{item.reason}</p>
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
  border: "1px solid rgba(108, 82, 56, 0.14)"
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
  background: "rgba(249, 245, 236, 0.96)",
  border: "1px solid rgba(108, 82, 56, 0.12)"
} as const;

const rowStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "10px",
  alignItems: "flex-start"
} as const;

const pathStyle = {
  display: "block",
  padding: "8px 10px",
  borderRadius: "12px",
  background: "rgba(44, 33, 19, 0.05)",
  color: "#4b3f33",
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
  fontSize: "0.82rem",
  overflowX: "auto" as const
} as const;

const reasonStyle = {
  margin: 0,
  color: "#5f5347"
} as const;

const eyebrowStyle = {
  margin: 0,
  textTransform: "uppercase" as const,
  letterSpacing: "0.12em",
  fontSize: "0.72rem",
  color: "#be7b2d"
} as const;

function badgeStyle(urgency: RunbookDockItem["urgency"]) {
  const background =
    urgency === "now"
      ? "rgba(191, 77, 48, 0.13)"
      : urgency === "soon"
        ? "rgba(214, 161, 81, 0.15)"
        : "rgba(92, 118, 143, 0.11)";
  const color = urgency === "now" ? "#8f311f" : urgency === "soon" ? "#7b541f" : "#496073";
  return {
    display: "inline-flex",
    alignItems: "center",
    padding: "5px 8px",
    borderRadius: "999px",
    background,
    color,
    fontSize: "0.72rem",
    textTransform: "uppercase" as const,
    letterSpacing: "0.08em"
  } as const;
}
