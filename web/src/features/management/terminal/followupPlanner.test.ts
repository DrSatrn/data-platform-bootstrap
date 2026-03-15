import { describe, expect, it } from "vitest";

import { completeCommand, createTerminalSession, startCommand } from "./sessionModel";
import { planFollowupActions, prioritizeActions } from "./followupPlanner";

describe("followupPlanner", () => {
  it("creates urgent failure follow-ups for pipeline sessions", () => {
    const failedSession = completeCommand(
      startCommand(
        createTerminalSession({
          id: "dbt-build",
          title: "DBT Build",
          scope: "pipelines",
          recommendedRunbook: "/docs/runbooks/optional-external-tools.md"
        }),
        "dbt build --select monthly_cashflow",
        "2026-03-15T14:00:00Z"
      ),
      {
        exitCode: 2,
        at: "2026-03-15T14:00:06Z",
        stderr: ["monthly_cashflow relation missing"]
      }
    );

    const actions = prioritizeActions(planFollowupActions(failedSession));

    expect(actions[0]?.priority).toBe("now");
    expect(actions.some((action) => action.title === "Retry the targeted pipeline command")).toBe(true);
    expect(actions.some((action) => action.runbookPath === "/docs/runbooks/optional-external-tools.md")).toBe(true);
  });

  it("suggests evidence promotion for successful recovery sessions", () => {
    const recoverySession = completeCommand(
      startCommand(
        createTerminalSession({
          id: "restore-drill",
          title: "Restore Drill",
          scope: "recovery",
          recommendedRunbook: "/docs/runbooks/backups.md"
        }),
        "backup verify latest.tar.gz",
        "2026-03-15T14:10:00Z"
      ),
      {
        exitCode: 0,
        at: "2026-03-15T14:10:05Z",
        stdout: ["bundle checksum ok"],
        artifacts: ["runs/run_1/recovery/verify-report.json"]
      }
    );

    const actions = planFollowupActions(recoverySession);

    expect(actions.some((action) => action.title === "Promote recovery evidence")).toBe(true);
    expect(actions.some((action) => action.artifactPath === "runs/run_1/recovery/verify-report.json")).toBe(true);
  });
});
