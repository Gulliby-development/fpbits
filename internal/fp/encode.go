package fp

import (
	"math/big"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/num"
)

// EncodeMag rounds a nonnegative rational to the nearest magnitude pattern,
// using even-pattern tie breaking. The formula continues through the first
// nonfinite pattern to define the overflow threshold.
func EncodeMag(f core.Format, r *big.Rat) (mag uint64, overflow bool) {
	if r == nil || r.Sign() <= 0 {
		return 0, false
	}

	firstNonFinite := f.FirstNonFinite()
	low, high := uint64(0), firstNonFinite
	for low < high {
		mid := low + (high-low)/2
		if MagRat(f, mid).Cmp(r) < 0 {
			low = mid + 1
		} else {
			high = mid
		}
	}

	upper := low
	upperValue := MagRat(f, upper)
	if upperValue.Cmp(r) == 0 || upper == 0 {
		return upper, upper >= firstNonFinite
	}

	lower := upper - 1
	lowerDistance := new(big.Rat).Sub(r, MagRat(f, lower))
	upperDistance := new(big.Rat).Sub(upperValue, r)
	switch lowerDistance.Cmp(upperDistance) {
	case -1:
		return lower, lower >= firstNonFinite
	case 1:
		return upper, upper >= firstNonFinite
	default:
		if lower&1 == 0 {
			return lower, lower >= firstNonFinite
		}
		return upper, upper >= firstNonFinite
	}
}

// Encode rounds a signed number into a format and returns its complete pattern.
func Encode(f core.Format, n num.Number) (bits uint64, overflow bool) {
	var mag uint64
	switch n.Kind {
	case num.Inf:
		mag = f.FirstNonFinite()
	case num.NaN:
		if f.IEEE {
			mag = f.FirstNonFinite() | (uint64(1) << uint(f.FracBits-1))
		} else {
			mag = f.FirstNonFinite()
		}
	default:
		mag, overflow = EncodeMag(f, n.Rat)
	}
	if n.Neg {
		mag |= uint64(1) << uint(f.TotalBits()-1)
	}
	return mag, overflow
}
