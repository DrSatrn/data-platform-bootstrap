// DashboardPage renders the reporting-oriented landing view. The page combines
// KPI summaries with a compact trend table so the first slice already feels
// like a useful operator-facing product.
import { StatCard } from "../components/StatCard";
import { useDashboardData } from "../features/dashboard/useDashboardData";

export function DashboardPage() {
  const { data, error } = useDashboardData();

  if (error) {
    return <section className="panel">Dashboard error: {error}</section>;
  }

  const series = data?.dashboard.series ?? [];
  const latest = series[series.length - 1];

  return (
    <section className="page-grid">
      <div className="hero card">
        <p className="eyebrow">Personal Finance Vertical Slice</p>
        <h2>Curated Financial Health</h2>
        <p className="lede">
          This dashboard is powered by curated metrics rather than raw tables, which keeps reporting
          stable, explainable, and fast.
        </p>
      </div>
      <div className="stats-grid">
        <StatCard label="Latest Savings Rate" value={formatPercent(latest?.savings_rate)} tone="good" />
        <StatCard label="Latest Income" value={formatCurrency(latest?.income)} />
        <StatCard label="Latest Expenses" value={formatCurrency(latest?.expenses)} tone="warn" />
      </div>
      <div className="card wide-card">
        <h3>Monthly Cashflow Trend</h3>
        <table className="data-table">
          <thead>
            <tr>
              <th>Month</th>
              <th>Income</th>
              <th>Expenses</th>
              <th>Savings Rate</th>
            </tr>
          </thead>
          <tbody>
            {series.map((row) => (
              <tr key={String(row.month)}>
                <td>{String(row.month)}</td>
                <td>{formatCurrency(row.income)}</td>
                <td>{formatCurrency(row.expenses)}</td>
                <td>{formatPercent(row.savings_rate)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function formatCurrency(value: unknown) {
  const number = typeof value === "number" ? value : 0;
  return new Intl.NumberFormat("en-AU", {
    style: "currency",
    currency: "AUD",
    maximumFractionDigits: 0
  }).format(number);
}

function formatPercent(value: unknown) {
  const number = typeof value === "number" ? value : 0;
  return `${Math.round(number * 100)}%`;
}
