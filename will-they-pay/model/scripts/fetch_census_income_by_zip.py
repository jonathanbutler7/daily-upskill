#!/usr/bin/env python3
"""Fetch Census ACS income data for ZIPs from USPS ZIP detail files.

Default behavior:
- Reads ZIPs from data/external/zip_locale_detail/ZIP_DETAIL-Table 1.csv using DELIVERY ZIPCODE.
- Fetches ACS profile fields for all ZCTAs in one request.
- Writes one output row per input ZIP with a census_match flag.

Useful references:
- Census profile table explorer: https://data.census.gov/table/ACSDP1Y2023.DP03
- ACS API field list (2023 ACS5): https://api.census.gov/data/2023/acs/acs5.html
- Quick explainer video: https://www.youtube.com/watch?v=A2HrsOS8omI

Sample single-ZIP request (with API key):
curl --location \
    'https://api.census.gov/data/2023/acs/acs5/profile?get=NAME,DP03_0062E,DP03_0063E&for=zip%20code%20tabulation%20area:78664&key=${CENSUS_API_KEY}'
"""

from __future__ import annotations

import argparse
import csv
import json
import os
import sys
from json import JSONDecodeError
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Dict, Iterable, List, Set

BASE_URL = "https://api.census.gov/data/2023/acs/acs5/profile"
DEFAULT_FIELDS = ["DP03_0062E", "DP03_0063E"]  # median + mean household income
FIELD_ALIASES = {
    "DP03_0062E": "median_household_income",
    "DP03_0063E": "mean_household_income",
    "DP03_0128PE": "poverty_rate_pct",
    "DP03_0005PE": "unemployment_rate_pct",
    "DP03_0097PE": "private_insurance_coverage_pct",
    "DP03_0098PE": "public_insurance_coverage_pct",
    "DP04_0134E": "median_gross_rent",
    "DP04_0089E": "median_home_value_owner_occupied",
    "DP02_0016E": "average_household_size_us",
    "DP02PR_0016E": "average_household_size_pr",
    "DP02_0068PE": "bachelors_or_higher_pct_us",
    "DP02PR_0068PE": "bachelors_or_higher_pct_pr",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Fetch Census income data for ZIPs in a local CSV file."
    )
    parser.add_argument(
        "--zip-file",
        default="data/external/zip_locale_detail/ZIP_DETAIL-Table 1.csv",
        help="Input CSV containing ZIP codes.",
    )
    parser.add_argument(
        "--zip-column",
        default="DELIVERY ZIPCODE",
        help="Column name in --zip-file that contains 5-digit ZIP codes.",
    )
    parser.add_argument(
        "--fields",
        default=",".join(DEFAULT_FIELDS),
        help="Comma-separated Census field codes to fetch.",
    )
    parser.add_argument(
        "--output",
        default="data/processed/census_zip_features_raw.csv",
        help="Output CSV path.",
    )
    parser.add_argument(
        "--year",
        default="2023",
        help="ACS year used in output metadata only.",
    )
    parser.add_argument(
        "--progress-every",
        type=int,
        default=1000,
        help="Print progress every N processed rows for load/write stages.",
    )
    parser.add_argument(
        "--raw-field-names",
        action="store_true",
        help="Use raw Census field codes as CSV headers instead of friendly names.",
    )
    return parser.parse_args()


def log_progress(message: str) -> None:
    print(message, flush=True)


def load_zip_codes(zip_file: Path, zip_column: str, progress_every: int) -> List[str]:
    zip_set: Set[str] = set()
    processed = 0

    log_progress(f"[1/3] Loading ZIPs from {zip_file}...")

    with zip_file.open("r", newline="", encoding="utf-8-sig") as f:
        reader = csv.DictReader(f)

        if reader.fieldnames is None or zip_column not in reader.fieldnames:
            available = ", ".join(reader.fieldnames or [])
            raise ValueError(
                f"Column '{zip_column}' not found in {zip_file}. Available columns: {available}"
            )

        for row in reader:
            processed += 1
            raw_zip = (row.get(zip_column) or "").strip()
            if len(raw_zip) == 5 and raw_zip.isdigit():
                zip_set.add(raw_zip)
            if progress_every > 0 and processed % progress_every == 0:
                log_progress(
                    f"  - load progress: {processed} rows scanned, {len(zip_set)} unique ZIPs"
                )

    log_progress(
        f"[1/3] ZIP load complete: {processed} rows scanned, {len(zip_set)} unique ZIPs"
    )

    return sorted(zip_set)


