package analysis

import (
	"strconv"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
)

// Steps returns the neighbouring representable values around v in value order.
func Steps(v fp.Value, count int) ([]string, error) {
	start := fp.Decode(v.Fmt, v.Bits)
	if start.Class == fp.NaN {
		return nil, core.Failf("cannot step from nan")
	}

	lowerRows := make([][]string, 0)
	current := start
	for distance := 1; distance <= count; distance++ {
		var ok bool
		current, ok = stepAdjacent(current, -1)
		if !ok {
			break
		}
		lowerRows = append(lowerRows, stepTableRow(current, "-"+strconv.Itoa(distance)))
	}

	rows := make([][]string, 0)
	for i := len(lowerRows) - 1; i >= 0; i-- {
		rows = append(rows, lowerRows[i])
	}

	if start.Mag == 0 && !start.Neg {
		start = fp.Decode(start.Fmt, 0)
	}
	rows = append(rows, stepTableRow(start, "0"))

	current = start
	for distance := 1; distance <= count; distance++ {
		var ok bool
		current, ok = stepAdjacent(current, 1)
		if !ok {
			break
		}
		rows = append(rows, stepTableRow(current, "+"+strconv.Itoa(distance)))
	}

	return core.RenderTable([]string{"offset", "bits", "value"}, rows), nil
}

func stepTableRow(v fp.Value, offset string) []string {
	return []string{offset, fp.Hex(v.Fmt, v.Bits), fp.Text(v)}
}

// stepAdjacent returns the immediately preceding or following value, collapsing
// the two signed zero encodings into one element.
func stepAdjacent(v fp.Value, direction int) (fp.Value, bool) {
	f := v.Fmt
	maxFinite := f.MaxFinite()
	firstNonFinite := f.FirstNonFinite()

	makeValue := func(negative bool, magnitude uint64) fp.Value {
		bits := magnitude
		if negative {
			bits |= uint64(1) << uint(f.TotalBits()-1)
		}
		return fp.Decode(f, bits)
	}
	zero := func() fp.Value {
		return fp.Decode(f, 0)
	}

	if direction < 0 {
		if v.Mag == 0 {
			return makeValue(true, 1), true
		}
		if v.Neg {
			if v.Class == fp.Inf {
				return fp.Value{}, false
			}
			if v.Mag == maxFinite {
				if f.IEEE {
					return makeValue(true, firstNonFinite), true
				}
				return fp.Value{}, false
			}
			return makeValue(true, v.Mag+1), true
		}
		if v.Class == fp.Inf {
			return makeValue(false, maxFinite), true
		}
		if v.Mag == 1 {
			return zero(), true
		}
		return makeValue(false, v.Mag-1), true
	}

	if v.Mag == 0 {
		return makeValue(false, 1), true
	}
	if v.Neg {
		if v.Class == fp.Inf {
			return makeValue(true, maxFinite), true
		}
		if v.Mag == 1 {
			return zero(), true
		}
		return makeValue(true, v.Mag-1), true
	}
	if v.Class == fp.Inf {
		return fp.Value{}, false
	}
	if v.Mag == maxFinite {
		if f.IEEE {
			return makeValue(false, firstNonFinite), true
		}
		return fp.Value{}, false
	}
	return makeValue(false, v.Mag+1), true
}
