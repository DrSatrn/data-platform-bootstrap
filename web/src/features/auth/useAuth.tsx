// This file provides the browser-side auth model for the platform. It keeps
// both native session login and the bootstrap-token escape hatch available so
// operators can recover a fresh environment without guessing at hidden flows.
import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from "react";

import { deleteJSON, fetchJSON, postJSON } from "../../lib/api";

type SessionPayload = {
  principal: {
    user_id?: string;
    subject: string;
    role: string;
    auth_source?: string;
  };
  capabilities: Record<string, boolean>;
};

type LoginPayload = {
  token: string;
  session: SessionPayload;
};

type AuthContextValue = {
  token: string;
  setToken: (value: string) => void;
  clearToken: () => void;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  session: SessionPayload | null;
  loading: boolean;
  error: string | null;
};

export const authStorageKey = "data-platform-auth-token";
const AuthContext = createContext<AuthContextValue | null>(null);

const anonymousSession: SessionPayload = {
  principal: { subject: "anonymous", role: "anonymous", auth_source: "anonymous" },
  capabilities: {
    view_platform: false,
    trigger_runs: false,
    edit_dashboards: false,
    run_admin_terminal: false,
    manage_users: false
  }
};

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setTokenState] = useState(() => window.localStorage.getItem(authStorageKey) ?? "");
  const [session, setSession] = useState<SessionPayload | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    window.localStorage.setItem(authStorageKey, token);
    setLoading(true);
    fetchJSON<SessionPayload>("/api/v1/session", token.trim() || undefined)
      .then((nextSession) => {
        setSession(nextSession);
        setError(null);
      })
      .catch(() => {
        setSession(anonymousSession);
      })
      .finally(() => setLoading(false));
  }, [token]);

  async function login(username: string, password: string) {
    const payload = await postJSON<LoginPayload, { username: string; password: string }>(
      "/api/v1/session",
      { username, password }
    );
    setTokenState(payload.token);
    setSession(payload.session);
    setError(null);
  }

  async function logout() {
    try {
      if (token.trim()) {
        await deleteJSON<{ status: string }>("/api/v1/session", token.trim());
      }
    } catch {
      // Clearing the local token is still the safest operator behavior.
    } finally {
      setTokenState("");
      setSession(anonymousSession);
      setError(null);
    }
  }

  const value = useMemo<AuthContextValue>(
    () => ({
      token,
      setToken: (value) => {
        setTokenState(value);
        setError(null);
      },
      clearToken: () => {
        setTokenState("");
        setError(null);
      },
      login: async (username, password) => {
        try {
          await login(username, password);
        } catch (err) {
          setError(err instanceof Error ? err.message : "Login failed");
          throw err;
        }
      },
      logout,
      session,
      loading,
      error
    }),
    [error, loading, session, token]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return value;
}
