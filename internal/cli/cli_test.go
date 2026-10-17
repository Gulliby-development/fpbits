package cli_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/cli"
)

type validationCounter struct {
	bytes.Buffer
	calls int
}

func (w *validationCounter) Write(p []byte) (int, error) {
	w.calls++
	return w.Buffer.Write(p)
}

type validationReader struct {
	data  []byte
	reads int
}

func (r *validationReader) Read(p []byte) (int, error) {
	r.reads++
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func validationRun(args []string, input []byte) (int, string, string, int, int, int) {
	stdout := &validationCounter{}
	stderr := &validationCounter{}
	stdin := &validationReader{data: input}
	status := cli.Run(args, stdin, stdout, stderr)
	return status, stdout.String(), stderr.String(), stdout.calls, stderr.calls, stdin.reads
}

func TestValidationDispatchAndPositionErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing command", nil, "fpbits: missing command (decode, encode, convert, ulp, next, range or sum)\n"},
		{"sanitized unknown command", []string{"bad\nλ"}, "fpbits: unknown command \"bad??\"\n"},
		{"misplaced help", []string{"--help", "range"}, "fpbits: unknown command \"--help\"\n"},
		{"decode missing format", []string{"decode"}, "fpbits: missing FORMAT\n"},
		{"decode missing hex before format lookup", []string{"decode", "bad-format"}, "fpbits: missing HEX\n"},
		{"encode missing format", []string{"encode"}, "fpbits: missing FORMAT\n"},
		{"encode missing number before format lookup", []string{"encode", "bad-format"}, "fpbits: missing NUMBER\n"},
		{"convert missing from", []string{"convert"}, "fpbits: missing FROM\n"},
		{"convert missing to", []string{"convert", "bad-from"}, "fpbits: missing TO\n"},
		{"convert missing hex before format lookup", []string{"convert", "bad-from", "bad-to"}, "fpbits: missing HEX\n"},
		{"ulp missing format", []string{"ulp"}, "fpbits: missing FORMAT\n"},
		{"ulp missing number before format lookup", []string{"ulp", "bad-format"}, "fpbits: missing NUMBER\n"},
		{"next missing format", []string{"next"}, "fpbits: missing FORMAT\n"},
		{"next missing hex before format lookup", []string{"next", "bad-format"}, "fpbits: missing HEX\n"},
		{"range missing format", []string{"range"}, "fpbits: missing FORMAT\n"},
		{"sum missing format", []string{"sum"}, "fpbits: missing FORMAT\n"},
		{"sum missing number before format lookup", []string{"sum", "bad-format"}, "fpbits: missing NUMBER\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, stdoutWrites, stderrWrites, _ := validationRun(test.args, nil)
			if status != 2 || stdout != "" || stderr != test.want {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), want (2, empty, %q)", test.args, status, stdout, stderr, test.want)
			}
			if stdoutWrites != 0 || stderrWrites != 1 {
				t.Errorf("writer calls = stdout %d, stderr %d; want 0 and 1", stdoutWrites, stderrWrites)
			}
		})
	}
}

func TestOptionErrorsPrecedeCommandArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"next unknown option before missing format", []string{"next", "--other"}, "fpbits: unknown option \"--other\"\n"},
		{"next invalid count before missing format", []string{"next", "--count=0"}, "fpbits: invalid --count \"0\" (expected an integer from 1 to 64)\n"},
		{"next missing option value before positions", []string{"next", "--count"}, "fpbits: --count needs a value\n"},
		{"unsupported option on decode", []string{"decode", "--count=2", "f16", "0000"}, "fpbits: unknown option \"--count=2\"\n"},
		{"unsupported option on encode", []string{"encode", "--count", "2", "f16", "1"}, "fpbits: unknown option \"--count\"\n"},
		{"unsupported option on convert", []string{"convert", "f16", "f32", "--count=2", "0000"}, "fpbits: unknown option \"--count=2\"\n"},
		{"unsupported option on ulp", []string{"ulp", "f16", "--count=2", "1"}, "fpbits: unknown option \"--count=2\"\n"},
		{"unsupported option on range", []string{"range", "--count=2", "f16"}, "fpbits: unknown option \"--count=2\"\n"},
		{"unsupported option on sum", []string{"sum", "f16", "--count=2", "1"}, "fpbits: unknown option \"--count=2\"\n"},
		{"bare double dash is unknown", []string{"next", "--"}, "fpbits: unknown option \"--\"\n"},
		{"help is an unknown option for next", []string{"next", "--help"}, "fpbits: unknown option \"--help\"\n"},
		{"unknown option beats deferred invalid count", []string{"next", "--count=0", "--bogus"}, "fpbits: unknown option \"--bogus\"\n"},
		{"option-looking value is consumed then invalid", []string{"next", "--count", "--help", "f16", "0000"}, "fpbits: invalid --count \"--help\" (expected an integer from 1 to 64)\n"},
		{"count zero is rejected", []string{"next", "--count=0", "f16", "0000"}, "fpbits: invalid --count \"0\" (expected an integer from 1 to 64)\n"},
		{"count 65 is rejected", []string{"next", "--count=65", "f16", "0000"}, "fpbits: invalid --count \"65\" (expected an integer from 1 to 64)\n"},
		{"count leading zero is rejected", []string{"next", "--count=01", "f16", "0000"}, "fpbits: invalid --count \"01\" (expected an integer from 1 to 64)\n"},
		{"count negative is rejected", []string{"next", "--count=-1", "f16", "0000"}, "fpbits: invalid --count \"-1\" (expected an integer from 1 to 64)\n"},
		{"count text is rejected", []string{"next", "--count=word", "f16", "0000"}, "fpbits: invalid --count \"word\" (expected an integer from 1 to 64)\n"},
		{"count VALUE is rejected", []string{"next", "--count=VALUE", "f16", "0000"}, "fpbits: invalid --count \"VALUE\" (expected an integer from 1 to 64)\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, _, _, _ := validationRun(test.args, nil)
			if status != 2 || stdout != "" || stderr != test.want {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), want (2, empty, %q)", test.args, status, stdout, stderr, test.want)
			}
		})
	}
}

func TestRepeatedCountAndOptionBoundaries(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "repeated option overwrites an invalid earlier value",
			args: []string{"next", "--count=0", "--count=1", "f16", "0000"},
			want: "offset  bits    value\n-1      0x8001  -6e-8\n0       0x0000  0e0\n+1      0x0001  6e-8\n",
		},
		{
			name: "equals option syntax accepts maximum",
			args: []string{"next", "--count=64", "e4m3", "7e"},
			want: "offset  bits  value\n-64     0x3e  1.8e0\n-63     0x3f  1.9e0\n-62     0x40  2e0\n-61     0x41  2.2e0\n-60     0x42  2.5e0\n-59     0x43  2.8e0\n-58     0x44  3e0\n-57     0x45  3.2e0\n-56     0x46  3.5e0\n-55     0x47  3.8e0\n-54     0x48  4e0\n-53     0x49  4.5e0\n-52     0x4a  5e0\n-51     0x4b  5.5e0\n-50     0x4c  6e0\n-49     0x4d  6.5e0\n-48     0x4e  7e0\n-47     0x4f  7.5e0\n-46     0x50  8e0\n-45     0x51  9e0\n-44     0x52  1e1\n-43     0x53  1.1e1\n-42     0x54  1.2e1\n-41     0x55  1.3e1\n-40     0x56  1.4e1\n-39     0x57  1.5e1\n-38     0x58  1.6e1\n-37     0x59  1.8e1\n-36     0x5a  2e1\n-35     0x5b  2.2e1\n-34     0x5c  2.4e1\n-33     0x5d  2.6e1\n-32     0x5e  2.8e1\n-31     0x5f  3e1\n-30     0x60  3.2e1\n-29     0x61  3.6e1\n-28     0x62  4e1\n-27     0x63  4.4e1\n-26     0x64  5e1\n-25     0x65  5.2e1\n-24     0x66  5.6e1\n-23     0x67  6e1\n-22     0x68  6.4e1\n-21     0x69  7e1\n-20     0x6a  8e1\n-19     0x6b  9e1\n-18     0x6c  1e2\n-17     0x6d  1.04e2\n-16     0x6e  1.1e2\n-15     0x6f  1.2e2\n-14     0x70  1.3e2\n-13     0x71  1.4e2\n-12     0x72  1.6e2\n-11     0x73  1.8e2\n-10     0x74  2e2\n-9      0x75  2.1e2\n-8      0x76  2.2e2\n-7      0x77  2.4e2\n-6      0x78  2.6e2\n-5      0x79  3e2\n-4      0x7a  3.2e2\n-3      0x7b  3.5e2\n-2      0x7c  4e2\n-1      0x7d  4.2e2\n0       0x7e  4.5e2\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, _, _, _ := validationRun(test.args, nil)
			if status != 0 || stderr != "" || stdout != test.want {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), want success and exact output %q", test.args, status, stdout, stderr, test.want)
			}
		})
	}
}

