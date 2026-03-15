import { afterEach, describe, expect, it, vi } from "vitest";

import { deleteJSON, fetchJSON, postJSON } from "./api";

describe("api helpers", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("fetchJSON sends bearer auth when a token is provided", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ status: "ok" })
    });
    vi.stubGlobal("fetch", fetchMock);

    const payload = await fetchJSON<{ status: string }>("/api/v1/session", "viewer-token");

    expect(payload.status).toBe("ok");
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/session", {
      headers: {
        Authorization: "Bearer viewer-token"
      }
    });
  });

  it("postJSON sends method, content type, body, and auth header", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ saved: true })
    });
    vi.stubGlobal("fetch", fetchMock);

    const payload = await postJSON<{ saved: boolean }, { id: string }>("/api/v1/reports", { id: "dashboard_1" }, "editor-token");

    expect(payload.saved).toBe(true);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/reports", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer editor-token"
      },
      body: JSON.stringify({ id: "dashboard_1" })
    });
  });

  it("deleteJSON surfaces the response text when a request fails", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      text: async () => "editor role required"
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(deleteJSON("/api/v1/reports?id=dashboard_1", "viewer-token")).rejects.toThrow("editor role required");
  });
});
