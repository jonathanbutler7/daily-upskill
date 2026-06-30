# WILL THEY PAY? - 5-Minute Hackathon Demo Narrative

## Problem Statement (30 seconds)
**"Healthcare offices are losing millions from unpaid invoices, and they have no way to see it coming."**

### The Numbers Tell the Story:
- **$1.52M in unpaid invoices** across 2,052 patients (from this one location cohort)
- **35.4% of invoices go unpaid** within the collection window
- **Top 1% of patients** owe up to $33K each
- Current workflow: **same collection approach for everyone** — no differentiation between high and low risk

### Why This Matters:
- Late/missed payments reduce office cash flow and profitability
- Collections teams waste effort on accounts that will eventually pay
- Financing offers aren't proactively offered when needed most
- Revenue visibility is poor until weeks/months after invoice creation

---

## The Solution: "Will They Pay?" ML Model

### What It Does (60 seconds)
A machine learning model that predicts payment likelihood within 30 days and surfaces risk scores at point-of-collection.

**Input:** When a patient invoice is created
- Patient demographics (name, DOB, contact)
- Financial metrics (invoice amount, account balance)
- Payment history (how often they've paid on time before)
- ZIP-level public census context (aggregate area indicators, not individual income)

**Output:** A risk prediction with confidence
```json
{
  "will_pay_in_30_days": false,
  "confidence_score": 0.78,
  "risk_band": "high",
  "expected_value_at_risk": 620.00
}
```

### How It Works (ML Pipeline)
1. **Data Assembly** → Combine invoices + payment history + ZIP-level public census context
2. **Feature Engineering** → Extract patterns (e.g., "patients with prior on-time payment rate > 80% are 3x more likely to pay")
3. **Train Baseline Models** → Compare Random Forest, Logistic Regression, XGBoost on historical invoices
4. **Evaluate** → Target >= 0.70 ROC-AUC; ensure high precision on "unlikely to pay" predictions
5. **Batch Score** → Run on all open balances; store predictions in database
6. **Serve at Collection** → API returns prediction + confidence for office staff UI

---

## Business Impact (120 seconds)

### Immediate Win: Revenue Recovery Opportunity
If we can **predict and proactively collect from just the top 30% of unpaid invoices:**
- **$1.19M revenue recovered** (78% of total loss)
- **678 high-risk accounts** need targeted action (calls + financing offers)
- **Avg unpaid amount:** $701.49 per person (2.4 unpaid invoices each)

### Workflows Enabled by the Prediction
Once deployed, office staff see the risk score at invoice creation or payment request:

**🔴 High Risk (Confidence > 0.7):** 
- → Proactive call with payment plan options
- → Offer financing if appropriate
- → Schedule follow-up before due date
- Expected: 15-20% lift in payment rate from this segment

**🟡 Medium Risk (Confidence 0.4-0.7):**
- → Standard reminder sequence (SMS then email)
- → Request payment method on file

**🟢 Low Risk (Confidence < 0.4):**
- → Passive workflow (single SMS reminder)
- → Allocate collections resources elsewhere

### ROI Calculation
- **Cost to train/deploy model:** ~$5K (hackathon to MVP)
- **Operational lift per averted non-payment:** $500-2000 in staff time + financing costs
- **Potential annual savings:** $5-40M depending on collection lift

---

## Technical Architecture

### Data Pipeline (Batch Scoring)
```
1. Ingest historical invoices + payments → SQL
2. Join with ZIP-level public census context (90%+ ZIP match rate)
3. Engineer features (on-time rate, balance trend, etc.)
4. Score with trained model → generate probabilities
5. Map to risk bands: high/medium/low
6. Store predictions in payment_predictions table
7. API query for real-time staff decision support
```

### Demo Tech Stack
- **Language:** Go (backend service)
- **ML Training:** Python + Pandas + scikit-learn
- **Data:** Weave payments database + public Census/ACS ZIP-level aggregates (CSV)
- **Model Evaluation:** ROC-AUC >= 0.70, Precision >= 0.65 on "unlikely to pay"

### Feature Importance Reference
**See [model/docs/FEATURE_IMPORTANCE.md](../model/docs/FEATURE_IMPORTANCE.md) for detailed feature weighting.**

**Top 5 drivers** (ranked by importance):
1. `average_days_to_pay` (18.3%) — Historical payment speed
2. `amount` (14.7%) — Invoice size
3. `appointment_reliability_score` (12.6%) — Behavioral consistency
4. `account_age_days` (10.2%) — Patient tenure
5. `tenure_months` (8.9%) — Relationship length

Payment history features account for **56%** of model importance; invoice context adds **35%**; census/demographic context contributes only **7.4%** (intentionally weak to avoid bias).

---

## How to Present This (Speaking Notes)

### Opening (15 seconds)
*"Imagine running a healthcare office. Every day you send out hundreds of invoices, but you don't know which ones will actually get paid. You treat every patient the same at collection time. **What if you could predict who's likely to pay, so you can proactively reach out to the risky ones first?** That's 'Will They Pay?'"*

### Show the Problem (30 seconds)
*"In just one location's data, we have **$1.52 million in unpaid invoices** across **2,000 patients**. That's real money. And look at the distribution — it's not evenly spread. **35% of all invoices go unpaid.** But here's the key: **if we can predict the top 30% of highest-risk invoices**, we can recover $1.19 million. That's 78% of the total loss. And it only requires action on 678 accounts."**

**Show the Python analysis output on screen.**

### Explain the Model (60 seconds)
*"The model learns from historical data. We ask: 'Which patients paid on time before? What do they have in common?' We combine three signals:*
- *1) Their payment history (if they've always paid on time, they're likely to pay again)*
- *2) The invoice amount (large invoices might have different dynamics)*
- *3) Their local area context (public ZIP-level Census/ACS indicators)*

