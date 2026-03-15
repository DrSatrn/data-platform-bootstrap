// The App component defines the primary information architecture for the
// platform UI. The design favors a focused internal-tool layout that surfaces
// operational context and curated analytics without decorative clutter.
import { useEffect, useState } from "react";

import { useAuth } from "../features/auth/useAuth";
import { DashboardPage } from "../pages/DashboardPage";
import { DatasetsPage } from "../pages/DatasetsPage";
import { MetricsPage } from "../pages/MetricsPage";
import { PipelinesPage } from "../pages/PipelinesPage";
import { SystemPage } from "../pages/SystemPage";

type Route = "dashboard" | "metrics" | "pipelines" | "datasets" | "system";

const routes: Array<{ id: Route; label: string }> = [
  { id: "dashboard", label: "Dashboard" },
  { id: "metrics", label: "Metrics" },
  { id: "pipelines", label: "Pipelines" },
  { id: "datasets", label: "Datasets" },
  { id: "system", label: "System" }
];

export function App() {
  const [route, setRoute] = useState<Route>("dashboard");
  const { token, setToken, clearToken, login, logout, session, loading, error } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loggingIn, setLoggingIn] = useState(false);

  useEffect(() => {
    document.title = `Data Platform | ${routes.find((item) => item.id === route)?.label}`;
  }, [route]);

  async function handleLogin() {
    if (!username.trim() || !password) {
      return;
    }
    setLoggingIn(true);
    try {
      await login(username.trim(), password);
      setPassword("");
    } finally {
      setLoggingIn(false);
    }
  }

  return (
    <div className="shell">
      <aside className="sidebar">
        <div>
          <p className="eyebrow">Local-First Control Plane</p>
          <h1>Data Platform</h1>
          <p className="lede">
            Orchestration, catalog, analytics, and reporting in one operator-focused surface.
          </p>
          <div className="stack auth-panel">
            <p className="muted">
              Session: {loading ? "loading..." : `${session?.principal.subject ?? "anonymous"} (${session?.principal.role ?? "anonymous"})`}
            </p>
            <input
              className="terminal-input"
              onChange={(event) => setUsername(event.target.value)}
              placeholder="Username"
              type="text"
              value={username}
            />
            <input
              className="terminal-input"
              onChange={(event) => setPassword(event.target.value)}
              placeholder="Password"
              type="password"
              value={password}
            />
            <div className="inline-actions">
              <button className="mini-button" onClick={() => void handleLogin()} type="button">
                {loggingIn ? "Signing in..." : "Sign in"}
              </button>
              <button className="mini-button" onClick={() => void logout()} type="button">
                Sign out
              </button>
            </div>
            <input
              className="terminal-input"
              onChange={(event) => setToken(event.target.value)}
              placeholder="Bootstrap/admin bearer token override"
              type="password"
              value={token}
            />
            <div className="inline-actions">
              <button className="mini-button" onClick={clearToken} type="button">
                Clear token
              </button>
              <span className="badge">{session?.principal.role ?? "anonymous"}</span>
            </div>
            {error ? <p className="muted">Auth error: {error}</p> : null}
            <p className="muted">
              Native sessions are the normal path. The bootstrap token remains available for first-run recovery and emergency admin access.
            </p>
          </div>
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
    case "metrics":
      return <MetricsPage />;
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
