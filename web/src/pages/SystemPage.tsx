// SystemPage gathers platform diagnostics so operators can quickly understand
// whether the stack is healthy and whether data trust signals need attention.
import { AdminTerminal } from "../components/AdminTerminal";
import { useAuth } from "../features/auth/useAuth";
import { useSystemData } from "../features/system/useSystemData";

export function SystemPage() {
  const { session } = useAuth();
  const { health, quality, overview, logs, audit, catalog, error, refreshing, refresh } = useSystemData();

  if (error) {
    return <section className="panel">System error: {error}</section>;
  }
  if (!health || !overview || !catalog) {
    return <section className="panel">Loading system view...</section>;
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
        <div className="row-between">
          <h2>Service Health</h2>
          <button className="mini-button" onClick={() => void refresh()} type="button">
            {refreshing ? "Refreshing..." : "Refresh"}
          </button>
        </div>
        <p className="muted">Environment: {health.environment}</p>
        <p>Status: {health.status}</p>
        <p className="muted">Session: {session?.principal.subject ?? "anonymous"} ({session?.principal.role ?? "anonymous"})</p>
        <p className="muted">Known pipelines: {overview.known_pipelines}</p>
        <p className="muted">Known assets: {overview.known_assets}</p>
        <p className="muted">Run history: {overview.run_history}</p>
      </article>
      <article className="card">
        <div className="row-between">
          <h2>Built-in Metrics</h2>
          <span className="badge">{refreshing ? "live refresh" : "10s polling"}</span>
        </div>
        <p className="muted">Uptime: {overview?.telemetry.uptime_seconds ?? 0}s</p>
        <p className="muted">Requests: {overview?.telemetry.total_requests ?? 0}</p>
        <p className="muted">Errors: {overview?.telemetry.total_errors ?? 0}</p>
        <p className="muted">Admin commands: {overview?.telemetry.total_commands ?? 0}</p>
        <p className="muted">Fresh assets: {freshnessStates.fresh ?? 0}</p>
        <p className="muted">Late assets: {freshnessStates.late ?? 0}</p>
        <p className="muted">Stale assets: {freshnessStates.stale ?? 0}</p>
      </article>
      <article className="card">
        <h2>Run Throughput</h2>
        <p className="muted">Total runs: {overview?.run_summary.total_runs ?? 0}</p>
        <p className="muted">Succeeded: {overview?.run_summary.succeeded_runs ?? 0}</p>
        <p className="muted">Failed: {overview?.run_summary.failed_runs ?? 0}</p>
        <p className="muted">Completed in 24h: {overview?.run_summary.completed_last_24_hours ?? 0}</p>
        <p className="muted">Failures in 24h: {overview?.run_summary.failed_last_24_hours ?? 0}</p>
        <p className="muted">Average duration: {overview?.run_summary.average_duration_seconds ?? 0}s</p>
      </article>
      <article className="card">
        <h2>Queue And Recovery</h2>
        <p className="muted">Queued: {overview?.queue_summary.queued ?? 0}</p>
        <p className="muted">Active: {overview?.queue_summary.active ?? 0}</p>
        <p className="muted">Completed: {overview?.queue_summary.completed ?? 0}</p>
        <p className="muted">Backups: {overview?.backup_summary.bundle_count ?? 0}</p>
        <p className="muted">Latest bundle bytes: {overview?.backup_summary.latest_bundle_bytes ?? 0}</p>
      </article>
      <article className="card">
        <h2>Catalog Trust</h2>
        <p className="muted">Missing docs: {catalog?.summary.assets_missing_docs ?? 0}</p>
        <p className="muted">Missing quality: {catalog?.summary.assets_missing_quality ?? 0}</p>
        <p className="muted">Assets with PII: {catalog?.summary.assets_containing_pii ?? 0}</p>
        <p className="muted">Lineage edges: {catalog?.summary.lineage_edges ?? 0}</p>
      </article>
      <article className="card wide-card">
        <h2>Failure Watch</h2>
        {overview?.run_summary.latest_failure_run_id ? (
          <div className="subcard">
            <div className="row-between">
              <strong>{overview.run_summary.latest_failure_run_id}</strong>
              <span className="badge">latest failure</span>
            </div>
            <p className="muted">{overview.run_summary.latest_failure_message || "No failure message captured."}</p>
          </div>
        ) : (
          <p className="muted">No failed runs are currently recorded in run history.</p>
        )}
        {overview?.backup_summary.latest_bundle_path ? (
          <p className="muted">Latest backup: {overview.backup_summary.latest_bundle_path}</p>
        ) : (
          <p className="muted">No backup bundles recorded yet.</p>
        )}
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
      <article className="card wide-card">
        <h2>Audit Trail</h2>
        <div className="stack">
          {(audit?.events ?? []).length === 0 ? (
            <p className="muted">No audit events have been recorded yet.</p>
          ) : (
            (audit?.events ?? []).map((event, index) => (
              <div className="subcard" key={`${event.time}-${event.action}-${index}`}>
                <div className="row-between">
                  <strong>{event.action}</strong>
                  <div className="inline-actions">
                    <span className="badge">{event.actor_role}</span>
                    <span className="badge">{event.outcome}</span>
                  </div>
                </div>
                <p className="muted">
                  {event.actor_subject} · {event.resource} · {new Date(event.time).toLocaleString()}
                </p>
              </div>
            ))
          )}
        </div>
      </article>
      <AdminTerminal />
    </section>
  );
}
