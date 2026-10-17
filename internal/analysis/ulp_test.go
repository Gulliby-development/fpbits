package analysis_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/analysis"
	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/num"
)

func TestUlpCompleteLiteralOutputs(t *testing.T) {
	tests := []struct {
		name   string
		format string
		text   string
		want   []string
	}{
		{
			name:   "positive zero uses minimum spacing",
			format: "f16",
			text:   "0",
			want: []string{
				"format: f16",
				"input: 0",
				"rounded: 0e0 (0x0000)",
				"ulp: 6e-8",
				"gap toward zero: 6e-8",
			},
		},
		{
			name:   "negative zero preserves its sign",
			format: "f16",
			text:   "-0",
			want: []string{
				"format: f16",
				"input: -0",
				"rounded: -0e0 (0x8000)",
				"ulp: 6e-8",
				"gap toward zero: 6e-8",
			},
		},
		{
			name:   "smallest subnormal",
			format: "f16",
			text:   "6e-8",
			want: []string{
				"format: f16",
				"input: 6e-8",
				"rounded: 6e-8 (0x0001)",
				"ulp: 6e-8",
				"gap toward zero: 6e-8",
			},
		},
		{
			name:   "smallest normal",
			format: "f16",
			text:   "6.104e-5",
			want: []string{
				"format: f16",
				"input: 6.104e-5",
				"rounded: 6.104e-5 (0x0400)",
				"ulp: 6e-8",
				"gap toward zero: 6e-8",
			},
		},
		{
			name:   "one has half-sized predecessor gap",
			format: "f16",
			text:   "1",
			want: []string{
				"format: f16",
				"input: 1",
				"rounded: 1e0 (0x3c00)",
				"ulp: 9.77e-4",
				"gap toward zero: 4.883e-4",
			},
		},
		{
			name:   "power of two exponent transition",
			format: "f16",
			text:   "2",
			want: []string{
				"format: f16",
				"input: 2",
				"rounded: 2e0 (0x4000)",
				"ulp: 1.953e-3",
				"gap toward zero: 9.77e-4",
			},
		},
		{
			name:   "negative value has positive spacing",
			format: "f16",
			text:   "-1",
			want: []string{
				"format: f16",
				"input: -1",
				"rounded: -1e0 (0xbc00)",
				"ulp: 9.77e-4",
				"gap toward zero: 4.883e-4",
			},
		},
		{
			name:   "largest finite retains formula spacing",
			format: "f16",
			text:   "65504",
			want: []string{
				"format: f16",
				"input: 65504",
				"rounded: 6.55e4 (0x7bff)",
				"ulp: 3.2e1",
				"gap toward zero: 3.2e1",
			},
		},
		{
			name:   "preserves the caller spelling",
			format: "f16",
			text:   "1.000",
			want: []string{
				"format: f16",
				"input: 1.000",
				"rounded: 1e0 (0x3c00)",
				"ulp: 9.77e-4",
				"gap toward zero: 4.883e-4",
			},
		},
		{
			name:   "bf16 spacing at one",
			format: "bf16",
			text:   "1",
			want: []string{
				"format: bf16",
				"input: 1",
				"rounded: 1e0 (0x3f80)",
				"ulp: 7.8e-3",
				"gap toward zero: 3.9e-3",
			},
		},
		{
			name:   "f32 spacing at one",
			format: "f32",
			text:   "1",
			want: []string{
				"format: f32",
				"input: 1",
				"rounded: 1e0 (0x3f800000)",
				"ulp: 1.1920929e-7",
				"gap toward zero: 5.9604645e-8",
			},
		},
		{
			name:   "f64 spacing at one",
			format: "f64",
			text:   "1",
			want: []string{
				"format: f64",
				"input: 1",
				"rounded: 1e0 (0x3ff0000000000000)",
				"ulp: 2.220446049250313e-16",
				"gap toward zero: 1.1102230246251565e-16",
			},
		},
		{
			name:   "e5m2 spacing at one",
			format: "e5m2",
			text:   "1",
			want: []string{
				"format: e5m2",
				"input: 1",
				"rounded: 1e0 (0x3c)",
				"ulp: 2.5e-1",
				"gap toward zero: 1.2e-1",
			},
		},
		{
			name:   "e4m3 finite maximum midpoint",
			format: "e4m3",
			text:   "464",
			want: []string{
				"format: e4m3",
				"input: 464",
				"rounded: 4.5e2 (0x7e)",
				"ulp: 3.2e1",
				"gap toward zero: 3.2e1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format, err := core.Lookup(test.format)
			if err != nil {
				t.Fatalf("Lookup(%q) returned error: %v", test.format, err)
			}
			number, err := num.Parse(test.text, false)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", test.text, err)
			}
			before := number.Rat.RatString()

			got, err := analysis.Ulp(format, test.text, number)
			if err != nil {
				t.Fatalf("Ulp(%q, %q) returned error: %v", test.format, test.text, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Ulp(%q, %q) = %#v, want %#v", test.format, test.text, got, test.want)
			}
			for _, line := range got {
				if strings.Contains(line, "\n") {
					t.Errorf("Ulp returned a line containing LF: %q", line)
				}
			}
			if after := number.Rat.RatString(); after != before {
				t.Errorf("Ulp mutated caller rational: got %s, want %s", after, before)
			}
		})
	}
}

func TestUlpRejectsNonfiniteRoundedResults(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		text    string
		message string
	}{
		{name: "infinity", format: "f16", text: "inf", message: "ulp needs a finite number"},
		{name: "NaN", format: "f16", text: "nan", message: "ulp needs a finite number"},
		{name: "finite overflow", format: "f16", text: "65520", message: "ulp needs a finite number"},
		{name: "e4m3 overflow", format: "e4m3", text: "464.01", message: "ulp needs a finite number"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format, err := core.Lookup(test.format)
			if err != nil {
				t.Fatalf("Lookup(%q) returned error: %v", test.format, err)
			}
			number, err := num.Parse(test.text, false)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", test.text, err)
			}

			lines, err := analysis.Ulp(format, test.text, number)
			if err == nil {
				t.Fatalf("Ulp(%q, %q) returned lines %#v, want an error", test.format, test.text, lines)
			}
			failure, ok := err.(core.Failure)
			if !ok {
				t.Fatalf("Ulp(%q, %q) error has type %T, want core.Failure", test.format, test.text, err)
			}
			if failure.Message != test.message {
				t.Errorf("Ulp(%q, %q) error = %q, want %q", test.format, test.text, failure.Message, test.message)
			}
			if len(lines) != 0 {
				t.Errorf("Ulp(%q, %q) returned lines on failure: %#v", test.format, test.text, lines)
			}
		})
	}
}
