import { useMemo, useState } from "react";

import { defaultCommandCatalog, mockRunbooks } from "../mockControlPlane";
import type { OperatorRole, TerminalCommandTemplate } from "../types";
import { filterCommands, groupCommandsByArea, topCommandSuggestions } from "./commandCatalog";

type OperatorWorkbenchProps = {
  role?: OperatorRole;
  recentCommands?: string[];
  onCommandSelect?: (command: TerminalCommandTemplate) => void;
};

export function OperatorWorkbench({
  role = "admin",
  recentCommands = ["status", "quality"],
  onCommandSelect
}: OperatorWorkbenchProps) {
  const [query, setQuery] = useState("");
  const visibleCommands = useMemo(() => filterCommands(defaultCommandCatalog, query), [query]);
  const grouped = useMemo(() => groupCommandsByArea(visibleCommands), [visibleCommands]);
  const suggestions = useMemo(
    () => topCommandSuggestions(defaultCommandCatalog, recentCommands, role, 4),
    [recentCommands, role]
  );

  return (
    <section style={shellStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Management Console Draft</p>
          <h2 style={{ margin: "6px 0 10px" }}>Operator Workbench</h2>
          <p style={mutedStyle}>
            Unwired prototype for a richer in-app terminal and management surface.
          </p>
        </div>
        <span style={badgeStyle}>{role}</span>
      </div>

      <label style={searchLabelStyle}>
        <span style={mutedStyle}>Search commands</span>
        <input
          onChange={(event) => setQuery(event.target.value)}
          placeholder="search status, backups, quality, assets..."
          style={inputStyle}
          value={query}
        />
      </label>

      <div style={gridStyle}>
        <article style={cardStyle}>
          <h3 style={sectionTitleStyle}>Suggested Commands</h3>
          <div style={listStyle}>
            {suggestions.map((command) => (
              <button key={command.id} onClick={() => onCommandSelect?.(command)} style={buttonCardStyle} type="button">
                <strong>{command.label}</strong>
                <span style={mutedStyle}>{command.command}</span>
                <span style={miniBadgeStyle}>{command.area}</span>
              </button>
            ))}
          </div>
        </article>

        <article style={cardStyle}>
          <h3 style={sectionTitleStyle}>Runbook Shortcuts</h3>
          <div style={listStyle}>
            {mockRunbooks.map((runbook) => (
              <div key={runbook.id} style={infoCardStyle}>
                <strong>{runbook.label}</strong>
                <span style={mutedStyle}>{runbook.path}</span>
                <p style={{ margin: 0, color: "#574d44" }}>{runbook.reason}</p>
              </div>
            ))}
          </div>
        </article>
      </div>

      <div style={listStyle}>
        {grouped.map((group) => (
          <article key={group.area} style={cardStyle}>
            <div style={rowBetweenStyle}>
              <h3 style={sectionTitleStyle}>{titleCase(group.area)}</h3>
              <span style={miniBadgeStyle}>{group.entries.length}</span>
            </div>
            <div style={listStyle}>
              {group.entries.map((command) => (
                <button key={command.id} onClick={() => onCommandSelect?.(command)} style={buttonCardStyle} type="button">
                  <div style={rowBetweenStyle}>
                    <strong>{command.label}</strong>
                    <span style={miniBadgeStyle}>{command.minimumRole}</span>
                  </div>
                  <span style={mutedStyle}>{command.command}</span>
                  <p style={{ margin: 0, color: "#574d44" }}>{command.summary}</p>
                </button>
              ))}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}

function titleCase(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1);
}

const shellStyle = {
  display: "grid",
  gap: "18px",
  padding: "24px",
  borderRadius: "24px",
  border: "1px solid rgba(104, 77, 48, 0.18)",
  background:
    "linear-gradient(180deg, rgba(255, 249, 241, 0.98), rgba(247, 239, 226, 0.96))",
  boxShadow: "0 18px 42px rgba(42, 30, 18, 0.09)"
} as const;

const headerStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "16px",
  alignItems: "flex-start"
} as const;

const gridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
  gap: "16px"
} as const;

const cardStyle = {
  display: "grid",
  gap: "12px",
  padding: "18px",
  borderRadius: "20px",
  background: "rgba(255, 255, 255, 0.82)",
  border: "1px solid rgba(112, 86, 58, 0.15)"
} as const;

const buttonCardStyle = {
  display: "grid",
  gap: "6px",
  textAlign: "left" as const,
  padding: "14px",
  borderRadius: "16px",
  border: "1px solid rgba(112, 86, 58, 0.15)",
  background: "rgba(250, 246, 239, 0.95)",
  cursor: "pointer"
};

const infoCardStyle = {
  display: "grid",
  gap: "6px",
  padding: "14px",
  borderRadius: "16px",
  border: "1px solid rgba(112, 86, 58, 0.15)",
  background: "rgba(250, 246, 239, 0.95)"
} as const;

const inputStyle = {
  width: "100%",
  padding: "12px 14px",
  borderRadius: "14px",
  border: "1px solid rgba(112, 86, 58, 0.2)",
  background: "rgba(255, 255, 255, 0.85)"
} as const;

const listStyle = {
  display: "grid",
  gap: "10px"
} as const;

const rowBetweenStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "12px",
  alignItems: "flex-start"
} as const;

const badgeStyle = {
  display: "inline-flex",
  alignItems: "center",
  borderRadius: "999px",
  background: "rgba(216, 134, 50, 0.16)",
  color: "#7b4316",
  padding: "8px 12px",
  textTransform: "uppercase" as const,
  fontSize: "0.76rem",
  letterSpacing: "0.08em"
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

const mutedStyle = {
  color: "#6c6155"
} as const;

const eyebrowStyle = {
  margin: 0,
  textTransform: "uppercase" as const,
  letterSpacing: "0.12em",
  fontSize: "0.72rem",
  color: "#c17a31"
} as const;

const searchLabelStyle = {
  display: "grid",
  gap: "8px"
} as const;

const sectionTitleStyle = {
  margin: 0
} as const;
