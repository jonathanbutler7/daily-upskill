package server

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	willtheypay "weavelab.xyz/schema-gen-go/schemas/payments-platform/will-they-pay/v1"
)

const recommendationPolicyVersion = "collection_policy_v1"

const predictionsCSVEnvVar = "WTP_PREDICTIONS_CSV_PATH"

type WillTheyPayServer struct {
	defaultModelVersion string
}

type predictionRow struct {
	willPayIn30  bool
	confidence   float32
	riskBand     string
	modelVersion string
	scoredAt     time.Time
	hasScoredAt  bool
}

func NewWillTheyPayServer(modelVersion string) *WillTheyPayServer {
	if modelVersion == "" {
		modelVersion = "unknown"
	}

	return &WillTheyPayServer{defaultModelVersion: modelVersion}
}

func (s *WillTheyPayServer) GetPaymentLikelihood(ctx context.Context, req *willtheypay.GetPaymentLikelihoodRequest) (*willtheypay.GetPaymentLikelihoodResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	if strings.TrimSpace(req.GetPatientId()) == "" || strings.TrimSpace(req.GetLocationId()) == "" {
		return nil, fmt.Errorf("patient_id and location_id are required")
	}

	prediction, err := s.getPredictionByPersonID(req.GetPatientId())
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	scoredAt := now
	if prediction.hasScoredAt {
		scoredAt = prediction.scoredAt
	}

	modelVersion := prediction.modelVersion
	if modelVersion == "" {
		modelVersion = s.defaultModelVersion
	}

	return &willtheypay.GetPaymentLikelihoodResponse{
		Likelihood: &willtheypay.PaymentLikelihood{
			HasPrediction:   true,
			WillPayIn_30:    prediction.willPayIn30,
			ConfidenceScore: prediction.confidence,
			RiskBand:        prediction.riskBand,
			ModelVersion:    modelVersion,
			ScoredAt:        timestamppb.New(scoredAt),
		},
		Recommendation: buildRecommendation(prediction.confidence, now),
	}, nil
}

func (s *WillTheyPayServer) getPredictionByPersonID(personID string) (predictionRow, error) {
	csvPath, err := csvPredictionsPath()
	if err != nil {
		return predictionRow{}, fmt.Errorf("failed to resolve predictions path: %w", err)
	}

	predictionsByPerson, err := loadPredictionsByPerson(csvPath)
	if err != nil {
		return predictionRow{}, fmt.Errorf("failed to load predictions: %w", err)
	}

	prediction, ok := predictionsByPerson[personID]
	if !ok {
		return predictionRow{}, fmt.Errorf("no prediction found for patient_id %q", personID)
	}

	return prediction, nil
}

func csvPredictionsPath() (string, error) {
	relativeCSVPath := filepath.Join("model", "data", "output", "scored_predictions_prod.csv")

	if explicitPath := strings.TrimSpace(os.Getenv(predictionsCSVEnvVar)); explicitPath != "" {
		if fileExists(explicitPath) {
			return explicitPath, nil
		}
		return "", fmt.Errorf("%s points to a missing file: %s", predictionsCSVEnvVar, explicitPath)
	}

	seen := map[string]struct{}{}
	var candidates []string

	addCandidatesFromRoot := func(root string) {
		if strings.TrimSpace(root) == "" {
			return
		}

		for _, base := range walkParents(root, 8) {
			candidate := filepath.Join(base, relativeCSVPath)
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			candidates = append(candidates, candidate)
		}
	}

	if wd, err := os.Getwd(); err == nil {
		addCandidatesFromRoot(wd)
	}

	if _, currentFile, _, ok := runtime.Caller(0); ok && filepath.IsAbs(currentFile) {
		addCandidatesFromRoot(filepath.Dir(currentFile))
	}

	if exePath, err := os.Executable(); err == nil {
		addCandidatesFromRoot(filepath.Dir(exePath))
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("predictions csv not found; looked for %s", relativeCSVPath)
}

func walkParents(start string, maxDepth int) []string {
	dirs := []string{}
	current := filepath.Clean(start)

	for i := 0; i <= maxDepth; i++ {
		dirs = append(dirs, current)
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return dirs
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func loadPredictionsByPerson(path string) (map[string]predictionRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	headers, err := r.Read()
	if err != nil {
		return nil, err
	}

	idx := make(map[string]int, len(headers))
	for i, header := range headers {
		idx[strings.TrimSpace(header)] = i
	}

	required := []string{"personid", "will_pay_in_30", "confidence_score", "risk_band", "model_version", "scored_at"}
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			return nil, fmt.Errorf("missing required column %q", col)
		}
	}

	predictions := make(map[string]predictionRow)
	for {
		record, readErr := r.Read()
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, readErr
		}

		personID := strings.TrimSpace(record[idx["personid"]])
		if personID == "" {
			continue
		}

		willPay, err := parseWillPay(record[idx["will_pay_in_30"]])
		if err != nil {
			return nil, fmt.Errorf("invalid will_pay_in_30 for personid %q: %w", personID, err)
		}

		confidence64, err := strconv.ParseFloat(strings.TrimSpace(record[idx["confidence_score"]]), 32)
		if err != nil {
			return nil, fmt.Errorf("invalid confidence_score for personid %q: %w", personID, err)
		}

		scoredAt, hasScoredAt, err := parseScoredAt(record[idx["scored_at"]])
		if err != nil {
			return nil, fmt.Errorf("invalid scored_at for personid %q: %w", personID, err)
		}

		predictions[personID] = predictionRow{
			willPayIn30:  willPay,
			confidence:   float32(confidence64),
			riskBand:     strings.TrimSpace(record[idx["risk_band"]]),
			modelVersion: strings.TrimSpace(record[idx["model_version"]]),
			scoredAt:     scoredAt,
			hasScoredAt:  hasScoredAt,
		}
	}

	return predictions, nil
}

