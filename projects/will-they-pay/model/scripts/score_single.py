"""
Score a single record against the production Will They Pay model.

Useful for spot-checking the model, debugging predictions, and verifying
the joblib artifact is working correctly — without running the full batch pipeline.

Usage (run from model/):
    python scripts/score_single.py \\
        --amount 150.00 \\
        --created-day-of-week 1 \\
        --created-day-of-month 15 \\
        --created-hour 10 \\
        --payment-origin-mode 2 \\
        --surchargingenabled 0 \\
        --is-guardian 0 \\
        --account-age-days 365 \\
        --tenure-months 12 \\
        --is-new-patient 0 \\
        --average-days-to-pay 14.5 \\
        --appointment-reliability-score 0.85 \\
        --median-household-income 65000 \\
        --poverty-rate-pct 0.12 \\
        --average-household-size 2.5 \\
        --bachelors-or-higher-pct 0.35 \\
        --unemployment-rate-pct 0.04

    # Or pass features as a JSON file:
    python scripts/score_single.py --json-file path/to/record.json

Output:
    Prints prediction result as JSON:
    {
        "will_pay_in_30": 1,
        "confidence_score": 0.843,
        "risk_band": "high",
        "model_version": "will_they_pay_prod_v1"
    }
"""

import argparse
import json
from pathlib import Path

import joblib
import pandas as pd

MODEL_VERSION = "will_they_pay_prod_v1"

FEATURES = [
    "amount",
    "created_day_of_week",
    "created_day_of_month",
    "created_hour",
    "payment_origin_mode",
    "surchargingenabled",
    "is_guardian",
    "account_age_days",
    "tenure_months",
    "is_new_patient",
    "average_days_to_pay",
    "appointment_reliability_score",
    "median_household_income",
    "poverty_rate_pct",
    "average_household_size",
    "bachelors_or_higher_pct",
    "unemployment_rate_pct",
]


def assign_risk_band(confidence_score: float) -> str:
    if confidence_score >= 0.70:
        return "high"
    elif confidence_score >= 0.40:
        return "medium"
    else:
        return "low"


def score_record(model, feature_values: dict) -> dict:
    missing = [f for f in FEATURES if f not in feature_values]
    if missing:
        raise ValueError(f"Missing required features: {missing}")

    X = pd.DataFrame([{f: feature_values[f] for f in FEATURES}])
    will_pay = int(model.predict(X)[0])
    confidence = float(model.predict_proba(X)[0, 1])

    return {
        "will_pay_in_30": will_pay,
        "confidence_score": round(confidence, 6),
        "risk_band": assign_risk_band(confidence),
        "model_version": MODEL_VERSION,
    }


def main() -> None:
    script_dir = Path(__file__).parent
    model_dir = script_dir.parent / "models"

    parser = argparse.ArgumentParser(
        description="Score a single record using the production Will They Pay model.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument(
        "--model",
        default=str(model_dir / "will_they_pay_prod_v1.joblib"),
        help="Path to the trained model joblib (default: %(default)s)",
    )
    parser.add_argument(
        "--json-file",
        help="Path to a JSON file containing feature values (alternative to individual flags)",
    )

    # Individual feature flags
    parser.add_argument("--amount", type=float)
    parser.add_argument("--created-day-of-week", type=int, dest="created_day_of_week")
    parser.add_argument("--created-day-of-month", type=int, dest="created_day_of_month")
    parser.add_argument("--created-hour", type=int, dest="created_hour")
    parser.add_argument("--payment-origin-mode", type=int, dest="payment_origin_mode")
    parser.add_argument("--surchargingenabled", type=int)
    parser.add_argument("--is-guardian", type=int, dest="is_guardian")
    parser.add_argument("--account-age-days", type=float, dest="account_age_days")
    parser.add_argument("--tenure-months", type=float, dest="tenure_months")
    parser.add_argument("--is-new-patient", type=int, dest="is_new_patient")
    parser.add_argument("--average-days-to-pay", type=float, dest="average_days_to_pay")
    parser.add_argument(
        "--appointment-reliability-score",
        type=float,
        dest="appointment_reliability_score",
    )
    parser.add_argument("--median-household-income", type=float, dest="median_household_income")
    parser.add_argument("--poverty-rate-pct", type=float, dest="poverty_rate_pct")
    parser.add_argument("--average-household-size", type=float, dest="average_household_size")
    parser.add_argument("--bachelors-or-higher-pct", type=float, dest="bachelors_or_higher_pct")
    parser.add_argument("--unemployment-rate-pct", type=float, dest="unemployment_rate_pct")

    args = parser.parse_args()

    # Build feature dict from JSON file or individual flags
    if args.json_file:
        with open(args.json_file) as f:
            feature_values = json.load(f)
    else:
        feature_values = {
            feat: getattr(args, feat)
            for feat in FEATURES
            if getattr(args, feat, None) is not None
        }
        missing = [f for f in FEATURES if f not in feature_values]
        if missing:
            parser.error(
                f"Missing features: {missing}\n"
                "Provide all 17 features as flags or use --json-file."
            )

    model = joblib.load(args.model)
    result = score_record(model, feature_values)
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
