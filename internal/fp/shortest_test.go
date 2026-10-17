package fp_test

import (
	"math/big"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
	"github.com/Gulliby-development/fpbits/internal/num"
)

func mustShortestFormat(t *testing.T, name string) core.Format {
	t.Helper()
	format, err := core.Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%q) returned error: %v", name, err)
	}
	return format
}

func mustShortestNumber(t *testing.T, text string) num.Number {
	t.Helper()
	number, err := num.Parse(text, false)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", text, err)
	}
	if number.Kind != num.Finite || number.Rat == nil {
		t.Fatalf("Parse(%q) = %#v, want a finite number", text, number)
	}
	return number
}

func mustShortestSource(t *testing.T, text string) *big.Rat {
	t.Helper()
	number := mustShortestNumber(t, text)
	r := new(big.Rat).Set(number.Rat)
	if number.Neg {
		r.Neg(r)
	}
	return r
}

func TestShortestDigitsAndFiniteText(t *testing.T) {
	tests := []struct {
		name   string
		format string
		mag    uint64
		digits string
		exp10  int
		text   string
	}{
		{name: "f16 zero", format: "f16", mag: 0x0000, digits: "0", exp10: 0, text: "0e0"},
		{name: "f16 decimal tenth has minimal tuple", format: "f16", mag: 0x2e66, digits: "1", exp10: -1, text: "1e-1"},
		{name: "f16 smallest subnormal", format: "f16", mag: 0x0001, digits: "6", exp10: -8, text: "6e-8"},
		{name: "f16 smallest normal", format: "f16", mag: 0x0400, digits: "6104", exp10: -5, text: "6.104e-5"},
		{name: "f16 value below one", format: "f16", mag: 0x3bff, digits: "9995", exp10: -1, text: "9.995e-1"},
		{name: "f16 one", format: "f16", mag: 0x3c00, digits: "1", exp10: 0, text: "1e0"},
		{name: "f16 value above one", format: "f16", mag: 0x3c01, digits: "1001", exp10: 0, text: "1.001e0"},
		{name: "f16 value below two", format: "f16", mag: 0x3fff, digits: "1999", exp10: 0, text: "1.999e0"},
		{name: "f16 exponent carry to two", format: "f16", mag: 0x4000, digits: "2", exp10: 0, text: "2e0"},
		{name: "f16 decimal decade carry candidate is rejected", format: "f16", mag: 0x48ff, digits: "999", exp10: 0, text: "9.99e0"},
		{name: "f16 exact decimal decade", format: "f16", mag: 0x4900, digits: "1", exp10: 1, text: "1e1"},
		{name: "f16 largest finite", format: "f16", mag: 0x7bff, digits: "655", exp10: 4, text: "6.55e4"},
		{name: "bf16 smallest subnormal", format: "bf16", mag: 0x0001, digits: "9", exp10: -41, text: "9e-41"},
		{name: "bf16 one", format: "bf16", mag: 0x3f80, digits: "1", exp10: 0, text: "1e0"},
		{name: "bf16 largest finite", format: "bf16", mag: 0x7f7f, digits: "339", exp10: 38, text: "3.39e38"},
		{name: "f32 smallest subnormal", format: "f32", mag: 0x00000001, digits: "1", exp10: -45, text: "1e-45"},
		{name: "f32 one", format: "f32", mag: 0x3f800000, digits: "1", exp10: 0, text: "1e0"},
		{name: "f32 next value after one", format: "f32", mag: 0x3f800001, digits: "10000001", exp10: 0, text: "1.0000001e0"},
		{name: "f32 largest finite", format: "f32", mag: 0x7f7fffff, digits: "34028235", exp10: 38, text: "3.4028235e38"},
		{name: "f64 one", format: "f64", mag: 0x3ff0000000000000, digits: "1", exp10: 0, text: "1e0"},
		{name: "f64 decimal conversion example", format: "f64", mag: 0x3fd3333333333334, digits: "30000000000000004", exp10: -1, text: "3.0000000000000004e-1"},
		{name: "f64 largest finite", format: "f64", mag: 0x7fefffffffffffff, digits: "17976931348623157", exp10: 308, text: "1.7976931348623157e308"},
		{name: "e5m2 one", format: "e5m2", mag: 0x3c, digits: "1", exp10: 0, text: "1e0"},
		{name: "e5m2 lower nearest-even shortest tie", format: "e5m2", mag: 0x3d, digits: "12", exp10: 0, text: "1.2e0"},
		{name: "e5m2 upper nearest-even shortest tie", format: "e5m2", mag: 0x3f, digits: "18", exp10: 0, text: "1.8e0"},
		{name: "e5m2 decimal-decade boundary", format: "e5m2", mag: 0x40, digits: "2", exp10: 0, text: "2e0"},
		{name: "e4m3 smallest subnormal", format: "e4m3", mag: 0x01, digits: "2", exp10: -3, text: "2e-3"},
		{name: "e4m3 smallest normal", format: "e4m3", mag: 0x08, digits: "16", exp10: -2, text: "1.6e-2"},
		{name: "e4m3 one", format: "e4m3", mag: 0x38, digits: "1", exp10: 0, text: "1e0"},
		{name: "e4m3 largest finite", format: "e4m3", mag: 0x7e, digits: "45", exp10: 2, text: "4.5e2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format := mustShortestFormat(t, test.format)
			gotDigits, gotExp10 := fp.ShortestDigits(format, test.mag)
			if gotDigits != test.digits || gotExp10 != test.exp10 {
				t.Errorf("ShortestDigits(%s, %#x) = (%q, %d), want (%q, %d)", test.format, test.mag, gotDigits, gotExp10, test.digits, test.exp10)
			}

			value := fp.Decode(format, test.mag)
			if got := fp.Text(value); got != test.text {
				t.Errorf("Text(%s, %#x) = %q, want %q", test.format, test.mag, got, test.text)
			}
			if got := fp.Text(value); got != test.text {
				t.Errorf("repeated Text(%s, %#x) = %q, want %q", test.format, test.mag, got, test.text)
			}
			gotBits, overflow := fp.Encode(format, mustShortestNumber(t, test.text))
			if gotBits != test.mag {
				t.Errorf("Encode(%s, Text(%#x)) bits = %#x, want %#x", test.format, test.mag, gotBits, test.mag)
			}
			if overflow {
				t.Errorf("Encode(%s, Text(%#x)) overflowed, want false", test.format, test.mag)
			}
		})
	}
}

