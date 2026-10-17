package analysis

import (
	"math/big"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
	"github.com/Gulliby-development/fpbits/internal/num"
)

// Ulp rounds n to f and reports its exact spacing above and below the result.
func Ulp(f core.Format, text string, n num.Number) ([]string, error) {
	bits, _ := fp.Encode(f, n)
	rounded := fp.Decode(f, bits)
	if !rounded.Finite() {
		return nil, core.Failf("ulp needs a finite number")
	}

	exponent := int(rounded.Mag >> uint(f.FracBits))
	if exponent < 1 {
		exponent = 1
	}
	ulpExponent := exponent - f.Bias - f.FracBits
	spacing := powerOfTwo(ulpExponent)

	gap := spacing
	if rounded.Mag != 0 {
		gap = new(big.Rat).Sub(
			fp.MagRat(f, rounded.Mag),
			fp.MagRat(f, rounded.Mag-1),
		)
	}

	spacingBits, _ := fp.Encode(f, num.Number{Kind: num.Finite, Rat: spacing})
	gapBits, _ := fp.Encode(f, num.Number{Kind: num.Finite, Rat: gap})

	return []string{
		"format: " + f.Name,
		"input: " + text,
		"rounded: " + fp.Text(rounded) + " (" + fp.Hex(f, rounded.Bits) + ")",
		"ulp: " + fp.Text(fp.Decode(f, spacingBits)),
		"gap toward zero: " + fp.Text(fp.Decode(f, gapBits)),
	}, nil
}

func powerOfTwo(exponent int) *big.Rat {
	if exponent >= 0 {
		value := new(big.Int).Lsh(big.NewInt(1), uint(exponent))
		return new(big.Rat).SetInt(value)
	}

	denominator := new(big.Int).Lsh(big.NewInt(1), uint(-exponent))
	return new(big.Rat).SetFrac(big.NewInt(1), denominator)
}
