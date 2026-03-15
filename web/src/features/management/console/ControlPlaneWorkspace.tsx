import { mockAttentionAssets, mockQueue, mockServiceStatus } from "../mockControlPlane";
import { buildAttentionQueue, summarizeAttention } from "../inventory/assetAttention";
import type { AssetAttentionCard, QueueSnapshotCard, ServiceStatusCard } from "../types";

type ControlPlaneWorkspaceProps = {
  attentionAssets?: AssetAttentionCard[];
  queue?: QueueSnapshotCard[];
  services?: ServiceStatusCard[];
};

export function ControlPlaneWorkspace({
  attentionAssets = mockAttentionAssets,
  queue = mockQueue,
  services = mockServiceStatus
}: ControlPlaneWorkspaceProps) {
  const attention = summarizeAttention(attentionAssets);
  const orderedAttention = buildAttentionQueue(attentionAssets);

  return (
    <section style={workspaceStyle}>
      <div style={heroStyle}>
        <div>
          <p style={eyebrowStyle}>Future Management Surface</p>
          <h2 style={{ margin: "6px 0 10px" }}>Control Plane Workspace</h2>
          <p style={mutedStyle}>
            Unwired management layout draft for queue visibility, service posture, and dataset attention.
          </p>
        </div>
        <div style={statRowStyle}>
          <StatCard label="Late Assets" value={String(attention.late)} />
          <StatCard label="Missing Assets" value={String(attention.missing)} />
          <StatCard label="Queue Depth" value={String(queue.length)} />
        </div>
      </div>

      <div style={serviceGridStyle}>
        {services.map((service) => (
          <article key={service.id} style={serviceCardStyle(service.state)}>
            <div style={rowBetweenStyle}>
              <strong>{service.label}</strong>
              <span style={miniBadgeStyle}>{service.state}</span>
            </div>
            <p style={mutedStyle}>{service.detail}</p>
          </article>
        ))}
      </div>

      <div style={panelGridStyle}>
        <article style={panelStyle}>
          <div style={rowBetweenStyle}>
            <h3 style={{ margin: 0 }}>Run Queue</h3>
            <span style={miniBadgeStyle}>{queue.length} active</span>
          </div>
          <div style={listStyle}>
            {queue.map((run) => (
              <div key={run.runID} style={listItemStyle}>
                <div style={rowBetweenStyle}>
                  <strong>{run.pipelineID}</strong>
                  <span style={miniBadgeStyle}>{run.status}</span>
                </div>
                <span style={mutedStyle}>{run.runID}</span>
                <span style={mutedStyle}>
                  {run.trigger} · requested {new Date(run.requestedAt).toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        </article>

        <article style={panelStyle}>
          <div style={rowBetweenStyle}>
            <h3 style={{ margin: 0 }}>Attention Queue</h3>
            <span style={miniBadgeStyle}>{orderedAttention.length} tracked</span>
          </div>
          <div style={listStyle}>
            {orderedAttention.map((asset) => (
              <div key={asset.assetID} style={listItemStyle}>
                <div style={rowBetweenStyle}>
                  <strong>{asset.assetID}</strong>
                  <span style={miniBadgeStyle}>{asset.freshnessState}</span>
                </div>
                <span style={mutedStyle}>layer: {asset.layer}</span>
                <span style={mutedStyle}>
                  docs: {asset.hasDocs ? "yes" : "no"} · quality: {asset.hasQuality ? "yes" : "no"}
                </span>
              </div>
            ))}
          </div>
        </article>
      </div>
    </section>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div style={statCardStyle}>
      <span style={mutedStyle}>{label}</span>
      <strong style={{ fontSize: "1.6rem" }}>{value}</strong>
    </div>
  );
}

const workspaceStyle = {
  display: "grid",
  gap: "18px"
} as const;

const heroStyle = {
  display: "grid",
  gap: "16px",
  padding: "24px",
  borderRadius: "24px",
  border: "1px solid rgba(101, 74, 47, 0.16)",
  background:
    "linear-gradient(135deg, rgba(245, 203, 135, 0.24), rgba(255, 246, 231, 0.94))"
} as const;

const serviceGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
  gap: "14px"
} as const;

const panelGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
  gap: "16px"
} as const;

const panelStyle = {
  display: "grid",
  gap: "12px",
  padding: "20px",
  borderRadius: "20px",
  border: "1px solid rgba(101, 74, 47, 0.14)",
  background: "rgba(255, 252, 247, 0.9)"
} as const;

const statRowStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(3, minmax(0, 1fr))",
  gap: "12px"
} as const;

const statCardStyle = {
  display: "grid",
  gap: "8px",
  padding: "16px",
  borderRadius: "16px",
  background: "rgba(255, 255, 255, 0.75)",
  border: "1px solid rgba(101, 74, 47, 0.14)"
} as const;

const rowBetweenStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "12px",
  alignItems: "flex-start"
} as const;

const listStyle = {
  display: "grid",
  gap: "10px"
} as const;

const listItemStyle = {
  display: "grid",
  gap: "6px",
  padding: "14px",
  borderRadius: "16px",
  border: "1px solid rgba(101, 74, 47, 0.14)",
  background: "rgba(250, 246, 239, 0.95)"
} as const;

const miniBadgeStyle = {
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
  color: "#c17a31"
} as const;

const mutedStyle = {
  margin: 0,
  color: "#6c6155"
} as const;

function serviceCardStyle(state: "healthy" | "degraded" | "offline") {
  const borderColor =
    state === "healthy" ? "rgba(59, 131, 85, 0.22)" : state === "degraded" ? "rgba(188, 130, 46, 0.24)" : "rgba(176, 72, 61, 0.24)";
  return {
    display: "grid",
    gap: "10px",
    padding: "16px",
    borderRadius: "18px",
    border: `1px solid ${borderColor}`,
    background: "rgba(255, 252, 247, 0.9)"
  } as const;
}
