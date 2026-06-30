package pghelpers

import (
	"math/big"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestText(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStr   string
		wantValid bool
	}{
		{"non-empty string", "hello", "hello", true},
		{"empty string", "", "", false},
		{"whitespace only", "   ", "   ", false},
		{"whitespace with content", "  hi  ", "  hi  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Text(tt.input)
			if got.String != tt.wantStr {
				t.Errorf("String = %q, want %q", got.String, tt.wantStr)
			}
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
		})
	}
}

func TestTimestamptz(t *testing.T) {
	t.Run("always valid", func(t *testing.T) {
		got := Timestamptz(time.Now())
		if !got.Valid {
			t.Error("expected Valid = true")
		}
	})

	t.Run("converts to UTC", func(t *testing.T) {
		loc, _ := time.LoadLocation("America/New_York")
		eastern := time.Date(2024, 1, 15, 10, 0, 0, 0, loc)
		got := Timestamptz(eastern)
		if got.Time.Location() != time.UTC {
			t.Errorf("expected UTC, got %v", got.Time.Location())
		}
		if !got.Time.Equal(eastern) {
			t.Errorf("time mismatch: got %v, want %v", got.Time, eastern.UTC())
		}
	})

	t.Run("UTC input unchanged", func(t *testing.T) {
		utc := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
		got := Timestamptz(utc)
		if !got.Time.Equal(utc) {
			t.Errorf("got %v, want %v", got.Time, utc)
		}
	})
}

func TestNumericToRat(t *testing.T) {
	mustScan := func(s string) pgtype.Numeric {
		var n pgtype.Numeric
		if err := n.Scan(s); err != nil {
			t.Fatalf("Scan(%q): %v", s, err)
		}
		return n
	}

	t.Run("invalid numeric returns nil", func(t *testing.T) {
		got := NumericToRat(pgtype.Numeric{Valid: false})
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("NaN returns nil", func(t *testing.T) {
		got := NumericToRat(pgtype.Numeric{Valid: true, NaN: true})
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("integer", func(t *testing.T) {
		got := NumericToRat(mustScan("100"))
		want := new(big.Rat).SetInt64(100)
		if got.Cmp(want) != 0 {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("decimal", func(t *testing.T) {
		got := NumericToRat(mustScan("123.45"))
		want, _ := new(big.Rat).SetString("123.45")
		if got.Cmp(want) != 0 {
			t.Errorf("got %v, want %v", got.FloatString(5), want.FloatString(5))
		}
	})

	t.Run("zero", func(t *testing.T) {
		got := NumericToRat(mustScan("0"))
		if got.Sign() != 0 {
			t.Errorf("expected 0, got %v", got)
		}
	})

	t.Run("negative", func(t *testing.T) {
		got := NumericToRat(mustScan("-50.25"))
		want, _ := new(big.Rat).SetString("-50.25")
		if got.Cmp(want) != 0 {
			t.Errorf("got %v, want %v", got.FloatString(5), want.FloatString(5))
		}
	})

	t.Run("trailing zeros preserved mathematically", func(t *testing.T) {
		a := NumericToRat(mustScan("100.00"))
		b := NumericToRat(mustScan("100.0"))
		if a.Cmp(b) != 0 {
			t.Errorf("100.00 and 100.0 should be equal, got %v vs %v", a, b)
		}
	})
}
