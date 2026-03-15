import { describe, expect, it } from "vitest";

import { mockAttentionAssets } from "../mockControlPlane";
import { buildAttentionQueue, summarizeAttention } from "./assetAttention";

describe("assetAttention", () => {
  it("summarizes freshness and coverage gaps", () => {
    const summary = summarizeAttention(mockAttentionAssets);

    expect(summary.totalAssets).toBe(3);
    expect(summary.late).toBe(1);
    expect(summary.missing).toBe(1);
    expect(summary.undocumented).toBe(1);
    expect(summary.missingQuality).toBe(3);
  });

  it("orders the attention queue by freshness severity first", () => {
    const queue = buildAttentionQueue(mockAttentionAssets);

    expect(queue[0]?.assetID).toBe("raw_account_balances");
    expect(queue[1]?.assetID).toBe("mart_budget_vs_actual");
  });
});
