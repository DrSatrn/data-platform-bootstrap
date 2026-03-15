// This file centralizes the frontend's API access so error handling and
// response shapes stay consistent across pages.
import { authStorageKey } from "../features/auth/useAuth";

export async function fetchJSON<T>(path: string, token?: string): Promise<T> {
  const resolvedToken = token ?? browserToken();
  const response = await fetch(path, {
    headers: {
      ...(resolvedToken ? { Authorization: `Bearer ${resolvedToken}` } : {})
    }
  });
  if (!response.ok) {
    throw new Error(`Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

export async function postJSON<TResponse, TRequest>(path: string, payload: TRequest, token?: string): Promise<TResponse> {
  const resolvedToken = token ?? browserToken();
  const response = await fetch(path, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(resolvedToken ? { Authorization: `Bearer ${resolvedToken}` } : {})
    },
    body: JSON.stringify(payload)
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Request failed for ${path}: ${response.status}`);
  }

  return (await response.json()) as TResponse;
}

export async function deleteJSON<TResponse>(path: string, token?: string): Promise<TResponse> {
  const resolvedToken = token ?? browserToken();
  const response = await fetch(path, {
    method: "DELETE",
    headers: {
      ...(resolvedToken ? { Authorization: `Bearer ${resolvedToken}` } : {})
    }
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Request failed for ${path}: ${response.status}`);
  }

  return (await response.json()) as TResponse;
}

function browserToken() {
  if (typeof window === "undefined") {
    return undefined;
  }
  const token = window.localStorage.getItem(authStorageKey)?.trim();
  return token || undefined;
}