func TestTextSignedZeroAndNonfiniteValuesAcrossFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		bits   uint64
		want   string
	}{
		{name: "positive f16 zero", format: "f16", bits: 0x0000, want: "0e0"},
		{name: "negative f16 zero", format: "f16", bits: 0x8000, want: "-0e0"},
		{name: "positive bf16 infinity", format: "bf16", bits: 0x7f80, want: "inf"},
		{name: "negative bf16 infinity", format: "bf16", bits: 0xff80, want: "-inf"},
		{name: "positive bf16 NaN", format: "bf16", bits: 0x7fc0, want: "nan"},
		{name: "negative bf16 NaN", format: "bf16", bits: 0xffc0, want: "nan"},
		{name: "positive f16 infinity", format: "f16", bits: 0x7c00, want: "inf"},
		{name: "negative f16 infinity", format: "f16", bits: 0xfc00, want: "-inf"},
		{name: "positive f16 NaN", format: "f16", bits: 0x7c01, want: "nan"},
		{name: "negative f16 NaN", format: "f16", bits: 0xfe00, want: "nan"},
		{name: "positive f32 infinity", format: "f32", bits: 0x7f800000, want: "inf"},
		{name: "negative f32 infinity", format: "f32", bits: 0xff800000, want: "-inf"},
		{name: "positive f32 NaN", format: "f32", bits: 0x7f800001, want: "nan"},
		{name: "negative f32 NaN", format: "f32", bits: 0xffc00000, want: "nan"},
		{name: "positive f64 infinity", format: "f64", bits: 0x7ff0000000000000, want: "inf"},
		{name: "negative f64 infinity", format: "f64", bits: 0xfff0000000000000, want: "-inf"},
		{name: "positive f64 NaN", format: "f64", bits: 0x7ff0000000000001, want: "nan"},
		{name: "negative f64 NaN", format: "f64", bits: 0xfff8000000000000, want: "nan"},
		{name: "positive e5m2 infinity", format: "e5m2", bits: 0x7c, want: "inf"},
		{name: "negative e5m2 infinity", format: "e5m2", bits: 0xfc, want: "-inf"},
		{name: "positive e5m2 NaN", format: "e5m2", bits: 0x7d, want: "nan"},
		{name: "negative e5m2 NaN", format: "e5m2", bits: 0xfe, want: "nan"},
		{name: "positive e4m3 NaN", format: "e4m3", bits: 0x7f, want: "nan"},
		{name: "negative e4m3 NaN", format: "e4m3", bits: 0xff, want: "nan"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := fp.Decode(mustShortestFormat(t, test.format), test.bits)
			if got := fp.Text(value); got != test.want {
				t.Errorf("Text(%s, %#x) = %q, want %q", test.format, test.bits, got, test.want)
			}
		})
	}
}

