import { mockFollowupDecks } from "./mockFollowups";
import { ArtifactFollowupPanel } from "./ArtifactFollowupPanel";
import type { FollowupAction } from "./followupPlanner";

type FollowupDeck = {
  sessionID: string;
  sessionTitle: string;
  actions: FollowupAction[];
};

type NextStepBoardProps = {
  decks?: FollowupDeck[];
};

export function NextStepBoard({ decks = mockFollowupDecks }: NextStepBoardProps) {
  return (
    <section style={shellStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Next Step Draft</p>
          <h2 style={{ margin: "6px 0 8px" }}>Operator Follow-up Board</h2>
          <p style={mutedStyle}>
            Unwired staging surface for deciding what artifact, runbook, or command should come
            next after a terminal session completes or fails.
          </p>
        </div>
      </div>

      <div style={gridStyle}>
        {decks.map((deck) => (
          <div key={deck.sessionID} style={deckStyle}>
            <div style={deckHeaderStyle}>
              <strong>{deck.sessionTitle}</strong>
              <span style={countStyle}>{deck.actions.length} actions</span>
            </div>
            <ArtifactFollowupPanel actions={deck.actions} />
          </div>
        ))}
      </div>
    </section>
  );
}

const shellStyle = {
  display: "grid",
  gap: "18px",
  padding: "24px",
  borderRadius: "24px",
  border: "1px solid rgba(90, 67, 44, 0.16)",
  background: "linear-gradient(180deg, rgba(248, 243, 234, 0.98), rgba(239, 231, 216, 0.96))"
} as const;

const headerStyle = {
  display: "grid",
  gap: "6px"
} as const;

const gridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
  gap: "16px"
} as const;

const deckStyle = {
  display: "grid",
  gap: "10px"
} as const;

const deckHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "10px",
  alignItems: "center"
} as const;

const countStyle = {
  color: "#6f6255",
  fontSize: "0.82rem"
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
