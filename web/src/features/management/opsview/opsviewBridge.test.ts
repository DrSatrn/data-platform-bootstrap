import { describe, expect, it } from "vitest";

import { buildOpsviewSignalCards, flattenOpsviewEvidence, type OpsviewExternalToolRunSummary } from "./opsviewBridge";

describe("opsviewBridge", () => {
  const summaries: OpsviewExternalToolRunSummary[] = [
    {
      run_id: "run_1",
      pipeline_id: "personal_finance_dbt_pipeline",
      job_id: "build_finance_dbt",
      tool: "dbt",
      action: "build",
      status: "failed",
      failure_class: "execution_failed",
      events: [],
      log_artifacts: [],
      output_artifacts: [],
      evidence: {
        run_id: "run_1",
        total_artifacts: 2,
        log_artifact_count: 1,
        output_artifact_count: 1,
        artifact_paths: [
          "external_tools/build_finance_dbt/logs/stderr.log",
          "external_tools/build_finance_dbt/target/run_results.json"
        ],
        log_paths: ["external_tools/build_finance_dbt/logs/stderr.log"],
        output_paths: ["external_tools/build_finance_dbt/target/run_results.json"]
      }
    }
  ];

  it("builds operator signal cards from attention and summaries", () => {
    const cards = buildOpsviewSignalCards(
      {
        total_jobs: 1,
        failed_jobs: 1,
        running_jobs: 0,
        succeeded_jobs: 0,
        jobs_missing_logs: 0,
        jobs_missing_outputs: 0
      },
      summaries
    );

    expect(cards[1]?.tone).toBe("critical");
    expect(cards[1]?.detail).toContain("build_finance_dbt");
    expect(cards[1]?.detail).toContain("execution_failed");
  });

  it("flattens opsview evidence while preserving exact paths", () => {
    const evidence = flattenOpsviewEvidence(summaries);

    expect(evidence).toHaveLength(2);
    expect(evidence[0]?.path).toBe("external_tools/build_finance_dbt/logs/stderr.log");
    expect(evidence[0]?.category).toBe("log");
    expect(evidence[1]?.category).toBe("output");
  });
});
