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
type Reason string

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

type ReversalCommand struct {
	TransactionID  TransactionID
	IdempotencyKey IdempotencyKey
	Reason         Reason
}

type Transaction struct {
	ID             TransactionID
	Type           LedgerTransactionType
	IdempotencyKey IdempotencyKey
	CreatedAt      string
	FromAccountID  AccountID
	ToAccountID    AccountID
	Amount         Amount
	CurrencyCode   CurrencyCode
}

type Entry struct {
	ID            int64
	TransactionID TransactionID
	AccountID     AccountID
	Amount        Amount
	CreatedAt     string
}

type LedgerEntryInput struct {
	AccountID AccountID
	Amount    Amount
}
