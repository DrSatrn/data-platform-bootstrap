import type { OpsviewAttentionSummary, OpsviewExternalToolRunSummary } from "./opsviewBridge";

export const mockOpsviewSummaries: OpsviewExternalToolRunSummary[] = [
  {
    run_id: "run_1",
    pipeline_id: "personal_finance_dbt_pipeline",
    job_id: "build_finance_dbt",
    tool: "dbt",
    action: "build",
    status: "failed",
    failure_class: "execution_failed",
    last_event_at: "2026-03-15T14:20:04Z",
    events: [],
    log_artifacts: [
      {
        run_id: "run_1",
        relative_path: "external_tools/build_finance_dbt/logs/stderr.log"
      }
    ],
    output_artifacts: [
      {
        run_id: "run_1",
        relative_path: "external_tools/build_finance_dbt/target/run_results.json"
      }
    ],
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
  },
  {
    run_id: "run_2",
    pipeline_id: "restore_validation_pipeline",
    job_id: "verify_restore_bundle",
    tool: "dbt",
    action: "build",
    status: "succeeded",
    last_event_at: "2026-03-15T14:24:05Z",
    events: [],
    log_artifacts: [
      {
        run_id: "run_2",
        relative_path: "external_tools/verify_restore_bundle/logs/stdout.log"
      }
    ],
    output_artifacts: [
      {
        run_id: "run_2",
        relative_path: "external_tools/verify_restore_bundle/target/run_results.json"
      }
    ],
    evidence: {
      run_id: "run_2",
      total_artifacts: 2,
      log_artifact_count: 1,
      output_artifact_count: 1,
      artifact_paths: [
        "external_tools/verify_restore_bundle/logs/stdout.log",
        "external_tools/verify_restore_bundle/target/run_results.json"
      ],
      log_paths: ["external_tools/verify_restore_bundle/logs/stdout.log"],
      output_paths: ["external_tools/verify_restore_bundle/target/run_results.json"]
    }
  }
];

export const mockOpsviewAttention: OpsviewAttentionSummary = {
  total_jobs: 2,
  failed_jobs: 1,
  running_jobs: 0,
  succeeded_jobs: 1,
  jobs_missing_logs: 0,
  jobs_missing_outputs: 0
};
