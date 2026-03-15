// This file provides a lightweight browser-side auth/session model backed by a
// locally stored bearer token. It keeps the current self-hosted product usable
// without introducing a heavyweight external auth dependency before the rest of
// the platform is ready.
import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from "react";

import { fetchJSON } from "../../lib/api";

type SessionPayload = {
  principal: {
    subject: string;
    role: string;
  };
  capabilities: Record<string, boolean>;
};

type AuthContextValue = {
  token: string;
  setToken: (value: string) => void;
  clearToken: () => void;
  session: SessionPayload | null;
  loading: boolean;
};

export const authStorageKey = "data-platform-auth-token";
const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setTokenState] = useState(() => window.localStorage.getItem(authStorageKey) ?? "");
  const [session, setSession] = useState<SessionPayload | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    window.localStorage.setItem(authStorageKey, token);
    setLoading(true);
    fetchJSON<SessionPayload>("/api/v1/session", token.trim() || undefined)
      .then(setSession)
      .catch(() =>
        setSession({
          principal: { subject: "anonymous", role: "anonymous" },
          capabilities: {
            view_platform: true,
            trigger_runs: false,
            edit_dashboards: false,
            run_admin_terminal: false
          }
        })
      )
      .finally(() => setLoading(false));
  }, [token]);

  const value = useMemo<AuthContextValue>(
    () => ({
      token,
      setToken: setTokenState,
      clearToken: () => setTokenState(""),
      session,
      loading
    }),
    [loading, session, token]
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
