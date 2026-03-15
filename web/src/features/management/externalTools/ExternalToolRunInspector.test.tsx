import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

import { ExternalToolRunInspector } from "./ExternalToolRunInspector";

describe("ExternalToolRunInspector", () => {
  it("renders grouped external tool visibility for operators", () => {
    const html = renderToStaticMarkup(
      <ExternalToolRunInspector
        events={[
          {
            time: "2026-03-15T00:00:00Z",
            level: "info",
            message: "external tool command started",
            fields: { job_id: "build_finance_dbt", tool: "dbt", action: "build" }
          },
          {
            time: "2026-03-15T00:00:02Z",
            level: "info",
            message: "external tool command finished",
            fields: { job_id: "build_finance_dbt", tool: "dbt", action: "build" }
          }
        ]}
        artifacts={[
          {
            run_id: "run_1",
            relative_path: "external_tools/build_finance_dbt/logs/stdout.log",
            size_bytes: 10,
            modified_at: "2026-03-15T00:00:01Z",
            content_type: "application/octet-stream"
          },
          {
            run_id: "run_1",
            relative_path: "external_tools/build_finance_dbt/target/run_results.json",
            size_bytes: 64,
            modified_at: "2026-03-15T00:00:02Z",
            content_type: "application/json"
          }
        ]}
      />
    );

    expect(html).toContain("External Tool Runs");
    expect(html).toContain("build_finance_dbt");
    expect(html).toContain("dbt build");
    expect(html).toContain("succeeded");
    expect(html).toContain("stdout.log");
    expect(html).toContain("run_results.json");
  });
});
