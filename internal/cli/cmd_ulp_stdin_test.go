package cli_test

import "testing"

func TestUlpReadsSingleNumberFromStdin(t *testing.T) {
	wantF16 := `format: f16
input: 0.1
rounded: 1e-1 (0x2e66)
ulp: 6.104e-5
gap toward zero: 6.104e-5
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
			args:       []string{"ulp", "f16", "-"},
			input:      []byte("0.1\n"),
			wantStatus: 0,
			wantStdout: wantF16,
		},
		{
			name:       "byte-order mark and CR LF",
			args:       []string{"ulp", "f16", "-"},
			input:      []byte{0xef, 0xbb, 0xbf, '0', '.', '1', '\r', '\n'},
			wantStatus: 0,
			wantStdout: wantF16,
		},
		{
			name:       "two words",
			args:       []string{"ulp", "f16", "-"},
			input:      []byte("1 2\n"),
			wantStatus: 2,
			wantStderr: "fpbits: <stdin>: expected 1 word, got 2\n",
		},
		{
			name:       "no words",
			args:       []string{"ulp", "f16", "-"},
			input:      []byte(" \t\r\n"),
			wantStatus: 2,
			wantStderr: "fpbits: <stdin>: no input\n",
		},
		{
			name:       "invalid UTF-8",
			args:       []string{"ulp", "f16", "-"},
			input:      []byte{0x31, 0xff},
			wantStatus: 2,
			wantStderr: "fpbits: <stdin>: input is not valid UTF-8\n",
		},
		{
			name:       "bad number word",
			args:       []string{"ulp", "f16", "-"},
			input:      []byte("x\n"),
			wantStatus: 2,
			wantStderr: "fpbits: invalid number \"x\"\n",
		},
		{
			name:       "overflowing word",
			args:       []string{"ulp", "f16", "-"},
			input:      []byte("1e5000\n"),
			wantStatus: 2,
			wantStderr: "fpbits: ulp needs a finite number\n",
		},
		{
			name:       "unknown format is reported before stdin is read",
			args:       []string{"ulp", "f17", "-"},
			input:      []byte{0xff},
			wantStatus: 2,
			wantStderr: "fpbits: unknown format \"f17\" (expected f16, bf16, f32, f64, e5m2 or e4m3)\n",
		},
		{
			name:       "too many arguments are reported before stdin is read",
			args:       []string{"ulp", "f16", "-", "1"},
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