def fetch_census_rows(fields: Iterable[str], api_key: str | None) -> Dict[str, Dict[str, str]]:
    field_list = [field.strip() for field in fields if field.strip()]
    if not field_list:
        raise ValueError("At least one Census field must be provided.")

    params = {
        "get": ",".join(["NAME", *field_list]),
        "for": "zip code tabulation area:*",
    }
    if api_key:
        params["key"] = api_key

    query = urllib.parse.urlencode(params)
    url = f"{BASE_URL}?{query}"
    req = urllib.request.Request(url, method="GET")
    log_progress("[2/3] Requesting Census ACS data for all ZCTAs...")
    with urllib.request.urlopen(req, timeout=90) as response:
        body_text = response.read().decode("utf-8", errors="replace")

    try:
        payload = json.loads(body_text)
    except JSONDecodeError as exc:
        preview = body_text.strip().replace("\n", " ")[:300]
        raise RuntimeError(
            f"Census response was not valid JSON. Response preview: {preview}"
        ) from exc

    if not payload or len(payload) < 2:
        raise RuntimeError("Census API returned no data rows.")

    headers = payload[0]
    try:
        zcta_col = "zip code tabulation area"
        zcta_idx = headers.index(zcta_col)
    except ValueError as exc:
        raise RuntimeError(
            f"Unexpected Census response format. Missing '{zcta_col}' column."
        ) from exc

    rows_by_zip: Dict[str, Dict[str, str]] = {}
    for row in payload[1:]:
        if len(row) != len(headers):
            continue
        row_dict = dict(zip(headers, row))
        zip_code = row[zcta_idx].strip()
        if len(zip_code) == 5 and zip_code.isdigit():
            rows_by_zip[zip_code] = row_dict

    log_progress(f"[2/3] Census fetch complete: {len(rows_by_zip)} ZIP rows available")

    return rows_by_zip


def write_output(
    output_file: Path,
    input_zips: List[str],
    census_rows_by_zip: Dict[str, Dict[str, str]],
    fields: List[str],
    year: str,
    progress_every: int,
    use_raw_field_names: bool,
) -> None:
    field_headers = []
    for field in fields:
        if use_raw_field_names:
            field_headers.append(field)
        else:
            field_headers.append(FIELD_ALIASES.get(field, field.lower()))

    output_headers = [
        "zip_code",
        "census_name",
        *field_headers,
        "census_match",
        "acs_year",
    ]

    with output_file.open("w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=output_headers)
        writer.writeheader()

        total = len(input_zips)
        log_progress(f"[3/3] Writing output CSV to {output_file} ({total} ZIPs)...")

        for i, zip_code in enumerate(input_zips, start=1):
            row = census_rows_by_zip.get(zip_code, {})
            out = {
                "zip_code": zip_code,
                "census_name": row.get("NAME", ""),
                "census_match": "1" if row else "0",
                "acs_year": year,
            }
            for field, header in zip(fields, field_headers):
                out[header] = row.get(field, "")
            writer.writerow(out)

            if progress_every > 0 and (i % progress_every == 0 or i == total):
                pct = (i / total) * 100 if total else 100.0
                log_progress(f"  - write progress: {i}/{total} rows ({pct:.1f}%)")

    log_progress("[3/3] CSV write complete.")


def main() -> int:
    args = parse_args()
    zip_file = Path(args.zip_file)
    output_file = Path(args.output)
    fields = [field.strip() for field in args.fields.split(",") if field.strip()]

    if not zip_file.exists():
        print(f"Input ZIP file not found: {zip_file}", file=sys.stderr)
        return 1

    try:
        input_zips = load_zip_codes(zip_file, args.zip_column, args.progress_every)
    except Exception as exc:
        print(f"Failed to parse ZIP file: {exc}", file=sys.stderr)
        return 1

    if not input_zips:
        print("No valid 5-digit ZIP codes found in input file.", file=sys.stderr)
        return 1

    api_key = os.getenv("CENSUS_API_KEY")

    try:
        census_rows = fetch_census_rows(fields, api_key)
    except urllib.error.HTTPError as exc:
        print(f"Census request failed: HTTP {exc.code} {exc.reason}", file=sys.stderr)
        try:
            print(exc.read().decode("utf-8", errors="replace"), file=sys.stderr)
        except Exception:
            pass
        return 1
    except urllib.error.URLError as exc:
        print(f"Census request failed: {exc}", file=sys.stderr)
        return 1
    except Exception as exc:
        print(f"Failed to fetch Census data: {exc}", file=sys.stderr)
        return 1

    try:
        write_output(
            output_file,
            input_zips,
            census_rows,
            fields,
            args.year,
            args.progress_every,
            args.raw_field_names,
        )
    except Exception as exc:
        print(f"Failed to write output CSV: {exc}", file=sys.stderr)
        return 1

    matched = sum(1 for z in input_zips if z in census_rows)
    missing = len(input_zips) - matched

    print(f"Input ZIP count: {len(input_zips)}")
    print(f"Matched Census ZCTA rows: {matched}")
    print(f"Missing Census matches: {missing}")
    print(f"Output written to: {output_file}")
    if not api_key:
        print("Note: CENSUS_API_KEY not set. Request was sent without an API key.")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
