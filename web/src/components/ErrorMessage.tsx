import type { ReactNode } from "react";

export function ErrorMessage({
  title = "Something went wrong",
  message,
  action
}: {
  title?: string;
  message: string;
  action?: ReactNode;
}) {
  return (
    <section className="panel error-panel" role="alert">
      <div className="stack">
        <div>
          <h2>{title}</h2>
          <p className="muted">{message}</p>
        </div>
        {action}
      </div>
    </section>
  );
}
