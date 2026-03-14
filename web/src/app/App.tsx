// The App component defines the primary information architecture for the
// platform UI. The design favors a focused internal-tool layout that surfaces
// operational context and curated analytics without decorative clutter.
import { useEffect, useState } from "react";

import { DashboardPage } from "../pages/DashboardPage";
import { DatasetsPage } from "../pages/DatasetsPage";
import { PipelinesPage } from "../pages/PipelinesPage";
import { SystemPage } from "../pages/SystemPage";

type Route = "dashboard" | "pipelines" | "datasets" | "system";

const routes: Array<{ id: Route; label: string }> = [
  { id: "dashboard", label: "Dashboard" },
  { id: "pipelines", label: "Pipelines" },
  { id: "datasets", label: "Datasets" },
  { id: "system", label: "System" }
];

export function App() {
  const [route, setRoute] = useState<Route>("dashboard");

  useEffect(() => {
    document.title = `Data Platform | ${routes.find((item) => item.id === route)?.label}`;
  }, [route]);

  return (
    <div className="shell">
      <aside className="sidebar">
        <div>
          <p className="eyebrow">Local-First Control Plane</p>
          <h1>Data Platform</h1>
          <p className="lede">
            Orchestration, catalog, analytics, and reporting in one operator-focused surface.
          </p>
        </div>
        <nav className="nav">
          {routes.map((item) => (
            <button
              key={item.id}
              className={item.id === route ? "nav-button active" : "nav-button"}
              onClick={() => setRoute(item.id)}
              type="button"
            >
              {item.label}
            </button>
          ))}
        </nav>
      </aside>
      <main className="content">{renderRoute(route)}</main>
    </div>
  );
}

function renderRoute(route: Route) {
  switch (route) {
    case "dashboard":
      return <DashboardPage />;
    case "pipelines":
      return <PipelinesPage />;
    case "datasets":
      return <DatasetsPage />;
    case "system":
      return <SystemPage />;
    default:
      return <DashboardPage />;
  }
}
