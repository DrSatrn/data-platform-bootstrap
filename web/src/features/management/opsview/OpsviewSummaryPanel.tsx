import { buildOpsviewSignalCards, flattenOpsviewEvidence, type OpsviewAttentionSummary, type OpsviewExternalToolRunSummary } from "./opsviewBridge";

type OpsviewSummaryPanelProps = {
  attention: OpsviewAttentionSummary;
  summaries: OpsviewExternalToolRunSummary[];
};

export function OpsviewSummaryPanel({ attention, summaries }: OpsviewSummaryPanelProps) {
  const cards = buildOpsviewSignalCards(attention, summaries);
  const evidence = flattenOpsviewEvidence(summaries).slice(0, 4);

  return (
    <section style={panelStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Opsview Bridge Draft</p>
          <h3 style={{ margin: "6px 0 0" }}>Backend Read-Model Snapshot</h3>
        </div>
      </div>

      <div style={cardGridStyle}>
        {cards.map((card) => (
          <div key={card.id} style={signalCardStyle(card.tone)}>
            <span style={labelStyle}>{card.label}</span>
            <strong style={valueStyle}>{card.value}</strong>
            <p style={detailStyle}>{card.detail}</p>
          </div>
        ))}
      </div>

      <div style={evidenceListStyle}>
        {evidence.map((item) => (
          <div key={`${item.jobID}:${item.path}`} style={evidenceItemStyle}>
            <strong>{item.jobID}</strong>
            <span style={metaStyle}>
              {item.status} · {item.category}
            </span>
            <code style={pathStyle}>{item.path}</code>
          </div>
        ))}
      </div>
    </section>
  );
}

const panelStyle = {
  display: "grid",
  gap: "16px",
  padding: "20px",
  borderRadius: "24px",
  background: "rgba(255, 254, 250, 0.86)",
  border: "1px solid rgba(94, 72, 51, 0.14)"
} as const;

const headerStyle = {
  display: "grid",
  gap: "4px"
} as const;

const cardGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
  gap: "12px"
} as const;

const labelStyle = {
  color: "#7a6856",
  fontSize: "0.8rem",
  textTransform: "uppercase" as const,
  letterSpacing: "0.08em"
} as const;

const valueStyle = {
  fontSize: "1.5rem"
} as const;

const detailStyle = {
  margin: 0,
  color: "#5c5044"
} as const;

const evidenceListStyle = {
  display: "grid",
  gap: "10px"
} as const;

const evidenceItemStyle = {
  display: "grid",
  gap: "6px",
  padding: "14px",
  borderRadius: "16px",
  background: "rgba(248, 244, 236, 0.95)",
  border: "1px solid rgba(94, 72, 51, 0.12)"
} as const;

const metaStyle = {
  color: "#7b6a58",
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

function signalCardStyle(tone: "critical" | "warning" | "healthy" | "neutral") {
  const border =
    tone === "critical"
      ? "rgba(191, 77, 48, 0.2)"
      : tone === "warning"
        ? "rgba(214, 161, 81, 0.22)"
        : tone === "healthy"
          ? "rgba(94, 146, 98, 0.2)"
          : "rgba(98, 112, 126, 0.18)";
  return {
    display: "grid",
    gap: "8px",
    padding: "14px",
    borderRadius: "16px",
    background: "rgba(255, 255, 255, 0.85)",
    border: `1px solid ${border}`
  } as const;
}
