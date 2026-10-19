package cli

import (
	"io"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/analysis"
	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
)

// RunNext lists the neighbouring representable values around a bit pattern.
func RunNext(argv []string, stdin io.Reader) (string, error) {
	parsed, err := core.ParseOptions(argv, []core.Spec{
		{Name: "--count", Lo: 1, Hi: 64, Fallback: 2},
	})
	if err != nil {
		return "", err
	}
	if len(parsed.Args) == 0 {
		return "", core.Failf("missing FORMAT")
	}
	if len(parsed.Args) == 1 {
		return "", core.Failf("missing HEX")
	}
	if len(parsed.Args) > 2 {
		return "", core.Failf("too many arguments")
	}

	format, err := core.Lookup(parsed.Args[0])
	if err != nil {
		return "", err
	}
	hex := parsed.Args[1]
	if hex == "-" {
		words, err := core.Words(stdin, parsed.Args[1:])
		if err != nil {
			return "", err
		}
		if len(words) != 1 {
			return "", core.Failf("<stdin>: expected 1 word, got %d", len(words))
		}
		hex = words[0]
	}
	bits, err := fp.ParseHex(format, hex)
	if err != nil {
		return "", err
	}
	lines, err := analysis.Steps(fp.Decode(format, bits), parsed.Values["--count"])
	if err != nil {
		return "", err
	}
	return strings.Join(lines, "\n") + "\n", nil
}
