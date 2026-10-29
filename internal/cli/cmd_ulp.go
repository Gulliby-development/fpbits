package cli

import (
	"io"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/analysis"
	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/num"
)

// RunUlp reports the spacing around a number rounded to the requested format.
func RunUlp(argv []string, stdin io.Reader) (string, error) {
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
	if len(parsed.Args) > 2 {
		return "", core.Failf("too many arguments")
	}

	format, err := core.Lookup(parsed.Args[0])
	if err != nil {
		return "", err
	}
	text := parsed.Args[1]
	if text == "-" {
		words, err := core.Words(stdin, parsed.Args[1:])
		if err != nil {
			return "", err
		}
		if len(words) != 1 {
			return "", core.Failf("<stdin>: expected 1 word, got %d", len(words))
		}
		text = words[0]
	}
	number, err := num.Parse(text, false)
	if err != nil {
		return "", err
	}
	lines, err := analysis.Ulp(format, text, number)
	if err != nil {
		return "", err
	}
	return strings.Join(lines, "\n") + "\n", nil
}
