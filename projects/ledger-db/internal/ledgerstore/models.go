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

type TransferCommand struct {
	FromAccountID  AccountID
	ToAccountID    AccountID
	Amount         Amount
	IdempotencyKey string
}

type PostTransferCommand struct {
	FromAccountID  AccountID
	ToAccountID    AccountID
	Amount         Amount
	IdempotencyKey IdempotencyKey
}

type PostExternalTransferCommand struct {
	ToAccountID       AccountID
	TransferAmount    Amount
	Rail              PaymentRail
	ExternalReference ExternalReference
	IdempotencyKey    IdempotencyKey
}
