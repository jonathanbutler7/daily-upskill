package ledgerstore

// Named types for values that cross ledgerstore helper boundaries.
type AccountID int64
type TransactionID int64
type Amount int64
type CurrencyCode string
type IdempotencyKey string
type PaymentRail string
type ExternalReference string
type ExternalTransferDirection string
type ExternalTransferStatus string
type LedgerTransactionType string

const (
	LedgerTransactionTypeTransfer LedgerTransactionType = "transfer"
	LedgerTransactionTypeDeposit  LedgerTransactionType = "deposit"
	LedgerTransactionTypeReversal LedgerTransactionType = "reversal"

	ExternalTransferDirectionDeposit    ExternalTransferDirection = "deposit"
	ExternalTransferDirectionWithdrawal ExternalTransferDirection = "withdrawal"

	ExternalTransferStatusPosted   ExternalTransferStatus = "posted"
	ExternalTransferStatusPending  ExternalTransferStatus = "pending"
	ExternalTransferStatusFailed   ExternalTransferStatus = "failed"
	ExternalTransferStatusCanceled ExternalTransferStatus = "canceled"
)

type TransferCommand struct {
	FromAccountID  AccountID
	ToAccountID    AccountID
	Amount         Amount
	IdempotencyKey string
}

type PostTransferCommand struct {
	IdempotencyKey IdempotencyKey
	Amount         Amount
	ToAccountID    AccountID
	FromAccountID  AccountID
}

type PostExternalTransferCommand struct {
	IdempotencyKey            IdempotencyKey
	TransferAmount            Amount
	UserAccountID             AccountID
	Rail                      PaymentRail
	ExternalReference         ExternalReference
	ExternalTransferDirection ExternalTransferDirection
}
