package cli

import (
	"io"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/analysis"
	"github.com/Gulliby-development/fpbits/internal/core"
)

// RunRange reports the representable limits of one floating-point format.
func RunRange(argv []string, stdin io.Reader) (string, error) {
	parsed, err := core.ParseOptions(argv, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) == 0 {
		return "", core.Failf("missing FORMAT")
	}
	if len(parsed.Args) > 1 {
		return "", core.Failf("too many arguments")
	}

	format, err := core.Lookup(parsed.Args[0])
	if err != nil {
		return "", err
	}
	return strings.Join(analysis.Limits(format), "\n") + "\n", nil
}