func TestExcessArgumentsPrecedeFormatAndWords(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"ulp excess wins over unknown format", []string{"ulp", "f17", "1", "extra"}, "fpbits: too many arguments\n"},
		{"next excess wins over unknown format", []string{"next", "f17", "0000", "extra"}, "fpbits: too many arguments\n"},
		{"range excess wins over unknown format", []string{"range", "f17", "extra"}, "fpbits: too many arguments\n"},
		{"convert looks up FROM before TO", []string{"convert", "f17", "bad-to", "bad-hex"}, "fpbits: unknown format \"f17\" (expected f16, bf16, f32, f64, e5m2 or e4m3)\n"},
		{"convert looks up TO before reading words", []string{"convert", "f16", "f17", "-"}, "fpbits: unknown format \"f17\" (expected f16, bf16, f32, f64, e5m2 or e4m3)\n"},
		{"unknown format precedes malformed stdin", []string{"decode", "f17", "-"}, "fpbits: unknown format \"f17\" (expected f16, bf16, f32, f64, e5m2 or e4m3)\n"},
		{"stdin decoding precedes hex parsing", []string{"decode", "f16", "-"}, "fpbits: <stdin>: input is not valid UTF-8\n"},
		{"leftmost invalid word wins", []string{"decode", "f16", "bad-first", "bad-second"}, "fpbits: invalid hex word \"bad-first\" for f16 (expected 4 hex digits)\n"},
		{"late invalid word yields no partial output", []string{"decode", "f16", "0000", "bad-late"}, "fpbits: invalid hex word \"bad-late\" for f16 (expected 4 hex digits)\n"},
		{"next rejects NaN only after valid format and hex", []string{"next", "f16", "fc01"}, "fpbits: cannot step from nan\n"},
		{"ulp rejects rounded nonfinite only after valid number", []string{"ulp", "f16", "1e5000"}, "fpbits: ulp needs a finite number\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, stdoutWrites, stderrWrites, _ := validationRun(test.args, []byte{0xff})
			if status != 2 || stdout != "" || stderr != test.want {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), want (2, empty, %q)", test.args, status, stdout, stderr, test.want)
			}
			if stdoutWrites != 0 || stderrWrites != 1 {
				t.Errorf("writer calls = stdout %d, stderr %d; want 0 and 1", stdoutWrites, stderrWrites)
			}
		})
	}
}

func TestStdinWordsAreExpandedOnlyForSoleDash(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		input []byte
		want  string
	}{
		{
			name:  "tabs split words",
			args:  []string{"decode", "f16", "-"},
			input: []byte("3c00\t0000"),
			want:  "format: f16\nbits: 0x3c00\nbinary: 0 01111 0000000000\nclass: normal\nvalue: 1e0\nexact: 1e0\n\nformat: f16\nbits: 0x0000\nbinary: 0 00000 0000000000\nclass: zero\nvalue: 0e0\nexact: 0e0\n",
		},
		{
			name:  "one BOM is removed but repeated BOM is content",
			args:  []string{"decode", "f16", "-"},
			input: []byte{0xef, 0xbb, 0xbf, 0xef, 0xbb, 0xbf, '3', 'c', '0', '0'},
			want:  "fpbits: invalid hex word \"?3c00\" for f16 (expected 4 hex digits)\n",
		},
		{
			name:  "nonseparator Unicode remains inside one word",
			args:  []string{"decode", "f16", "-"},
			input: []byte("3c00\u00a00000"),
			want:  "fpbits: invalid hex word \"3c00?0000\" for f16 (expected 4 hex digits)\n",
		},
		{
			name:  "whitespace-only input has no words",
			args:  []string{"decode", "f16", "-"},
			input: []byte(" \t\r\n"),
			want:  "fpbits: <stdin>: no input\n",
		},
		{
			name:  "malformed UTF-8 is rejected by core through cli.Run",
			args:  []string{"decode", "f16", "-"},
			input: []byte{0x33, 0x63, 0x30, 0x30, 0x20, 0xff, 0x0a},
			want:  "fpbits: <stdin>: input is not valid UTF-8\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, stdoutWrites, stderrWrites, _ := validationRun(test.args, test.input)
			if status == 0 {
				if stdout != test.want || stderr != "" || stdoutWrites != 1 || stderrWrites != 0 {
					t.Errorf("cli.Run(%q) = (%d, %q, %q), writes %d/%d; want exact success output and one stdout write", test.args, status, stdout, stderr, stdoutWrites, stderrWrites)
				}
			} else if status != 2 || stdout != "" || stderr != test.want || stdoutWrites != 0 || stderrWrites != 1 {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), writes %d/%d; want exact failure and no partial stdout", test.args, status, stdout, stderr, stdoutWrites, stderrWrites)
			}
		})
	}
}

