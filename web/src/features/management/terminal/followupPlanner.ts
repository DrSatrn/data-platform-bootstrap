import type { TerminalSession } from "./sessionModel";

export type FollowupPriority = "now" | "soon" | "watch";

export type FollowupAction = {
  id: string;
  title: string;
  priority: FollowupPriority;
  rationale: string;
  suggestedCommand?: string;
  artifactPath?: string;
  runbookPath?: string;
};

export function planFollowupActions(session: TerminalSession): FollowupAction[] {
  const actions: FollowupAction[] = [];

  if (session.status === "failed") {
    actions.push({
      id: `${session.id}:inspect-failure`,
      title: "Inspect failure output",
      priority: "now",
      rationale: summarizeFailure(session),
      artifactPath: session.pinnedArtifacts[0],
      runbookPath: session.recommendedRunbook
    });
  }

  if (session.status === "failed" && session.scope === "pipelines") {
    actions.push({
      id: `${session.id}:rerun-targeted-command`,
      title: "Retry the targeted pipeline command",
      priority: "soon",
      rationale: "A failed pipeline-oriented terminal session should offer a narrow rerun before broad recovery work.",
      suggestedCommand: session.currentCommand,
      runbookPath: session.recommendedRunbook
    });
  }

  if (session.status === "completed" && session.scope === "recovery") {
    actions.push({
      id: `${session.id}:record-recovery-evidence`,
      title: "Promote recovery evidence",
      priority: "soon",
      rationale: "Successful recovery work should keep a visible verification artifact for later audit.",
      artifactPath: session.pinnedArtifacts[0],
      runbookPath: session.recommendedRunbook
    });
  }

  if (session.status === "completed" && session.pinnedArtifacts.length === 0) {
    actions.push({
      id: `${session.id}:capture-artifact`,
      title: "Promote a primary artifact",
      priority: "watch",
      rationale: "A completed operational command is easier to audit when at least one artifact is pinned to the session.",
      suggestedCommand: session.currentCommand
    });
  }

  if (session.recommendedRunbook) {
    actions.push({
      id: `${session.id}:open-runbook`,
      title: "Open the linked runbook",
      priority: session.status === "failed" ? "now" : "watch",
      rationale: "Terminal work should carry the matching operating guide so the next operator step is explicit.",
      runbookPath: session.recommendedRunbook
    });
  }

  return dedupeActions(actions);
}

export function prioritizeActions(actions: FollowupAction[]): FollowupAction[] {
  const rank: Record<FollowupPriority, number> = {
    now: 0,
    soon: 1,
    watch: 2
  };

  return [...actions].sort((left, right) => rank[left.priority] - rank[right.priority] || left.title.localeCompare(right.title));
}

function summarizeFailure(session: TerminalSession) {
  const latestFailureLine = [...session.entries]
    .reverse()
    .find((entry) => entry.kind === "stderr" || (entry.kind === "system" && entry.content.includes("failed")));

  if (latestFailureLine) {
    return latestFailureLine.content;
  }

  if (session.lastExitCode !== undefined) {
    return `Session exited with code ${session.lastExitCode}.`;
  }

  return "The session failed and needs operator review.";
}

function dedupeActions(actions: FollowupAction[]) {
  const seen = new Set<string>();
  return actions.filter((action) => {
    if (seen.has(action.id)) {
      return false;
    }
    seen.add(action.id);
    return true;
  });
}
