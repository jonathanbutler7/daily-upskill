#!/usr/bin/env python3
"""Export raw payment rows for invoice IDs from a source CSV.

Flow:
1) Read invoice IDs from input CSV.
2) Query payments.invoice_payments by invoiceid.
3) Join to payments.payment_log by paymentid.
4) Write raw query results to output CSV.

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
        description="Export raw payment rows for invoice IDs from a CSV input."
    )
    parser.add_argument(
        "--input",
        default="data/raw/prod_invoices_dataset.csv",
        help="Input invoice CSV path.",
    )
    parser.add_argument(
        "--output",
        default="data/processed/prod_invoice_payment_rows.csv",
        help="Output CSV path.",
    )
    parser.add_argument(
        "--invoice-id-column",
        default="invoice_id",
        help="Invoice ID column name in input CSV.",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=100,
        help="Number of invoice IDs per DB query.",
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


def load_invoice_ids(input_path: Path, invoice_id_column: str) -> List[str]:
    with input_path.open("r", newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        fieldnames = reader.fieldnames or []
        if invoice_id_column not in fieldnames:
            raise ValueError(
                f"Invoice ID column '{invoice_id_column}' not found in input CSV."
            )

        invoice_ids = [
            (row.get(invoice_id_column) or "").strip()
            for row in reader
            if (row.get(invoice_id_column) or "").strip()
        ]

    return invoice_ids


def fetch_rows(
    conn,
    invoice_ids: List[str],
    batch_size: int,
    batch_delay_seconds: float,
):
    query = """
        SELECT
            ip.invoiceid,
            ip.paymentid,
            pl.id,
            pl.merchantid,
            pl.amount,
            pl.weavefee,
            pl.receiptemail,
            pl.paymentstatus,
            pl.statusreason,
            pl.confirmationcode,
            pl.submittedat,
            pl.createdat,
            pl.recordedat,
            pl.processorid,
            pl.paymenttype,
            pl.pricingid,
            pl.pricingrate,
            pl.pricingtransactioncost,
            pl.updatedat,
            pl.origin,
            pl.popupnotificationsent,
            pl.userid,
            pl.expires_at,
            pl.processortype
        FROM payments.invoice_payments ip
        JOIN payments.payment_log pl
          ON pl.id = ip.paymentid
        WHERE ip.invoiceid = ANY(%s::uuid[])
    """

    rows_out = []
    headers = None

    total_batches = (len(invoice_ids) + batch_size - 1) // batch_size
    print(f"Fetching rows in {total_batches} batches...", flush=True)

    with conn.cursor() as cur:
        for idx, batch in enumerate(chunked(invoice_ids, batch_size), start=1):
            cur.execute(query, (batch,))
            batch_rows = cur.fetchall()

            if headers is None:
                headers = [d.name for d in cur.description]

            rows_out.extend(batch_rows)

            print(
                f"  batch {idx}/{total_batches}: queried {len(batch)} invoice IDs, "
                f"returned {len(batch_rows)} rows",
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

    invoice_ids = load_invoice_ids(input_path, args.invoice_id_column)
    print(f"Loaded {len(invoice_ids)} invoice IDs from {input_path}", flush=True)
    
    # get one connection to be used by all the batched queries
    conn = get_connection(args)
    try:
        headers, rows = fetch_rows(
            conn=conn,
            invoice_ids=invoice_ids,
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

    print(f"Wrote raw query CSV: {output_path}")
    print(f"Rows written: {len(rows)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
