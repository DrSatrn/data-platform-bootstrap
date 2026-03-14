// SystemPage gathers platform diagnostics so operators can quickly understand
// whether the stack is healthy and whether data trust signals need attention.
import { useSystemData } from "../features/system/useSystemData";

export function SystemPage() {
  const { health, quality, error } = useSystemData();

  if (error) {
    return <section className="panel">System error: {error}</section>;
  }

  return (
    <section className="page-grid">
      <article className="card">
        <h2>Service Health</h2>
        <p className="muted">Environment: {health?.environment ?? "unknown"}</p>
        <p className="muted">API: {health?.http_addr ?? "unknown"}</p>
        <p className="muted">Web: {health?.web_addr ?? "unknown"}</p>
        <p>Status: {health?.status ?? "unknown"}</p>
      </article>
      <article className="card wide-card">
        <h2>Quality Signals</h2>
        <div className="stack">
          {(quality?.checks ?? []).map((check) => (
            <div className="subcard" key={check.id}>
              <div className="row-between">
                <strong>{check.name}</strong>
                <span className="badge">{check.status}</span>
              </div>
              <p className="muted">{check.message}</p>
            </div>
          ))}
        </div>
      </article>
    </section>
  );
}
