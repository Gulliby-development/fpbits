package cli

import (
	"io"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
	"github.com/Gulliby-development/fpbits/internal/num"
)

// RunConvert re-encodes one or more bit patterns from one format to another.
func RunConvert(argv []string, stdin io.Reader) (string, error) {
	parsed, err := core.ParseOptions(argv, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) == 0 {
		return "", core.Failf("missing FROM")
	}
	if len(parsed.Args) == 1 {
		return "", core.Failf("missing TO")
	}
	if len(parsed.Args) == 2 {
		return "", core.Failf("missing HEX")
	}

	from, err := core.Lookup(parsed.Args[0])
	if err != nil {
		return "", err
	}
	to, err := core.Lookup(parsed.Args[1])
	if err != nil {
		return "", err
	}
	words, err := core.Words(stdin, parsed.Args[2:])
	if err != nil {
		return "", err
	}
	if len(words) > 100 {
		return "", core.Failf("at most 100 values")
	}

	lines := make([]string, 0)
	for i, word := range words {
		bits, err := fp.ParseHex(from, word)
		if err != nil {
			return "", err
		}
		original := fp.Decode(from, bits)

		input := num.Number{Neg: original.Neg}
		source := original.Signed()
		switch original.Class {
		case fp.Zero, fp.Subnormal, fp.Normal:
			input.Kind = num.Finite
			input.Rat = fp.MagRat(from, original.Mag)
			source = original.Signed()
		case fp.Inf:
			input.Kind = num.Inf
		case fp.NaN:
			input.Kind = num.NaN
		}

		convertedBits, overflow := fp.Encode(to, input)
		converted := fp.Decode(to, convertedBits)

		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "input: "+fp.Hex(from, bits)+" ("+from.Name+")")
		lines = append(lines, fp.Describe(converted)...)
		if errorLine := fp.ErrorLine(source, converted, overflow); errorLine != "" {
			lines = append(lines, errorLine)
		}
	}

	return strings.Join(lines, "\n") + "\n", nil
}
