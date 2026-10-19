package cli_test

import (
	"bytes"
	"io"

	"testing"

	"github.com/Gulliby-development/fpbits/internal/cli"
)

type writeCounter struct {
	bytes.Buffer
	calls int
}

func (w *writeCounter) Write(p []byte) (int, error) {
	w.calls++
	return w.Buffer.Write(p)
}

func runCLI(args []string, input []byte) (int, string, string, int, int) {
	stdout := &writeCounter{}
	stderr := &writeCounter{}
	status := cli.Run(args, bytes.NewReader(input), stdout, stderr)
	return status, stdout.String(), stderr.String(), stdout.calls, stderr.calls
}

func TestWorkedCommandOutputs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "decode f16 finite, negative zero, and negative infinity",
			args: []string{"decode", "f16", "7bff", "8000", "fc00"},
			want: `format: f16
bits: 0x7bff
binary: 0 11110 1111111111
class: normal
value: 6.55e4
exact: 6.5504e4

format: f16
bits: 0x8000
binary: 1 00000 0000000000
class: zero
value: -0e0
exact: -0e0

format: f16
bits: 0xfc00
binary: 1 11111 0000000000
class: inf
value: -inf
`,
		},
		{
			name: "decode e5m2 shortest decimal ties",
			args: []string{"decode", "e5m2", "3d", "3f"},
			want: `format: e5m2
bits: 0x3d
binary: 0 01111 01
class: normal
value: 1.2e0
exact: 1.25e0

format: e5m2
bits: 0x3f
binary: 0 01111 11
class: normal
value: 1.8e0
exact: 1.75e0
`,
		},
		{
			name: "encode f16 finite rounding underflow and overflow",
			args: []string{"encode", "f16", "0.1", "-1e-30", "1e10"},
			want: `input: 0.1
format: f16
bits: 0x2e66
binary: 0 01011 1001100110
class: normal
value: 1e-1
exact: 9.99755859375e-2
error: -2.44140625e-5

input: -1e-30
format: f16
bits: 0x8000
binary: 1 00000 0000000000
class: zero
value: -0e0
exact: -0e0
error: 1e-30

input: 1e10
format: f16
bits: 0x7c00
binary: 0 11111 0000000000
class: inf
value: inf
error: overflow
`,
		},
		{
			name: "encode e4m3 overflow midpoint",
			args: []string{"encode", "e4m3", "464", "464.01"},
			want: `input: 464
format: e4m3
bits: 0x7e
binary: 0 1111 110
class: normal
value: 4.5e2
exact: 4.48e2
error: -1.6e1

input: 464.01
format: e4m3
bits: 0x7f
binary: 0 1111 111
class: nan
value: nan
error: overflow
`,
		},
		{
			name: "convert finite overflow and infinity to e4m3",
			args: []string{"convert", "f16", "e4m3", "7bff", "7c00"},
			want: `input: 0x7bff (f16)
format: e4m3
bits: 0x7f
binary: 0 1111 111
class: nan
value: nan
error: overflow

input: 0x7c00 (f16)
format: e4m3
bits: 0x7f
binary: 0 1111 111
class: nan
value: nan
`,
		},
		{
			name: "ulp f16 one",
			args: []string{"ulp", "f16", "1"},
			want: `format: f16
input: 1
rounded: 1e0 (0x3c00)
ulp: 9.77e-4
gap toward zero: 4.883e-4
`,
		},
		{
			name: "next from f16 negative zero",
			args: []string{"next", "--count", "2", "f16", "8000"},
			want: `offset  bits    value
-2      0x8002  -1e-7
-1      0x8001  -6e-8
0       0x8000  -0e0
+1      0x0001  6e-8
+2      0x0002  1e-7
`,
		},
		{
			name: "next e4m3 upper finite endpoint clips",
			args: []string{"next", "--count", "3", "e4m3", "7e"},
			want: `offset  bits  value
-3      0x7b  3.5e2
-2      0x7c  4e2
-1      0x7d  4.2e2
0       0x7e  4.5e2
`,
		},
		{
			name: "range f16",
			args: []string{"range", "f16"},
			want: `format: f16
layout: 1 sign bit, 5 exponent bits, 10 fraction bits
bias: 15
smallest subnormal: 6e-8 (0x0001)
smallest normal: 6.104e-5 (0x0400)
largest: 6.55e4 (0x7bff)
epsilon: 9.77e-4
infinities: 2
nan encodings: 2046
finite encodings: 63488
`,
		},
		{
			name: "sum f64 preserves exact and sequential totals",
			args: []string{"sum", "f64", "0.1", "0.2"},
			want: `step  term  sum
1     1e-1  1e-1
2     2e-1  3.0000000000000004e-1
exact: 3e-1
result: 3.0000000000000004e-1 (0x3fd3333333333334)
error: 4.44089209850062616169452667236328125e-17
`,
		},
		{
			name: "sum f16 sequential rounding loses three increments",
			args: []string{"sum", "f16", "2048", "1", "1", "1"},
			want: `step  term     sum
1     2.048e3  2.048e3
2     1e0      2.048e3
3     1e0      2.048e3
4     1e0      2.048e3
exact: 2.051e3
result: 2.048e3 (0x6800)
error: -3e0
`,
		},
		{
			name: "sum two negative zero terms",
			args: []string{"sum", "f16", "-0", "-0"},
			want: `step  term  sum
1     -0e0  0e0
2     -0e0  0e0
exact: 0e0
result: 0e0 (0x0000)
error: 0e0
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, stdoutWrites, stderrWrites := runCLI(test.args, nil)
			if status != 0 {
				t.Fatalf("cli.Run(%q) status = %d, want 0; stderr = %q", test.args, status, stderr)
			}
			if stdout != test.want {
				t.Errorf("cli.Run(%q) stdout = %q, want %q", test.args, stdout, test.want)
			}
			if stderr != "" {
				t.Errorf("cli.Run(%q) stderr = %q, want empty", test.args, stderr)
			}
			if stdoutWrites != 1 || stderrWrites != 0 {
				t.Errorf("cli.Run(%q) writes = stdout %d, stderr %d; want 1 and 0", test.args, stdoutWrites, stderrWrites)
			}
		})
	}
}

func TestAllFormatsHaveLiteralRangeReports(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "bf16",
			want: `format: bf16
layout: 1 sign bit, 8 exponent bits, 7 fraction bits
bias: 127
smallest subnormal: 9e-41 (0x0001)
smallest normal: 1.18e-38 (0x0080)
largest: 3.39e38 (0x7f7f)
epsilon: 7.8e-3
infinities: 2
nan encodings: 254
finite encodings: 65280
`,
		},
		{
			name: "f32",
			want: `format: f32
layout: 1 sign bit, 8 exponent bits, 23 fraction bits
bias: 127
smallest subnormal: 1e-45 (0x00000001)
smallest normal: 1.1754944e-38 (0x00800000)
largest: 3.4028235e38 (0x7f7fffff)
epsilon: 1.1920929e-7
infinities: 2
nan encodings: 16777214
finite encodings: 4278190080
`,
		},
		{
			name: "f64",
			want: `format: f64
layout: 1 sign bit, 11 exponent bits, 52 fraction bits
bias: 1023
smallest subnormal: 5e-324 (0x0000000000000001)
smallest normal: 2.2250738585072014e-308 (0x0010000000000000)
largest: 1.7976931348623157e308 (0x7fefffffffffffff)
epsilon: 2.220446049250313e-16
infinities: 2
nan encodings: 9007199254740990
finite encodings: 18437736874454810624
`,
		},
		{
			name: "e5m2",
			want: `format: e5m2
layout: 1 sign bit, 5 exponent bits, 2 fraction bits
bias: 15
smallest subnormal: 2e-5 (0x01)
smallest normal: 6e-5 (0x04)
largest: 6e4 (0x7b)
epsilon: 2.5e-1
infinities: 2
nan encodings: 6
finite encodings: 248
`,
		},
		{
			name: "e4m3",
			want: `format: e4m3
layout: 1 sign bit, 4 exponent bits, 3 fraction bits
bias: 7
smallest subnormal: 2e-3 (0x01)
smallest normal: 1.6e-2 (0x08)
largest: 4.5e2 (0x7e)
epsilon: 1.3e-1
infinities: 0
nan encodings: 2
finite encodings: 254
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, _, _ := runCLI([]string{"range", test.name}, nil)
			if status != 0 || stderr != "" || stdout != test.want {
				t.Errorf("range %s = status %d, stdout %q, stderr %q; want status 0 and exact report %q", test.name, status, stdout, stderr, test.want)
			}
		})
	}
}

