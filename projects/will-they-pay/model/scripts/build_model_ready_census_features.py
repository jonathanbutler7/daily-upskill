#!/usr/bin/env python3
"""Build model-ready Census ZIP features from the raw Census output CSV."""

from __future__ import annotations

import argparse
import csv
from pathlib import Path

MISSING_MARKERS = {"", "null", "-666666666", "-999999999", "N", "(X)"}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Transform raw Census ZIP features into a model-ready table."
    )
    parser.add_argument(
        "--input",
        default="data/processed/census_zip_features_raw.csv",
        help="Raw Census CSV path.",
    )
    parser.add_argument(
        "--output",
        default="data/processed/census_zip_features_model_ready.csv",
        help="Model-ready CSV path.",
    )
    return parser.parse_args()


def to_num(value: str) -> str:
    text = (value or "").strip()
    if text in MISSING_MARKERS:
        return ""
    try:
        return str(float(text))
    except ValueError:
        return ""


def pick_numeric(*values: str) -> str:
    for value in values:
        parsed = to_num(value)
        if parsed != "":
            return parsed
    return ""


def build_model_ready(input_path: Path, output_path: Path) -> tuple[int, int, int]:
    fieldnames = [
        "zip_code",
        "census_name",
        "median_household_income",
        "mean_household_income",
        "poverty_rate_pct",
        "unemployment_rate_pct",
        "private_insurance_coverage_pct",
        "public_insurance_coverage_pct",
        "median_gross_rent",
        "median_home_value_owner_occupied",
        "average_household_size",
        "bachelors_or_higher_pct",
        "census_match",
        "acs_year",
    ]

    rows = 0
    matched = 0

    with input_path.open("r", newline="", encoding="utf-8") as src, output_path.open(
        "w", newline="", encoding="utf-8"
    ) as dst:
        reader = csv.DictReader(src)
        writer = csv.DictWriter(dst, fieldnames=fieldnames)
        writer.writeheader()

        for row in reader:
            rows += 1
            output_row = {
                "zip_code": (row.get("zip_code") or "").strip(),
                "census_name": (row.get("census_name") or "").strip(),
                "median_household_income": to_num(row.get("median_household_income") or ""),
                "mean_household_income": to_num(row.get("mean_household_income") or ""),
                "poverty_rate_pct": to_num(row.get("poverty_rate_pct") or ""),
                "unemployment_rate_pct": to_num(row.get("unemployment_rate_pct") or ""),
                "private_insurance_coverage_pct": to_num(
                    row.get("private_insurance_coverage_pct") or ""
                ),
                "public_insurance_coverage_pct": to_num(
                    row.get("public_insurance_coverage_pct") or ""
                ),
                "median_gross_rent": to_num(row.get("median_gross_rent") or ""),
                "median_home_value_owner_occupied": to_num(
                    row.get("median_home_value_owner_occupied") or ""
                ),
                "average_household_size": pick_numeric(
                    row.get("average_household_size_us") or "",
                    row.get("average_household_size_pr") or "",
                ),
                "bachelors_or_higher_pct": pick_numeric(
                    row.get("bachelors_or_higher_pct_us") or "",
                    row.get("bachelors_or_higher_pct_pr") or "",
                ),
                "census_match": "1"
                if (row.get("census_match") or "").strip() == "1"
                else "0",
                "acs_year": (row.get("acs_year") or "").strip(),
            }

            matched += 1 if output_row["census_match"] == "1" else 0
            writer.writerow(output_row)

    return rows, matched, rows - matched


def main() -> int:
    args = parse_args()
    input_path = Path(args.input)
    output_path = Path(args.output)

    if not input_path.exists():
        raise FileNotFoundError(f"Raw input file does not exist: {input_path}")

    output_path.parent.mkdir(parents=True, exist_ok=True)
    rows, matched, unmatched = build_model_ready(input_path, output_path)

    print(f"Model-ready file written: {output_path}")
    print(f"Rows: {rows}")
    print(f"Matched rows: {matched}")
    print(f"Unmatched rows: {unmatched}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
