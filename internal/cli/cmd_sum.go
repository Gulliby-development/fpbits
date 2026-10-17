package cli

import (
	"io"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/analysis"
	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/num"
)

// RunSum computes the sequentially rounded sum of finite input numbers.
func RunSum(argv []string, stdin io.Reader) (string, error) {
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
	if len(words) > 50 {
		return "", core.Failf("at most 50 terms")
	}

	nums := make([]num.Number, 0, len(words))
	for _, word := range words {
		number, err := num.Parse(word, true)
		if err != nil {
			return "", err
		}
		nums = append(nums, number)
	}

	return strings.Join(analysis.Summing(format, words, nums), "\n") + "\n", nil
}
