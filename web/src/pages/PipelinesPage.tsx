// PipelinesPage shows pipeline definitions and recent run state in one place.
// That tight coupling is useful in an internal platform because operators
// usually need both the static shape and the latest execution outcome together.
import { useState } from "react";

import { useAuth } from "../features/auth/useAuth";
import { usePipelines } from "../features/pipelines/usePipelines";
import { useRunArtifacts } from "../features/pipelines/useRunArtifacts";

export function PipelinesPage() {
  const { session } = useAuth();
  const { data, error, pendingPipelineID, triggerPipeline } = usePipelines();
  const [selectedRunID, setSelectedRunID] = useState<string | null>(null);
  const { artifacts, error: artifactError } = useRunArtifacts(selectedRunID);

  if (error) {
    return <section className="panel">Pipelines error: {error}</section>;
  }

  return (
    <section className="page-grid">
      <div className="card wide-card">
        <h2>Pipelines</h2>
        {!session?.capabilities.trigger_runs ? (
          <p className="muted">Editor token required to queue manual runs from the UI.</p>
        ) : null}
        <div className="stack">
          {(data?.pipelines ?? []).map((pipeline) => (
            <article className="subcard" key={pipeline.id}>
              <div className="row-between">
                <div>
                  <h3>{pipeline.name}</h3>
                  <p>{pipeline.description}</p>
                </div>
                <div className="inline-actions">
                  <span className="badge">{pipeline.owner}</span>
                  <button
                    className="mini-button"
                    disabled={pendingPipelineID === pipeline.id || !session?.capabilities.trigger_runs}
                    onClick={() => void triggerPipeline(pipeline.id)}
                    type="button"
                  >
                    {pendingPipelineID === pipeline.id ? "Queueing..." : "Run now"}
                  </button>
                </div>
              </div>
              <p className="muted">Jobs: {pipeline.jobs.map((job) => `${job.id} (${job.type})`).join(", ")}</p>
            </article>
          ))}
        </div>
      </div>
      <div className="card">
        <h2>Recent Runs</h2>
        <div className="stack">
          {(data?.runs ?? []).length === 0 ? (
            <p className="muted">No runs have been recorded yet.</p>
          ) : (
            data?.runs.map((run) => (
              <article className="subcard" key={run.id}>
                <div className="row-between">
                  <div>
                    <strong>{run.pipeline_id}</strong>
                    <p className="muted">Trigger: {run.trigger}</p>
                    <p className="muted">Updated: {new Date(run.updated_at).toLocaleString()}</p>
                  </div>
                  <div className="inline-actions">
                    <span className="badge">{run.status}</span>
                    <button className="mini-button" onClick={() => setSelectedRunID(run.id)} type="button">
                      Artifacts
                    </button>
                  </div>
                </div>
                <p className="muted">
                  Jobs: {run.job_runs.map((job) => `${job.job_id}=${job.status}`).join(", ")}
                </p>
                <p className="muted">
                  Events: {run.events.slice(-2).map((event) => `${event.level}: ${event.message}`).join(" | ")}
                </p>
              </article>
            ))
          )}
        </div>
      </div>
      <div className="card wide-card">
        <h2>Run Artifacts</h2>
        <p className="muted">
          {selectedRunID ? `Showing artifacts for ${selectedRunID}` : "Select a recent run to inspect worker outputs."}
        </p>
        {artifactError ? <p className="muted">Artifact error: {artifactError}</p> : null}
        <div className="stack">
          {artifacts.length === 0 ? (
            <p className="muted">No artifacts loaded.</p>
          ) : (
            artifacts.map((artifact) => (
              <div className="subcard" key={artifact.relative_path}>
                <div className="row-between">
                  <strong>{artifact.relative_path}</strong>
                  <span className="badge">{Math.max(1, Math.round(artifact.size_bytes / 1024))} KB</span>
                </div>
                <p className="muted">
                  Updated: {new Date(artifact.modified_at).toLocaleString()} | Type: {artifact.content_type}
                </p>
                <a
                  className="artifact-link"
                  href={`/api/v1/artifacts?run_id=${encodeURIComponent(artifact.run_id)}&path=${encodeURIComponent(artifact.relative_path)}`}
                  rel="noreferrer"
                  target="_blank"
                >
                  Open artifact
                </a>
              </div>
            ))
          )}
        </div>
      </div>
    </section>
  );
}
