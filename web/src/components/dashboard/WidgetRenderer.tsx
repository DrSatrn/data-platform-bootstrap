import { StatCard } from "../StatCard";
import type { DashboardWidget } from "../../features/dashboard/useDashboardData";
import { deriveColumns, firstMetricField, formatValue, widgetPlacementStyle } from "./widgetUtils";

export function WidgetRenderer({
  widget,
  series,
  onExportCSV
}: {
  widget: DashboardWidget;
  series: Array<Record<string, string | number>>;
  onExportCSV: (widgetID: string) => Promise<void>;
}) {
  const columns = deriveColumns(series);

  return (
    <article className="card dashboard-widget" style={widgetPlacementStyle(widget)}>
      <div className="row-between">
        <h3>{widget.name}</h3>
        <div className="inline-actions">
          <span className="badge">{widget.dataset_ref ?? widget.metric_ref}</span>
          <button className="mini-button" onClick={() => void onExportCSV(widget.id)} type="button">
            Export CSV
          </button>
        </div>
      </div>
      {widget.description ? <p className="muted">{widget.description}</p> : null}
      {series.length === 0 ? (
        <p className="muted">No data is available yet for this widget.</p>
      ) : widget.type === "kpi" ? (
        <KPIWidget widget={widget} series={series} />
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

function KPIWidget({
  widget,
  series
}: {
  widget: DashboardWidget;
  series: Array<Record<string, string | number>>;
}) {
  const latest = series[series.length - 1] ?? {};
  const valueField = widget.value_field ?? firstMetricField(latest);

  return (
    <StatCard
      label={widget.name}
      tone={valueField.includes("rate") ? "good" : "neutral"}
      value={formatValue(latest[valueField], valueField)}
    />
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
