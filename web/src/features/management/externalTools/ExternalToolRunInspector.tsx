import { buildExternalToolRunSummaries, type ExternalToolArtifact, type ExternalToolEvent } from "./externalToolRunSummary";

type Props = {
  events: ExternalToolEvent[];
  artifacts: ExternalToolArtifact[];
};

export function ExternalToolRunInspector({ events, artifacts }: Props) {
  const summaries = buildExternalToolRunSummaries(events, artifacts);

  if (summaries.length === 0) {
    return (
      <section>
        <h2>External Tool Runs</h2>
        <p>No external tool activity recorded.</p>
      </section>
    );
  }

  return (
    <section>
      <h2>External Tool Runs</h2>
      {summaries.map((summary) => (
        <article key={summary.jobID}>
          <h3>{summary.jobID}</h3>
          <p>
            {summary.tool} {summary.action || "run"} · {summary.status}
          </p>
          <p>{summary.events.length} lifecycle events</p>
          <p>{summary.logArtifacts.length} log artifacts</p>
          <p>{summary.outputArtifacts.length} declared outputs</p>
          <ul>
            {summary.logArtifacts.map((artifact) => (
              <li key={artifact.relative_path}>{artifact.relative_path}</li>
            ))}
            {summary.outputArtifacts.map((artifact) => (
              <li key={artifact.relative_path}>{artifact.relative_path}</li>
            ))}
          </ul>
        </article>
      ))}
    </section>
  );
}
