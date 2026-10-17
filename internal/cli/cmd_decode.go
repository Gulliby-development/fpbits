package cli

import (
	"io"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
)

// RunDecode decodes one or more hexadecimal words in the selected format.
func RunDecode(argv []string, stdin io.Reader) (string, error) {
	parsed, err := core.ParseOptions(argv, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) == 0 {
		return "", core.Failf("missing FORMAT")
	}
	if len(parsed.Args) == 1 {
		return "", core.Failf("missing HEX")
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
		bits, err := fp.ParseHex(format, word)
		if err != nil {
			return "", err
		}
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, fp.Describe(fp.Decode(format, bits))...)
	}

	return strings.Join(lines, "\n") + "\n", nil
}
