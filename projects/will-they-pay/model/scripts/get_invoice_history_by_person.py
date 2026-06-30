#!/usr/bin/env python3
"""Export raw invoice history for person IDs derived from an invoice CSV.

Flow:
1) Read person IDs from an input invoice CSV.
2) Query payments.invoices by personid in batches.
3) Write raw invoice history rows to output CSV.

No label logic, no renaming, no feature transformations.
"""

from __future__ import annotations

import argparse
import csv
import os
import time
from pathlib import Path
from typing import Iterable, List

try:
    import psycopg2
except ImportError as exc:  # pragma: no cover
    raise SystemExit(
        "psycopg2 is required. Install with: pip install psycopg2-binary"
    ) from exc


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Export raw invoice history from payments.invoices for person IDs "
            "found in a seed invoice CSV."
        )
    )
    parser.add_argument(
        "--input",
        default="data/raw/invoice_cohorts/invoices_seed.csv",
        help="Seed invoice CSV path.",
    )
    parser.add_argument(
        "--output",
        default="data/raw/person_invoice_history.csv",
        help="Output CSV path for invoice history rows.",
    )
    parser.add_argument(
        "--person-id-column",
        default="personid",
        help=(
            "Person ID column name in input CSV. "
            "Use 'personid' for DB invoice exports or 'person_id_obfuscated' for local datasets."
        ),
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=100,
        help="Number of person IDs per DB query.",
    )
    parser.add_argument(
        "--batch-delay-seconds",
        type=float,
        default=1.0,
        help="Delay between DB batch queries.",
    )
    parser.add_argument(
        "--dsn",
        default=os.getenv("DATABASE_URL", ""),
        help=(
            "Postgres DSN. If omitted, script uses PGHOST/PGPORT/PGDATABASE/"
            "PGUSER/PGPASSWORD env vars."
        ),
    )
    return parser.parse_args()


def chunked(values: List[str], size: int) -> Iterable[List[str]]:
    for i in range(0, len(values), size):
        yield values[i : i + size]


def get_connection(args: argparse.Namespace):
    if args.dsn:
        return psycopg2.connect(args.dsn)
    return psycopg2.connect(
        host="localhost",
        port="5433",
        dbname="payments",
        user="postgres",
        password=os.getenv("PGPASSWORD"),
    )


def load_person_ids(input_path: Path, person_id_column: str) -> List[str]:
    with input_path.open("r", newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        fieldnames = reader.fieldnames or []
        if person_id_column not in fieldnames:
            raise ValueError(
                f"Person ID column '{person_id_column}' not found in input CSV."
            )

        person_ids = [
            (row.get(person_id_column) or "").strip()
            for row in reader
            if (row.get(person_id_column) or "").strip()
        ]

    # Preserve order while deduping.
    unique_ids = list(dict.fromkeys(person_ids))
    return unique_ids


def fetch_invoice_history_rows(
    conn,
    person_ids: List[str],
    batch_size: int,
    batch_delay_seconds: float,
):
    query = """
        SELECT
            id,
            merchantid,
            personid,
            amount,
            createdat,
            expiresat,
            status,
            hasattachment,
            surchargingenabled,
            providername,
            userid,
            username
        FROM payments.invoices
        WHERE personid = ANY(%s::uuid[])
        ORDER BY personid ASC, createdat ASC
    """

    rows_out = []
    headers = None

    total_batches = (len(person_ids) + batch_size - 1) // batch_size
    print(f"Fetching invoice history in {total_batches} batches...", flush=True)

    with conn.cursor() as cur:
        for idx, batch in enumerate(chunked(person_ids, batch_size), start=1):
            cur.execute(query, (batch,))
            batch_rows = cur.fetchall()

            if headers is None:
                headers = [d.name for d in cur.description]

            rows_out.extend(batch_rows)

            print(
                f"  batch {idx}/{total_batches}: queried {len(batch)} person IDs, "
                f"returned {len(batch_rows)} invoice rows",
                flush=True,
            )

            if idx < total_batches and batch_delay_seconds > 0:
                time.sleep(batch_delay_seconds)

    if headers is None:
        headers = []

    return headers, rows_out


def main() -> int:
    args = parse_args()
    input_path = Path(args.input)
    output_path = Path(args.output)

    if not input_path.exists():
        raise FileNotFoundError(f"Input CSV not found: {input_path}")

    if args.batch_size <= 0:
        raise ValueError("--batch-size must be > 0")

    person_ids = load_person_ids(input_path, args.person_id_column)
    print(f"Loaded {len(person_ids)} unique person IDs from {input_path}", flush=True)

    conn = get_connection(args)
    try:
        headers, rows = fetch_invoice_history_rows(
            conn=conn,
            person_ids=person_ids,
            batch_size=args.batch_size,
            batch_delay_seconds=args.batch_delay_seconds,
        )
    finally:
        conn.close()

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        if headers:
            writer.writerow(headers)
        writer.writerows(rows)

    print(f"Wrote raw invoice history CSV: {output_path}")
    print(f"Rows written: {len(rows)}")
    print(
        "Next: run scripts/get_payment_log_data.py with --input pointing to this output and --invoice-id-column id",
        flush=True,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
