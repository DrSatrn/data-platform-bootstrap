// SystemPage gathers platform diagnostics so operators can quickly understand
// whether the stack is healthy and whether data trust signals need attention.
import { AdminTerminal } from "../components/AdminTerminal";
import { useAuth } from "../features/auth/useAuth";
import { useSystemData } from "../features/system/useSystemData";

export function SystemPage() {
  const { session } = useAuth();
  const { health, quality, overview, logs, catalog, error } = useSystemData();

  if (error) {
    return <section className="panel">System error: {error}</section>;
  }

  const freshnessStates = (catalog?.assets ?? []).reduce(
    (accumulator, asset) => {
      accumulator[asset.freshness_status.state] = (accumulator[asset.freshness_status.state] ?? 0) + 1;
      return accumulator;
    },
    {} as Record<string, number>
  );

  return (
    <section className="page-grid">
      <article className="card">
        <h2>Service Health</h2>
        <p className="muted">Environment: {health?.environment ?? "unknown"}</p>
        <p>Status: {health?.status ?? "unknown"}</p>
        <p className="muted">Session: {session?.principal.subject ?? "anonymous"} ({session?.principal.role ?? "anonymous"})</p>
        <p className="muted">Known pipelines: {overview?.known_pipelines ?? 0}</p>
        <p className="muted">Known assets: {overview?.known_assets ?? 0}</p>
        <p className="muted">Run history: {overview?.run_history ?? 0}</p>
      </article>
      <article className="card">
        <h2>Built-in Metrics</h2>
        <p className="muted">Uptime: {overview?.telemetry.uptime_seconds ?? 0}s</p>
        <p className="muted">Requests: {overview?.telemetry.total_requests ?? 0}</p>
        <p className="muted">Errors: {overview?.telemetry.total_errors ?? 0}</p>
        <p className="muted">Admin commands: {overview?.telemetry.total_commands ?? 0}</p>
        <p className="muted">Fresh assets: {freshnessStates.fresh ?? 0}</p>
        <p className="muted">Late assets: {freshnessStates.late ?? 0}</p>
        <p className="muted">Stale assets: {freshnessStates.stale ?? 0}</p>
      </article>
      <article className="card">
        <h2>Catalog Trust</h2>
        <p className="muted">Missing docs: {catalog?.summary.assets_missing_docs ?? 0}</p>
        <p className="muted">Missing quality: {catalog?.summary.assets_missing_quality ?? 0}</p>
        <p className="muted">Assets with PII: {catalog?.summary.assets_containing_pii ?? 0}</p>
        <p className="muted">Lineage edges: {catalog?.summary.lineage_edges ?? 0}</p>
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
      <article className="card wide-card">
        <h2>Recent Platform Logs</h2>
        <div className="stack">
          {(logs?.logs ?? []).slice(-8).map((entry, index) => (
            <div className="subcard" key={`${entry.time}-${index}`}>
              <div className="row-between">
                <strong>{entry.level}</strong>
                <span className="badge">{new Date(entry.time).toLocaleTimeString()}</span>
              </div>
              <p className="muted">{entry.message}</p>
            </div>
          ))}
        </div>
      </article>
      <AdminTerminal />
    </section>
  );
}
