import type { ExternalToolArtifact, ExternalToolEvent } from "../externalTools/externalToolRunSummary";

export const previewExternalToolEvents: ExternalToolEvent[] = [
  {
    time: "2026-03-15T14:20:00Z",
    level: "info",
    message: "external tool command started",
    fields: {
      job_id: "build_finance_dbt",
      tool: "dbt",
      action: "build"
    }
  },
  {
    time: "2026-03-15T14:20:04Z",
    level: "error",
    message: "external tool failed",
    fields: {
      job_id: "build_finance_dbt",
      tool: "dbt",
      action: "build"
    }
  },
  {
    time: "2026-03-15T14:24:00Z",
    level: "info",
    message: "external tool command started",
    fields: {
      job_id: "verify_restore_bundle",
      tool: "dbt",
      action: "build"
    }
  },
  {
    time: "2026-03-15T14:24:05Z",
    level: "info",
    message: "external tool command finished",
    fields: {
      job_id: "verify_restore_bundle",
      tool: "dbt",
      action: "build"
    }
  }
];

export const previewExternalToolArtifacts: ExternalToolArtifact[] = [
  {
    run_id: "run_1",
    relative_path: "external_tools/build_finance_dbt/logs/stdout.log",
    size_bytes: 1240,
    modified_at: "2026-03-15T14:20:02Z",
    content_type: "application/octet-stream"
  },
  {
    run_id: "run_1",
    relative_path: "external_tools/build_finance_dbt/logs/stderr.log",
    size_bytes: 488,
    modified_at: "2026-03-15T14:20:04Z",
    content_type: "application/octet-stream"
  },
  {
    run_id: "run_2",
    relative_path: "external_tools/verify_restore_bundle/logs/stdout.log",
    size_bytes: 216,
    modified_at: "2026-03-15T14:24:05Z",
    content_type: "application/octet-stream"
  },
  {
    run_id: "run_2",
    relative_path: "external_tools/verify_restore_bundle/target/run_results.json",
    size_bytes: 96,
    modified_at: "2026-03-15T14:24:05Z",
    content_type: "application/json"
  }
];
