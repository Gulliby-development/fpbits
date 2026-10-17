package fp

import (
	"math/big"

	"github.com/Gulliby-development/fpbits/internal/num"
)

// Add adds two same-format values exactly, then rounds finite sums once.
func Add(a, b Value) Value {
	f := a.Fmt

	if a.Class == NaN || b.Class == NaN {
		bits, _ := Encode(f, num.Number{Kind: num.NaN})
		return Decode(f, bits)
	}
	if a.Class == Inf || b.Class == Inf {
		if a.Class == Inf && b.Class == Inf && a.Neg != b.Neg {
			bits, _ := Encode(f, num.Number{Kind: num.NaN})
			return Decode(f, bits)
		}
		infinity := a
		if infinity.Class != Inf {
			infinity = b
		}
		bits, _ := Encode(f, num.Number{Kind: num.Inf, Neg: infinity.Neg})
		return Decode(f, bits)
	}

	sum := new(big.Rat).Add(a.Signed(), b.Signed())
	negativeZero := sum.Sign() == 0 && a.Class == Zero && a.Neg && b.Class == Zero && b.Neg
	negative := sum.Sign() < 0 || negativeZero
	magnitude := new(big.Rat).Abs(sum)
	bits, _ := Encode(f, num.Number{Kind: num.Finite, Neg: negative, Rat: magnitude})
	return Decode(f, bits)
}
