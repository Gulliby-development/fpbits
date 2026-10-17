package cli

import (
	"io"
	"math/big"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
	"github.com/Gulliby-development/fpbits/internal/num"
)

// RunEncode rounds one or more decimal numbers to the selected format.
func RunEncode(argv []string, stdin io.Reader) (string, error) {
	parsed, err := core.ParseOptions(argv, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) == 0 {
		return "", core.Failf("missing FORMAT")
	}
	if len(parsed.Args) == 1 {
		return "", core.Failf("missing NUMBER")
	}

	format, err := core.Lookup(parsed.Args[0])
	if err != nil {
		return "", err
	}
	words, err := core.Words(stdin, parsed.Args[1:])
	if err != nil {
		return "", err
	}
	if len(words) > 100 {
		return "", core.Failf("at most 100 values")
	}

	lines := make([]string, 0)
	for i, word := range words {
		number, err := num.Parse(word, false)
		if err != nil {
			return "", err
		}
		bits, overflow := fp.Encode(format, number)
		value := fp.Decode(format, bits)

		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "input: "+word)
		lines = append(lines, fp.Describe(value)...)

		if number.Kind == num.Finite {
			source := new(big.Rat).Set(number.Rat)
			if number.Neg {
				source.Neg(source)
			}
			if errorLine := fp.ErrorLine(source, value, overflow); errorLine != "" {
				lines = append(lines, errorLine)
			}
		}
	}

	return strings.Join(lines, "\n") + "\n", nil
}
