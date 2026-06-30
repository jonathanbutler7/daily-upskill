#!/usr/bin/env python3
"""Export raw invoice rows from a start date with row limit.

Flow:
1) Query payments.invoices using createdat >= start_date.
2) Limit result set size for manageable cohort extracts.
3) Write one raw CSV output.
4) Reuse scripts/get_payment_log_data.py on this CSV.

No label logic, no renaming, no feature transformations.
"""

from __future__ import annotations

import argparse
import csv
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence

try:
    import psycopg2
except ImportError as exc:  # pragma: no cover
    raise SystemExit(
        "psycopg2 is required. Install with: pip install psycopg2-binary"
    ) from exc


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Export raw invoice rows from payments.invoices starting at a date."
    )
    parser.add_argument(
        "--output",
        default="data/raw/invoice_cohorts/invoices_start_window.csv",
        help="Output CSV path.",
    )
    parser.add_argument(
        "--start-date",
        required=True,
        help=(
            "Inclusive start date in ISO format (YYYY-MM-DD or full timestamp)."
        ),
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=5000,
        help="Maximum number of rows to export.",
    )
    parser.add_argument(
        "--merchant-id",
        required=True,
        help="Merchant UUID to scope the extract to a single location.",
    )
    parser.add_argument(
        "--status",
        type=int,
        default=None,
        help="Optional invoice status filter.",
    )
    parser.add_argument(
        "--print-explain",
        action="store_true",
        help="Print EXPLAIN plan for the query before executing.",
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


def parse_iso_datetime(value: str) -> datetime:
    if not value:
        raise ValueError("empty datetime value")

    normalized = value.strip().replace("Z", "+00:00")
    dt = datetime.fromisoformat(normalized)
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


def fetch_invoices_from_start(
    conn,
    start: datetime,
    limit: int,
    merchant_id: str,
    status: int | None,
) -> tuple[Sequence[str], Sequence[tuple]]:
    filters = ["createdat >= %s", "createdat <= NOW() - INTERVAL '30 days'"]
    params = [start]

    if merchant_id:
        filters.append("merchantid = %s::uuid")
        params.append(merchant_id)

    where_clause = " AND ".join(filters)

    query = f"""
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
        WHERE {where_clause}
        ORDER BY createdat DESC, id DESC
        LIMIT %s
    """
    params.append(limit)

    with conn.cursor() as cur:
        cur.execute(query, tuple(params))
        rows = cur.fetchall()
        headers = [d.name for d in cur.description]

    return headers, rows


def print_explain_plan(
    conn,
    start: datetime,
    limit: int,
    merchant_id: str,
    status: int | None,
) -> None:
    filters = ["createdat >= %s", "createdat <= NOW() - INTERVAL '30 days'"]
    params = [start]

    if merchant_id:
        filters.append("merchantid = %s::uuid")
        params.append(merchant_id)

    if status is not None:
        filters.append("status = %s")
        params.append(status)

    where_clause = " AND ".join(filters)

    explain_query = f"""
        EXPLAIN
        SELECT id, merchantid, personid, createdat, status
        FROM payments.invoices
        WHERE {where_clause}
        LIMIT %s
    """
    params.append(limit)

    with conn.cursor() as cur:
        cur.execute(explain_query, tuple(params))
        lines = [row[0] for row in cur.fetchall()]

    print("EXPLAIN plan:", flush=True)
    for line in lines:
        print(f"  {line}", flush=True)


def write_csv(output_path: Path, headers: Sequence[str], rows: Sequence[tuple]) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(headers)
        writer.writerows(rows)


def main() -> int:
    args = parse_args()

    if args.limit <= 0:
        raise ValueError("--limit must be > 0")

    output_path = Path(args.output)
    start = parse_iso_datetime(args.start_date)

    conn = get_connection(args)
    try:
        if args.print_explain:
            print_explain_plan(
                conn=conn,
                start=start,
                limit=args.limit,
                merchant_id=args.merchant_id,
                status=args.status,
            )

        headers, rows = fetch_invoices_from_start(
            conn=conn,
            start=start,
            limit=args.limit,
            merchant_id=args.merchant_id,
            status=args.status,
        )
    finally:
        conn.close()

    write_csv(output_path, headers, rows)

    print(f"Start date: {start.isoformat()}", flush=True)
    print(f"Limit: {args.limit}", flush=True)
    if args.merchant_id:
        print(f"Merchant ID: {args.merchant_id}", flush=True)
    if args.status is not None:
        print(f"Status: {args.status}", flush=True)
    print(f"Wrote raw query CSV: {output_path}", flush=True)
    print(f"Rows written: {len(rows)}", flush=True)
    print(
        "Next: run scripts/get_payment_log_data.py with this CSV and "
        "--invoice-id-column id",
        flush=True,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
