// This file centralizes the frontend's API access so error handling and
// response shapes stay consistent across pages.
export async function fetchJSON<T>(path: string): Promise<T> {
  const response = await fetch(path);
  if (!response.ok) {
    throw new Error(`Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}
