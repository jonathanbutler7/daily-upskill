# Scripts

Utility scripts for the Will They Pay data pipeline and model scoring.
Run all scripts from the `model/` directory (or use `make` targets from the repo root).

---

## Data Pipeline (run in order to refresh training data)

### 1. `get_invoice_cohorts_data.py`

Pulls raw invoice rows from `payments.invoices` using:
- `createdat >= --start-date`
- `LIMIT --limit` (default `5000`)
- optional `--merchant-id` scope
- optional `--status` scope

Output: `data/raw/invoice_cohorts/invoices_*.csv`

Run this first, then `get_payment_log_data.py` to fetch matching payment rows.

Tip: pass `--print-explain` to print the query plan and verify index usage.

### 2. `get_payment_log_data.py`

Pulls payment log rows for a set of invoices.

Pass `--invoice-id-column id` to match against the invoice CSV from step 1.

Output: `data/raw/invoice_cohorts/payment_rows_*.csv`

### 3. `get_invoice_history_by_person.py`

Person-centric extractor. Reads a seed invoice CSV, extracts unique person IDs, then pulls
full invoice history for those people from `payments.invoices`.

Use this to make person the primary entity. After exporting, reuse `get_payment_log_data.py`
with `--invoice-id-column id` to fetch payment rows for historical invoices.

Output: `data/raw/invoice_cohorts/person_invoice_history_*.csv`

### 4. `fetch_census_income_by_zip.py`

Fetches ZIP-level demographic features from the Census API.
Output: `data/processed/census_zip_features_raw.csv`

Run via: `make census-raw`

### 5. `build_model_ready_census_features.py`

Transforms raw census output into cleaned, model-ready feature columns.
Output: `data/processed/census_zip_features_model_ready.csv`

Run via: `make census-model`

---

## Scoring Scripts

### `score_batch.py`

Loads `data/processed/training_data_prod.csv` (the flat feature table), filters to open
invoices, runs them through the trained joblib, and writes
`data/output/scored_predictions_prod.csv`.

Prerequisites: run both production training notebooks first.

```bash
make score-batch
# or:
python scripts/score_batch.py
```

### `score_single.py`

Scores a single record against the production model. Useful for spot-checking
the joblib artifact without running the full pipeline.

```bash
make score-single ARGS="--amount 150 --average-days-to-pay 14.5 ..."
# or:
python scripts/score_single.py --help
```