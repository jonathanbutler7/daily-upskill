package processor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	stripe "github.com/stripe/stripe-go/v82"
)

// StripeConfig holds configuration for the Stripe payment processor.
type StripeConfig struct {
	APIKey string
}

// StripeProcessor implements PaymentProcessor using the Stripe API.
type StripeProcessor struct {
	sc *stripe.Client
}

// NewStripeProcessor creates a StripeProcessor using the provided API key.
func NewStripeProcessor(cfg StripeConfig) *StripeProcessor {
	return &StripeProcessor{sc: stripe.NewClient(cfg.APIKey)}
}

// CreatePaymentMethod tokenises card data into a Stripe PaymentMethod.
// Accepts either raw card numbers (PCI-scoped environments) or Stripe test tokens (tok_visa, etc.).
// The caller must clear CardNumber and CVV from memory immediately after this call.
func (p *StripeProcessor) CreatePaymentMethod(ctx context.Context, req CreatePaymentMethodRequest) (*PaymentMethod, error) {
	var cardParams *stripe.PaymentMethodCreateCardParams

	if strings.HasPrefix(req.CardNumber, "tok_") {
		cardParams = &stripe.PaymentMethodCreateCardParams{
			Token: stripe.String(req.CardNumber),
		}
	} else {
		expMonth, err := strconv.ParseInt(req.ExpMonth, 10, 64)
		if err != nil {
			return nil, &ProcessorError{Code: "invalid_expiration", Message: fmt.Sprintf("invalid exp_month %q: %v", req.ExpMonth, err)}
		}
		expYear, err := strconv.ParseInt(req.ExpYear, 10, 64)
		if err != nil {
			return nil, &ProcessorError{Code: "invalid_expiration", Message: fmt.Sprintf("invalid exp_year %q: %v", req.ExpYear, err)}
		}
		cardParams = &stripe.PaymentMethodCreateCardParams{
			Number:   stripe.String(req.CardNumber),
			ExpMonth: stripe.Int64(expMonth),
			ExpYear:  stripe.Int64(expYear),
		}
		if strings.TrimSpace(req.CVV) != "" {
			cardParams.CVC = stripe.String(req.CVV)
		}
	}

	pm, err := p.sc.V1PaymentMethods.Create(ctx, &stripe.PaymentMethodCreateParams{
		Type: stripe.String("card"),
		Card: cardParams,
	})
	if err != nil {
		return nil, mapStripeError(err)
	}

	last4 := ""
	if pm.Card != nil {
		last4 = pm.Card.Last4
	}
	return &PaymentMethod{ID: pm.ID, Last4: last4}, nil
}

// CreatePaymentIntent creates a Stripe PaymentIntent without immediately confirming it.
func (p *StripeProcessor) CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*PaymentIntent, error) {
	params := &stripe.PaymentIntentCreateParams{
		Params: stripe.Params{
			IdempotencyKey: stripe.String(req.IdempotencyKey + "-create"),
		},
		Amount:        stripe.Int64(req.AmountCents),
		Currency:      stripe.String(req.Currency),
		PaymentMethod: stripe.String(req.PaymentMethodID),
		Confirm:       stripe.Bool(false),
		// Explicitly restrict to card only. Using automatic_payment_methods with
		// allow_redirects=never is insufficient when the Stripe account has a dashboard
		// Payment Method Configuration (PMC) — the PMC overrides the allow_redirects
		// setting and Stripe rejects confirm without a return_url.
		PaymentMethodTypes: []*string{stripe.String("card")},
	}
	for k, v := range req.Metadata {
		params.AddMetadata(k, v)
	}

	pi, err := p.sc.V1PaymentIntents.Create(ctx, params)
	if err != nil {
		return nil, mapStripeError(err)
	}

	return stripePaymentIntentToPaymentIntent(pi), nil
}

// ConfirmPaymentIntent confirms an existing PaymentIntent, triggering the charge.
func (p *StripeProcessor) ConfirmPaymentIntent(ctx context.Context, paymentIntentID, idempotencyKey string) (*PaymentIntent, error) {
	params := &stripe.PaymentIntentConfirmParams{
		Params: stripe.Params{
			IdempotencyKey: stripe.String(idempotencyKey + "-confirm"),
		},
	}

	pi, err := p.sc.V1PaymentIntents.Confirm(ctx, paymentIntentID, params)
	if err != nil {
		return nil, mapStripeError(err)
	}

	return stripePaymentIntentToPaymentIntent(pi), nil
}

func stripePaymentIntentToPaymentIntent(pi *stripe.PaymentIntent) *PaymentIntent {
	result := &PaymentIntent{
		ID:     pi.ID,
		Status: string(pi.Status),
	}
	if pi.LatestCharge != nil {
		result.ChargeID = pi.LatestCharge.ID
	}
	return result
}

// mapStripeError converts a Stripe API error into a ProcessorError with a
// canonical Code that IsRetryable() understands.
func mapStripeError(err error) *ProcessorError {
	var stripeErr *stripe.Error
	if !errors.As(err, &stripeErr) {
		return &ProcessorError{Code: "network_failure", Message: err.Error()}
	}

	// Decline codes are the most specific signal for card errors.
	if stripeErr.DeclineCode != "" {
		return &ProcessorError{
			Code:    string(stripeErr.DeclineCode),
			Message: stripeErr.Msg,
		}
	}

	// Map Stripe error codes that we treat as retryable.
	if stripeErr.Code == stripe.ErrorCodeRateLimit {
		return &ProcessorError{Code: "rate_limit", Message: stripeErr.Msg}
	}

	// api_error type signals a Stripe-side problem — treat as transient.
	if stripeErr.Type == stripe.ErrorTypeAPI {
		return &ProcessorError{Code: "processor_unavailable", Message: stripeErr.Msg}
	}

	// Fall back to the raw error code or type string.
	code := string(stripeErr.Code)
	if code == "" {
		code = string(stripeErr.Type)
	}
	return &ProcessorError{Code: code, Message: stripeErr.Msg}
}
