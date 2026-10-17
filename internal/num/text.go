package num

import (
	"math/big"
	"strconv"
	"strings"
)

// Sci renders significant decimal digits using exp10 as the exponent of the
// first significant digit.
func Sci(neg bool, digits string, exp10 int) string {
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		if neg {
			return "-0e0"
		}
		return "0e0"
	}
	digits = strings.TrimRight(digits, "0")

	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.WriteByte(digits[0])
	if len(digits) > 1 {
		b.WriteByte('.')
		b.WriteString(digits[1:])
	}
	b.WriteByte('e')
	b.WriteString(strconv.Itoa(exp10))
	return b.String()
}

// Exact renders a nonnegative rational with a terminating decimal expansion.
func Exact(neg bool, r *big.Rat) string {
	if r == nil || r.Sign() == 0 {
		return Sci(neg, "0", 0)
	}

	numerator := new(big.Int).Abs(r.Num())
	denominator := new(big.Int).Set(r.Denom())

	twos := factorCount(denominator, 2)
	fives := factorCount(denominator, 5)
	if denominator.Cmp(big.NewInt(1)) != 0 {
		// Nonterminating rationals are outside Exact's contract.
		return ""
	}

	scale := twos
	if fives > scale {
		scale = fives
	}
	if scale > twos {
		twoPower := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(scale-twos)), nil)
		numerator.Mul(numerator, twoPower)
	}
	if scale > fives {
		fivePower := new(big.Int).Exp(big.NewInt(5), big.NewInt(int64(scale-fives)), nil)
		numerator.Mul(numerator, fivePower)
	}

	digits := numerator.String()
	return Sci(neg, digits, len(digits)-scale-1)
}

// factorCount removes all factors of factor from n and returns their count.
// Both n and the receiver values used during division are owned by this call.
func factorCount(n *big.Int, factor int64) int {
	divisor := big.NewInt(factor)
	quotient := new(big.Int)
	remainder := new(big.Int)
	count := 0
	for {
		quotient.QuoRem(n, divisor, remainder)
		if remainder.Sign() != 0 {
			return count
		}
		n.Set(quotient)
		count++
	}
}
