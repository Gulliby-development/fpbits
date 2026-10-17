package analysis

import (
	"math/big"
	"strconv"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
)

// Limits returns the exact representable limits and encoding counts for f.
func Limits(f core.Format) []string {
	smallestSubnormal := fp.Decode(f, 1)
	smallestNormal := fp.Decode(f, uint64(1)<<uint(f.FracBits))
	largest := fp.Decode(f, f.MaxFinite())

	epsilonRat := new(big.Rat).SetFrac(
		big.NewInt(1),
		new(big.Int).Lsh(big.NewInt(1), uint(f.FracBits)),
	)
	epsilonMag, _ := fp.EncodeMag(f, epsilonRat)
	epsilon := fp.Decode(f, epsilonMag)

	infinities := uint64(0)
	nanEncodings := uint64(2)
	if f.IEEE {
		infinities = 2
		nanEncodings = 2 * ((uint64(1) << uint(f.FracBits)) - 1)
	}
	finiteEncodings := 2 * f.FirstNonFinite()

	return []string{
		"format: " + f.Name,
		"layout: 1 sign bit, " +
			core.Plural(f.ExpBits, "exponent bit", "exponent bits") + ", " +
			core.Plural(f.FracBits, "fraction bit", "fraction bits"),
		"bias: " + strconv.Itoa(f.Bias),
		"smallest subnormal: " + fp.Text(smallestSubnormal) + " (" + fp.Hex(f, smallestSubnormal.Bits) + ")",
		"smallest normal: " + fp.Text(smallestNormal) + " (" + fp.Hex(f, smallestNormal.Bits) + ")",
		"largest: " + fp.Text(largest) + " (" + fp.Hex(f, largest.Bits) + ")",
		"epsilon: " + fp.Text(epsilon),
		"infinities: " + strconv.FormatUint(infinities, 10),
		"nan encodings: " + strconv.FormatUint(nanEncodings, 10),
		"finite encodings: " + strconv.FormatUint(finiteEncodings, 10),
	}
}
