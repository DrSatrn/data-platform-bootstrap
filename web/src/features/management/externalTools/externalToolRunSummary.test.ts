import { describe, expect, it } from "vitest";

import { buildExternalToolRunSummaries } from "./externalToolRunSummary";

describe("externalToolRunSummary", () => {
  it("groups external tool events and artifacts by job", () => {
    const summaries = buildExternalToolRunSummaries(
      [
        {
          time: "2026-03-15T00:00:00Z",
          level: "info",
          message: "external tool command started",
          fields: { job_id: "build_finance_dbt", tool: "dbt", action: "build" }
        },
        {
          time: "2026-03-15T00:00:03Z",
          level: "info",
          message: "external tool command finished",
          fields: { job_id: "build_finance_dbt", tool: "dbt", action: "build" }
        }
      ],
      [
        {
          run_id: "run_1",
          relative_path: "external_tools/build_finance_dbt/logs/stdout.log",
          size_bytes: 12,
          modified_at: "2026-03-15T00:00:01Z",
          content_type: "application/octet-stream"
        },
        {
          run_id: "run_1",
          relative_path: "external_tools/build_finance_dbt/logs/stderr.log",
          size_bytes: 10,
          modified_at: "2026-03-15T00:00:01Z",
          content_type: "application/octet-stream"
        },
        {
          run_id: "run_1",
          relative_path: "external_tools/build_finance_dbt/target/run_results.json",
          size_bytes: 42,
          modified_at: "2026-03-15T00:00:02Z",
          content_type: "application/json"
        }
      ]
    );

    expect(summaries).toHaveLength(1);
    expect(summaries[0]?.jobID).toBe("build_finance_dbt");
    expect(summaries[0]?.tool).toBe("dbt");
    expect(summaries[0]?.action).toBe("build");
    expect(summaries[0]?.status).toBe("succeeded");
    expect(summaries[0]?.logArtifacts).toHaveLength(2);
    expect(summaries[0]?.outputArtifacts).toHaveLength(1);
  });
});
