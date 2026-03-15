// This hook manages the admin-only identity view so operators can create and
// maintain platform users from the built-in System page.
import { useEffect, useState } from "react";

import { patchJSON, postJSON, fetchJSON } from "../../lib/api";
import { useAuth } from "../auth/useAuth";

type User = {
  id: string;
  username: string;
  display_name: string;
  role: string;
  is_active: boolean;
  is_bootstrap: boolean;
};

type UserListPayload = {
  users: User[];
};

export function useIdentityAdmin() {
  const { session } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canManageUsers = Boolean(session?.capabilities.manage_users);

  async function load() {
    if (!canManageUsers) {
      setUsers([]);
      return;
    }
    setLoading(true);
    try {
      const payload = await fetchJSON<UserListPayload>("/api/v1/admin/users");
      setUsers(payload.users);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load users");
    } finally {
      setLoading(false);
    }
  }

  async function createUser(payload: { username: string; display_name: string; role: string; password: string }) {
    await postJSON<{ user: User }, typeof payload>("/api/v1/admin/users", payload);
    await load();
  }

  async function setUserActive(username: string, active: boolean) {
    await patchJSON<{ user: User }, { username: string; action: string; active: boolean }>("/api/v1/admin/users", {
      username,
      action: "set_active",
      active
    });
    await load();
  }

  async function resetPassword(username: string, password: string) {
    await patchJSON<{ user: User }, { username: string; action: string; password: string }>("/api/v1/admin/users", {
      username,
      action: "reset_password",
      password
    });
    await load();
  }

  useEffect(() => {
    void load();
  }, [canManageUsers]);

  return { users, loading, error, canManageUsers, reload: load, createUser, setUserActive, resetPassword };
}
