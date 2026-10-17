package num

import (
	"math/big"

	"github.com/Gulliby-development/fpbits/internal/core"
)

// Kind identifies the kind of number represented by a Number.
type Kind int

const (
	// Finite identifies a finite rational number.
	Finite Kind = iota
	// Inf identifies an infinity.
	Inf
	// NaN identifies a not-a-number value.
	NaN
)

// Number contains a number's kind, sign, and absolute finite value.
type Number struct {
	Kind Kind
	Neg  bool
	Rat  *big.Rat
}

// Parse parses an exact decimal or a lowercase infinity/NaN spelling.
// When finiteOnly is true, recognized special values are rejected.
func Parse(text string, finiteOnly bool) (Number, error) {
	neg := false
	start := 0
	if len(text) > 0 {
		switch text[0] {
		case '+':
			start++
		case '-':
			neg = true
			start++
		}
	}
	if start == len(text) {
		return Number{}, invalidNumber(text)
	}

	body := text[start:]
	if body == "inf" || body == "nan" {
		if finiteOnly {
			return Number{}, core.Failf("invalid number %s (sum takes finite numbers)", core.Show(text))
		}
		kind := Inf
		if body == "nan" {
			kind = NaN
		}
		return Number{Kind: kind, Neg: neg}, nil
	}

	// Retain at most the permitted significand length. Once over the limit,
	// continue scanning so malformed syntax still takes precedence.
	significand := make([]byte, 0, 400)
	digitCount := 0
	fractionDigits := 0
	seenPoint := false
	seenDigit := false
	i := start
	for i < len(text) {
		c := text[i]
		if c >= '0' && c <= '9' {
			seenDigit = true
			if digitCount < 401 {
				digitCount++
			}
			if digitCount <= 400 {
				significand = append(significand, c)
			}
			if seenPoint {
				fractionDigits++
			}
			i++
			continue
		}
		if c == '.' && !seenPoint {
			seenPoint = true
			i++
			continue
		}
		break
	}
	if !seenDigit {
		return Number{}, invalidNumber(text)
	}

	explicitExponent := 0
	exponentOutOfRange := false
	if i < len(text) && (text[i] == 'e' || text[i] == 'E') {
		i++
		exponentNegative := false
		if i < len(text) && (text[i] == '+' || text[i] == '-') {
			exponentNegative = text[i] == '-'
			i++
		}
		exponentStart := i
		exponentMagnitude := 0
		for i < len(text) {
			c := text[i]
			if c < '0' || c > '9' {
				break
			}
			digit := int(c - '0')
			if !exponentOutOfRange {
				if exponentMagnitude > 500 || (exponentMagnitude == 500 && digit > 0) ||
					(exponentMagnitude < 500 && exponentMagnitude*10+digit > 5000) {
					exponentOutOfRange = true
				} else {
					exponentMagnitude = exponentMagnitude*10 + digit
				}
			}
			i++
		}
		if i == exponentStart {
			return Number{}, invalidNumber(text)
		}
		if exponentNegative {
			explicitExponent = -exponentMagnitude
		} else {
			explicitExponent = exponentMagnitude
		}
	}
	if i != len(text) {
		return Number{}, invalidNumber(text)
	}

	if digitCount > 400 || exponentOutOfRange {
		return Number{}, outOfRange(text)
	}
	// The magnitude is capped at 5000 while scanning. The explicit range check
	// is kept here to make the inclusive accepted endpoints explicit.
	if explicitExponent < -5000 || explicitExponent > 5000 {
		return Number{}, outOfRange(text)
	}

	significandInt, ok := new(big.Int).SetString(string(significand), 10)
	if !ok {
		return Number{}, invalidNumber(text)
	}
	rat := new(big.Rat)
	if significandInt.Sign() == 0 {
		return Number{Kind: Finite, Neg: neg, Rat: rat}, nil
	}

	scale := explicitExponent - fractionDigits
	if scale >= 0 {
		power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		significandInt.Mul(significandInt, power)
		rat.SetInt(significandInt)
	} else {
		denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-scale)), nil)
		rat.SetFrac(significandInt, denominator)
	}
	return Number{Kind: Finite, Neg: neg, Rat: rat}, nil
}

func invalidNumber(text string) error {
	return core.Failf("invalid number %s", core.Show(text))
}

func outOfRange(text string) error {
	return core.Failf(
		"number %s is out of range (at most 400 digits, exponent from -5000 to 5000)",
		core.Show(text),
	)
}
