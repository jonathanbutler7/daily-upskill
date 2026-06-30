"""
Placeholder tests for score_single.py.

These cover the two pure functions that have no external dependencies (no model
artifact required). Add model-integration tests here once CI can supply the
.joblib artifact via DVC.
"""

import pytest

from score_single import assign_risk_band, score_record


# ---------------------------------------------------------------------------
# assign_risk_band
# ---------------------------------------------------------------------------


class TestAssignRiskBand:
    def test_high_at_threshold(self):
        assert assign_risk_band(0.70) == "high"

    def test_high_above_threshold(self):
        assert assign_risk_band(0.95) == "high"

    def test_medium_at_lower_threshold(self):
        assert assign_risk_band(0.40) == "medium"

    def test_medium_mid_range(self):
        assert assign_risk_band(0.55) == "medium"

    def test_medium_just_below_high(self):
        assert assign_risk_band(0.699) == "medium"

    def test_low_below_threshold(self):
        assert assign_risk_band(0.39) == "low"

    def test_low_at_zero(self):
        assert assign_risk_band(0.0) == "low"


# ---------------------------------------------------------------------------
# score_record — missing feature validation
# ---------------------------------------------------------------------------


class TestScoreRecordValidation:
    def test_raises_on_missing_features(self):
        """score_record should raise ValueError if any feature is absent."""
        with pytest.raises(ValueError, match="Missing required features"):
            score_record(model=None, feature_values={"amount": 100.0})

    def test_raises_lists_missing_feature_names(self):
        """The error message should name the missing features."""
        with pytest.raises(ValueError, match="average_days_to_pay"):
            score_record(model=None, feature_values={"amount": 100.0})
