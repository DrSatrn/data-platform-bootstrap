export type RunbookDockItem = {
  id: string;
  label: string;
  path: string;
  reason: string;
  urgency: "now" | "soon" | "reference";
};

type SourceSession = {
  id: string;
  title: string;
  status: "ready" | "running" | "waiting" | "failed" | "completed";
  recommendedRunbook?: string;
};

type SourceFollowup = {
  id: string;
  title: string;
  priority: "now" | "soon" | "watch";
  rationale: string;
  runbookPath?: string;
};

export function buildRunbookDock(sessions: SourceSession[], followups: SourceFollowup[]): RunbookDockItem[] {
  const items = new Map<string, RunbookDockItem>();

  for (const session of sessions) {
    if (!session.recommendedRunbook) {
      continue;
    }
    const urgency = session.status === "failed" ? "now" : session.status === "completed" ? "soon" : "reference";
    items.set(session.recommendedRunbook, {
      id: `${session.id}:runbook`,
      label: session.title,
      path: session.recommendedRunbook,
      reason: `Linked from terminal session ${session.title}.`,
      urgency
    });
  }

  for (const followup of followups) {
    if (!followup.runbookPath) {
      continue;
    }
    const urgency = followup.priority === "watch" ? "reference" : followup.priority;
    const existing = items.get(followup.runbookPath);
    if (existing) {
      if (rankUrgency(urgency) < rankUrgency(existing.urgency)) {
        existing.urgency = urgency;
      }
      existing.reason = `${existing.reason} ${followup.title}: ${followup.rationale}`;
      continue;
    }
    items.set(followup.runbookPath, {
      id: followup.id,
      label: followup.title,
      path: followup.runbookPath,
      reason: followup.rationale,
      urgency
    });
  }

  return [...items.values()].sort((left, right) => rankUrgency(left.urgency) - rankUrgency(right.urgency));
}

function rankUrgency(value: RunbookDockItem["urgency"]) {
  switch (value) {
    case "now":
      return 0;
    case "soon":
      return 1;
    default:
      return 2;
  }
}
