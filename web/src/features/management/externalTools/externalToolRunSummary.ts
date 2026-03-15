export type ExternalToolEvent = {
  time: string;
  level: string;
  message: string;
  fields?: Record<string, string>;
};

export type ExternalToolArtifact = {
  run_id: string;
  relative_path: string;
  size_bytes: number;
  modified_at: string;
  content_type: string;
};

export type ExternalToolRunSummary = {
  jobID: string;
  tool: string;
  action: string;
  status: "succeeded" | "failed" | "running" | "unknown";
  events: ExternalToolEvent[];
  logArtifacts: ExternalToolArtifact[];
  outputArtifacts: ExternalToolArtifact[];
};

export function buildExternalToolRunSummaries(
  events: ExternalToolEvent[],
  artifacts: ExternalToolArtifact[]
): ExternalToolRunSummary[] {
  const summaries = new Map<string, ExternalToolRunSummary>();

  for (const event of events) {
    const jobID = event.fields?.job_id;
    const tool = event.fields?.tool;
    if (!jobID || !tool) {
      continue;
    }
    const summary = getOrCreateSummary(summaries, jobID, tool, event.fields?.action ?? "");
    if (event.fields?.action) {
      summary.action = event.fields.action;
    }
    summary.events.push(event);
  }

  for (const artifact of artifacts) {
    const parts = artifact.relative_path.split("/");
    if (parts[0] !== "external_tools" || parts.length < 3) {
      continue;
    }
    const jobID = parts[1];
    const summary = getOrCreateSummary(summaries, jobID, "unknown", "");
    if (parts[2] === "logs") {
      summary.logArtifacts.push(artifact);
      continue;
    }
    summary.outputArtifacts.push(artifact);
  }

  return Array.from(summaries.values())
    .map((summary) => ({
      ...summary,
      status: summarizeStatus(summary.events)
    }))
    .sort((left, right) => left.jobID.localeCompare(right.jobID));
}

function getOrCreateSummary(
  summaries: Map<string, ExternalToolRunSummary>,
  jobID: string,
  tool: string,
  action: string
): ExternalToolRunSummary {
  const existing = summaries.get(jobID);
  if (existing) {
    if (existing.tool === "unknown" && tool) {
      existing.tool = tool;
    }
    if (!existing.action && action) {
      existing.action = action;
    }
    return existing;
  }

  const created: ExternalToolRunSummary = {
    jobID,
    tool: tool || "unknown",
    action,
    status: "unknown",
    events: [],
    logArtifacts: [],
    outputArtifacts: []
  };
  summaries.set(jobID, created);
  return created;
}

function summarizeStatus(events: ExternalToolEvent[]): ExternalToolRunSummary["status"] {
  if (events.some((event) => event.level === "error" || event.message.includes("failed"))) {
    return "failed";
  }
  if (events.some((event) => event.message.includes("finished"))) {
    return "succeeded";
  }
  if (events.length > 0) {
    return "running";
  }
  return "unknown";
}
