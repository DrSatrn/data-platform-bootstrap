// The App component defines the primary information architecture for the
// platform UI. The design favors a focused internal-tool layout that surfaces
// operational context and curated analytics without decorative clutter.
import { useEffect, useMemo, useState } from "react";
import { Navigate, NavLink, Route, Routes, useLocation } from "react-router-dom";

import { ErrorBoundary } from "../components/ErrorBoundary";
import { useAuth } from "../features/auth/useAuth";
import { DashboardPage } from "../pages/DashboardPage";
import { DatasetsPage } from "../pages/DatasetsPage";
import { ManagementPage } from "../pages/ManagementPage";
import { MetricsPage } from "../pages/MetricsPage";
import { PipelinesPage } from "../pages/PipelinesPage";
import { SystemPage } from "../pages/SystemPage";

const routes = [
  { path: "/dashboard", label: "Dashboard" },
  { path: "/management", label: "Management" },
  { path: "/metrics", label: "Metrics" },
  { path: "/pipelines", label: "Pipelines" },
  { path: "/datasets", label: "Datasets" },
  { path: "/system", label: "System" }
];

export function App() {
  const location = useLocation();
  const { token, setToken, clearToken, login, logout, session, loading, error } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loggingIn, setLoggingIn] = useState(false);
  const currentLabel = useMemo(
    () => routes.find((item) => location.pathname === "/" ? item.path === "/dashboard" : item.path === location.pathname)?.label ?? "Dashboard",
    [location.pathname]
  );

  useEffect(() => {
    document.title = `Data Platform | ${currentLabel}`;
  }, [currentLabel]);

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
            <NavLink
              key={item.path}
              className={({ isActive }) => (isActive ? "nav-button active" : "nav-button")}
              to={item.path}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="content">
        <ErrorBoundary>
          <Routes>
            <Route element={<Navigate replace to="/dashboard" />} path="/" />
            <Route element={<DashboardPage />} path="/dashboard" />
            <Route element={<ManagementPage />} path="/management" />
            <Route element={<MetricsPage />} path="/metrics" />
            <Route element={<PipelinesPage />} path="/pipelines" />
            <Route element={<DatasetsPage />} path="/datasets" />
            <Route element={<SystemPage />} path="/system" />
            <Route element={<Navigate replace to="/dashboard" />} path="*" />
          </Routes>
        </ErrorBoundary>
      </main>
    </div>
  );
}
