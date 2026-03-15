import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

import { mockOpsviewAttention, mockOpsviewSummaries } from "./mockOpsview";
import { OpsviewSummaryPanel } from "./OpsviewSummaryPanel";

describe("OpsviewSummaryPanel", () => {
  it("renders backend read-model signals for the staged management console", () => {
    const html = renderToStaticMarkup(
      <OpsviewSummaryPanel attention={mockOpsviewAttention} summaries={mockOpsviewSummaries} />
    );

    expect(html).toContain("Backend Read-Model Snapshot");
    expect(html).toContain("Tracked Jobs");
    expect(html).toContain("Failed Jobs");
    expect(html).toContain("build_finance_dbt");
    expect(html).toContain("execution_failed");
  });
});
