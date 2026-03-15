// This file centralizes the frontend's API access so error handling and
// response shapes stay consistent across pages.
export async function fetchJSON<T>(path: string, token?: string): Promise<T> {
  const response = await fetch(path, {
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
  if (!response.ok) {
    throw new Error(`Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

export async function postJSON<TResponse, TRequest>(path: string, payload: TRequest, token?: string): Promise<TResponse> {
  const response = await fetch(path, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {})
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
  const response = await fetch(path, {
    method: "DELETE",
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Request failed for ${path}: ${response.status}`);
  }

  return (await response.json()) as TResponse;
}
