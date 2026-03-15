export function LoadingSpinner({ label = "Loading..." }: { label?: string }) {
  return (
    <section className="panel" aria-busy="true" aria-live="polite">
      <div className="loading-state">
        <span className="loading-spinner" />
        <p className="muted">{label}</p>
      </div>
    </section>
  );
}
