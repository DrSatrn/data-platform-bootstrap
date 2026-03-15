import { buildExternalToolRunSummaries, type ExternalToolArtifact, type ExternalToolEvent } from "./externalToolRunSummary";
import type { OpsviewExternalToolRunSummary } from "../opsview/opsviewBridge";

type Props = {
  events?: ExternalToolEvent[];
  artifacts?: ExternalToolArtifact[];
  summaries?: OpsviewExternalToolRunSummary[];
};

export function ExternalToolRunInspector({ events = [], artifacts = [], summaries }: Props) {
  const resolvedSummaries =
    summaries ??
    buildExternalToolRunSummaries(events, artifacts).map((summary) => ({
      run_id: "unknown",
      pipeline_id: "unknown",
      job_id: summary.jobID,
      tool: summary.tool,
      action: summary.action,
      status: summary.status,
      failure_class: undefined,
      events: summary.events,
      log_artifacts: summary.logArtifacts,
      output_artifacts: summary.outputArtifacts,
      evidence: {
        run_id: "unknown",
        total_artifacts: summary.logArtifacts.length + summary.outputArtifacts.length,
        log_artifact_count: summary.logArtifacts.length,
        output_artifact_count: summary.outputArtifacts.length,
        artifact_paths: [...summary.logArtifacts.map((artifact) => artifact.relative_path), ...summary.outputArtifacts.map((artifact) => artifact.relative_path)],
        log_paths: summary.logArtifacts.map((artifact) => artifact.relative_path),
        output_paths: summary.outputArtifacts.map((artifact) => artifact.relative_path)
      }
    }));

  if (resolvedSummaries.length === 0) {
    return (
      <section>
        <h2>External Tool Runs</h2>
        <p>No external tool activity recorded in the current run snapshot window.</p>
      </section>
    );
  }

  return (
    <section>
      <h2>External Tool Runs</h2>
      {resolvedSummaries.map((summary) => (
        <article key={`${summary.run_id}:${summary.job_id}`}>
          <h3>{summary.job_id}</h3>
          <p>
            {summary.tool} {summary.action || "run"} · {summary.status}
          </p>
          <p>run: {summary.run_id} · pipeline: {summary.pipeline_id}</p>
          {summary.failure_class ? <p>failure class: {summary.failure_class}</p> : null}
          <p>{summary.events.length} lifecycle events</p>
          <p>{summary.log_artifacts.length} log artifacts</p>
          <p>{summary.output_artifacts.length} declared outputs</p>
          <ul>
            {summary.log_artifacts.map((artifact) => (
              <li key={artifact.relative_path}>{artifact.relative_path}</li>
            ))}
            {summary.output_artifacts.map((artifact) => (
              <li key={artifact.relative_path}>{artifact.relative_path}</li>
            ))}
          </ul>
        </article>
      ))}
    </section>
  );
}
