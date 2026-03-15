import { describe, expect, it } from "vitest";

import { buildRunbookDock } from "./runbookDock";

describe("runbookDock", () => {
  it("promotes failed-session runbooks ahead of reference docs", () => {
    const items = buildRunbookDock(
      [
        {
          id: "dbt-run",
          title: "DBT Build",
          status: "failed",
          recommendedRunbook: "/docs/runbooks/optional-external-tools.md"
        },
        {
          id: "restore",
          title: "Restore Drill",
          status: "completed",
          recommendedRunbook: "/docs/runbooks/backups.md"
        }
      ],
      []
    );

    expect(items[0]?.path).toBe("/docs/runbooks/optional-external-tools.md");
    expect(items[0]?.urgency).toBe("now");
  });

  it("merges follow-up references into the dock", () => {
    const items = buildRunbookDock([], [
      {
        id: "open-runbook",
        title: "Open linked runbook",
        priority: "soon",
        rationale: "Review the dbt checklist before rerunning the job.",
        runbookPath: "/docs/runbooks/dbt-operator-checklist.md"
      }
    ]);

    expect(items).toHaveLength(1);
    expect(items[0]?.path).toBe("/docs/runbooks/dbt-operator-checklist.md");
    expect(items[0]?.urgency).toBe("soon");
  });
});
