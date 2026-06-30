# Feature Importance Analysis: Will They Pay? XGBoost Model

**Model:** XGBoost (Production)  
**Test ROC-AUC:** 0.9469  
**Training Records:** 19,480  
**Test Records:** 4,871  

---

## Feature Importance Ranking (by Gain)

| Rank | Feature | Importance | Category | Interpretation |
|---:|---|---:|---|---|
| **1** | `average_days_to_pay` | **18.3%** | Payment History | **Strongest predictor**. Patients with longer historical payment delays are more likely to have unpaid invoices. This signal alone carries the most predictive weight. |
| **2** | `amount` | **14.7%** | Invoice Context | Larger invoice amounts correlate with payment likelihood. High-value invoices may receive prioritized payment or differ in payment dynamics. |
| **3** | `appointment_reliability_score` | **12.6%** | Payment History | Behavioral consistency (appointment attendance/no-shows) is strongly predictive of invoice payment. **⚠️ This feature is currently derived from synthetic dummy data** (see [Model Card](../MODEL_CARD.md#known-limitations)). The signal is a placeholder until real scheduler data is available for training. |
| **4** | `account_age_days` | **10.2%** | Payment History | Established patients (longer tenure) are more reliable payers. New patients show higher default risk. |
| **5** | `tenure_months` | **8.9%** | Payment History | Related to account age; longer relationships predict better payment. |
| **6** | `created_day_of_week` | **7.5%** | Invoice Context | Temporal pattern: invoices created on certain days of the week have different payment rates (e.g., Monday vs. Friday effects). |
| **7** | `created_hour` | **6.4%** | Invoice Context | Time-of-day effect on payment behavior (e.g., business hours vs. after-hours). |
| **8** | `is_new_patient` | **5.8%** | Patient Profile | New patients flag; captures risk tier. |
| **9** | `payment_origin_mode` | **4.2%** | Invoice Context | How payment was initiated (online, phone, automated) affects collection risk. |
| **10** | `median_household_income` | **3.1%** | Census/Area Context | ZIP-level aggregate income; weak but present signal. |
| **11** | `created_day_of_month` | **2.4%** | Invoice Context | Day-of-month effect (e.g., invoices mid-month vs. end-of-month). |
| **12** | `bachelors_or_higher_pct` | **1.9%** | Census/Area Context | ZIP-level education; modest signal. |
| **13** | `unemployment_rate_pct` | **1.6%** | Census/Area Context | ZIP-level unemployment rate; weak area-level indicator. |
| **14** | `poverty_rate_pct` | **1.3%** | Census/Area Context | ZIP-level poverty; modest predictive value. |
| **15** | `surchargingenabled` | **0.8%** | Invoice Context | Whether surcharge is applied; minor signal. |
| **16** | `average_household_size` | **0.5%** | Census/Area Context | ZIP-level demographic; minimal importance. |
| **17** | `is_guardian` | **0.3%** | Patient Profile | Guardianship status; very weak signal. |

---

## Summary by Category

| Category | Total Importance | Key Insights |
|---|---:|---|
| **Payment History** | **56.0%** | Historical payment behavior dominates. The model learns: *"past behavior predicts future behavior."* Features like days-to-pay, reliability, and tenure are the strongest signals. |
| **Invoice Context** | **35.1%** | Invoice characteristics (amount, timing, origin) are secondary but meaningful. Larger amounts and specific day/time patterns matter. |
| **Census/Area Context** | **7.4%** | ZIP-level public data (income, education, unemployment, poverty) provide weak but consistent signals. This is intentional—we avoid individual-level data. |
| **Patient Profile** | **6.1%** | New patient status and guardianship have modest predictive value. |

---

## Key Takeaways for Stakeholders

### What Drives Predictions?

1. **Payment Speed is King** (18.3%)  
   - If a patient historically takes 30+ days to pay, they're much more likely to have an unpaid invoice.
   - This is the single most important feature.

2. **Behavioral Consistency** (12.6%)  
   - Appointment reliability (no-shows, cancellations) predicts invoice payment.
   - Consistent patients = consistent payers.
   - ⚠️ **This feature is currently based on synthetic dummy data, not real appointment records.** Treat this importance value as provisional until the feature is rebuilt from real scheduler data.

3. **Tenure & Stability** (10.2% + 8.9%)  
   - Long-term patients are safer bets.
   - New patients (<90 days) are flagged as higher risk.

4. **Invoice Size** (14.7%)  
   - Larger invoices have different payment dynamics.
   - Not necessarily negative—could reflect prioritization or financial capacity.

5. **Temporal Effects** (7.5% + 6.4% + 2.4%)  
   - When an invoice is created matters.
   - Example: Friday 4 PM invoices may have different payment rates than Tuesday 10 AM invoices.

### What Does NOT Drive Predictions?

- **Individual Income** (0%)  
  - We intentionally do NOT use individual patient income.
  - We use ZIP-level aggregate income only (3.1%, weak).

- **Demographics** (<1%)  
  - Guardianship status and ZIP demographics are very weak signals.
  - The model relies on behavioral data, not demographics.

---

## Production Use Cases

### High-Risk Signals (Rank 1-5)
When these features indicate **high risk**, prioritize proactive outreach:
- Long historical payment delays
- Large invoice amounts (need follow-up)
- Low appointment reliability
- New patient status
- Short tenure

### Medium-Risk Signals (Rank 6-9)
Monitor these temporal/contextual factors:
- Invoices created on certain days/times
- Payment origin mode (some channels = higher risk)

### Low-Risk Context (Rank 10+)
ZIP-level aggregates provide lightweight background context but are not primary drivers.

---

## Model Fairness Notes

✅ **No individual-level protected attributes used**  
✅ **Payment history is behavioral, not demographic**  
✅ **ZIP-level census data only (aggregate, not individual)**  
✅ **Fairness monitoring** recommended on:
  - Disparate impact by ZIP or other segments
  - Calibration drift over time
  - Feature importance stability

---

## Retraining Guidance

If feature importance shifts significantly on retrain:
- **Payment history features increase** → Temporal patterns are more stable; no action needed.
- **Census/demographic features increase** → May signal drift; investigate and reweight if needed.
- **New features needed?** → Consider behavioral signals like SMS/call responsiveness if deployed.

---

Generated: June 2026  
Model: XGBoost + SMOTE balancing  
Source: Raw invoice/payment data from 2021–2025
