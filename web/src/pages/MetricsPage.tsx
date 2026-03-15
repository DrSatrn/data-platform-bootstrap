// MetricsPage exposes the semantic metric registry, filter controls, and
// preview data so analysts can browse the serving layer without editing a
// dashboard first.
import { useMetrics } from "../features/metrics/useMetrics";

export function MetricsPage() {
  const { metrics, error, selectedMetricID, selectedMetric, setSelectedMetricID, filters, updateFilter, preview } = useMetrics();

  if (error) {
    return <section className="panel">Metrics error: {error}</section>;
  }

  return (
    <section className="page-grid">
      <article className="hero card wide-card">
        <p className="eyebrow">Semantic Layer Browser</p>
        <div className="row-between">
          <div>
            <h2>Metrics Registry</h2>
            <p className="lede">Browse curated metrics, dimensions, measures, and preview the chart-ready responses the UI consumes.</p>
          </div>
          <span className="badge">{metrics.length} metrics</span>
        </div>
      </article>

      <article className="card">
        <div className="row-between">
          <h2>Available Metrics</h2>
          <span className="badge">repo-managed</span>
        </div>
        <div className="stack">
          {metrics.map((metric) => (
            <button
              className={`nav-button ${selectedMetricID === metric.definition.id ? "active" : ""}`}
              key={metric.definition.id}
              onClick={() => setSelectedMetricID(metric.definition.id)}
              type="button"
            >
              <div className="row-between">
                <strong>{metric.definition.name}</strong>
                <span className="badge">{metric.definition.default_visualization}</span>
              </div>
              <p className="muted">{metric.definition.dataset_ref}</p>
            </button>
          ))}
        </div>
      </article>

      <article className="card">
        {selectedMetric ? (
          <>
            <div className="row-between">
              <div>
                <h2>{selectedMetric.definition.name}</h2>
                <p>{selectedMetric.definition.description}</p>
              </div>
              <span className="badge">{selectedMetric.definition.id}</span>
            </div>

            <div className="form-grid">
              <div className="subcard">
                <p className="muted">Owner</p>
                <strong>{selectedMetric.definition.owner}</strong>
              </div>
              <div className="subcard">
                <p className="muted">Dataset</p>
                <strong>{selectedMetric.definition.dataset_ref}</strong>
              </div>
              <div className="subcard">
                <p className="muted">Dimensions</p>
                <div className="inline-actions">
                  {selectedMetric.definition.dimensions.map((value) => (
                    <span className="badge" key={value}>
                      {value}
                    </span>
                  ))}
                </div>
              </div>
              <div className="subcard">
                <p className="muted">Measures</p>
                <div className="inline-actions">
                  {selectedMetric.definition.measures.map((value) => (
                    <span className="badge" key={value}>
                      {value}
                    </span>
                  ))}
                </div>
              </div>
            </div>

            <div className="form-grid">
              <label className="stack">
                <span className="muted">From month</span>
                <input
                  className="terminal-input"
                  onChange={(event) => updateFilter("fromMonth", event.target.value)}
                  placeholder="2026-01"
                  value={filters.fromMonth}
                />
              </label>
              <label className="stack">
                <span className="muted">To month</span>
                <input
                  className="terminal-input"
                  onChange={(event) => updateFilter("toMonth", event.target.value)}
                  placeholder="2026-12"
                  value={filters.toMonth}
                />
              </label>
              <label className="stack wide-field">
                <span className="muted">Category filter</span>
                <input
                  className="terminal-input"
                  onChange={(event) => updateFilter("category", event.target.value)}
                  placeholder="Optional category for dimensional metrics"
                  value={filters.category}
                />
              </label>
            </div>

            <div className="chart-panel">
              {selectedMetric.definition.default_visualization === "line" ? <LinePreview rows={preview} /> : null}
              {selectedMetric.definition.default_visualization !== "line" ? <BarPreview rows={preview} /> : null}
              <table className="data-table">
                <thead>
                  <tr>
                    {Object.keys(preview[0] ?? { empty: "" }).map((key) => (
                      <th key={key}>{key}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {preview.length > 0 ? (
                    preview.map((row, index) => (
                      <tr key={`${selectedMetric.definition.id}-${index}`}>
                        {Object.entries(row).map(([key, value]) => (
                          <td key={key}>{String(value)}</td>
                        ))}
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={4}>No preview rows returned for the current filters.</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </>
        ) : (
          <>
            <h2>No metric selected</h2>
            <p className="muted">Load the metric registry to inspect the semantic layer.</p>
          </>
        )}
      </article>
    </section>
  );
}

function LinePreview({ rows }: { rows: Array<Record<string, string | number>> }) {
  if (rows.length === 0) {
    return <div className="subcard">No line preview available.</div>;
  }
  const values = rows.map((row) => Number(row[Object.keys(row).find((key) => typeof row[key] === "number") ?? ""] ?? 0));
  const max = Math.max(...values, 1);
  const points = values
    .map((value, index) => `${20 + index * (280 / Math.max(values.length - 1, 1))},${180 - (value / max) * 140}`)
    .join(" ");

  return (
    <svg className="chart-surface" viewBox="0 0 320 200">
      <line className="chart-axis" x1="20" x2="300" y1="180" y2="180" />
      <line className="chart-axis" x1="20" x2="20" y1="20" y2="180" />
      <polyline className="chart-line" points={points} />
    </svg>
  );
}

function BarPreview({ rows }: { rows: Array<Record<string, string | number>> }) {
  if (rows.length === 0) {
    return <div className="subcard">No bar preview available.</div>;
  }
  const values = rows.map((row) => Number(row[Object.keys(row).find((key) => typeof row[key] === "number") ?? ""] ?? 0));
  const max = Math.max(...values, 1);
  const width = 240 / Math.max(values.length, 1);

  return (
    <svg className="chart-surface" viewBox="0 0 320 200">
      <line className="chart-axis" x1="20" x2="300" y1="180" y2="180" />
      <line className="chart-axis" x1="20" x2="20" y1="20" y2="180" />
      {values.map((value, index) => {
        const height = (value / max) * 140;
        return <rect className="chart-bar" height={height} key={`${value}-${index}`} width={Math.max(width - 8, 14)} x={30 + index * width} y={180 - height} />;
      })}
    </svg>
  );
}
