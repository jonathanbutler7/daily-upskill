# Demo Metrics And Guardrails

This is a presentation cheat sheet for rollout, trust, safety, and business-value questions.

## 1. North Star

Primary goal: recover revenue that would otherwise go uncollected, without creating customer harm.

Primary business metric:
- Incremental collected revenue versus current workflow

Primary safety principle:
- The score is for payment workflow prioritization only, never for care access or appointment decisions.

## 2. Business Value Metrics

Use these when asked, "How do we know this is working?"

Primary metrics:
- Incremental collected revenue versus control
- Recovered revenue from 30+ day unpaid balances
- Recovery rate on model-prioritized accounts

Supporting metrics:
- Days to payment
- Dollars collected per staff touch
- Payment plan acceptance rate
- Financing offer acceptance rate
- Follow-up completion rate on recommended actions

Useful talking point:
- In one location cohort, about $1.52M of revenue went uncollected over roughly two years. Even partial recovery is meaningful.

## 3. Rollout Metrics

Use these when asked, "What tells you to expand?"

Expand signals:
- Statistically meaningful lift in collected revenue versus control
- Higher recovery rate on prioritized balances than non-prioritized balances
- Stable or improved days-to-payment
- Strong staff adoption of recommendations

Suggested first-pass thresholds:
- 10%+ lift in recovered revenue on 30+ day unpaid balances in pilot locations
- 70%+ recommendation view rate in the collections modal
- 50%+ action completion rate on high-priority recommendations

## 4. Pause / Stop Metrics

Use these when asked, "What makes you pause rollout immediately?"

Immediate stop conditions:
- Any evidence the model is influencing care access or appointment decisions
- Any evidence of offices refusing care based on score
- Confirmed data leakage or corrupted feature pipeline

Pause-and-investigate conditions:
- Precision on the high-risk bucket drops below an agreed threshold
- Complaint rate increases after rollout
- Office staff report repeated obviously wrong reason codes or scores
- Drift in calibration or feature behavior after retrain

Suggested first-pass thresholds:
- Complaint rate increases by 10%+ versus control
- High-risk precision drops below 0.65
- Calibration drift exceeds agreed tolerance on pilot monitoring

## 5. Trust Metrics

Use these when asked, "Why would staff trust this?"

Signals trust is improving:
- Recommendation action rate increases over time
- Staff override rate decreases over time
- Staff feedback says reason codes match account reality
- High-risk recommendations convert to payment more often than lower-risk ones

Signals trust is degrading:
- High override rate on recommended actions
- Staff ignore or dismiss the score in the modal
- Feedback that reason codes are vague or incorrect
- High false-positive rate in high-touch outreach

Suggested first-pass thresholds:
- Staff override rate under 25% after pilot onboarding
- Positive staff feedback on explanation usefulness in post-pilot interviews

## 6. Reason Code Metrics

Use these when asked, "How do we know explanations are helping?"

Helpful metrics:
- Higher action rate when reason codes are shown versus hidden
- Lower override rate when reason codes are shown
- Better agreement from staff that the score matches what they know about the account

Guardrails:
- If reason codes are repeatedly inaccurate, vague, or contradicted by account reality, trust will drop fast
- Reason codes should explain the score, not expose raw internals or sensitive fields

## 7. False Positive Controls

Use these when asked, "How do you avoid wasting staff time or bothering customers?"

Model controls:
- Oversample the minority class during training so the model learns non-pay behavior
- Tune threshold based on precision/recall tradeoff rather than using the default cutoff blindly

Product controls:
- Keep high-touch outreach gated behind both score and value-at-risk
- Use lighter-touch workflows for lower-confidence cases
- Never use score alone to deny financing, care, or service

Concrete example:
- A low payment-likelihood score may qualify an account for proactive financing outreach, but not for denial of options.

## 8. Control Setup For Pilot

Use these when asked, "How do you prove lift if today is one-size-fits-all?"

Suggested setup:
- Roll out by location or team in phases
- Pilot group gets model-prioritized outreach
- Control group keeps current one-size-fits-all process for a fixed window
- Compare revenue recovery, days-to-payment, complaint rate, and touches per dollar collected

Why this matters:
- Without a control, you cannot tell whether collections would have happened anyway.

## 9. Workload Metrics

Use these when asked, "Does this create more work for staff?"

Helpful metrics:
- Dollars collected per staff touch
- Manual calls per dollar recovered
- Share of low-risk accounts kept on automated or low-touch workflows

Desired outcome:
- Total effort stays flat or declines while recovery improves because staff focus moves to the highest-value accounts.

## 10. Monitoring And Maintenance

Use these when asked, "How do you keep this from becoming set-it-and-forget-it?"

Monitor continuously:
- Precision by risk band
- Recovery lift versus control
- Complaint / safety signals
- Calibration drift over time
- Feature importance stability after retrains

Operational rule:
- If drift or safety signals breach threshold, tighten policy thresholds immediately and retrain before wider rollout.

## 11. Short Answers To Memorize

If asked about false positives:
- We control false positives with threshold tuning, action gating, and monitoring precision on the high-risk bucket.

If asked about customer harm:
- Worst-case harm is extra outreach, not denial of care. The score is only used in payment workflows.

If asked about financing:
- The score should help offices offer financing proactively, not deny it automatically.

If asked about trust:
- Trust is earned through cautious rollout, clear reason codes, and feedback loops with office staff.

If asked about business value:
- Success is incremental collected revenue, not just a strong model metric.