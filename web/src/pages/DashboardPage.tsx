// DashboardPage renders the reporting-oriented landing view. The page is now
// driven by saved dashboard definitions plus constrained analytics queries so
// the UI feels closer to a lightweight internal BI tool than a hardcoded demo.
import { StatCard } from "../components/StatCard";
import { useDashboardData } from "../features/dashboard/useDashboardData";

export function DashboardPage() {
  const { data, error } = useDashboardData();

  if (error) {
    return <section className="panel">Dashboard error: {error}</section>;
  }

  const dashboard = data?.dashboard;
  const widgetData = data?.widgetData ?? {};
  const kpiWidgets = (dashboard?.widgets ?? []).filter((widget) => widget.type === "kpi");
  const detailWidgets = (dashboard?.widgets ?? []).filter((widget) => widget.type !== "kpi");

  return (
    <section className="page-grid">
      <div className="hero card">
        <p className="eyebrow">Personal Finance Vertical Slice</p>
        <h2>{dashboard?.name ?? "Finance Overview"}</h2>
        <p className="lede">
          {dashboard?.description ??
            "Curated reporting widgets are driven by saved dashboard definitions rather than hardcoded page assumptions."}
        </p>
      </div>
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
      {detailWidgets.map((widget) => {
        const series = widgetData[widget.id]?.series ?? [];
        const columns = deriveColumns(series);

        return (
          <article className="card wide-card" key={widget.id}>
            <div className="row-between">
              <h3>{widget.name}</h3>
              <span className="badge">{widget.dataset_ref ?? widget.metric_ref}</span>
            </div>
            {widget.description ? <p className="muted">{widget.description}</p> : null}
            {series.length === 0 ? (
              <p className="muted">No data is available yet for this widget.</p>
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
      })}
    </section>
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
  if (typeof value === "number" && (field.includes("income") || field.includes("expense") || field.includes("spend") || field.includes("budget") || field.includes("variance"))) {
    return new Intl.NumberFormat("en-AU", {
      style: "currency",
      currency: "AUD",
      maximumFractionDigits: 0
    }).format(value);
  }
  return String(value ?? "");
}
