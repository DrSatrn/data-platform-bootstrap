// PipelinesPage shows pipeline definitions and recent run state in one place.
// That tight coupling is useful in an internal platform because operators
// usually need both the static shape and the latest execution outcome together.
import { usePipelines } from "../features/pipelines/usePipelines";

export function PipelinesPage() {
  const { data, error } = usePipelines();

  if (error) {
    return <section className="panel">Pipelines error: {error}</section>;
  }

  return (
    <section className="page-grid">
      <div className="card wide-card">
        <h2>Pipelines</h2>
        <div className="stack">
          {(data?.pipelines ?? []).map((pipeline) => (
            <article className="subcard" key={pipeline.id}>
              <div className="row-between">
                <div>
                  <h3>{pipeline.name}</h3>
                  <p>{pipeline.description}</p>
                </div>
                <span className="badge">{pipeline.owner}</span>
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
              <div className="row-between" key={run.id}>
                <span>{run.pipeline_id}</span>
                <span className="badge">{run.status}</span>
              </div>
            ))
          )}
        </div>
      </div>
    </section>
  );
}
