import { describe, expect, it } from "vitest";

import { buildEvidenceBoard } from "./evidenceBoard";

describe("evidenceBoard", () => {
  it("prioritizes failed-session evidence and deduplicates by path", () => {
    const items = buildEvidenceBoard(
      [
        {
          id: "dbt-run",
          title: "DBT Build",
          status: "failed",
          pinnedArtifacts: ["external_tools/build_finance_dbt/logs/stderr.log"]
        }
      ],
      [
        {
          jobID: "build_finance_dbt",
          status: "failed",
          logArtifacts: [{ relative_path: "external_tools/build_finance_dbt/logs/stderr.log" }],
          outputArtifacts: [{ relative_path: "external_tools/build_finance_dbt/target/run_results.json" }]
        }
      ]
    );

    expect(items[0]?.path).toBe("external_tools/build_finance_dbt/logs/stderr.log");
    expect(items).toHaveLength(2);
  });
});
