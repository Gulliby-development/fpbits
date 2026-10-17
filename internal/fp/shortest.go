package fp

import (
	"math/big"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/num"
)

// ShortestDigits returns the shortest decimal significand that rounds back to
// mag. exp10 is the exponent of the first significant digit.
func ShortestDigits(f core.Format, mag uint64) (digits string, exp10 int) {
	if mag == 0 {
		return "0", 0
	}

	value := MagRat(f, mag)
	decimalExp := decimalExponent10(value)
	for significantDigits := 1; ; significantDigits++ {
		scale := decimalExp - significantDigits + 1
		lowerInteger, hasRemainder := scaledFloor(value, scale)
		upperInteger := new(big.Int).Set(lowerInteger)
		if hasRemainder {
			upperInteger.Add(upperInteger, big.NewInt(1))
		}

		lower := decimalValue(lowerInteger, scale)
		lowerMatches := roundsTo(f, lower, mag)
		if !hasRemainder {
			if lowerMatches {
				return decimalDigits(lowerInteger, scale)
			}
			continue
		}

		upper := decimalValue(upperInteger, scale)
		upperMatches := roundsTo(f, upper, mag)
		switch {
		case lowerMatches && !upperMatches:
			return decimalDigits(lowerInteger, scale)
		case upperMatches && !lowerMatches:
			return decimalDigits(upperInteger, scale)
		case lowerMatches && upperMatches:
			lowerDistance := new(big.Rat).Sub(value, lower)
			upperDistance := new(big.Rat).Sub(upper, value)
			switch lowerDistance.Cmp(upperDistance) {
			case -1:
				return decimalDigits(lowerInteger, scale)
			case 1:
				return decimalDigits(upperInteger, scale)
			default:
				if decimalLastDigitEven(lowerInteger) {
					return decimalDigits(lowerInteger, scale)
				}
				return decimalDigits(upperInteger, scale)
			}
		}
	}
}

// Text returns the shortest canonical decimal text for a decoded value.
func Text(v Value) string {
	switch v.Class {
	case Zero:
		return num.Sci(v.Neg, "0", 0)
	case Subnormal, Normal:
		digits, exp10 := ShortestDigits(v.Fmt, v.Mag)
		return num.Sci(v.Neg, digits, exp10)
	case Inf:
		if v.Neg {
			return "-inf"
		}
		return "inf"
	case NaN:
		return "nan"
	default:
		return ""
	}
}

func decimalExponent10(value *big.Rat) int {
	numeratorDigits := len(new(big.Int).Abs(value.Num()).String())
	denominatorDigits := len(value.Denom().String())
	exponent := numeratorDigits - denominatorDigits
	if exponent >= 0 {
		threshold := new(big.Int).Mul(value.Denom(), powerOfTen(exponent))
		if value.Num().Cmp(threshold) < 0 {
			exponent--
		}
	} else {
		scaledNumerator := new(big.Int).Mul(value.Num(), powerOfTen(-exponent))
		if scaledNumerator.Cmp(value.Denom()) < 0 {
			exponent--
		}
	}
	return exponent
}

func scaledFloor(value *big.Rat, decimalScale int) (*big.Int, bool) {
	numerator := new(big.Int).Set(value.Num())
	denominator := new(big.Int).Set(value.Denom())
	if decimalScale >= 0 {
		denominator.Mul(denominator, powerOfTen(decimalScale))
	} else {
		numerator.Mul(numerator, powerOfTen(-decimalScale))
	}

	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	return quotient, remainder.Sign() != 0
}

func decimalValue(significand *big.Int, decimalScale int) *big.Rat {
	numerator := new(big.Int).Set(significand)
	if decimalScale >= 0 {
		numerator.Mul(numerator, powerOfTen(decimalScale))
		return new(big.Rat).SetInt(numerator)
	}
	return new(big.Rat).SetFrac(numerator, powerOfTen(-decimalScale))
}

func roundsTo(f core.Format, candidate *big.Rat, mag uint64) bool {
	rounded, overflow := EncodeMag(f, candidate)
	return !overflow && rounded == mag
}

func decimalDigits(significand *big.Int, decimalScale int) (string, int) {
	digits := significand.String()
	exp10 := decimalScale + len(digits) - 1
	for len(digits) > 1 && digits[len(digits)-1] == '0' {
		digits = digits[:len(digits)-1]
	}
	return digits, exp10
}

func decimalLastDigitEven(significand *big.Int) bool {
	digits := significand.String()
	last := digits[len(digits)-1]
	return (last-'0')%2 == 0
}

func powerOfTen(exponent int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
}
