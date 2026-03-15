import { describe, expect, it } from "vitest";

import { completeCommand, createTerminalSession, sessionNeedsAttention, startCommand, summarizeSession } from "./sessionModel";

describe("sessionModel", () => {
  it("tracks a successful command and preserves artifacts", () => {
    const started = startCommand(
      createTerminalSession({
        id: "restore-check",
        title: "Restore Check",
        scope: "recovery"
      }),
      "backup verify latest.tar.gz",
      "2026-03-15T13:00:00Z"
    );

    const completed = completeCommand(started, {
      exitCode: 0,
      at: "2026-03-15T13:00:04Z",
      stdout: ["bundle checksum ok"],
      artifacts: ["runs/run_1/recovery/verify-report.json"]
    });

    expect(completed.status).toBe("completed");
    expect(completed.lastExitCode).toBe(0);
    expect(completed.pinnedArtifacts).toContain("runs/run_1/recovery/verify-report.json");
    expect(summarizeSession(completed)).toBe("command completed successfully");
  });

  it("flags failed sessions for attention", () => {
    const failed = completeCommand(
      startCommand(
        createTerminalSession({
          id: "dbt-run",
          title: "DBT Finance Build",
          scope: "pipelines"
        }),
        "dbt build --select monthly_cashflow",
        "2026-03-15T13:10:00Z"
      ),
      {
        exitCode: 2,
        at: "2026-03-15T13:10:08Z",
        stderr: ["model monthly_cashflow failed"]
      }
    );

    expect(sessionNeedsAttention(failed)).toBe(true);
    expect(summarizeSession(failed)).toContain("failed with exit code 2");
  });
});
