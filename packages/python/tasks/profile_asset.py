#!/usr/bin/env python3
"""
This utility profiles a materialized asset for the catalog detail page. It is
intentionally lightweight: a few high-signal statistics per column, plus row
count and file format, so operators can reason about dataset quality without
leaving the platform.
"""

from __future__ import annotations

import csv
import json
import os
from collections import OrderedDict
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


def main() -> None:
    request_path = os.environ["PLATFORM_TASK_REQUEST_PATH"]
    result_path = os.environ["PLATFORM_TASK_RESULT_PATH"]

    with open(request_path, "r", encoding="utf-8") as handle:
        request = json.load(handle)

    source_path = Path(request["source_path"])
    rows, fmt = load_rows(source_path)
    columns = build_column_profiles(rows)
    result = {
        "format": fmt,
        "row_count": len(rows),
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "columns": columns,
    }

    with open(result_path, "w", encoding="utf-8") as handle:
        json.dump(result, handle, indent=2)


def load_rows(source_path: Path) -> tuple[list[dict[str, Any]], str]:
    if source_path.suffix.lower() == ".csv":
        with open(source_path, "r", encoding="utf-8") as handle:
            return list(csv.DictReader(handle)), "csv"

    with open(source_path, "r", encoding="utf-8") as handle:
        payload = json.load(handle)

    if isinstance(payload, list):
        rows = [row for row in payload if isinstance(row, dict)]
    elif isinstance(payload, dict):
        if isinstance(payload.get("series"), list):
            rows = [row for row in payload["series"] if isinstance(row, dict)]
        else:
            rows = [payload]
    else:
        rows = []
    return rows, "json"


def build_column_profiles(rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    columns: OrderedDict[str, list[Any]] = OrderedDict()
    for row in rows:
        for key in row.keys():
            columns.setdefault(key, [])
        for key in columns.keys():
            columns[key].append(row.get(key))

    return [profile_column(name, values) for name, values in columns.items()]


def profile_column(name: str, values: list[Any]) -> dict[str, Any]:
    non_null = [value for value in values if value not in (None, "", [])]
    samples = []
    seen_samples = set()
    for value in non_null:
        rendered = render_value(value)
        if rendered in seen_samples:
            continue
        seen_samples.add(rendered)
        samples.append(rendered)
        if len(samples) == 3:
            break

    observed_type = "unknown"
    if non_null:
        observed_type = infer_type(non_null)

    scalar_values = [render_value(value) for value in non_null]
    numeric_values = [coerce_number(value) for value in non_null]
    numeric_values = [value for value in numeric_values if value is not None]

    profile = {
        "name": name,
        "observed_type": observed_type,
        "null_count": len(values) - len(non_null),
        "unique_count": len({render_value(value) for value in non_null}),
        "sample_values": samples,
    }

    if numeric_values:
        profile["min_value"] = render_value(min(numeric_values))
        profile["max_value"] = render_value(max(numeric_values))
    elif scalar_values:
        profile["min_value"] = min(scalar_values)
        profile["max_value"] = max(scalar_values)
    return profile


def infer_type(values: list[Any]) -> str:
    if all(isinstance(value, bool) for value in values):
        return "boolean"
    if all(coerce_number(value) is not None for value in values):
        return "number"
    if all(isinstance(value, (dict, list)) for value in values):
        return "json"
    return "string"


def coerce_number(value: Any) -> float | None:
    if isinstance(value, bool):
        return None
    if isinstance(value, (int, float)):
        return float(value)
    if isinstance(value, str):
        try:
            return float(value.strip())
        except ValueError:
            return None
    return None


def render_value(value: Any) -> str:
    if isinstance(value, float) and value.is_integer():
        return str(int(value))
    if isinstance(value, (dict, list)):
        return json.dumps(value, sort_keys=True)
    return str(value)


if __name__ == "__main__":
    main()