func parseWillPay(v string) (bool, error) {
	value := strings.TrimSpace(v)
	if value == "1" {
		return true, nil
	}
	if value == "0" {
		return false, nil
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}
	return b, nil
}

func parseScoredAt(v string) (time.Time, bool, error) {
	value := strings.TrimSpace(v)
	if value == "" {
		return time.Time{}, false, nil
	}

	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, false, err
	}

	return t.UTC(), true, nil
}

func deriveRiskBand(confidenceScore float32) string {
	if confidenceScore >= 0.70 {
		return "high"
	}

	if confidenceScore >= 0.40 {
		return "medium"
	}

	return "low"
}

func buildRecommendation(confidenceScore float32, now time.Time) *willtheypay.CollectionRecommendation {
	recommendation := &willtheypay.CollectionRecommendation{
		HasRecommendation: true,
		PolicyVersion:     recommendationPolicyVersion,
	}

	if confidenceScore < 0.40 {
		recommendation.ActionCode = willtheypay.RecommendationActionCode_CALL_AND_OFFER_PLAN
		recommendation.Channel = willtheypay.RecommendationChannel_PHONE
		recommendation.DueBy = timestamppb.New(now.Add(24 * time.Hour))
		recommendation.FallbackActionCode = willtheypay.RecommendationActionCode_SEND_SMS_PAY_LINK
		recommendation.FallbackAfterDays = 2
		recommendation.ReasonCodes = []willtheypay.RecommendationReasonCode{
			willtheypay.RecommendationReasonCode_LOW_PAYMENT_PROBABILITY,
		}
		recommendation.ScriptHint = "High-risk account. Call now, offer a payment plan, and collect the first payment today."
		return recommendation
	}

	if confidenceScore < 0.70 {
		recommendation.ActionCode = willtheypay.RecommendationActionCode_STANDARD_REMINDER_SEQUENCE
		recommendation.Channel = willtheypay.RecommendationChannel_SMS
		recommendation.DueBy = timestamppb.New(now.Add(48 * time.Hour))
		recommendation.FallbackActionCode = willtheypay.RecommendationActionCode_CALL_REMINDER
		recommendation.FallbackAfterDays = 3
		recommendation.ReasonCodes = []willtheypay.RecommendationReasonCode{
			willtheypay.RecommendationReasonCode_MEDIUM_PAYMENT_PROBABILITY,
		}
		recommendation.ScriptHint = "Moderate-risk account. Send reminder flow now and follow up with a call if unpaid."
		return recommendation
	}

	recommendation.ActionCode = willtheypay.RecommendationActionCode_LIGHT_TOUCH_REMINDER
	recommendation.Channel = willtheypay.RecommendationChannel_SMS
	recommendation.DueBy = timestamppb.New(now.Add(72 * time.Hour))
	recommendation.FallbackActionCode = willtheypay.RecommendationActionCode_NONE
	recommendation.FallbackAfterDays = 0
	recommendation.ReasonCodes = []willtheypay.RecommendationReasonCode{
		willtheypay.RecommendationReasonCode_HIGH_PAYMENT_PROBABILITY,
	}
	recommendation.ScriptHint = "Low-risk account. Send a light reminder and avoid manual outreach unless it becomes overdue."

	return recommendation
}
