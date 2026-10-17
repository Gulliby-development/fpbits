package analysis_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/analysis"
	"github.com/Gulliby-development/fpbits/internal/core"
)

func mustLimitsFormat(t *testing.T, name string) core.Format {
	t.Helper()
	format, err := core.Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%q) returned error: %v", name, err)
	}
	return format
}

func TestLimitsCompleteLiteralReports(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   []string
	}{
		{
			name:   "f16 exact worked range report",
			format: "f16",
			want: []string{
				"format: f16",
				"layout: 1 sign bit, 5 exponent bits, 10 fraction bits",
				"bias: 15",
				"smallest subnormal: 6e-8 (0x0001)",
				"smallest normal: 6.104e-5 (0x0400)",
				"largest: 6.55e4 (0x7bff)",
				"epsilon: 9.77e-4",
				"infinities: 2",
				"nan encodings: 2046",
				"finite encodings: 63488",
			},
		},
		{
			name:   "bf16 counts and endpoints",
			format: "bf16",
			want: []string{
				"format: bf16",
				"layout: 1 sign bit, 8 exponent bits, 7 fraction bits",
				"bias: 127",
				"smallest subnormal: 9e-41 (0x0001)",
				"smallest normal: 1.18e-38 (0x0080)",
				"largest: 3.39e38 (0x7f7f)",
				"epsilon: 7.8e-3",
				"infinities: 2",
				"nan encodings: 254",
				"finite encodings: 65280",
			},
		},
		{
			name:   "f32 counts and endpoints",
			format: "f32",
			want: []string{
				"format: f32",
				"layout: 1 sign bit, 8 exponent bits, 23 fraction bits",
				"bias: 127",
				"smallest subnormal: 1e-45 (0x00000001)",
				"smallest normal: 1.1754944e-38 (0x00800000)",
				"largest: 3.4028235e38 (0x7f7fffff)",
				"epsilon: 1.1920929e-7",
				"infinities: 2",
				"nan encodings: 16777214",
				"finite encodings: 4278190080",
			},
		},
		{
			name:   "f64 counts do not overflow",
			format: "f64",
			want: []string{
				"format: f64",
				"layout: 1 sign bit, 11 exponent bits, 52 fraction bits",
				"bias: 1023",
				"smallest subnormal: 5e-324 (0x0000000000000001)",
				"smallest normal: 2.2250738585072014e-308 (0x0010000000000000)",
				"largest: 1.7976931348623157e308 (0x7fefffffffffffff)",
				"epsilon: 2.220446049250313e-16",
				"infinities: 2",
				"nan encodings: 9007199254740990",
				"finite encodings: 18437736874454810624",
			},
		},
		{
			name:   "e5m2 IEEE special encodings",
			format: "e5m2",
			want: []string{
				"format: e5m2",
				"layout: 1 sign bit, 5 exponent bits, 2 fraction bits",
				"bias: 15",
				"smallest subnormal: 2e-5 (0x01)",
				"smallest normal: 6e-5 (0x04)",
				"largest: 6e4 (0x7b)",
				"epsilon: 2.5e-1",
				"infinities: 2",
				"nan encodings: 6",
				"finite encodings: 248",
			},
		},
		{
			name:   "e4m3 has no infinity encodings",
			format: "e4m3",
			want: []string{
				"format: e4m3",
				"layout: 1 sign bit, 4 exponent bits, 3 fraction bits",
				"bias: 7",
				"smallest subnormal: 2e-3 (0x01)",
				"smallest normal: 1.6e-2 (0x08)",
				"largest: 4.5e2 (0x7e)",
				"epsilon: 1.3e-1",
				"infinities: 0",
				"nan encodings: 2",
				"finite encodings: 254",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := analysis.Limits(mustLimitsFormat(t, test.format))
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Limits(%s) = %#v, want %#v", test.format, got, test.want)
			}
			for _, line := range got {
				if strings.Contains(line, "\n") {
					t.Errorf("Limits(%s) returned a line containing LF: %q", test.format, line)
				}
			}
		})
	}
}
