package cli

import (
	"io"
	"strings"

	"github.com/Gulliby-development/fpbits/internal/core"
)

// Usage contains the command-line usage lines.
var Usage = []string{
	"usage: fpbits decode FORMAT HEX...",
	"       fpbits encode FORMAT NUMBER...",
	"       fpbits convert FROM TO HEX...",
	"       fpbits ulp FORMAT NUMBER",
	"       fpbits next [--count N] FORMAT HEX",
	"       fpbits range FORMAT",
	"       fpbits sum FORMAT NUMBER...",
}

// Run dispatches one command and writes its complete output or diagnostic.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = stderr.Write([]byte("fpbits: missing command (decode, encode, convert, ulp, next, range or sum)\n"))
		return 2
	}
	if len(args) == 1 && args[0] == "--help" {
		output := strings.Join(Usage, "\n") + "\n"
		_, _ = stdout.Write([]byte(output))
		return 0
	}

	var output string
	var err error
	switch args[0] {
	case "decode":
		output, err = RunDecode(args[1:], stdin)
	case "encode":
		output, err = RunEncode(args[1:], stdin)
	case "convert":
		output, err = RunConvert(args[1:], stdin)
	case "ulp":
		output, err = RunUlp(args[1:], stdin)
	case "next":
		output, err = RunNext(args[1:], stdin)
	case "range":
		output, err = RunRange(args[1:], stdin)
	case "sum":
		output, err = RunSum(args[1:], stdin)
	default:
		err = core.Failf("unknown command %s", core.Show(args[0]))
	}

	if err != nil {
		_, _ = stderr.Write([]byte("fpbits: " + err.Error() + "\n"))
		return 2
	}
	_, _ = stdout.Write([]byte(output))
	return 0
}
