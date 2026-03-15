// DatasetsPage now behaves more like a metadata workbench than a flat list.
// Operators can scan the catalog quickly on the left, then inspect ownership,
// freshness, lineage references, and column documentation for the selected
// asset on the right.
import { Asset, AssetProfile, useDatasets } from "../features/datasets/useDatasets";

export function DatasetsPage() {
  const { data, error, profile, profileError, profileLoading, selectedAsset, selectedAssetID, setSelectedAssetID } = useDatasets();

  if (error) {
    return <section className="panel">Datasets error: {error}</section>;
  }

  return (
    <section className="page-grid">
      <article className="card wide-card">
        <div className="row-between">
          <h2>Catalog Trust Summary</h2>
          <span className="badge">{data?.summary.total_assets ?? 0} assets</span>
        </div>
        <div className="stats-grid">
          <div className="subcard">
            <p className="muted">Documented columns</p>
            <strong>
              {data?.summary.documented_columns ?? 0}/{data?.summary.total_columns ?? 0}
            </strong>
          </div>
          <div className="subcard">
            <p className="muted">Assets missing docs</p>
            <strong>{data?.summary.assets_missing_docs ?? 0}</strong>
          </div>
          <div className="subcard">
            <p className="muted">Assets missing quality</p>
            <strong>{data?.summary.assets_missing_quality ?? 0}</strong>
          </div>
        </div>
      </article>
      <article className="card">
        <div className="row-between">
          <h2>Catalog</h2>
          <span className="badge">{(data?.assets ?? []).length} assets</span>
        </div>
        <div className="stack">
          {(data?.assets ?? []).map((asset) => (
            <button
              className={`nav-button ${selectedAssetID === asset.id ? "active" : ""}`}
              key={asset.id}
              onClick={() => setSelectedAssetID(asset.id)}
              type="button"
            >
              <div className="row-between">
                <strong>{asset.name}</strong>
                <span className="badge">{asset.layer}</span>
              </div>
              <p className="muted">
                {asset.owner} · {asset.freshness_status.state}
              </p>
            </button>
          ))}
        </div>
      </article>

      {selectedAsset ? (
        <DatasetDetail asset={selectedAsset} profile={profile} profileError={profileError} profileLoading={profileLoading} />
      ) : (
        <article className="card">
          <h2>No dataset selected</h2>
          <p className="muted">Load manifests or run a sample pipeline to populate the catalog.</p>
        </article>
      )}
    </section>
  );
}

