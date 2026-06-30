package pghelpers

import (
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func Text(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: strings.TrimSpace(s) != ""}
}

func Timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func NumericToRat(n pgtype.Numeric) *big.Rat {
	if !n.Valid || n.NaN {
		return nil
	}
	r := new(big.Rat)
	if n.Int != nil {
		r.SetInt(n.Int)
	}
	if n.Exp > 0 {
		exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)
		r.Mul(r, new(big.Rat).SetInt(exp))
	} else if n.Exp < 0 {
		exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)
		r.Quo(r, new(big.Rat).SetInt(exp))
	}
	return r
}
