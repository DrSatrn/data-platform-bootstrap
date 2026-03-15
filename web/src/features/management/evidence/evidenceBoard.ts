export type EvidenceItem = {
  id: string;
  title: string;
  kind: "log" | "output" | "report";
  path: string;
  source: string;
  importance: "high" | "medium" | "low";
};

type SourceSession = {
  id: string;
  title: string;
  status: "ready" | "running" | "waiting" | "failed" | "completed";
  pinnedArtifacts: string[];
};

type SourceInspector = {
  jobID: string;
  status: "succeeded" | "failed" | "running" | "unknown";
  logArtifacts: Array<{ relative_path: string }>;
  outputArtifacts: Array<{ relative_path: string }>;
};

export function buildEvidenceBoard(
  sessions: SourceSession[],
  inspectorSummaries: SourceInspector[]
): EvidenceItem[] {
  const items: EvidenceItem[] = [];

  for (const session of sessions) {
    for (const path of session.pinnedArtifacts) {
      items.push({
        id: `${session.id}:${path}`,
        title: session.title,
        kind: classifyPath(path),
        path,
        source: "terminal_session",
        importance: session.status === "failed" ? "high" : "medium"
      });
    }
  }

  for (const summary of inspectorSummaries) {
    for (const artifact of summary.logArtifacts) {
      items.push({
        id: `${summary.jobID}:${artifact.relative_path}`,
        title: `${summary.jobID} logs`,
        kind: "log",
        path: artifact.relative_path,
        source: "external_tool",
        importance: summary.status === "failed" ? "high" : "medium"
      });
    }
    for (const artifact of summary.outputArtifacts) {
      items.push({
        id: `${summary.jobID}:${artifact.relative_path}`,
        title: `${summary.jobID} outputs`,
        kind: classifyPath(artifact.relative_path),
        path: artifact.relative_path,
        source: "external_tool",
        importance: summary.status === "failed" ? "medium" : "low"
      });
    }
  }

  return dedupeByPath(items).sort((left, right) => rankImportance(left.importance) - rankImportance(right.importance));
}

function classifyPath(path: string): EvidenceItem["kind"] {
  if (path.endsWith(".log")) {
    return "log";
  }
  if (path.endsWith(".json") && path.includes("report")) {
    return "report";
  }
  return "output";
}

function rankImportance(value: EvidenceItem["importance"]) {
  switch (value) {
    case "high":
      return 0;
    case "medium":
      return 1;
    default:
      return 2;
  }
}

function dedupeByPath(items: EvidenceItem[]) {
  const seen = new Set<string>();
  return items.filter((item) => {
    if (seen.has(item.path)) {
      return false;
    }
    seen.add(item.path);
    return true;
  });
}
