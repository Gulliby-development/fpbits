package fp

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/core"
)

// Class identifies a decoded floating-point value's category.
type Class int

const (
	// Zero identifies either signed zero.
	Zero Class = iota
	// Subnormal identifies a finite value with a zero exponent field.
	Subnormal
	// Normal identifies a finite value with a nonzero exponent field.
	Normal
	// Inf identifies an infinity.
	Inf
	// NaN identifies a not-a-number value.
	NaN
)

// String returns the canonical name of the class.
func (c Class) String() string {
	switch c {
	case Zero:
		return "zero"
	case Subnormal:
		return "subnormal"
	case Normal:
		return "normal"
	case Inf:
		return "inf"
	case NaN:
		return "nan"
	default:
		return ""
	}
}

// Value is a decoded floating-point bit pattern.
type Value struct {
	Fmt   core.Format
	Bits  uint64
	Neg   bool
	Mag   uint64
	Class Class
}

// Decode extracts the sign, magnitude and class of a width-valid bit pattern.
func Decode(f core.Format, bits uint64) Value {
	signPosition := f.TotalBits() - 1
	signMask := uint64(1) << uint(signPosition)
	magnitudeMask := signMask - 1
	mag := bits & magnitudeMask
	neg := bits&signMask != 0

	var class Class
	switch {
	case mag == 0:
		class = Zero
	case mag >= f.FirstNonFinite():
		if f.IEEE && mag == f.FirstNonFinite() {
			class = Inf
		} else {
			class = NaN
		}
	case mag>>uint(f.FracBits) == 0:
		class = Subnormal
	default:
		class = Normal
	}

	return Value{Fmt: f, Bits: bits, Neg: neg, Mag: mag, Class: class}
}

// ParseHex parses one fixed-width hexadecimal pattern, with an optional 0x prefix.
func ParseHex(f core.Format, text string) (uint64, error) {
	digits := text
	if len(digits) >= 2 && digits[0] == '0' && (digits[1] == 'x' || digits[1] == 'X') {
		digits = digits[2:]
	}
	if len(digits) != f.HexDigits() {
		return 0, invalidHex(f, text)
	}

	var bits uint64
	for i := 0; i < len(digits); i++ {
		c := digits[i]
		var nibble byte
		switch {
		case c >= '0' && c <= '9':
			nibble = c - '0'
		case c >= 'a' && c <= 'f':
			nibble = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			nibble = c - 'A' + 10
		default:
			return 0, invalidHex(f, text)
		}
		bits = bits<<4 | uint64(nibble)
	}
	return bits, nil
}

func invalidHex(f core.Format, text string) error {
	return core.Failf(
		"invalid hex word %s for %s (expected %d hex digits)",
		core.Show(text),
		f.Name,
		f.HexDigits(),
	)
}

// Hex renders a pattern as a lowercase, zero-padded hexadecimal word.
func Hex(f core.Format, bits uint64) string {
	digits := strconv.FormatUint(bits, 16)
	if len(digits) < f.HexDigits() {
		digits = strings.Repeat("0", f.HexDigits()-len(digits)) + digits
	}
	return "0x" + digits
}

// Binary renders a pattern as sign, exponent and fraction bit fields.
func Binary(f core.Format, bits uint64) string {
	fracMask := (uint64(1) << uint(f.FracBits)) - 1
	expMask := (uint64(1) << uint(f.ExpBits)) - 1
	frac := bits & fracMask
	exp := bits >> uint(f.FracBits) & expMask
	sign := bits >> uint(f.TotalBits()-1) & 1
	return binaryField(sign, 1) + " " +
		binaryField(exp, f.ExpBits) + " " +
		binaryField(frac, f.FracBits)
}

func binaryField(value uint64, width int) string {
	text := strconv.FormatUint(value, 2)
	if len(text) < width {
		text = strings.Repeat("0", width-len(text)) + text
	}
	return text
}

// MagRat returns the exact nonnegative rational represented by a finite magnitude.
func MagRat(f core.Format, mag uint64) *big.Rat {
	fracMask := (uint64(1) << uint(f.FracBits)) - 1
	expMask := (uint64(1) << uint(f.ExpBits)) - 1
	fraction := mag & fracMask
	exponent := mag >> uint(f.FracBits) & expMask

	significand := new(big.Int).SetUint64(fraction)
	binaryExponent := 1 - f.Bias - f.FracBits
	if exponent != 0 {
		significand.Add(significand, new(big.Int).Lsh(big.NewInt(1), uint(f.FracBits)))
		binaryExponent = int(exponent) - f.Bias - f.FracBits
	}
	if binaryExponent >= 0 {
		scaled := new(big.Int).Lsh(significand, uint(binaryExponent))
		return new(big.Rat).SetInt(scaled)
	}
	denominator := new(big.Int).Lsh(big.NewInt(1), uint(-binaryExponent))
	return new(big.Rat).SetFrac(significand, denominator)
}

// Finite reports whether the value is zero, subnormal or normal.
func (v Value) Finite() bool {
	return v.Class == Zero || v.Class == Subnormal || v.Class == Normal
}

// Signed returns an owned exact signed rational for a finite value.
// A nil result indicates a nonfinite value.
func (v Value) Signed() *big.Rat {
	if !v.Finite() {
		return nil
	}
	r := MagRat(v.Fmt, v.Mag)
	if v.Neg {
		r.Neg(r)
	}
	return r
}