func TestExportedAdaptersReturnLiteralLFReports(t *testing.T) {
	tests := []struct {
		name string
		run  func([]string, io.Reader) (string, error)
		args []string
		want string
	}{
		{
			name: "RunDecode",
			run:  cli.RunDecode,
			args: []string{"f16", "0000"},
			want: `format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0
`,
		},
		{
			name: "RunEncode",
			run:  cli.RunEncode,
			args: []string{"f16", ".5"},
			want: `input: .5
format: f16
bits: 0x3800
binary: 0 01110 0000000000
class: normal
value: 5e-1
exact: 5e-1
error: 0e0
`,
		},
		{
			name: "RunConvert",
			run:  cli.RunConvert,
			args: []string{"f16", "f32", "3c00"},
			want: `input: 0x3c00 (f16)
format: f32
bits: 0x3f800000
binary: 0 01111111 00000000000000000000000
class: normal
value: 1e0
exact: 1e0
error: 0e0
`,
		},
		{
			name: "RunUlp",
			run:  cli.RunUlp,
			args: []string{"f16", ".5"},
			want: `format: f16
input: .5
rounded: 5e-1 (0x3800)
ulp: 4.883e-4
gap toward zero: 2.441e-4
`,
		},
		{
			name: "RunNext",
			run:  cli.RunNext,
			args: []string{"--count=1", "f16", "0000"},
			want: `offset  bits    value
-1      0x8001  -6e-8
0       0x0000  0e0
+1      0x0001  6e-8
`,
		},
		{
			name: "RunRange",
			run:  cli.RunRange,
			args: []string{"e4m3"},
			want: `format: e4m3
layout: 1 sign bit, 4 exponent bits, 3 fraction bits
bias: 7
smallest subnormal: 2e-3 (0x01)
smallest normal: 1.6e-2 (0x08)
largest: 4.5e2 (0x7e)
epsilon: 1.3e-1
infinities: 0
nan encodings: 2
finite encodings: 254
`,
		},
		{
			name: "RunSum",
			run:  cli.RunSum,
			args: []string{"f16", "1"},
			want: `step  term  sum
1     1e0   1e0
exact: 1e0
result: 1e0 (0x3c00)
error: 0e0
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.run(test.args, bytes.NewReader(nil))
			if err != nil {
				t.Fatalf("%s(%q) returned error: %v", test.name, test.args, err)
			}
			if got != test.want {
				t.Errorf("%s(%q) = %q, want %q", test.name, test.args, got, test.want)
			}
		})
	}
}

func TestDecodeStdinBOMAndCRLFAreLiteralEquivalent(t *testing.T) {
	args := []string{"decode", "f16", "-"}
	input := []byte{0xef, 0xbb, 0xbf, '3', 'c', '0', '0', '\r', '\n', '0', 'x', '7', 'B', 'F', 'F', '\n'}
	want := `format: f16
bits: 0x3c00
binary: 0 01111 0000000000
class: normal
value: 1e0
exact: 1e0

format: f16
bits: 0x7bff
binary: 0 11110 1111111111
class: normal
value: 6.55e4
exact: 6.5504e4
`

	status, stdout, stderr, _, _ := runCLI(args, input)
	if status != 0 || stderr != "" || stdout != want {
		t.Errorf("decode stdin returned status %d, stdout %q, stderr %q; want exact two-block report %q", status, stdout, stderr, want)
	}
}

func TestCLIInputGrammarSpecialsAndBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		input      []byte
		wantStatus int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "hex accepts uppercase optional prefix",
			args:       []string{"decode", "f16", "0X7BFF"},
			wantStatus: 0,
			wantStdout: `format: f16
bits: 0x7bff
binary: 0 11110 1111111111
class: normal
value: 6.55e4
exact: 6.5504e4
`,
		},
		{
			name:       "negative zero and subnormal decode",
			args:       []string{"decode", "f16", "8000", "0001"},
			wantStatus: 0,
			wantStdout: `format: f16
bits: 0x8000
binary: 1 00000 0000000000
class: zero
value: -0e0
exact: -0e0

format: f16
bits: 0x0001
binary: 0 00000 0000000001
class: subnormal
value: 6e-8
exact: 5.9604644775390625e-8
`,
		},
		{
			name:       "IEEE infinity and quiet NaN",
			args:       []string{"decode", "f16", "7c00", "7e00"},
			wantStatus: 0,
			wantStdout: `format: f16
bits: 0x7c00
binary: 0 11111 0000000000
class: inf
value: inf

format: f16
bits: 0x7e00
binary: 0 11111 1000000000
class: nan
value: nan
`,
		},
		{
			name:       "e4m3 has signed NaN but no infinity",
			args:       []string{"decode", "e4m3", "7f", "ff"},
			wantStatus: 0,
			wantStdout: `format: e4m3
bits: 0x7f
binary: 0 1111 111
class: nan
value: nan

format: e4m3
bits: 0xff
binary: 1 1111 111
class: nan
value: nan
`,
		},
		{
			name:       "finite conversion keeps exact error",
			args:       []string{"convert", "f16", "e5m2", "3c00"},
			wantStatus: 0,
			wantStdout: `input: 0x3c00 (f16)
format: e5m2
bits: 0x3c
binary: 0 01111 00
class: normal
value: 1e0
exact: 1e0
error: 0e0
`,
		},
		{
			name:       "finite inexact conversion reports signed rational error",
			args:       []string{"convert", "f16", "e5m2", "2e66"},
			wantStatus: 0,
			wantStdout: `input: 0x2e66 (f16)
format: e5m2
bits: 0x2e
binary: 0 01011 10
class: normal
value: 9e-2
exact: 9.375e-2
error: -6.2255859375e-3
`,
		},
		{
			name:       "conversion preserves negative zero",
			args:       []string{"convert", "f16", "f32", "8000"},
			wantStatus: 0,
			wantStdout: `input: 0x8000 (f16)
format: f32
bits: 0x80000000
binary: 1 00000000 00000000000000000000000
class: zero
value: -0e0
exact: -0e0
error: 0e0
`,
		},
		{
			name:       "finite decimal grammar accepts integer and fractional forms",
			args:       []string{"encode", "f16", "12", "1.5"},
			wantStatus: 0,
			wantStdout: `input: 12
format: f16
bits: 0x4a00
binary: 0 10010 1000000000
class: normal
value: 1.2e1
exact: 1.2e1
error: 0e0

input: 1.5
format: f16
bits: 0x3e00
binary: 0 01111 1000000000
class: normal
value: 1.5e0
exact: 1.5e0
error: 0e0
`,
		},
		{
			name:       "finite decimal grammar accepts a trailing decimal point and exponent case",
			args:       []string{"encode", "f16", "5.", "2E+2"},
			wantStatus: 0,
			wantStdout: `input: 5.
format: f16
bits: 0x4500
binary: 0 10001 0100000000
class: normal
value: 5e0
exact: 5e0
error: 0e0

input: 2E+2
format: f16
bits: 0x5a40
binary: 0 10110 1001000000
class: normal
value: 2e2
exact: 2e2
error: 0e0
`,
		},
		{
			name:       "decimal grammar accepts leading point and explicit positive sign",
			args:       []string{"encode", "f16", "+.5", "1e-3"},
			wantStatus: 0,
			wantStdout: `input: +.5
format: f16
bits: 0x3800
binary: 0 01110 0000000000
class: normal
value: 5e-1
exact: 5e-1
error: 0e0

input: 1e-3
format: f16
bits: 0x1419
binary: 0 00101 0000011001
class: normal
value: 1e-3
exact: 1.00040435791015625e-3
error: 4.0435791015625e-7
`,
		},
		{
			name:       "invalid UTF-8 bytes from stdin are diagnosed by the CLI",
			args:       []string{"decode", "f16", "-"},
			input:      []byte{0x33, 0x63, 0x30, 0x30, 0x20, 0xff, 0x0a},
			wantStatus: 2,
			wantStderr: "fpbits: <stdin>: input is not valid UTF-8\n",
		},
		{
			name:       "blank stdin has no words",
			args:       []string{"decode", "f16", "-"},
			input:      []byte(" \t\r\n"),
			wantStatus: 2,
			wantStderr: "fpbits: <stdin>: no input\n",
		},
		{
			name:       "dash among arguments is an ordinary invalid word",
			args:       []string{"decode", "f16", "-", "0000"},
			input:      []byte("3c00"),
			wantStatus: 2,
			wantStderr: "fpbits: invalid hex word \"-\" for f16 (expected 4 hex digits)\n",
		},
		{
			name:       "ulp never reads stdin",
			args:       []string{"ulp", "f16", "1"},
			input:      []byte{0xff},
			wantStatus: 0,
			wantStdout: `format: f16
input: 1
rounded: 1e0 (0x3c00)
ulp: 9.77e-4
gap toward zero: 4.883e-4
`,
		},
		{
			name:       "unknown format lists supported formats in order",
			args:       []string{"decode", "f17", "0000"},
			wantStatus: 2,
			wantStderr: "fpbits: unknown format \"f17\" (expected f16, bf16, f32, f64, e5m2 or e4m3)\n",
		},
		{
			name:       "malformed hex word",
			args:       []string{"decode", "f16", "3c0"},
			wantStatus: 2,
			wantStderr: "fpbits: invalid hex word \"3c0\" for f16 (expected 4 hex digits)\n",
		},
		{
			name:       "malformed decimal is not treated as a number",
			args:       []string{"encode", "f16", "0x10"},
			wantStatus: 2,
			wantStderr: "fpbits: invalid number \"0x10\"\n",
		},
		{
			name:       "sum rejects special spellings before analysis",
			args:       []string{"sum", "f16", "1", "inf"},
			wantStatus: 2,
			wantStderr: "fpbits: invalid number \"inf\" (sum takes finite numbers)\n",
		},
		{
			name:       "out of range positive exponent",
			args:       []string{"encode", "f16", "1e5001"},
			wantStatus: 2,
			wantStderr: "fpbits: number \"1e5001\" is out of range (at most 400 digits, exponent from -5000 to 5000)\n",
		},
		{
			name:       "out of range negative exponent",
			args:       []string{"encode", "f16", "1e-5001"},
			wantStatus: 2,
			wantStderr: "fpbits: number \"1e-5001\" is out of range (at most 400 digits, exponent from -5000 to 5000)\n",
		},
		{
			name:       "accepted positive exponent endpoint parses before ULP finite restriction",
			args:       []string{"ulp", "f16", "1e5000"},
			wantStatus: 2,
			wantStderr: "fpbits: ulp needs a finite number\n",
		},
		{
			name:       "accepted negative exponent endpoint rounds to signed zero",
			args:       []string{"ulp", "f16", "-1e-5000"},
			wantStatus: 0,
			wantStdout: `format: f16
input: -1e-5000
rounded: -0e0 (0x8000)
ulp: 6e-8
gap toward zero: 6e-8
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, _, _ := runCLI(test.args, test.input)
			if status != test.wantStatus {
				t.Errorf("cli.Run(%q) status = %d, want %d", test.args, status, test.wantStatus)
			}
			if stdout != test.wantStdout {
				t.Errorf("cli.Run(%q) stdout = %q, want %q", test.args, stdout, test.wantStdout)
			}
			if stderr != test.wantStderr {
				t.Errorf("cli.Run(%q) stderr = %q, want %q", test.args, stderr, test.wantStderr)
			}
		})
	}
}

