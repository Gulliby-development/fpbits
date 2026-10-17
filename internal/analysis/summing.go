package analysis

import (
	"math/big"
	"strconv"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
	"github.com/Gulliby-development/fpbits/internal/num"
)

// Summing returns the sequentially rounded sums and the exact total of the
// original finite inputs.
func Summing(f core.Format, texts []string, nums []num.Number) []string {
	accumulator := fp.Decode(f, 0)
	exactTotal := new(big.Rat)
	rows := make([][]string, 0, len(nums))

	for i, number := range nums {
		bits, _ := fp.Encode(f, number)
		term := fp.Decode(f, bits)
		accumulator = fp.Add(accumulator, term)

		// Rat values are absolute; make a signed copy so caller-owned inputs
		// remain unchanged during exact accumulation.
		signedInput := new(big.Rat).Set(number.Rat)
		if number.Neg {
			signedInput.Neg(signedInput)
		}
		exactTotal.Add(exactTotal, signedInput)

		rows = append(rows, []string{
			strconv.Itoa(i + 1),
			fp.Text(term),
			fp.Text(accumulator),
		})
	}

	lines := core.RenderTable([]string{"step", "term", "sum"}, rows)
	exactMagnitude := new(big.Rat).Abs(new(big.Rat).Set(exactTotal))
	lines = append(lines, "exact: "+num.Exact(exactTotal.Sign() < 0, exactMagnitude))

	resultBits := accumulator.Bits
	lines = append(lines, "result: "+fp.Text(accumulator)+" ("+fp.Hex(f, resultBits)+")")
	if errorLine := fp.ErrorLine(exactTotal, accumulator, false); errorLine != "" {
		lines = append(lines, errorLine)
	}
	return lines
}
