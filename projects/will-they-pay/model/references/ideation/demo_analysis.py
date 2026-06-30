#!/usr/bin/env python3
"""
Demo analysis: Calculate revenue loss from unpaid invoices
This shows the business value of the "Will They Pay?" model
"""

import pandas as pd
import json
from pathlib import Path

# Load data
data_dir = Path("model/data/raw/invoice_cohorts")
invoices = pd.read_csv(data_dir / "invoices_2025-04-20.csv")
payment_rows = pd.read_csv(data_dir / "payment_rows_2025-04-20.csv")
person_history = pd.read_csv(data_dir / "person_invoice_history_2025-04-20.csv")

print("=" * 70)
print("WILL THEY PAY? - Revenue Loss Analysis")
print("=" * 70)
print()

# Basic stats
print(f"📊 DATA SNAPSHOT")
print(f"  Total invoices: {len(invoices):,}")
print(f"  Total payment records: {len(payment_rows):,}")
print(f"  Unique people: {invoices['personid'].nunique():,}")
print()

# Find invoices without payments
invoices_with_payment = set(payment_rows['invoiceid'].unique())
unpaid_invoices = invoices[~invoices['id'].isin(invoices_with_payment)]

print(f"⚠️  UNPAID INVOICES")
print(f"  Count: {len(unpaid_invoices):,} invoices ({len(unpaid_invoices)/len(invoices)*100:.1f}%)")
print(f"  Unique people affected: {unpaid_invoices['personid'].nunique():,}")
print()

# Calculate revenue at risk (convert cents to dollars)
total_unpaid = unpaid_invoices['amount'].sum() / 100
avg_unpaid_invoice = unpaid_invoices['amount'].mean() / 100
median_unpaid_invoice = unpaid_invoices['amount'].median() / 100
max_unpaid_invoice = unpaid_invoices['amount'].max() / 100

print(f"💰 REVENUE AT RISK")
print(f"  Total: ${total_unpaid:,.2f}")
print(f"  Average per invoice: ${avg_unpaid_invoice:,.2f}")
print(f"  Median per invoice: ${median_unpaid_invoice:,.2f}")
print(f"  Max single invoice: ${max_unpaid_invoice:,.2f}")
print()

# Distribution analysis
print(f"📈 DISTRIBUTION")
unpaid_by_person = unpaid_invoices.groupby('personid')['amount'].agg(['sum', 'count']).reset_index()
print(f"  Avg unpaid amount per person: ${unpaid_by_person['sum'].mean()/100:,.2f}")
print(f"  Avg invoices per person: {unpaid_by_person['count'].mean():.1f}")
print(f"  Max owed by one person: ${unpaid_by_person['sum'].max()/100:,.2f}")
print()

# Calculate if we could predict just 30% of the highest-risk unpaid invoices
high_value_unpaid = unpaid_invoices.nlargest(int(len(unpaid_invoices) * 0.3), 'amount')
recoverable = high_value_unpaid['amount'].sum() / 100

print(f"💡 POTENTIAL IMPACT")
print(f"  If we could predict & collect just top 30% of unpaid invoices:")
print(f"    → Revenue recovered: ${recoverable:,.2f}")
print(f"    → People positively impacted: {high_value_unpaid['personid'].nunique():,}")
print(f"    → % of total loss recovered: {recoverable/total_unpaid*100:.0f}%")
print()

# High-risk customer segments
print(f"🎯 HIGH-RISK SEGMENTS")
paid_invoices = invoices[invoices['id'].isin(invoices_with_payment)]
print(f"  Invoices that got paid: {len(paid_invoices):,}")
print(f"  Pay rate: {len(paid_invoices)/len(invoices)*100:.1f}%")
print(f"  Non-pay rate: {len(unpaid_invoices)/len(invoices)*100:.1f}%")
print()

# The ask
print(f"🎯 THE OPPORTUNITY")
print(f"  → Build a predictive model that identifies high-risk non-payers")
print(f"  → Use: patient demographics + payment history + zipcode income data")
print(f"  → Output: Risk score + confidence for each open balance")
print(f"  → Benefit: Proactive collection via calls/plans for high-risk cases")
print(f"  → Expected ROI: Can recover {recoverable/total_unpaid*100:.0f}% of ${total_unpaid:,.2f}")
print()

print("=" * 70)
print(f"BOTTOM LINE: ${total_unpaid:,.2f} revenue at risk across {unpaid_invoices['personid'].nunique():,} people")
print("=" * 70)
