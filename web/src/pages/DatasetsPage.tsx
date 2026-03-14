// DatasetsPage presents the metadata catalog in a compact, readable format.
// The page highlights layer, ownership, and column shape because those are the
// first things operators usually need to understand an asset.
import { useDatasets } from "../features/datasets/useDatasets";

export function DatasetsPage() {
  const { data, error } = useDatasets();

  if (error) {
    return <section className="panel">Datasets error: {error}</section>;
  }

  return (
    <section className="page-grid">
      {(data?.assets ?? []).map((asset) => (
        <article className="card" key={asset.id}>
          <div className="row-between">
            <h2>{asset.name}</h2>
            <span className="badge">{asset.layer}</span>
          </div>
          <p>{asset.description}</p>
          <p className="muted">Owner: {asset.owner}</p>
          <table className="data-table">
            <thead>
              <tr>
                <th>Column</th>
                <th>Type</th>
              </tr>
            </thead>
            <tbody>
              {asset.columns.map((column) => (
                <tr key={column.name}>
                  <td>{column.name}</td>
                  <td>{column.type}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </article>
      ))}
    </section>
  );
}
