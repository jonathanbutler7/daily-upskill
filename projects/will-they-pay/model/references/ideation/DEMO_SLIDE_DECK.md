# Will They Pay? Demo Slide Deck

## Slide 1 - Business Problem + Revenue Opportunity
**Title:** $1.52M Unpaid Today; $1.19M Phase-1 Opportunity

- **$1.52M unpaid** in one location cohort (~2 years)
- **5,317 unpaid invoices** (**35.4%**, no-payment-row definition)
- **$1.19M addressable** by prioritizing top **30%** balances
- That is **78%** of unpaid dollars in this cohort

**Speaker note:**
"Lead with impact: this is a revenue recovery product, not a model demo."

---

## Slide 2 - Why This Works
**Title:** Payment Speed Strongly Predicts Collection Risk

- Mean time-to-pay: **0.59 days** (~**14h 16m**)
- Long tail extends to **104 days**
- % with unpaid balance by speed band:
- **<15 min: 16.9%** -> **30+ days: 56.3%**

**Speaker note:**
"As payment speed slows, unpaid incidence rises materially."

---

## Slide 3 - How It Works + Model Selection
**Title:** ML Pipeline & Why XGBoost

- Pipeline: labeled data → feature engineering → XGBoost training → risk scores
- Why XGBoost: evaluated LR, RF, XGBoost; XGBoost won on discrimination + calibration
- Production eval: **ROC-AUC 0.946** (strong separation of payers vs non-payers)
- Output: `confidence_score` + `risk_band` fed directly to product

**Speaker note:**
"Simple pipeline, strong model, actionable output."

---

## Slide 4 - Revenue Action in Product
**Title:** Risk Score → Workflow → Collected Dollars

- High risk: proactive outreach + payment plan
- Medium/low risk: lighter-touch cadence
- Integration: TTP scheduling + Tasks ownership tracking
- Goal: highest-ROI staff effort on highest-risk balances

**Speaker note:**
"Prediction is useful only when it changes actions in product."

---

## Slide 5 - Business Impact Summary
**Title:** Customer Value

- Revenue at risk: **$1.52M**
- Phase-1 target: **$1.19M**
- Focused strategy: top **30%** prioritized first
- Outcome: higher collection efficiency + better patient experience

**Speaker note:**
"Same team, better prioritization, materially better recovery potential."

---

## Slide 6 - Future Roadmap
**Title:** Next Product Investments

- Manual single-patient model reruns in product UI
- More intentional dogfooding (especially task workflow usage)
- Better AI next-best-action suggestions to collect more revenue

**Speaker note:**
"Next phase is tighter product loops and higher-quality AI actioning."

---

## Slide 7 - Ethics & Guardrails
**Title:** Patient-Safe, Fairness-Aware Deployment

- Use case: prioritize outreach, **never** deny care
- ZIP-level Census/ACS aggregates only (not individual income)
- Monitor segment-level calibration + error rates each retrain cycle
- Remove/reweight features if fairness thresholds are breached

**15-sec if-asked response:**
"We use aggregate ZIP context as a small signal, never individual income, and never to deny care. We monitor fairness every retrain and can remove/reweight any feature if bias appears."

---

## Slide 8 - Q&A Support: Metric Definitions
**Title:** Why Unpaid Counts Can Differ

| Definition | Count | Rate |
|---|---:|---:|
| No payment row exists | 5,317 | 35.45% |
| No successful payment row (`paymentstatus = 1`) | 5,386 | 35.91% |
| Not paid within 30 days | 5,440 | 36.27% |

Use this table for Q&A on metric definitions.