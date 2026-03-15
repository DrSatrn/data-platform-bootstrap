#!/usr/bin/env python3
"""
This task enriches the landed raw transaction CSV into a staging JSON file with
normalized descriptions, resolved categories, and broader category groups. The
Go worker remains in charge of orchestration; this script only owns the bounded
enrichment logic.
"""

from __future__ import annotations

import csv
import json
import os
from pathlib import Path


def load_request() -> dict:
    request_path = Path(os.environ["PLATFORM_TASK_REQUEST_PATH"])
    return json.loads(request_path.read_text())


def infer_category(description: str, original_category: str) -> tuple[str, str, bool]:
    if original_category.strip():
        return original_category.strip(), classify_group(original_category.strip()), False

    normalized = description.lower()
    if "woolworths" in normalized or "coles" in normalized or "aldi" in normalized:
        category = "Groceries"
    elif "uber" in normalized or "translink" in normalized or "opal" in normalized:
        category = "Transport"
    elif "rent" in normalized or "real estate" in normalized:
        category = "Housing"
    elif "netflix" in normalized or "spotify" in normalized:
        category = "Entertainment"
    elif "salary" in normalized or "payroll" in normalized:
        category = "Income"
    else:
        category = "Uncategorized"
    return category, classify_group(category), True


def classify_group(category: str) -> str:
    if category in {"Income"}:
        return "Income"
    if category in {"Groceries", "Dining", "Entertainment"}:
        return "Lifestyle"
    if category in {"Transport"}:
        return "Mobility"
    if category in {"Housing", "Utilities"}:
        return "Essentials"
    return "Other"


def main() -> None:
    request = load_request()
    data_root = Path(request["data_root"])
    source = data_root / "raw" / "raw_transactions.csv"
    target = data_root / "staging" / "staging_transactions_enriched.json"
    target.parent.mkdir(parents=True, exist_ok=True)

    rows = []
    inferred = 0
    with source.open(newline="") as handle:
        reader = csv.DictReader(handle)
        for row in reader:
            normalized_description = " ".join((row.get("description") or row.get("account_name") or row["transaction_id"]).strip().lower().split())
            resolved_category, category_group, was_inferred = infer_category(normalized_description, row.get("category") or "")
            if was_inferred:
                inferred += 1
            rows.append(
                {
                    "transaction_id": row["transaction_id"],
                    "occurred_at": row["occurred_at"],
                    "account_name": row["account_name"],
                    "amount": float(row["amount"]),
                    "category": resolved_category,
                    "category_group": category_group,
                    "normalized_description": normalized_description,
                    "inferred_category": was_inferred,
                }
            )

    target.write_text(json.dumps(rows, indent=2))

    result = {
        "message": "enriched raw transactions into a staging JSON dataset",
        "outputs": [
            {
                "relative_path": "staging/staging_transactions_enriched.json",
                "source_path": str(target),
                "content_type": "application/json",
            }
        ],
        "metadata": {
            "rows_written": len(rows),
            "categories_inferred": inferred,
        },
        "log_lines": [
            f"source={source}",
            f"target={target}",
            f"rows_written={len(rows)}",
            f"categories_inferred={inferred}",
        ],
    }

    Path(os.environ["PLATFORM_TASK_RESULT_PATH"]).write_text(json.dumps(result, indent=2))
    print("python task completed: staging enrichment")


if __name__ == "__main__":
    main()
