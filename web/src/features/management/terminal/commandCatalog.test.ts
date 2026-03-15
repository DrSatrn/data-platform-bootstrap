import { describe, expect, it } from "vitest";

import { defaultCommandCatalog } from "../mockControlPlane";
import { filterCommands, groupCommandsByArea, roleCanExecute, topCommandSuggestions } from "./commandCatalog";

describe("commandCatalog", () => {
  it("groups commands by area", () => {
    const groups = groupCommandsByArea(defaultCommandCatalog);

    expect(groups.length).toBeGreaterThan(1);
    expect(groups.some((group) => group.area === "recovery")).toBe(true);
  });

  it("filters commands by label, command, or summary", () => {
    const result = filterCommands(defaultCommandCatalog, "recovery bundle");

    expect(result.map((entry) => entry.id)).toContain("backup-create");
  });

  it("applies role checks correctly", () => {
    expect(roleCanExecute("admin", "admin")).toBe(true);
    expect(roleCanExecute("editor", "admin")).toBe(false);
    expect(roleCanExecute("viewer", "viewer")).toBe(true);
  });

  it("prioritizes recent commands in suggestions", () => {
    const suggestions = topCommandSuggestions(defaultCommandCatalog, ["quality", "status"], "admin", 3);

    expect(suggestions[0]?.command).toBe("quality");
    expect(suggestions[1]?.command).toBe("status");
  });
});
