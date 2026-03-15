// DashboardPage renders the reporting-oriented landing view and now includes a
// lightweight dashboard editor. The goal is to make the reporting layer feel
// like a real internal tool where operators can shape the UI, not just consume
// a fixed dashboard.
import { StatCard } from "../components/StatCard";
import { useAuth } from "../features/auth/useAuth";
import { DashboardWidget, useDashboardData } from "../features/dashboard/useDashboardData";

const datasetOptions = [
  "mart_monthly_cashflow",
  "mart_category_spend",
  "mart_budget_vs_actual"
];

const metricOptions = ["metrics_savings_rate", "metrics_category_variance"];
const axisOptions = ["month", "category", "total_income", "total_expense", "net_cashflow", "actual_spend", "budget_amount", "variance_amount"];

export function DashboardPage() {
  const { session } = useAuth();
  const {
    dashboard,
    dashboards,
    draft,
    widgetData,
    isEditing,
    isSaving,
    error,
    saveError,
    selectedDashboardID,
    selectDashboard,
    startEditing,
    cancelEditing,
    updateDraft,
    updateWidget,
    updateWidgetFilter,
    addWidget,
    removeWidget,
    moveWidget,
    createDashboard,
    duplicateDashboard,
    deleteDashboard,
    saveDashboard
  } = useDashboardData();

  if (error) {
    return <section className="panel">Dashboard error: {error}</section>;
  }

  const activeDashboard = isEditing && draft ? draft : dashboard;
  if (!activeDashboard && dashboards.length === 0) {
    return <section className="panel">Loading dashboards...</section>;
  }
  const kpiWidgets = (activeDashboard?.widgets ?? []).filter((widget) => widget.type === "kpi");
  const detailWidgets = (activeDashboard?.widgets ?? []).filter((widget) => widget.type !== "kpi");

  return (
    <section className="page-grid">
      <div className="hero card wide-card">
        <p className="eyebrow">Saved Reporting Surface</p>
        <div className="row-between">
          <div>
            <h2>{activeDashboard?.name ?? "Dashboard Workspace"}</h2>
            <p className="lede">
              {activeDashboard?.description ??
                "Saved dashboards now drive the reporting experience and can be edited directly from the browser."}
            </p>
          </div>
          <div className="inline-actions">
            <select
              className="terminal-input compact-input"
              onChange={(event) => selectDashboard(event.target.value)}
              value={selectedDashboardID ?? ""}
            >
              {dashboards.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.name}
                </option>
              ))}
            </select>
            <button
              className="mini-button"
              disabled={!session?.capabilities.edit_dashboards}
              onClick={createDashboard}
              type="button"
            >
              New dashboard
            </button>
            <button
              className="mini-button"
              disabled={!session?.capabilities.edit_dashboards}
              onClick={duplicateDashboard}
              type="button"
            >
              Duplicate
            </button>
            <button
              className="mini-button"
              disabled={isSaving || !dashboard || !session?.capabilities.edit_dashboards}
              onClick={() => void deleteDashboard()}
              type="button"
            >
              Delete
            </button>
            {!isEditing ? (
              <button
                className="mini-button"
                disabled={!session?.capabilities.edit_dashboards}
                onClick={startEditing}
                type="button"
              >
                Edit dashboard
              </button>
            ) : (
              <>
                <button className="mini-button" onClick={cancelEditing} type="button">
                  Cancel
                </button>
                <button
                  className="mini-button"
                  disabled={isSaving || !session?.capabilities.edit_dashboards}
                  onClick={() => void saveDashboard()}
                  type="button"
                >
                  {isSaving ? "Saving..." : "Save dashboard"}
                </button>
              </>
            )}
          </div>
        </div>
        {!session?.capabilities.edit_dashboards ? (
          <p className="muted">Editor token required to create or modify saved dashboards.</p>
        ) : null}
        {saveError ? <p className="muted">Save error: {saveError}</p> : null}
      </div>

      {isEditing && draft ? (
        <article className="card wide-card">
          <div className="row-between">
            <h3>Dashboard Editor</h3>
            <button
              className="mini-button"
              disabled={!session?.capabilities.edit_dashboards}
              onClick={addWidget}
              type="button"
            >
              Add widget
            </button>
          </div>
          <div className="form-grid">
            <label className="stack">
              <span className="muted">Dashboard name</span>
              <input
                className="terminal-input"
                onChange={(event) => updateDraft("name", event.target.value)}
                value={draft.name}
              />
            </label>
            <label className="stack wide-field">
              <span className="muted">Description</span>
              <textarea
                className="terminal-input"
                onChange={(event) => updateDraft("description", event.target.value)}
                rows={3}
                value={draft.description}
              />
            </label>
          </div>
          <div className="stack">
            {draft.widgets.map((widget, index) => (
              <div className="subcard" key={widget.id}>
                <div className="row-between">
                  <strong>{widget.name || `Widget ${index + 1}`}</strong>
                  <div className="inline-actions">
                    <button
                      className="mini-button"
                      disabled={!session?.capabilities.edit_dashboards}
                      onClick={() => moveWidget(widget.id, -1)}
                      type="button"
                    >
                      Up
                    </button>
                    <button
                      className="mini-button"
                      disabled={!session?.capabilities.edit_dashboards}
                      onClick={() => moveWidget(widget.id, 1)}
                      type="button"
                    >
                      Down
                    </button>
                    <button
                      className="mini-button"
                      disabled={!session?.capabilities.edit_dashboards}
                      onClick={() => removeWidget(widget.id)}
                      type="button"
                    >
                      Remove
                    </button>
                  </div>
                </div>
                <div className="widget-editor-grid">
                  <label className="stack">
                    <span className="muted">Widget name</span>
                    <input
                      className="terminal-input"
                      onChange={(event) => updateWidget(widget.id, "name", event.target.value)}
                      value={widget.name}
                    />
                  </label>
                  <label className="stack">
                    <span className="muted">Type</span>
                    <select
                      className="terminal-input"
                      onChange={(event) => updateWidget(widget.id, "type", event.target.value)}
                      value={widget.type}
                    >
                      <option value="kpi">KPI</option>
                      <option value="table">Table</option>
                      <option value="line">Line</option>
                      <option value="bar">Bar</option>
                    </select>
                  </label>
                  <label className="stack">
                    <span className="muted">Dataset</span>
                    <select
                      className="terminal-input"
                      onChange={(event) => updateWidget(widget.id, "dataset_ref", event.target.value)}
                      value={widget.dataset_ref ?? ""}
                    >
                      <option value="">None</option>
                      {datasetOptions.map((option) => (
                        <option key={option} value={option}>
                          {option}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="stack">
                    <span className="muted">Metric</span>
                    <select
                      className="terminal-input"
                      onChange={(event) => updateWidget(widget.id, "metric_ref", event.target.value)}
                      value={widget.metric_ref ?? ""}
                    >
                      <option value="">None</option>
                      {metricOptions.map((option) => (
                        <option key={option} value={option}>
                          {option}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="stack">
                    <span className="muted">Value field</span>
                    <input
                      className="terminal-input"
                      onChange={(event) => updateWidget(widget.id, "value_field", event.target.value)}
                      value={widget.value_field ?? ""}
                    />
                  </label>
                  <label className="stack">
                    <span className="muted">X axis</span>
                    <select
                      className="terminal-input"
                      onChange={(event) => updateWidget(widget.id, "x_axis", event.target.value)}
                      value={widget.x_axis ?? "month"}
                    >
                      {axisOptions.map((option) => (
                        <option key={option} value={option}>
                          {option}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="stack">
                    <span className="muted">Y axis</span>
                    <select
                      className="terminal-input"
                      onChange={(event) => updateWidget(widget.id, "y_axis", event.target.value)}
                      value={widget.y_axis ?? "net_cashflow"}
                    >
                      {axisOptions.map((option) => (
                        <option key={option} value={option}>
                          {option}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="stack">
                    <span className="muted">Limit</span>
                    <input
                      className="terminal-input"
                      min={1}
                      onChange={(event) => updateWidget(widget.id, "limit", Number(event.target.value) || 0)}
                      type="number"
                      value={widget.limit ?? 12}
                    />
                  </label>
                  <label className="stack wide-field">
                    <span className="muted">Description</span>
                    <input
                      className="terminal-input"
                      onChange={(event) => updateWidget(widget.id, "description", event.target.value)}
                      value={widget.description ?? ""}
                    />
                  </label>
                  <label className="stack">
                    <span className="muted">From month</span>
                    <input
                      className="terminal-input"
                      onChange={(event) => updateWidgetFilter(widget.id, "from_month", event.target.value)}
                      placeholder="YYYY-MM"
                      value={widget.filters?.from_month ?? ""}
                    />
                  </label>
                  <label className="stack">
                    <span className="muted">To month</span>
                    <input
                      className="terminal-input"
                      onChange={(event) => updateWidgetFilter(widget.id, "to_month", event.target.value)}
                      placeholder="YYYY-MM"
                      value={widget.filters?.to_month ?? ""}
                    />
                  </label>
                  <label className="stack">
                    <span className="muted">Category filter</span>
                    <input
                      className="terminal-input"
                      onChange={(event) => updateWidgetFilter(widget.id, "category", event.target.value)}
                      placeholder="Category"
                      value={widget.filters?.category ?? ""}
                    />
                  </label>
                </div>
              </div>
            ))}
          </div>
        </article>
      ) : null}

      <div className="stats-grid">
        {kpiWidgets.map((widget) => {
          const series = widgetData[widget.id]?.series ?? [];
          const latest = series[series.length - 1] ?? {};
          const valueField = widget.value_field ?? firstMetricField(latest);

          return (
            <StatCard
              key={widget.id}
              label={widget.name}
              value={formatValue(latest[valueField], valueField)}
              tone={valueField.includes("rate") ? "good" : "neutral"}
            />
          );
        })}
      </div>
      {detailWidgets.map((widget) => (
        <WidgetPreview key={widget.id} widget={widget} series={widgetData[widget.id]?.series ?? []} />
      ))}
    </section>
  );
}

function WidgetPreview({
  widget,
  series
}: {
  widget: DashboardWidget;
  series: Array<Record<string, string | number>>;
}) {
  const columns = deriveColumns(series);

  return (
    <article className="card wide-card">
      <div className="row-between">
        <h3>{widget.name}</h3>
        <span className="badge">{widget.dataset_ref ?? widget.metric_ref}</span>
      </div>
      {widget.description ? <p className="muted">{widget.description}</p> : null}
      {series.length === 0 ? (
        <p className="muted">No data is available yet for this widget.</p>
      ) : widget.type === "line" || widget.type === "bar" ? (
        <ChartPreview widget={widget} series={series} />
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              {columns.map((column) => (
                <th key={column}>{column}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {series.map((row, index) => (
              <tr key={`${widget.id}-${index}`}>
                {columns.map((column) => (
                  <td key={column}>{formatValue(row[column], column)}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </article>
  );
}

function ChartPreview({
  widget,
  series
}: {
  widget: DashboardWidget;
  series: Array<Record<string, string | number>>;
}) {
  const xKey = widget.x_axis ?? "month";
  const yKey = widget.y_axis ?? widget.value_field ?? firstMetricField(series[0] ?? {});
  const points = series
    .map((row, index) => ({
      label: String(row[xKey] ?? index + 1),
      value: Number(row[yKey] ?? 0),
      index
    }))
    .filter((point) => Number.isFinite(point.value));

  if (points.length === 0) {
    return <p className="muted">This chart widget does not have plottable numeric data yet.</p>;
  }

  const width = 640;
  const height = 240;
  const padding = 28;
  const maxValue = Math.max(...points.map((point) => point.value), 0);
  const minValue = Math.min(...points.map((point) => point.value), 0);
  const range = Math.max(maxValue - minValue, 1);
  const baselineY = height - padding - ((0 - minValue) / range) * (height - padding * 2);
  const stepX = points.length > 1 ? (width - padding * 2) / (points.length - 1) : 0;
  const path = points
    .map((point, index) => {
      const x = padding + index * stepX;
      const y = height - padding - ((point.value - minValue) / range) * (height - padding * 2);
      return `${index === 0 ? "M" : "L"} ${x} ${y}`;
    })
    .join(" ");

  return (
    <div className="chart-panel">
      <svg className="chart-surface" viewBox={`0 0 ${width} ${height}`} role="img">
        <line className="chart-axis" x1={padding} x2={padding} y1={padding} y2={height - padding} />
        <line className="chart-axis" x1={padding} x2={width - padding} y1={baselineY} y2={baselineY} />
        {widget.type === "line" ? (
          <>
            <path className="chart-line" d={path} />
            {points.map((point, index) => {
              const x = padding + index * stepX;
              const y = height - padding - ((point.value - minValue) / range) * (height - padding * 2);
              return <circle className="chart-dot" cx={x} cy={y} key={point.label} r={4} />;
            })}
          </>
        ) : (
          points.map((point, index) => {
            const barWidth = Math.max((width - padding * 2) / Math.max(points.length * 1.8, 1), 16);
            const x = padding + index * ((width - padding * 2) / Math.max(points.length, 1)) + 6;
            const valueY = height - padding - ((point.value - minValue) / range) * (height - padding * 2);
            const barHeight = Math.abs(baselineY - valueY);
            const y = Math.min(baselineY, valueY);
            return (
              <rect
                className="chart-bar"
                height={barHeight}
                key={point.label}
                rx={6}
                width={barWidth}
                x={x}
                y={y}
              />
            );
          })
        )}
      </svg>
      <div className="chart-legend">
        {points.map((point) => (
          <div className="chart-legend-item" key={point.label}>
            <span>{point.label}</span>
            <strong>{formatValue(point.value, yKey)}</strong>
          </div>
        ))}
      </div>
    </div>
  );
}

function deriveColumns(series: Array<Record<string, string | number>>) {
  const first = series[0];
  return first ? Object.keys(first) : [];
}

function firstMetricField(row: Record<string, string | number>) {
  return Object.keys(row).find((key) => key !== "month" && key !== "category") ?? "value";
}

function formatValue(value: string | number | undefined, field: string) {
  if (typeof value === "number" && field.includes("rate")) {
    return `${Math.round(value * 100)}%`;
  }
  if (
    typeof value === "number" &&
    (field.includes("income") ||
      field.includes("expense") ||
      field.includes("spend") ||
      field.includes("budget") ||
      field.includes("variance"))
  ) {
    return new Intl.NumberFormat("en-AU", {
      style: "currency",
      currency: "AUD",
      maximumFractionDigits: 0
    }).format(value);
  }
  return String(value ?? "");
}