function DatasetDetail({
  asset,
  profile,
  profileError,
  profileLoading
}: {
  asset: Asset;
  profile: AssetProfile | null;
  profileError: string | null;
  profileLoading: boolean;
}) {
  return (
    <article className="card">
      <div className="row-between">
        <h2>{asset.name}</h2>
        <div className="inline-actions">
          <span className="badge">{asset.layer}</span>
          <span className="badge">{asset.kind}</span>
          <span className="badge">{asset.freshness_status.state}</span>
        </div>
      </div>
      <p>{asset.description}</p>
      <div className="form-grid">
        <div className="subcard">
          <p className="muted">Owner</p>
          <strong>{asset.owner}</strong>
        </div>
        <div className="subcard">
          <p className="muted">Freshness</p>
          <strong>{asset.freshness_status.message}</strong>
          {asset.freshness_status.last_updated ? (
            <p className="muted">Last updated {new Date(asset.freshness_status.last_updated).toLocaleString()}</p>
          ) : null}
        </div>
        <div className="subcard">
          <p className="muted">Coverage</p>
          <strong>
            {asset.coverage.documented_columns}/{asset.coverage.total_columns} documented columns
          </strong>
          <p className="muted">
            Docs {asset.coverage.has_documentation ? "present" : "missing"} · Quality{" "}
            {asset.coverage.has_quality_checks ? "present" : "missing"} · PII{" "}
            {asset.coverage.contains_pii ? "detected" : "none"}
          </p>
        </div>
        <div className="subcard">
          <p className="muted">Lineage</p>
          <strong>{asset.lineage.upstream.length} upstream</strong>
          <p className="muted">{asset.lineage.downstream.length} downstream</p>
        </div>
      </div>
      <div className="stack">
        <ReferenceStrip label="Sources" values={asset.source_refs} />
        <ReferenceStrip label="Upstream assets" values={asset.lineage.upstream} />
        <ReferenceStrip label="Downstream assets" values={asset.lineage.downstream} />
        <ReferenceStrip label="Quality checks" values={asset.quality_check_refs} />
        <ReferenceStrip label="Docs" values={asset.documentation_refs} />
      </div>
      <div className="form-grid">
        <div className="subcard">
          <p className="muted">Runtime profile</p>
          <strong>{profileLoading ? "Profiling..." : `${profile?.row_count ?? 0} rows`}</strong>
          <p className="muted">
            {profile
              ? `${profile.format.toUpperCase()} · ${formatBytes(profile.file_bytes)} · ${profile.profile_state}`
              : profileError ?? "No runtime profile available yet."}
          </p>
        </div>
        <div className="subcard">
          <p className="muted">Observed at</p>
          <strong>{profile?.observed_at ? new Date(profile.observed_at).toLocaleString() : "Not observed yet"}</strong>
          <p className="muted">
            Profile generated {profile?.generated_at ? new Date(profile.generated_at).toLocaleString() : "pending"}
          </p>
        </div>
      </div>
      <table className="data-table">
        <thead>
          <tr>
            <th>Column</th>
            <th>Type</th>
            <th>Description</th>
            <th>Flags</th>
          </tr>
        </thead>
        <tbody>
          {asset.columns.map((column) => (
            <tr key={column.name}>
              <td>{column.name}</td>
              <td>{column.type}</td>
              <td>{column.description || "No column documentation yet."}</td>
              <td>{column.is_pii ? "PII" : "-"}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <div className="stack">
        <div className="row-between">
          <h3>Observed Column Profile</h3>
          <span className="badge">{profile?.columns.length ?? 0} profiled columns</span>
        </div>
        {profileError ? <p className="muted">{profileError}</p> : null}
        {profileLoading ? <p className="muted">Generating profile from the current materialized asset...</p> : null}
        {profile ? (
          <table className="data-table">
            <thead>
              <tr>
                <th>Column</th>
                <th>Observed type</th>
                <th>Nulls</th>
                <th>Unique</th>
                <th>Range</th>
                <th>Samples</th>
              </tr>
            </thead>
            <tbody>
              {profile.columns.map((column) => (
                <tr key={column.name}>
                  <td>{column.name}</td>
                  <td>{column.observed_type}</td>
                  <td>{column.null_count}</td>
                  <td>{column.unique_count}</td>
                  <td>
                    {column.min_value || column.max_value
                      ? `${column.min_value ?? "-"} to ${column.max_value ?? "-"}`
                      : "-"}
                  </td>
                  <td>{column.sample_values.length > 0 ? column.sample_values.join(", ") : "-"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : null}
      </div>
    </article>
  );
}

function ReferenceStrip({ label, values }: { label: string; values: string[] }) {
  return (
    <div className="stack">
      <span className="muted">{label}</span>
      <div className="inline-actions">
        {(values.length > 0 ? values : ["None recorded"]).map((value) => (
          <span className="badge" key={value}>
            {value}
          </span>
        ))}
      </div>
    </div>
  );
}

function formatBytes(value: number) {
  if (value <= 0) {
    return "0 B";
  }
  if (value < 1024) {
    return `${value} B`;
  }
  if (value < 1024 * 1024) {
    return `${(value / 1024).toFixed(1)} KB`;
  }
  return `${(value / (1024 * 1024)).toFixed(1)} MB`;
}
