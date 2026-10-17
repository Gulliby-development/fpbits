package fp

import (
	"math/big"

	"github.com/Gulliby-development/fpbits/internal/num"
)

// Describe returns the stable, newline-free lines describing v.
func Describe(v Value) []string {
	lines := []string{
		"format: " + v.Fmt.Name,
		"bits: " + Hex(v.Fmt, v.Bits),
		"binary: " + Binary(v.Fmt, v.Bits),
		"class: " + v.Class.String(),
		"value: " + Text(v),
	}
	if v.Finite() {
		lines = append(lines, "exact: "+num.Exact(v.Neg, MagRat(v.Fmt, v.Mag)))
	}
	return lines
}

// ErrorLine reports the exact rounding error, or an empty string when source
// is nil. The source rational is read-only and is not retained.
func ErrorLine(source *big.Rat, result Value, overflow bool) string {
	if source == nil {
		return ""
	}
	if overflow {
		return "error: overflow"
	}
	if !result.Finite() {
		return ""
	}

	difference := new(big.Rat).Sub(result.Signed(), source)
	magnitude := new(big.Rat).Abs(difference)
	return "error: " + num.Exact(difference.Sign() < 0, magnitude)
}
