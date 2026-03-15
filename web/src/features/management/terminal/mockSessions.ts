import { completeCommand, createTerminalSession, startCommand, type TerminalSession } from "./sessionModel";

const restoreStarted = startCommand(
  createTerminalSession({
    id: "restore-drill",
    title: "Restore Drill",
    scope: "recovery",
    recommendedRunbook: "/docs/runbooks/backups.md"
  }),
  "backup verify latest.tar.gz",
  "2026-03-15T13:00:00Z"
);

const dbtStarted = startCommand(
  createTerminalSession({
    id: "dbt-finance",
    title: "DBT Finance Build",
    scope: "pipelines",
    recommendedRunbook: "/docs/runbooks/optional-external-tools.md"
  }),
  "dbt build --select monthly_cashflow",
  "2026-03-15T13:04:00Z"
);

export const mockTerminalSessions: TerminalSession[] = [
  completeCommand(restoreStarted, {
    exitCode: 0,
    at: "2026-03-15T13:00:06Z",
    stdout: ["bundle checksum ok", "archive contents verified"],
    artifacts: ["runs/run_1/recovery/verify-report.json"]
  }),
  completeCommand(dbtStarted, {
    exitCode: 2,
    at: "2026-03-15T13:04:09Z",
    stdout: ["1 of 3 models built"],
    stderr: ["monthly_cashflow.sql failed because source relation was missing"]
  })
];
