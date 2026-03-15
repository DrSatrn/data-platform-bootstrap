export type TerminalEntryKind = "command" | "stdout" | "stderr" | "system";

export type TerminalSessionStatus = "ready" | "running" | "waiting" | "failed" | "completed";

export type TerminalTranscriptEntry = {
  id: string;
  kind: TerminalEntryKind;
  content: string;
  at: string;
};

export type TerminalSession = {
  id: string;
  title: string;
  scope: string;
  status: TerminalSessionStatus;
  currentCommand: string;
  lastExitCode?: number;
  pinnedArtifacts: string[];
  recommendedRunbook?: string;
  entries: TerminalTranscriptEntry[];
};

export function createTerminalSession(input: {
  id: string;
  title: string;
  scope: string;
  currentCommand?: string;
  status?: TerminalSessionStatus;
  entries?: TerminalTranscriptEntry[];
  pinnedArtifacts?: string[];
  recommendedRunbook?: string;
}): TerminalSession {
  return {
    id: input.id,
    title: input.title,
    scope: input.scope,
    status: input.status ?? "ready",
    currentCommand: input.currentCommand ?? "",
    pinnedArtifacts: input.pinnedArtifacts ?? [],
    recommendedRunbook: input.recommendedRunbook,
    entries: input.entries ?? []
  };
}

export function startCommand(session: TerminalSession, command: string, at: string): TerminalSession {
  return {
    ...session,
    status: "running",
    currentCommand: command,
    entries: [
      ...session.entries,
      {
        id: `${session.id}:${session.entries.length + 1}`,
        kind: "command",
        content: command,
        at
      }
    ]
  };
}

export function appendOutput(
  session: TerminalSession,
  kind: Exclude<TerminalEntryKind, "command">,
  content: string,
  at: string
): TerminalSession {
  return {
    ...session,
    entries: [
      ...session.entries,
      {
        id: `${session.id}:${session.entries.length + 1}`,
        kind,
        content,
        at
      }
    ]
  };
}

export function completeCommand(
  session: TerminalSession,
  input: {
    exitCode: number;
    at: string;
    artifacts?: string[];
    stdout?: string[];
    stderr?: string[];
  }
): TerminalSession {
  let next = session;

  for (const line of input.stdout ?? []) {
    next = appendOutput(next, "stdout", line, input.at);
  }
  for (const line of input.stderr ?? []) {
    next = appendOutput(next, "stderr", line, input.at);
  }

  next = appendOutput(
    next,
    "system",
    input.exitCode === 0 ? "command completed successfully" : `command failed with exit code ${input.exitCode}`,
    input.at
  );

  return {
    ...next,
    status: input.exitCode === 0 ? "completed" : "failed",
    lastExitCode: input.exitCode,
    pinnedArtifacts: dedupeStrings([...next.pinnedArtifacts, ...(input.artifacts ?? [])])
  };
}

export function sessionNeedsAttention(session: TerminalSession): boolean {
  if (session.status === "failed") {
    return true;
  }
  return session.pinnedArtifacts.length === 0 && session.status === "completed";
}

export function summarizeSession(session: TerminalSession): string {
  const lastLine = [...session.entries].reverse().find((entry) => entry.kind !== "command");
  if (lastLine) {
    return lastLine.content;
  }
  if (session.currentCommand) {
    return `ready to run ${session.currentCommand}`;
  }
  return "no command executed yet";
}

function dedupeStrings(values: string[]) {
  return [...new Set(values)];
}
