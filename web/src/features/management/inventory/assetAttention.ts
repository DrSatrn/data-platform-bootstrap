import type { AssetAttentionCard } from "../types";

export type AttentionSummary = {
  totalAssets: number;
  fresh: number;
  late: number;
  stale: number;
  missing: number;
  undocumented: number;
  missingQuality: number;
};

export function summarizeAttention(assets: AssetAttentionCard[]): AttentionSummary {
  const summary: AttentionSummary = {
    totalAssets: assets.length,
    fresh: 0,
    late: 0,
    stale: 0,
    missing: 0,
    undocumented: 0,
    missingQuality: 0
  };

  for (const asset of assets) {
    switch (asset.freshnessState) {
      case "fresh":
        summary.fresh++;
        break;
      case "late":
        summary.late++;
        break;
      case "stale":
        summary.stale++;
        break;
      case "missing":
        summary.missing++;
        break;
      default:
        break;
    }

    if (!asset.hasDocs) {
      summary.undocumented++;
    }
    if (!asset.hasQuality) {
      summary.missingQuality++;
    }
  }

  return summary;
}

export function buildAttentionQueue(assets: AssetAttentionCard[]) {
  const severity = (asset: AssetAttentionCard) => {
    switch (asset.freshnessState) {
      case "missing":
        return 4;
      case "stale":
        return 3;
      case "late":
        return 2;
      case "unknown":
        return 1;
      default:
        return 0;
    }
  };

  return [...assets].sort((left, right) => {
    const severityDelta = severity(right) - severity(left);
    if (severityDelta !== 0) {
      return severityDelta;
    }
    const docsDelta = Number(left.hasDocs) - Number(right.hasDocs);
    if (docsDelta !== 0) {
      return docsDelta;
    }
    const qualityDelta = Number(left.hasQuality) - Number(right.hasQuality);
    if (qualityDelta !== 0) {
      return qualityDelta;
    }
    return left.assetID.localeCompare(right.assetID);
  });
}
