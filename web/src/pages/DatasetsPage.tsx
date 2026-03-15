// DatasetsPage now behaves more like a metadata workbench than a flat list.
// Operators can scan the catalog quickly on the left, then inspect ownership,
// freshness, lineage references, and column documentation for the selected
// asset on the right.
import { useEffect, useState } from "react";

import { ErrorMessage } from "../components/ErrorMessage";
import { LoadingSpinner } from "../components/LoadingSpinner";
import { useAuth } from "../features/auth/useAuth";
import { Asset, AssetProfile, DrilldownQuery, useDatasets } from "../features/datasets/useDatasets";

export function DatasetsPage() {
  const {
    data,
    drilldown,
    drilldownError,
    drilldownFilters,
    drilldownLoading,
    error,
    profile,
    profileError,
    profileLoading,
    saveAnnotations,
    saveError,
    savePending,
    selectedAsset,
    selectedAssetID,
    setSelectedAssetID,
    updateDrilldownFilter
  } = useDatasets();

  if (error) {
    return <ErrorMessage message={error} title="Datasets error" />;
  }
  if (!data) {
    return <LoadingSpinner label="Loading datasets..." />;
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
        <DatasetDetail
          asset={selectedAsset}
          profile={profile}
          profileError={profileError}
          profileLoading={profileLoading}
          saveAnnotations={saveAnnotations}
          saveError={saveError}
          savePending={savePending}
          drilldown={drilldown}
          drilldownError={drilldownError}
          drilldownFilters={drilldownFilters}
          drilldownLoading={drilldownLoading}
          updateDrilldownFilter={updateDrilldownFilter}
        />
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
  profileLoading,
  saveAnnotations,
  saveError,
  savePending,
  drilldown,
  drilldownError,
  drilldownFilters,
  drilldownLoading,
  updateDrilldownFilter
}: {
  asset: Asset;
  profile: AssetProfile | null;
  profileError: string | null;
  profileLoading: boolean;
  saveAnnotations: ReturnType<typeof useDatasets>["saveAnnotations"];
  saveError: string | null;
  savePending: boolean;
  drilldown: DrilldownQuery | null;
  drilldownError: string | null;
  drilldownFilters: {
    fromMonth: string;
    toMonth: string;
    category: string;
    groupBy: string;
    drillDimension: string;
    drillValue: string;
    sortBy: string;
    sortDirection: string;
  };
  drilldownLoading: boolean;
  updateDrilldownFilter: ReturnType<typeof useDatasets>["updateDrilldownFilter"];
}) {
  const { session } = useAuth();
  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState(() => buildDraft(asset));

  useEffect(() => {
    setDraft(buildDraft(asset));
    setIsEditing(false);
  }, [asset]);

  async function submitAnnotations() {
    await saveAnnotations({
      asset_id: asset.id,
      owner: draft.owner,
      description: draft.description,
      quality_check_refs: splitLines(draft.qualityCheckRefs),
      documentation_refs: splitLines(draft.documentationRefs),
      column_descriptions: asset.columns.map((column, index) => ({
        name: column.name,
        description: draft.columnDescriptions[index] ?? ""
      }))
    });
    setIsEditing(false);
  }

  return (
    <article className="card">
      <div className="row-between">
        <h2>{asset.name}</h2>
        <div className="inline-actions">
          <span className="badge">{asset.layer}</span>
          <span className="badge">{asset.kind}</span>
          <span className="badge">{asset.freshness_status.state}</span>
          {session?.capabilities.edit_metadata ? (
            <button className="mini-button" onClick={() => setIsEditing((value) => !value)} type="button">
              {isEditing ? "Cancel edit" : "Edit annotations"}
            </button>
          ) : null}
        </div>
      </div>
      <p>{asset.description}</p>
      {!session?.capabilities.edit_metadata ? (
        <p className="muted">Editor role required to update owners, docs, and column descriptions.</p>
      ) : null}
      {saveError ? <p className="muted">Save error: {saveError}</p> : null}
      <div className="form-grid">
        <div className="subcard">
          <p className="muted">Owner</p>
          {isEditing ? (
            <input
              className="terminal-input"
              onChange={(event) => setDraft((current) => ({ ...current, owner: event.target.value }))}
              value={draft.owner}
            />
          ) : (
            <strong>{asset.owner}</strong>
          )}
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
      {isEditing ? (
        <div className="stack">
          <label className="stack">
            <span className="muted">Description</span>
            <textarea
              className="terminal-input"
              onChange={(event) => setDraft((current) => ({ ...current, description: event.target.value }))}
              rows={3}
              value={draft.description}
            />
          </label>
          <div className="form-grid">
            <label className="stack">
              <span className="muted">Documentation refs (one per line)</span>
              <textarea
                className="terminal-input"
                onChange={(event) => setDraft((current) => ({ ...current, documentationRefs: event.target.value }))}
                rows={4}
                value={draft.documentationRefs}
              />
            </label>
            <label className="stack">
              <span className="muted">Quality check refs (one per line)</span>
              <textarea
                className="terminal-input"
                onChange={(event) => setDraft((current) => ({ ...current, qualityCheckRefs: event.target.value }))}
                rows={4}
                value={draft.qualityCheckRefs}
              />
            </label>
          </div>
          <div className="row-between">
            <h3>Column Documentation Overrides</h3>
            <button className="button" disabled={savePending} onClick={() => void submitAnnotations()} type="button">
              {savePending ? "Saving..." : "Save annotations"}
            </button>
          </div>
        </div>
      ) : null}
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
      {asset.id.startsWith("mart_") || asset.id.startsWith("metrics_") ? (
        <div className="stack">
          <div className="row-between">
            <h3>Curated Drill-down</h3>
            <span className="badge">{drilldown?.series.length ?? 0} rows</span>
          </div>
          <p className="muted">Use the analytics service to slice the curated dataset directly from the catalog workbench.</p>
          <div className="form-grid">
            <label className="stack">
              <span className="muted">From month</span>
              <input
                className="terminal-input"
                onChange={(event) => updateDrilldownFilter("fromMonth", event.target.value)}
                placeholder="YYYY-MM"
                value={drilldownFilters.fromMonth}
              />
            </label>
            <label className="stack">
              <span className="muted">To month</span>
              <input
                className="terminal-input"
                onChange={(event) => updateDrilldownFilter("toMonth", event.target.value)}
                placeholder="YYYY-MM"
                value={drilldownFilters.toMonth}
              />
            </label>
            <label className="stack">
              <span className="muted">Category</span>
              <input
                className="terminal-input"
                onChange={(event) => updateDrilldownFilter("category", event.target.value)}
                placeholder="Food"
                value={drilldownFilters.category}
              />
            </label>
            <label className="stack">
              <span className="muted">Group by</span>
              <select className="terminal-input" onChange={(event) => updateDrilldownFilter("groupBy", event.target.value)} value={drilldownFilters.groupBy}>
                <option value="">No grouping</option>
                {(drilldown?.available_dimensions ?? []).map((dimension) => (
                  <option key={dimension} value={dimension}>
                    {dimension}
                  </option>
                ))}
              </select>
            </label>
            <label className="stack">
              <span className="muted">Drill dimension</span>
              <select
                className="terminal-input"
                onChange={(event) => updateDrilldownFilter("drillDimension", event.target.value)}
                value={drilldownFilters.drillDimension}
              >
                <option value="">No drill filter</option>
                {(drilldown?.available_dimensions ?? []).map((dimension) => (
                  <option key={dimension} value={dimension}>
                    {dimension}
                  </option>
                ))}
              </select>
            </label>
            <label className="stack">
              <span className="muted">Drill value</span>
              <input
                className="terminal-input"
                onChange={(event) => updateDrilldownFilter("drillValue", event.target.value)}
                placeholder="2026-02 or Food"
                value={drilldownFilters.drillValue}
              />
            </label>
            <label className="stack">
              <span className="muted">Sort by</span>
              <select className="terminal-input" onChange={(event) => updateDrilldownFilter("sortBy", event.target.value)} value={drilldownFilters.sortBy}>
                <option value="">Default</option>
                {[...(drilldown?.available_dimensions ?? []), ...(drilldown?.available_measures ?? [])].map((field) => (
                  <option key={field} value={field}>
                    {field}
                  </option>
                ))}
              </select>
            </label>
            <label className="stack">
              <span className="muted">Sort direction</span>
              <select
                className="terminal-input"
                onChange={(event) => updateDrilldownFilter("sortDirection", event.target.value)}
                value={drilldownFilters.sortDirection}
              >
                <option value="asc">asc</option>
                <option value="desc">desc</option>
              </select>
            </label>
          </div>
          {drilldownError ? <p className="muted">{drilldownError}</p> : null}
          {drilldownLoading ? <p className="muted">Running curated drill-down query...</p> : null}
          {drilldown && drilldown.series.length > 0 ? (
            <table className="data-table">
              <thead>
                <tr>
                  {Object.keys(drilldown.series[0]).map((column) => (
                    <th key={column}>{column}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {drilldown.series.map((row, index) => (
                  <tr key={`${asset.id}-drill-${index}`}>
                    {Object.keys(drilldown.series[0]).map((column) => (
                      <td key={column}>{String(row[column] ?? "")}</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <p className="muted">No drill-down rows yet for the current filters.</p>
          )}
        </div>
      ) : null}
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
              <td>
                {isEditing ? (
                  <textarea
                    className="terminal-input"
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        columnDescriptions: current.columnDescriptions.map((value, index) =>
                          asset.columns[index].name === column.name ? event.target.value : value
                        )
                      }))
                    }
                    rows={2}
                    value={draft.columnDescriptions[asset.columns.findIndex((item) => item.name === column.name)] ?? ""}
                  />
                ) : (
                  column.description || "No column documentation yet."
                )}
              </td>
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

function buildDraft(asset: Asset) {
  return {
    owner: asset.owner,
    description: asset.description,
    documentationRefs: asset.documentation_refs.join("\n"),
    qualityCheckRefs: asset.quality_check_refs.join("\n"),
    columnDescriptions: asset.columns.map((column) => column.description ?? "")
  };
}

function splitLines(value: string) {
  return value
    .split("\n")
    .map((item) => item.trim())
    .filter(Boolean);
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