func TestNonExpandingArgumentsNeverReadStdin(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"dash mixed with decode data is ordinary", []string{"decode", "f16", "-", "0000"}, "fpbits: invalid hex word \"-\" for f16 (expected 4 hex digits)\n"},
		{"dash mixed with encode data is ordinary", []string{"encode", "f16", "-", "1"}, "fpbits: invalid number \"-\"\n"},
		{"dash mixed with convert data is ordinary", []string{"convert", "f16", "f32", "-", "0000"}, "fpbits: invalid hex word \"-\" for f16 (expected 4 hex digits)\n"},
		{"dash mixed with sum data is ordinary", []string{"sum", "f16", "-", "1"}, "fpbits: invalid number \"-\"\n"},
		{"ulp does not expand stdin", []string{"ulp", "f16", "1"}, ""},
		{"next does not expand stdin", []string{"next", "f16", "0000"}, ""},
		{"range does not expand stdin", []string{"range", "f16"}, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdin := &validationReader{data: []byte{0xff}}
			stdout := &validationCounter{}
			stderr := &validationCounter{}
			status := cli.Run(test.args, stdin, stdout, stderr)
			if stdin.reads != 0 {
				t.Fatalf("stdin was read %d times for %q", stdin.reads, test.args)
			}
			if test.want != "" {
				if status != 2 || stdout.String() != "" || stderr.String() != test.want {
					t.Errorf("cli.Run(%q) = (%d, %q, %q), want exact failure %q", test.args, status, stdout.String(), stderr.String(), test.want)
				}
				return
			}
			if status != 0 || stdout.Len() == 0 || stderr.Len() != 0 {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), want success", test.args, status, stdout.String(), stderr.String())
			}
		})
	}
}

func TestSuccessfulDispatchWritesStdoutOnce(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"decode", []string{"decode", "f16", "0000"}},
		{"encode", []string{"encode", "f16", "0"}},
		{"convert", []string{"convert", "f16", "f32", "0000"}},
		{"ulp", []string{"ulp", "f16", "1"}},
		{"next", []string{"next", "f16", "0000"}},
		{"range", []string{"range", "f16"}},
		{"sum", []string{"sum", "f16", "0"}},
		{"help", []string{"--help"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr, stdoutWrites, stderrWrites, _ := validationRun(test.args, nil)
			if status != 0 || stdout == "" || !strings.HasSuffix(stdout, "\n") || stderr != "" {
				t.Errorf("cli.Run(%q) = (%d, %q, %q), want successful LF-terminated output", test.args, status, stdout, stderr)
			}
			if stdoutWrites != 1 || stderrWrites != 0 {
				t.Errorf("writer calls = stdout %d, stderr %d; want 1 and 0", stdoutWrites, stderrWrites)
			}
		})
	}
}

func TestEveryValidationFailureIsAtomic(t *testing.T) {
	status, stdout, stderr, stdoutWrites, stderrWrites, _ := validationRun([]string{"encode", "f16", "1", "bad-number"}, nil)
	want := "fpbits: invalid number \"bad-number\"\n"
	if status != 2 || stdout != "" || stderr != want {
		t.Errorf("late encode failure = (%d, %q, %q), want (2, empty, %q)", status, stdout, stderr, want)
	}
	if stdoutWrites != 0 || stderrWrites != 1 {
		t.Errorf("writer calls = stdout %d, stderr %d; want 0 and 1", stdoutWrites, stderrWrites)
	}
}

func TestValidationDiagnosticLineEndingsAreLF(t *testing.T) {
	status, stdout, stderr, _, _, _ := validationRun([]string{"decode", "f16", "bad"}, nil)
	if status != 2 || stdout != "" || stderr != "fpbits: invalid hex word \"bad\" for f16 (expected 4 hex digits)\n" {
		t.Fatalf("failure = (%d, %q, %q), want exact LF diagnostic", status, stdout, stderr)
	}
	if strings.Contains(stderr, "\r") || !strings.HasSuffix(stderr, "\n") {
		t.Fatalf("stderr is not LF terminated: %q", stderr)
	}
}

func TestNoPanicOnMalformedArgumentBytes(t *testing.T) {
	args := []string{"unknown\xff"}
	status, stdout, stderr, _, _, _ := validationRun(args, nil)
	if status != 2 || stdout != "" || stderr != "fpbits: unknown command \"unknown?\"\n" {
		t.Fatalf("malformed argument result = (%d, %q, %q)", status, stdout, stderr)
	}
}
