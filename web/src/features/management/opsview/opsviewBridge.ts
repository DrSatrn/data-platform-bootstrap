export type OpsviewArtifact = {
  run_id: string;
  relative_path: string;
  size_bytes?: number;
  modified_at?: string;
  content_type?: string;
};

export type OpsviewRunEvent = {
  time: string;
  level: string;
  message: string;
  fields?: Record<string, string>;
};

export type OpsviewEvidenceSummary = {
  run_id: string;
  total_artifacts: number;
  log_artifact_count: number;
  output_artifact_count: number;
  artifact_paths: string[];
  log_paths: string[];
  output_paths: string[];
};

export type OpsviewExternalToolRunSummary = {
  run_id: string;
  pipeline_id: string;
  job_id: string;
  tool: string;
  action: string;
  status: "failed" | "running" | "succeeded" | "unknown";
  failure_class?: string;
  last_event_at?: string;
  events: OpsviewRunEvent[];
  log_artifacts: OpsviewArtifact[];
  output_artifacts: OpsviewArtifact[];
  evidence: OpsviewEvidenceSummary;
};

export type OpsviewAttentionSummary = {
  total_jobs: number;
  failed_jobs: number;
  running_jobs: number;
  succeeded_jobs: number;
  jobs_missing_logs: number;
  jobs_missing_outputs: number;
};

export type OperatorSignalCard = {
  id: string;
  label: string;
  tone: "critical" | "warning" | "healthy" | "neutral";
  value: string;
  detail: string;
};

export function buildOpsviewSignalCards(
  attention: OpsviewAttentionSummary,
  summaries: OpsviewExternalToolRunSummary[]
): OperatorSignalCard[] {
  const mostRecentFailure = summaries.find((summary) => summary.status === "failed");

  return [
    {
      id: "opsview-total-jobs",
      label: "Tracked Jobs",
      tone: attention.total_jobs === 0 ? "neutral" : "healthy",
      value: String(attention.total_jobs),
      detail: "External-tool jobs summarized by the backend read-model layer."
    },
    {
      id: "opsview-failed-jobs",
      label: "Failed Jobs",
      tone: attention.failed_jobs > 0 ? "critical" : "healthy",
      value: String(attention.failed_jobs),
      detail: mostRecentFailure
        ? `${mostRecentFailure.job_id} failed${mostRecentFailure.failure_class ? ` (${mostRecentFailure.failure_class})` : ""}.`
        : "No failed external-tool jobs in this snapshot."
    },
    {
      id: "opsview-missing-logs",
      label: "Missing Logs",
      tone: attention.jobs_missing_logs > 0 ? "warning" : "healthy",
      value: String(attention.jobs_missing_logs),
      detail: "Jobs without stdout/stderr evidence are harder to diagnose."
    },
    {
      id: "opsview-missing-outputs",
      label: "Missing Outputs",
      tone: attention.jobs_missing_outputs > 0 ? "warning" : "healthy",
      value: String(attention.jobs_missing_outputs),
      detail: "Jobs without declared outputs need operator follow-up before trust."
    }
  ];
}

export function flattenOpsviewEvidence(summaries: OpsviewExternalToolRunSummary[]) {
  return summaries.flatMap((summary) =>
    summary.evidence.artifact_paths.map((path) => ({
      jobID: summary.job_id,
      runID: summary.run_id,
      status: summary.status,
      failureClass: summary.failure_class,
      path,
      category: summary.evidence.log_paths.includes(path) ? "log" : "output"
    }))
  );
}
