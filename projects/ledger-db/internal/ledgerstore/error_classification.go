package ledgerstore

import "errors"

type LedgerErrorCode string

const (
	LedgerErrorCodeValidationAmountRequired            LedgerErrorCode = "validation.amount_required"
	LedgerErrorCodeValidationAccountRequired           LedgerErrorCode = "validation.account_required"
	LedgerErrorCodeValidationExternalReferenceRequired LedgerErrorCode = "validation.external_reference_required"
	LedgerErrorCodeValidationRailRequired              LedgerErrorCode = "validation.rail_required"
	LedgerErrorCodeValidationIdempotencyKeyRequired    LedgerErrorCode = "validation.idempotency_key_required"
	LedgerErrorCodeValidationTransactionIDRequired     LedgerErrorCode = "validation.transaction_id_required"
	LedgerErrorCodeValidationReasonRequired            LedgerErrorCode = "validation.reason_required"
	LedgerErrorCodeValidationDirectionRequired         LedgerErrorCode = "validation.direction_required"

	LedgerErrorCodeNotFoundFromAccount           LedgerErrorCode = "not_found.from_account"
	LedgerErrorCodeNotFoundToAccount             LedgerErrorCode = "not_found.to_account"
	LedgerErrorCodeNotFoundCashSettlementAccount LedgerErrorCode = "not_found.cash_settlement_account"
	LedgerErrorCodeNotFoundTransaction           LedgerErrorCode = "not_found.transaction"

	LedgerErrorCodeBusinessInsufficientFunds     LedgerErrorCode = "business.insufficient_funds"
	LedgerErrorCodeBusinessCurrencyMismatch      LedgerErrorCode = "business.currency_mismatch"
	LedgerErrorCodeBusinessIdempotencyConflict   LedgerErrorCode = "business.idempotency_conflict"
	LedgerErrorCodeBusinessReversalAlreadyExists LedgerErrorCode = "business.reversal_already_exists"

	LedgerErrorCodeInvariantTransactionNotBalanced LedgerErrorCode = "invariant.transaction_not_balanced"
	LedgerErrorCodeInvariantBalanceMismatch        LedgerErrorCode = "invariant.balance_mismatch"
	LedgerErrorCodeInvariantInvalidReversal        LedgerErrorCode = "invariant.invalid_reversal"

	LedgerErrorCodeDBUniqueViolation      LedgerErrorCode = "db.unique_violation"
	LedgerErrorCodeDBForeignKeyViolation  LedgerErrorCode = "db.foreign_key_violation"
	LedgerErrorCodeDBCheckViolation       LedgerErrorCode = "db.check_violation"
	LedgerErrorCodeDBSerializationFailure LedgerErrorCode = "db.serialization_failure"
	LedgerErrorCodeDBDeadlock             LedgerErrorCode = "db.deadlock"
	LedgerErrorCodeDBUnavailable          LedgerErrorCode = "db.unavailable"

	LedgerErrorCodeInternalUnknown LedgerErrorCode = "internal.unknown"
)

const (
	LedgerErrorCategoryValidation = "validation"
	LedgerErrorCategoryNotFound   = "not_found"
	LedgerErrorCategoryBusiness   = "business"
	LedgerErrorCategoryInvariant  = "invariant"
	LedgerErrorCategoryDB         = "db"
	LedgerErrorCategoryInternal   = "internal"
)

type LedgerErrorInfo struct {
	Code      LedgerErrorCode
	Category  string
	Retryable bool
	Expected  bool
	Message   string
}

type ledgerErrorInfoProvider interface {
	ledgerErrorInfo() LedgerErrorInfo
	ledgerErrorInfoWithMessage(message string) LedgerErrorInfo
}

type ledgerError struct {
	Code      LedgerErrorCode
	Category  string
	Retryable bool
	Expected  bool
	Message   string
}

func (e *ledgerError) Error() string {
	return e.Message
}

func (e *ledgerError) ledgerErrorInfo() LedgerErrorInfo {
	return LedgerErrorInfo{
		Code:      e.Code,
		Category:  e.Category,
		Retryable: e.Retryable,
		Expected:  e.Expected,
		Message:   e.Message,
	}
}

func (e *ledgerError) ledgerErrorInfoWithMessage(message string) LedgerErrorInfo {
	info := e.ledgerErrorInfo()
	info.Message = message
	return info
}

func newLedgerError(code LedgerErrorCode, category string, expected bool, message string) error {
	return &ledgerError{
		Code:     code,
		Category: category,
		Expected: expected,
		Message:  message,
	}
}

func ClassifyError(err error) LedgerErrorInfo {
	if err == nil {
		return LedgerErrorInfo{}
	}

	message := err.Error()
	var ledgerErr ledgerErrorInfoProvider
	if errors.As(err, &ledgerErr) {
		return ledgerErr.ledgerErrorInfoWithMessage(message)
	}
	if info, ok := classifyPostgresError(err, message); ok {
		return info
	}

	return LedgerErrorInfo{
		Code:     LedgerErrorCodeInternalUnknown,
		Category: LedgerErrorCategoryInternal,
		Message:  message,
	}
}

func classifyPostgresError(err error, message string) (LedgerErrorInfo, bool) {
	var sqlStateErr interface {
		SQLState() string
	}
	if !errors.As(err, &sqlStateErr) {
		return LedgerErrorInfo{}, false
	}

	switch sqlStateErr.SQLState() {
	case "23505":
		return dbErrorInfo(LedgerErrorCodeDBUniqueViolation, false, false, message), true
	case "23503":
		return dbErrorInfo(LedgerErrorCodeDBForeignKeyViolation, false, false, message), true
	case "23514":
		return dbErrorInfo(LedgerErrorCodeDBCheckViolation, false, false, message), true
	case "40001":
		return dbErrorInfo(LedgerErrorCodeDBSerializationFailure, true, true, message), true
	case "40P01":
		return dbErrorInfo(LedgerErrorCodeDBDeadlock, true, true, message), true
	case "08006":
		return dbErrorInfo(LedgerErrorCodeDBUnavailable, true, false, message), true
	default:
		return LedgerErrorInfo{}, false
	}
}

func dbErrorInfo(code LedgerErrorCode, retryable bool, expected bool, message string) LedgerErrorInfo {
	return LedgerErrorInfo{
		Code:      code,
		Category:  LedgerErrorCategoryDB,
		Retryable: retryable,
		Expected:  expected,
		Message:   message,
	}
}
