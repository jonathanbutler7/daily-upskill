package ledgerstore

var (
	ErrAmountGreaterThanZero = newLedgerError(
		LedgerErrorCodeValidationAmountRequired,
		LedgerErrorCategoryValidation,
		true,
		"amount must be greater than 0",
	)
	ErrCashSettlementAccountNotFound = newLedgerError(
		LedgerErrorCodeNotFoundCashSettlementAccount,
		LedgerErrorCategoryNotFound,
		true,
		"Cash Settlement account not found",
	)
	ErrCurrencyMismatch = newLedgerError(
		LedgerErrorCodeBusinessCurrencyMismatch,
		LedgerErrorCategoryBusiness,
		true,
		"currency mismatch",
	)
	ErrExternalReferenceEmpty = newLedgerError(
		LedgerErrorCodeValidationExternalReferenceRequired,
		LedgerErrorCategoryValidation,
		true,
		"external reference must not be empty",
	)
	ErrExternalReferenceRequired = newLedgerError(
		LedgerErrorCodeValidationExternalReferenceRequired,
		LedgerErrorCategoryValidation,
		true,
		"externalReference is required",
	)
	ErrFromAccountIDRequired = newLedgerError(
		LedgerErrorCodeValidationAccountRequired,
		LedgerErrorCategoryValidation,
		true,
		"from account id is required",
	)
	ErrFromAccountNotFound = newLedgerError(
		LedgerErrorCodeNotFoundFromAccount,
		LedgerErrorCategoryNotFound,
		true,
		"from account not found",
	)
	ErrIdempotencyConflict = newLedgerError(
		LedgerErrorCodeBusinessIdempotencyConflict,
		LedgerErrorCategoryBusiness,
		true,
		"idempotency conflict",
	)
	ErrIdempotencyKeyRequired = newLedgerError(
		LedgerErrorCodeValidationIdempotencyKeyRequired,
		LedgerErrorCategoryValidation,
		true,
		"idempotency key is required",
	)
	ErrInsufficientFunds = newLedgerError(
		LedgerErrorCodeBusinessInsufficientFunds,
		LedgerErrorCategoryBusiness,
		true,
		"insufficient funds",
	)
	ErrMustBeWithdrawalOrDeposit = newLedgerError(
		LedgerErrorCodeValidationDirectionRequired,
		LedgerErrorCategoryValidation,
		true,
		"direction must be withdrawal or deposit",
	)
	ErrNoRowsFound = newLedgerError(
		LedgerErrorCodeNotFoundTransaction,
		LedgerErrorCategoryNotFound,
		true,
		"no rows found",
	)
	ErrRailValueRequired = newLedgerError(
		LedgerErrorCodeValidationRailRequired,
		LedgerErrorCategoryValidation,
		true,
		"rail value is required",
	)
	ErrReversalAlreadyExists = newLedgerError(
		LedgerErrorCodeBusinessReversalAlreadyExists,
		LedgerErrorCategoryBusiness,
		true,
		"reversal already exists",
	)
	ErrToAccountIDRequired = newLedgerError(
		LedgerErrorCodeValidationAccountRequired,
		LedgerErrorCategoryValidation,
		true,
		"to account id is required",
	)
	ErrToAccountNotFound = newLedgerError(
		LedgerErrorCodeNotFoundToAccount,
		LedgerErrorCategoryNotFound,
		true,
		"to account not found",
	)
	ErrTransactionIDRequired = newLedgerError(
		LedgerErrorCodeValidationTransactionIDRequired,
		LedgerErrorCategoryValidation,
		true,
		"transaction id is required",
	)
	ErrTransactionNotBalanced = newLedgerError(
		LedgerErrorCodeInvariantTransactionNotBalanced,
		LedgerErrorCategoryInvariant,
		false,
		"transaction is not balanced",
	)
	ErrTransferAmountRequired = newLedgerError(
		LedgerErrorCodeValidationAmountRequired,
		LedgerErrorCategoryValidation,
		true,
		"transfer amount is required",
	)
	ErrReasonIsRequired = newLedgerError(
		LedgerErrorCodeValidationReasonRequired,
		LedgerErrorCategoryValidation,
		true,
		"reason is required",
	)
)
