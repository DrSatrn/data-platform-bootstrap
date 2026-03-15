// DatasetsPage now behaves more like a metadata workbench than a flat list.
// Operators can scan the catalog quickly on the left, then inspect ownership,
// freshness, lineage references, and column documentation for the selected
// asset on the right.
import { Asset, useDatasets } from "../features/datasets/useDatasets";

export function DatasetsPage() {
  const { data, error, selectedAsset, selectedAssetID, setSelectedAssetID } = useDatasets();

  if (error) {
    return <section className="panel">Datasets error: {error}</section>;
  }

  return (
    <section className="page-grid">
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
        <DatasetDetail asset={selectedAsset} />
      ) : (
        <article className="card">
          <h2>No dataset selected</h2>
          <p className="muted">Load manifests or run a sample pipeline to populate the catalog.</p>
        </article>
      )}
    </section>
  );
}

function DatasetDetail({ asset }: { asset: Asset }) {
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
      </div>
      <div className="stack">
        <ReferenceStrip label="Sources" values={asset.source_refs} />
        <ReferenceStrip label="Quality checks" values={asset.quality_check_refs} />
        <ReferenceStrip label="Docs" values={asset.documentation_refs} />
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
