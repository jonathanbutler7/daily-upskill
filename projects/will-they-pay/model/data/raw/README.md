# Raw Data

This folder is the source snapshot layer pulled from real production database queries.

## Files

- `invoice_cohorts/invoices_*.csv`: raw invoice rows from `payments.invoices` table
- `invoice_cohorts/payment_rows_*.csv`: payment_log table rows for those invoices.
- `invoice_cohorts/person_invoice_history_*.csv`: full invoice history for the people in scope.

## Quick Notes

- Files are date-stamped snapshots. Keep them immutable.
- Refresh with scripts in `model/scripts/` (see `model/scripts/README.md`).
- Feature engineering starts in `model/notebooks/production/build_training_data_prod.ipynb`.
- All the queries are scoped to one location, for the purpose of getting a good amount of history for a person