func TestDescribeReturnsCompleteLiteralLines(t *testing.T) {
	tests := []struct {
		name   string
		format string
		bits   uint64
		want   []string
	}{
		{
			name:   "negative zero keeps sign and exact line",
			format: "f16",
			bits:   0x8000,
			want: []string{
				"format: f16",
				"bits: 0x8000",
				"binary: 1 00000 0000000000",
				"class: zero",
				"value: -0e0",
				"exact: -0e0",
			},
		},
		{
			name:   "ordinary finite value includes exact line",
			format: "f16",
			bits:   0x2e66,
			want: []string{
				"format: f16",
				"bits: 0x2e66",
				"binary: 0 01011 1001100110",
				"class: normal",
				"value: 1e-1",
				"exact: 9.99755859375e-2",
			},
		},
		{
			name:   "infinity omits exact line",
			format: "f16",
			bits:   0xfc00,
			want: []string{
				"format: f16",
				"bits: 0xfc00",
				"binary: 1 11111 0000000000",
				"class: inf",
				"value: -inf",
			},
		},
		{
			name:   "e4m3 NaN omits exact line",
			format: "e4m3",
			bits:   0xff,
			want: []string{
				"format: e4m3",
				"bits: 0xff",
				"binary: 1 1111 111",
				"class: nan",
				"value: nan",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := fp.Describe(fp.Decode(mustShortestFormat(t, test.format), test.bits))
			if len(got) != len(test.want) {
				t.Fatalf("Describe returned %d lines, want %d: %#v", len(got), len(test.want), got)
			}
			for i, want := range test.want {
				if got[i] != want {
					t.Errorf("Describe line %d = %q, want %q", i, got[i], want)
				}
			}
			for _, line := range got {
				if containsLineFeed(line) {
					t.Errorf("Describe line contains LF: %q", line)
				}
			}
		})
	}
}

func TestErrorLineExactErrorsAndNilPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		format   string
		bits     uint64
		overflow bool
		want     string
	}{
		{name: "decimal tenth rounded down", source: "0.1", format: "f16", bits: 0x2e66, want: "error: -2.44140625e-5"},
		{name: "negative underflow rounded to signed zero", source: "-1e-30", format: "f16", bits: 0x8000, want: "error: 1e-30"},
		{name: "e4m3 maximum midpoint rounded down", source: "464", format: "e4m3", bits: 0x7e, want: "error: -1.6e1"},
		{name: "finite overflow", source: "65520", format: "f16", bits: 0x7c00, overflow: true, want: "error: overflow"},
		{name: "exact positive result has positive zero error", source: "1", format: "f16", bits: 0x3c00, want: "error: 0e0"},
		{name: "positive error", source: "0.999", format: "f16", bits: 0x3c00, want: "error: 1e-3"},
		{name: "nil source precedes overflow and nonfinite result", format: "f16", bits: 0x7c00, overflow: true, want: ""},
		{name: "nil source for finite result", format: "f16", bits: 0x3c00, want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := fp.Decode(mustShortestFormat(t, test.format), test.bits)
			var source *big.Rat
			if test.source != "" {
				source = mustShortestSource(t, test.source)
			}
			before := ""
			if source != nil {
				before = source.RatString()
			}
			first := fp.ErrorLine(source, value, test.overflow)
			if first != test.want {
				t.Errorf("ErrorLine(%q, %s, %#x, %t) = %q, want %q", test.source, test.format, test.bits, test.overflow, first, test.want)
			}
			if containsLineFeed(first) {
				t.Errorf("ErrorLine returned LF: %q", first)
			}
			second := fp.ErrorLine(source, value, test.overflow)
			if second != first {
				t.Errorf("repeated ErrorLine = %q, want %q", second, first)
			}
			if source != nil && source.RatString() != before {
				t.Errorf("ErrorLine mutated source: got %s, want %s", source.RatString(), before)
			}
		})
	}
}

func containsLineFeed(text string) bool {
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			return true
		}
	}
	return false
}
