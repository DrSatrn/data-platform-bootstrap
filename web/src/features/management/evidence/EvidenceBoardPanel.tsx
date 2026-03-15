import type { EvidenceItem } from "./evidenceBoard";

type EvidenceBoardPanelProps = {
  items: EvidenceItem[];
};

export function EvidenceBoardPanel({ items }: EvidenceBoardPanelProps) {
  return (
    <article style={panelStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Evidence Board Draft</p>
          <h3 style={{ margin: "6px 0 0" }}>Operator Evidence</h3>
        </div>
      </div>

      <div style={listStyle}>
        {items.map((item) => (
          <div key={item.id} style={cardStyle}>
            <div style={rowStyle}>
              <strong>{item.title}</strong>
              <span style={badgeStyle(item.importance)}>{item.importance}</span>
            </div>
            <span style={metaStyle}>
              {item.source} · {item.kind}
            </span>
            <code style={pathStyle}>{item.path}</code>
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

const metaStyle = {
  color: "#7b6b5a",
  fontSize: "0.8rem"
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

const eyebrowStyle = {
  margin: 0,
  textTransform: "uppercase" as const,
  letterSpacing: "0.12em",
  fontSize: "0.72rem",
  color: "#be7b2d"
} as const;

function badgeStyle(importance: EvidenceItem["importance"]) {
  const background =
    importance === "high"
      ? "rgba(191, 77, 48, 0.13)"
      : importance === "medium"
        ? "rgba(214, 161, 81, 0.15)"
        : "rgba(92, 118, 143, 0.11)";
  const color = importance === "high" ? "#8f311f" : importance === "medium" ? "#7b541f" : "#496073";
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
