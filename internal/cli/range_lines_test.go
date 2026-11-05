package cli_test

import (
	"strings"
	"testing"
)

const f16Report = `format: f16
layout: 1 sign bit, 5 exponent bits, 10 fraction bits
bias: 15
smallest subnormal: 6e-8 (0x0001)
smallest normal: 6.104e-5 (0x0400)
largest: 6.55e4 (0x7bff)
epsilon: 9.77e-4
infinities: 2
nan encodings: 2046
finite encodings: 63488
`

func TestRangeLinesKeepsTheFirstLines(t *testing.T) {
	all := strings.SplitAfter(f16Report, "\n")
	for n := 1; n <= 10; n++ {
		args := []string{"range", "--lines", string(rune('0' + n%10)), "f16"}
		if n == 10 {
			args[2] = "10"
		}
		want := strings.Join(all[:n], "")
		status, stdout, stderr, _, _ := runCLI(args, nil)
		if status != 0 || stderr != "" || stdout != want {
			t.Errorf("range --lines %d = status %d, stdout %q, stderr %q; want %q", n, status, stdout, stderr, want)
		}
	}
}

func TestRangeLinesAcceptsTheEqualsFormAndAnyPosition(t *testing.T) {
	want := "format: f16\nlayout: 1 sign bit, 5 exponent bits, 10 fraction bits\nbias: 15\n"
	for _, args := range [][]string{
		{"range", "--lines=3", "f16"},
		{"range", "f16", "--lines=3"},
		{"range", "f16", "--lines", "3"},
		{"range", "--lines", "9", "--lines=3", "f16"},
	} {
		status, stdout, stderr, _, _ := runCLI(args, nil)
		if status != 0 || stderr != "" || stdout != want {
			t.Errorf("%v = status %d, stdout %q, stderr %q; want %q", args, status, stdout, stderr, want)
		}
	}
}

func TestRangeLinesDefaultsToTheWholeReport(t *testing.T) {
	status, stdout, stderr, _, _ := runCLI([]string{"range", "f16"}, nil)
	if status != 0 || stderr != "" || stdout != f16Report {
		t.Errorf("range f16 = status %d, stdout %q, stderr %q", status, stdout, stderr)
	}
}

func TestRangeLinesErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"zero", []string{"range", "--lines=0", "f16"}, "fpbits: invalid --lines \"0\" (expected an integer from 1 to 10)\n"},
		{"eleven", []string{"range", "--lines", "11", "f16"}, "fpbits: invalid --lines \"11\" (expected an integer from 1 to 10)\n"},
		{"leading zero", []string{"range", "--lines=03", "f16"}, "fpbits: invalid --lines \"03\" (expected an integer from 1 to 10)\n"},
		{"text", []string{"range", "--lines=x", "f16"}, "fpbits: invalid --lines \"x\" (expected an integer from 1 to 10)\n"},
		{"needs a value", []string{"range", "f16", "--lines"}, "fpbits: --lines needs a value\n"},
		{"unknown option first", []string{"range", "--lines=0", "--bogus"}, "fpbits: unknown option \"--bogus\"\n"},
		{"invalid value before missing format", []string{"range", "--lines=99"}, "fpbits: invalid --lines \"99\" (expected an integer from 1 to 10)\n"},
		{"missing format", []string{"range", "--lines=2"}, "fpbits: missing FORMAT\n"},
		{"too many before the format lookup", []string{"range", "--lines=2", "f17", "f16"}, "fpbits: too many arguments\n"},
		{"unknown format", []string{"range", "--lines=2", "f17"}, "fpbits: unknown format \"f17\" (expected f16, bf16, f32, f64, e5m2 or e4m3)\n"},
		{"count is still unknown", []string{"range", "--count=2", "f16"}, "fpbits: unknown option \"--count=2\"\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, _, _ := runCLI(test.args, nil)
			if status != 2 || stdout != "" || stderr != test.want {
				t.Errorf("%v = status %d, stdout %q, stderr %q; want status 2 and %q", test.args, status, stdout, stderr, test.want)
			}
		})
	}
}
