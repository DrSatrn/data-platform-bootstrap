import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

import { ManagementConsolePreview } from "./ManagementConsolePreview";

describe("ManagementConsolePreview", () => {
  it("renders the unwired composite management console staging surface", () => {
    const html = renderToStaticMarkup(<ManagementConsolePreview />);

    expect(html).toContain("Management Console Preview");
    expect(html).toContain("Operator Workbench");
    expect(html).toContain("Control Plane Workspace");
    expect(html).toContain("External Tool Runs");
    expect(html).toContain("Operator Session Deck");
    expect(html).toContain("Backend Read-Model Snapshot");
    expect(html).toContain("Operator Runbooks");
    expect(html).toContain("Operator Evidence");
    expect(html).toContain("Operator Follow-up Board");
  });
});
