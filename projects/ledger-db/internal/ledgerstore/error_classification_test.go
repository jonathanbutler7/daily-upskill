package ledgerstore

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestClassifyErrorSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want LedgerErrorInfo
	}{
		{
			name: "business insufficient funds",
			err:  ErrInsufficientFunds,
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeBusinessInsufficientFunds,
				Category: LedgerErrorCategoryBusiness,
				Expected: true,
				Message:  ErrInsufficientFunds.Error(),
			},
		},
		{
			name: "business idempotency conflict",
			err:  ErrIdempotencyConflict,
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeBusinessIdempotencyConflict,
				Category: LedgerErrorCategoryBusiness,
				Expected: true,
				Message:  ErrIdempotencyConflict.Error(),
			},
		},
		{
			name: "business reversal already exists",
			err:  ErrReversalAlreadyExists,
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeBusinessReversalAlreadyExists,
				Category: LedgerErrorCategoryBusiness,
				Expected: true,
				Message:  ErrReversalAlreadyExists.Error(),
			},
		},
		{
			name: "invariant transaction not balanced",
			err:  ErrTransactionNotBalanced,
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeInvariantTransactionNotBalanced,
				Category: LedgerErrorCategoryInvariant,
				Message:  ErrTransactionNotBalanced.Error(),
			},
		},
		{
			name: "wrapped validation account required",
			err:  fmt.Errorf("post transfer: %w", ErrFromAccountIDRequired),
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeValidationAccountRequired,
				Category: LedgerErrorCategoryValidation,
				Expected: true,
				Message:  "post transfer: from account id is required",
			},
		},
		{
			name: "external reference aliases share one code",
			err:  ErrExternalReferenceEmpty,
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeValidationExternalReferenceRequired,
				Category: LedgerErrorCategoryValidation,
				Expected: true,
				Message:  ErrExternalReferenceEmpty.Error(),
			},
		},
		{
			name: "not found to account",
			err:  ErrToAccountNotFound,
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeNotFoundToAccount,
				Category: LedgerErrorCategoryNotFound,
				Expected: true,
				Message:  ErrToAccountNotFound.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyError(tt.err)
			if got != tt.want {
				t.Fatalf("ClassifyError() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestClassifyErrorPostgresSQLStates(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want LedgerErrorInfo
	}{
		{
			name: "unique violation",
			err:  pgErr("23505", "duplicate key value violates unique constraint"),
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeDBUniqueViolation,
				Category: LedgerErrorCategoryDB,
				Message:  "ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)",
			},
		},
		{
			name: "foreign key violation",
			err:  pgErr("23503", "insert or update violates foreign key constraint"),
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeDBForeignKeyViolation,
				Category: LedgerErrorCategoryDB,
				Message:  "ERROR: insert or update violates foreign key constraint (SQLSTATE 23503)",
			},
		},
		{
			name: "check violation",
			err:  pgErr("23514", "new row violates check constraint"),
			want: LedgerErrorInfo{
				Code:     LedgerErrorCodeDBCheckViolation,
				Category: LedgerErrorCategoryDB,
				Message:  "ERROR: new row violates check constraint (SQLSTATE 23514)",
			},
		},
		{
			name: "serialization failure is retryable",
			err:  pgErr("40001", "could not serialize access due to concurrent update"),
			want: LedgerErrorInfo{
				Code:      LedgerErrorCodeDBSerializationFailure,
				Category:  LedgerErrorCategoryDB,
				Retryable: true,
				Expected:  true,
				Message:   "ERROR: could not serialize access due to concurrent update (SQLSTATE 40001)",
			},
		},
		{
			name: "deadlock is retryable",
			err:  pgErr("40P01", "deadlock detected"),
			want: LedgerErrorInfo{
				Code:      LedgerErrorCodeDBDeadlock,
				Category:  LedgerErrorCategoryDB,
				Retryable: true,
				Expected:  true,
				Message:   "ERROR: deadlock detected (SQLSTATE 40P01)",
			},
		},
		{
			name: "connection failure is retryable",
			err:  fmt.Errorf("commit: %w", pgErr("08006", "connection failure")),
			want: LedgerErrorInfo{
				Code:      LedgerErrorCodeDBUnavailable,
				Category:  LedgerErrorCategoryDB,
				Retryable: true,
				Message:   "commit: ERROR: connection failure (SQLSTATE 08006)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyError(tt.err)
			if got != tt.want {
				t.Fatalf("ClassifyError() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestClassifyErrorUnknownAndNil(t *testing.T) {
	if got := ClassifyError(nil); got != (LedgerErrorInfo{}) {
		t.Fatalf("ClassifyError(nil) = %#v, want zero value", got)
	}

	err := errors.New("disk is on fire")
	want := LedgerErrorInfo{
		Code:     LedgerErrorCodeInternalUnknown,
		Category: LedgerErrorCategoryInternal,
		Message:  err.Error(),
	}
	if got := ClassifyError(err); got != want {
		t.Fatalf("ClassifyError() = %#v, want %#v", got, want)
	}
}

func pgErr(code string, message string) error {
	return &pgconn.PgError{
		Severity: "ERROR",
		Code:     code,
		Message:  message,
	}
}