*We train a Random Forest — it's like an ensemble of decision trees that vote on the prediction. Then we score all open invoices. The output is simple: 'likely to pay' or 'unlikely to pay' with a confidence score between 0 and 1."*

### Show the Impact (45 seconds)
*"When our model predicts 'unlikely to pay' with high confidence, that's the signal for the office staff to:*
- *Call the patient with a payment plan*
- *Offer financing options*
- *Schedule a follow-up before the due date*

*For the 'likely to pay' group, we do minimal outreach. This prioritization could recover 15-20% more revenue from the high-risk segment alone — which adds millions in annual revenue for a large office network."*

### Close with Next Steps (30 seconds)
*"For MVP, we've trained the model and batch-scored the open invoices. The next phase is integrating this prediction into the office staff's payment workflow UI, so they see the risk score in real-time when collecting. We also want to test automated next-best-action recommendations — not just risk, but 'call with 3-part payment plan' or 'offer financing' based on the risk score."*

---

## Key Data Points to Memorize

| Metric | Value |
|--------|-------|
| **Total Unpaid Invoices** | 5,317 (35.4% of all invoices) |
| **Total Revenue at Risk** | $1.52M |
| **Unique Affected Patients** | 2,052 |
| **Avg Unpaid per Patient** | $701.49 |
| **Avg Unpaid per Invoice** | $286.20 |
| **Max Single Invoice** | $10,027.74 |
| **Top 30% Recoverable** | $1.19M (78% of total) |
| **Pay Rate** | 64.6% |
| **Non-Pay Rate** | 35.4% |

---

## Demo Talking Points (Rapid Fire)

✅ **Problem is real:** $1.52M at stake, 35% non-payment rate, affects 2K patients  
✅ **Model is simple:** Just 3-4 feature groups (payment history, amount, local area context)  
✅ **ML approach is standard:** Random Forest baseline with 0.70+ ROC-AUC target  
✅ **Impact is quantifiable:** 78% recovery on top 30% of unpaid invoices  
✅ **Deployment is feasible:** Batch scoring in < 5 mins, API serves predictions to UI  
✅ **Business case is strong:** $5-40M annual savings vs. $5K to build MVP  

---

## Questions You'll Get (& Answers)

**Q: How do you get 90%+ ZIP code match rate?**  
A: We use USPS ZIP code tables + patient address ZIP from Weave, then attach only ZIP-level public aggregates. Unmatchable records (no address or non-US) fall back to neutral defaults or are flagged for manual review.

**Q: What if a patient moves?**  
A: Address ZIP is refreshed at collection/update workflows and the model can be re-scored. We use area-level context only, not individual income or protected-class attributes.

**Q: How often does the model need retraining?**  
A: For MVP, monthly or quarterly on new historical outcomes. Long-term, we'd monitor for drift and retrain every 3-6 months or when payment patterns shift significantly.

**Q: Will patients feel targeted by high-risk calls?**  
A: Not if we frame it right. "Hi, we wanted to catch up on your account and offer a payment plan" is much friendlier than a late-payment dunning notice weeks later.

**Q: What about fairness — could the model discriminate by ZIP code income?**  
A: Great question. We position ZIP context as a small, aggregate signal and never as a reason to deny care. We also run fairness checks (disparate impact and calibration by segment), monitor drift, and can remove/reweight ZIP-derived features if bias signals appear.

---

## One-Slide Summary (If Needed)

**Problem:** $1.52M unpaid; 35% non-payment rate; no visibility at collection time  
**Solution:** ML model predicts payment likelihood (high/medium/low) using payment history + invoice amount + local ZIP-level context  
**Impact:** Enable proactive calls + financing offers for 678 high-risk accounts → recover $1.19M (78% of loss)  
**MVP:** Trained model, batch-scored open invoices, ready to integrate into office staff UI  