func TestUlpNextAndSumRestrictions(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "rounded positive infinity is not a finite ULP input",
			args:       []string{"ulp", "f16", "65520"},
			wantStatus: 2,
			wantStderr: "fpbits: ulp needs a finite number\n",
		},
		{
			name:       "e4m3 overflow is not a finite ULP input",
			args:       []string{"ulp", "e4m3", "464.01"},
			wantStatus: 2,
			wantStderr: "fpbits: ulp needs a finite number\n",
		},
		{
			name:       "cannot step from positive NaN",
			args:       []string{"next", "f16", "7c01"},
			wantStatus: 2,
			wantStderr: "fpbits: cannot step from nan\n",
		},
		{
			name:       "cannot step from negative NaN",
			args:       []string{"next", "f16", "fc01"},
			wantStatus: 2,
			wantStderr: "fpbits: cannot step from nan\n",
		},
		{
			name:       "invalid count sixty-five is rejected",
			args:       []string{"next", "--count", "65", "f16", "0000"},
			wantStatus: 2,
			wantStderr: "fpbits: invalid --count \"65\" (expected an integer from 1 to 64)\n",
		},
		{
			name:       "count one is accepted and crosses positive zero",
			args:       []string{"next", "--count=1", "f16", "0000"},
			wantStatus: 0,
			wantStdout: `offset  bits    value
-1      0x8001  -6e-8
0       0x0000  0e0
+1      0x0001  6e-8
`,
		},
		{
			name:       "positive IEEE infinity endpoint clips above",
			args:       []string{"next", "--count=1", "f16", "7c00"},
			wantStatus: 0,
			wantStdout: `offset  bits    value
-1      0x7bff  6.55e4
0       0x7c00  inf
`,
		},
		{
			name:       "negative IEEE infinity endpoint clips below",
			args:       []string{"next", "--count=1", "f16", "fc00"},
			wantStatus: 0,
			wantStdout: `offset  bits    value
0       0xfc00  -inf
+1      0xfbff  -6.55e4
`,
		},
		{
			name:       "negative e4m3 maximum clips below the finite endpoint",
			args:       []string{"next", "--count=3", "e4m3", "fe"},
			wantStatus: 0,
			wantStdout: `offset  bits  value
0       0xfe  -4.5e2
+1      0xfd  -4.2e2
+2      0xfc  -4e2
+3      0xfb  -3.5e2
`,
		},
		{
			name:       "count sixty-four is accepted and yields sixty-five neighbours at the clipped endpoint",
			args:       []string{"next", "--count=64", "e4m3", "7e"},
			wantStatus: 0,
			wantStdout: `offset  bits  value
-64     0x3e  1.8e0
-63     0x3f  1.9e0
-62     0x40  2e0
-61     0x41  2.2e0
-60     0x42  2.5e0
-59     0x43  2.8e0
-58     0x44  3e0
-57     0x45  3.2e0
-56     0x46  3.5e0
-55     0x47  3.8e0
-54     0x48  4e0
-53     0x49  4.5e0
-52     0x4a  5e0
-51     0x4b  5.5e0
-50     0x4c  6e0
-49     0x4d  6.5e0
-48     0x4e  7e0
-47     0x4f  7.5e0
-46     0x50  8e0
-45     0x51  9e0
-44     0x52  1e1
-43     0x53  1.1e1
-42     0x54  1.2e1
-41     0x55  1.3e1
-40     0x56  1.4e1
-39     0x57  1.5e1
-38     0x58  1.6e1
-37     0x59  1.8e1
-36     0x5a  2e1
-35     0x5b  2.2e1
-34     0x5c  2.4e1
-33     0x5d  2.6e1
-32     0x5e  2.8e1
-31     0x5f  3e1
-30     0x60  3.2e1
-29     0x61  3.6e1
-28     0x62  4e1
-27     0x63  4.4e1
-26     0x64  5e1
-25     0x65  5.2e1
-24     0x66  5.6e1
-23     0x67  6e1
-22     0x68  6.4e1
-21     0x69  7e1
-20     0x6a  8e1
-19     0x6b  9e1
-18     0x6c  1e2
-17     0x6d  1.04e2
-16     0x6e  1.1e2
-15     0x6f  1.2e2
-14     0x70  1.3e2
-13     0x71  1.4e2
-12     0x72  1.6e2
-11     0x73  1.8e2
-10     0x74  2e2
-9      0x75  2.1e2
-8      0x76  2.2e2
-7      0x77  2.4e2
-6      0x78  2.6e2
-5      0x79  3e2
-4      0x7a  3.2e2
-3      0x7b  3.5e2
-2      0x7c  4e2
-1      0x7d  4.2e2
0       0x7e  4.5e2
`,
		},
		{
			name:       "sum ending in infinity omits an error line",
			args:       []string{"sum", "f16", "1e10", "1"},
			wantStatus: 0,
			wantStdout: `step  term  sum
1     inf   inf
2     1e0   inf
exact: 1.0000000001e10
result: inf (0x7c00)
`,
		},
		{
			name:       "sum ending in NaN omits an error line",
			args:       []string{"sum", "e4m3", "464.01", "1"},
			wantStatus: 0,
			wantStdout: `step  term  sum
1     nan   nan
2     1e0   nan
exact: 4.6501e2
result: nan (0x7f)
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, _, _ := runCLI(test.args, nil)
			if status != test.wantStatus || stderr != test.wantStderr {
				t.Errorf("cli.Run(%q) = status %d, stderr %q; want status %d, stderr %q", test.args, status, stderr, test.wantStatus, test.wantStderr)
			}
			if stdout != test.wantStdout {
				t.Errorf("cli.Run(%q) stdout = %q, want %q", test.args, stdout, test.wantStdout)
			}
		})
	}
}

func TestSumRejectsEverySignedSpecialNumber(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{text: "inf", want: "fpbits: invalid number \"inf\" (sum takes finite numbers)\n"},
		{text: "-inf", want: "fpbits: invalid number \"-inf\" (sum takes finite numbers)\n"},
		{text: "nan", want: "fpbits: invalid number \"nan\" (sum takes finite numbers)\n"},
		{text: "-nan", want: "fpbits: invalid number \"-nan\" (sum takes finite numbers)\n"},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			status, stdout, stderr, _, _ := runCLI([]string{"sum", "f16", test.text}, nil)
			if status != 2 || stdout != "" || stderr != test.want {
				t.Errorf("sum %s returned (%d, %q, %q), want status 2 and %q", test.text, status, stdout, stderr, test.want)
			}
		})
	}
}

func TestAtomicFailureWritesAndUsage(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantStdout string
		wantStderr string
		outWrites  int
		errWrites  int
	}{
		{
			name:       "sole help",
			args:       []string{"--help"},
			wantStatus: 0,
			wantStdout: `usage: fpbits decode FORMAT HEX...
       fpbits encode FORMAT NUMBER...
       fpbits convert FROM TO HEX...
       fpbits ulp FORMAT NUMBER
       fpbits next [--count N] FORMAT HEX
       fpbits range FORMAT
       fpbits sum FORMAT NUMBER...
`,
			outWrites: 1,
		},
		{
			name:       "missing command",
			args:       nil,
			wantStatus: 2,
			wantStderr: "fpbits: missing command (decode, encode, convert, ulp, next, range or sum)\n",
			errWrites:  1,
		},
		{
			name:       "unknown command",
			args:       []string{"x"},
			wantStatus: 2,
			wantStderr: "fpbits: unknown command \"x\"\n",
			errWrites:  1,
		},
		{
			name:       "help is not a command when additional arguments follow",
			args:       []string{"--help", "range"},
			wantStatus: 2,
			wantStderr: "fpbits: unknown command \"--help\"\n",
			errWrites:  1,
		},
		{
			name:       "valid first word is not written before a later bad word",
			args:       []string{"decode", "f16", "3c00", "bad"},
			wantStatus: 2,
			wantStderr: "fpbits: invalid hex word \"bad\" for f16 (expected 4 hex digits)\n",
			errWrites:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, stdoutWrites, stderrWrites := runCLI(test.args, nil)
			if status != test.wantStatus || stdout != test.wantStdout || stderr != test.wantStderr {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), want (%d, %q, %q)", test.args, status, stdout, stderr, test.wantStatus, test.wantStdout, test.wantStderr)
			}
			if stdoutWrites != test.outWrites || stderrWrites != test.errWrites {
				t.Errorf("cli.Run(%q) writes = stdout %d, stderr %d; want stdout %d, stderr %d", test.args, stdoutWrites, stderrWrites, test.outWrites, test.errWrites)
			}
		})
	}
}

func TestOneHundredDataWordsAcceptedAndOneHundredOneRejected(t *testing.T) {
	args := []string{
		"decode", "f16",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
		"0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000",
	}

	status, stdout, stderr, stdoutWrites, stderrWrites := runCLI(args, nil)
	want := `format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0

format: f16
bits: 0x0000
binary: 0 00000 0000000000
class: zero
value: 0e0
exact: 0e0
`
	if status != 0 || stderr != "" || stdout != want {
		t.Errorf("decode of exactly 100 words returned (%d, %q, %q), want the complete literal report", status, stdout, stderr)
	}
	if stdoutWrites != 1 || stderrWrites != 0 {
		t.Errorf("decode of exactly 100 words wrote stdout %d times and stderr %d times, want 1 and 0", stdoutWrites, stderrWrites)
	}

	args = append(args, "not-a-hex-word")
	status, stdout, stderr, _, _ = runCLI(args, nil)
	if status != 2 || stdout != "" || stderr != "fpbits: at most 100 values\n" {
		t.Errorf("decode of 101 words returned (%d, %q, %q), want (2, empty, at most 100 values)", status, stdout, stderr)
	}
}

func TestFiftySumTermsAcceptedAndFiftyOneRejected(t *testing.T) {
	args := []string{
		"sum", "f16",
		"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
		"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
		"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
		"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
		"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
	}
	status, stdout, stderr, stdoutWrites, stderrWrites := runCLI(args, nil)
	want := `step  term  sum
1     0e0   0e0
2     0e0   0e0
3     0e0   0e0
4     0e0   0e0
5     0e0   0e0
6     0e0   0e0
7     0e0   0e0
8     0e0   0e0
9     0e0   0e0
10    0e0   0e0
11    0e0   0e0
12    0e0   0e0
13    0e0   0e0
14    0e0   0e0
15    0e0   0e0
16    0e0   0e0
17    0e0   0e0
18    0e0   0e0
19    0e0   0e0
20    0e0   0e0
21    0e0   0e0
22    0e0   0e0
23    0e0   0e0
24    0e0   0e0
25    0e0   0e0
26    0e0   0e0
27    0e0   0e0
28    0e0   0e0
29    0e0   0e0
30    0e0   0e0
31    0e0   0e0
32    0e0   0e0
33    0e0   0e0
34    0e0   0e0
35    0e0   0e0
36    0e0   0e0
37    0e0   0e0
38    0e0   0e0
39    0e0   0e0
40    0e0   0e0
41    0e0   0e0
42    0e0   0e0
43    0e0   0e0
44    0e0   0e0
45    0e0   0e0
46    0e0   0e0
47    0e0   0e0
48    0e0   0e0
49    0e0   0e0
50    0e0   0e0
exact: 0e0
result: 0e0 (0x0000)
error: 0e0
`
	if status != 0 || stderr != "" || stdout != want {
		t.Errorf("sum of exactly 50 terms returned (%d, %q, %q), want the complete literal report", status, stdout, stderr)
	}
	if stdoutWrites != 1 || stderrWrites != 0 {
		t.Errorf("sum of exactly 50 terms wrote stdout %d times and stderr %d times, want 1 and 0", stdoutWrites, stderrWrites)
	}

	args = append(args, "not-a-number")
	status, stdout, stderr, _, _ = runCLI(args, nil)
	if status != 2 || stdout != "" || stderr != "fpbits: at most 50 terms\n" {
		t.Errorf("sum of 51 terms returned (%d, %q, %q), want (2, empty, at most 50 terms)", status, stdout, stderr)
	}
}

const oneHundredZeroDigits = "0000000000" +
	"0000000000" +
	"0000000000" +
	"0000000000" +
	"0000000000" +
	"0000000000" +
	"0000000000" +
	"0000000000" +
	"0000000000" +
	"0000000000"

const fourHundredZeroDigits = oneHundredZeroDigits + oneHundredZeroDigits +
	oneHundredZeroDigits + oneHundredZeroDigits
const fourHundredOneZeroDigits = fourHundredZeroDigits + "0"

func TestNumberDigitLimitIsFourHundred(t *testing.T) {
	if len(fourHundredZeroDigits) != 400 || len(fourHundredOneZeroDigits) != 401 {
		t.Fatalf("digit fixtures have lengths %d and %d, want 400 and 401", len(fourHundredZeroDigits), len(fourHundredOneZeroDigits))
	}

	status, stdout, stderr, _, _ := runCLI([]string{"sum", "f16", fourHundredZeroDigits}, nil)
	want := `step  term  sum
1     0e0   0e0
exact: 0e0
result: 0e0 (0x0000)
error: 0e0
`
	if status != 0 || stderr != "" || stdout != want {
		t.Errorf("sum with 400 significand digits returned (%d, %q, %q), want exact zero report", status, stdout, stderr)
	}

	status, stdout, stderr, _, _ = runCLI([]string{"sum", "f16", fourHundredOneZeroDigits}, nil)
	want = "fpbits: number \"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\" is out of range (at most 400 digits, exponent from -5000 to 5000)\n"
	if status != 2 || stdout != "" || stderr != want {
		t.Errorf("sum with 401 significand digits returned (%d, %q, %q), want the complete literal diagnostic", status, stdout, stderr)
	}
}

func TestNextReadsSingleHexWordFromStdin(t *testing.T) {
	wantOutput := `offset  bits    value
-1      0x3bff  9.995e-1
0       0x3c00  1e0
+1      0x3c01  1.001e0
`
	tests := []struct {
		name       string
		args       []string
		input      []byte
		wantStatus int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "one word",
			args:       []string{"next", "--count", "1", "f16", "-"},
			input:      []byte("3c00\n"),
			wantStatus: 0,
			wantStdout: wantOutput,
		},
		{
			name:       "byte-order mark, prefix, and CR LF",
			args:       []string{"next", "--count", "1", "f16", "-"},
			input:      []byte{0xef, 0xbb, 0xbf, '0', 'x', '3', 'c', '0', '0', '\r', '\n'},
			wantStatus: 0,
			wantStdout: wantOutput,
		},
		{
			name:       "two words",
			args:       []string{"next", "f16", "-"},
			input:      []byte("3c00 3c01\n"),
			wantStatus: 2,
			wantStderr: "fpbits: <stdin>: expected 1 word, got 2\n",
		},
		{
			name:       "no words",
			args:       []string{"next", "f16", "-"},
			input:      []byte(" \t\r\n"),
			wantStatus: 2,
			wantStderr: "fpbits: <stdin>: no input\n",
		},
		{
			name:       "invalid UTF-8",
			args:       []string{"next", "f16", "-"},
			input:      []byte{0x33, 0xff},
			wantStatus: 2,
			wantStderr: "fpbits: <stdin>: input is not valid UTF-8\n",
		},
		{
			name:       "bad hex word",
			args:       []string{"next", "f16", "-"},
			input:      []byte("3c0\n"),
			wantStatus: 2,
			wantStderr: "fpbits: invalid hex word \"3c0\" for f16 (expected 4 hex digits)\n",
		},
		{
			name:       "unknown format is reported before stdin is read",
			args:       []string{"next", "f17", "-"},
			input:      []byte{0xff},
			wantStatus: 2,
			wantStderr: "fpbits: unknown format \"f17\" (expected f16, bf16, f32, f64, e5m2 or e4m3)\n",
		},
		{
			name:       "too many arguments are reported before stdin is read",
			args:       []string{"next", "f16", "-", "3c00"},
			input:      []byte{0xff},
			wantStatus: 2,
			wantStderr: "fpbits: too many arguments\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, _, _ := runCLI(test.args, test.input)
			if status != test.wantStatus || stdout != test.wantStdout || stderr != test.wantStderr {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), want (%d, %q, %q)", test.args, status, stdout, stderr, test.wantStatus, test.wantStdout, test.wantStderr)
			}
		})
	}
}
