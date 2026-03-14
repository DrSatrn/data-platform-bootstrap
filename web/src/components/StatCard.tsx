// StatCard renders a compact KPI panel used throughout the internal dashboard.
// The component stays intentionally plain so the UI can remain fast and easy to
// evolve.
type StatCardProps = {
  label: string;
  value: string;
  tone?: "neutral" | "good" | "warn";
};

export function StatCard({ label, value, tone = "neutral" }: StatCardProps) {
  return (
    <article className={`card stat-card tone-${tone}`}>
      <p className="stat-label">{label}</p>
      <strong className="stat-value">{value}</strong>
    </article>
  );
}
