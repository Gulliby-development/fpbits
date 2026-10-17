package fp_test

import (
	"testing"

	"github.com/Gulliby-development/fpbits/internal/fp"
	"github.com/Gulliby-development/fpbits/internal/num"
)

func mustFiniteNumber(t *testing.T, text string) num.Number {
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

func TestEncodeMagRoundsExactRationals(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		text     string
		wantMag  uint64
		overflow bool
	}{
		{name: "f16 decimal tenth", format: "f16", text: "0.1", wantMag: 0x2e66},
		{name: "f16 just below maximum overflow midpoint", format: "f16", text: "65519.99", wantMag: 0x7bff},
		{name: "f16 last integer below overflow midpoint", format: "f16", text: "65519", wantMag: 0x7bff},
		{name: "f16 overflow midpoint tie chooses infinity", format: "f16", text: "65520", wantMag: 0x7c00, overflow: true},
		{name: "f16 exponent carry below tie", format: "f16", text: "1.99951171874", wantMag: 0x3fff},
		{name: "f16 exponent carry midpoint chooses even", format: "f16", text: "1.99951171875", wantMag: 0x4000},
		{name: "f16 exponent carry above tie", format: "f16", text: "1.99951171876", wantMag: 0x4000},
		{name: "f16 subnormal normal midpoint below", format: "f16", text: "0.0000610053", wantMag: 0x03ff},
		{name: "f16 subnormal normal midpoint tie", format: "f16", text: "0.0000610053539276123046875", wantMag: 0x0400},
		{name: "f16 subnormal normal midpoint above", format: "f16", text: "0.0000610054", wantMag: 0x0400},
		{name: "f16 half minimum subnormal tie chooses zero", format: "f16", text: "0.0000000298023223876953125", wantMag: 0x0000},
		{name: "f16 above half minimum subnormal", format: "f16", text: "0.0000000298023223876953126", wantMag: 0x0001},
		{name: "f16 exact maximum finite", format: "f16", text: "65504", wantMag: 0x7bff},
		{name: "bf16 midpoint at one chooses even lower", format: "bf16", text: "1.00390625", wantMag: 0x3f80},
		{name: "bf16 just above midpoint at one", format: "bf16", text: "1.0039062501", wantMag: 0x3f81},
		{name: "bf16 minimum subnormal rounds up", format: "bf16", text: "1e-40", wantMag: 0x0001},
		{name: "bf16 underflow remains zero", format: "bf16", text: "1e-42", wantMag: 0x0000},
		{name: "bf16 maximum finite", format: "bf16", text: "3.3895313892515355e38", wantMag: 0x7f7f},
		{name: "bf16 finite value just below overflow threshold", format: "bf16", text: "3.39e38", wantMag: 0x7f7f},
		{name: "bf16 overflow above threshold", format: "bf16", text: "3.4028236e38", wantMag: 0x7f80, overflow: true},
		{name: "f32 midpoint at one chooses even lower", format: "f32", text: "1.000000059604644775390625", wantMag: 0x3f800000},
		{name: "f32 just above midpoint at one", format: "f32", text: "1.000000059604644775390626", wantMag: 0x3f800001},
		{name: "f32 minimum subnormal rounds up", format: "f32", text: "1e-45", wantMag: 0x00000001},
		{name: "f32 underflow remains zero", format: "f32", text: "1e-46", wantMag: 0x00000000},
		{name: "f32 maximum finite", format: "f32", text: "3.4028234663852886e38", wantMag: 0x7f7fffff},
		{name: "f32 overflow", format: "f32", text: "3.5e38", wantMag: 0x7f800000, overflow: true},
		{name: "f64 midpoint at one chooses even lower", format: "f64", text: "1.00000000000000011102230246251565404236316680908203125", wantMag: 0x3ff0000000000000},
		{name: "f64 just above midpoint at one", format: "f64", text: "1.00000000000000011102230246251565404236316680908203126", wantMag: 0x3ff0000000000001},
		{name: "f64 maximum finite", format: "f64", text: "1.7976931348623157e308", wantMag: 0x7fefffffffffffff},
		{name: "f64 overflow", format: "f64", text: "1.8e308", wantMag: 0x7ff0000000000000, overflow: true},
		{name: "e5m2 lower-even midpoint", format: "e5m2", text: "1.125", wantMag: 0x3c},
		{name: "e5m2 above lower-even midpoint", format: "e5m2", text: "1.1251", wantMag: 0x3d},
		{name: "e5m2 below carry midpoint", format: "e5m2", text: "1.8749", wantMag: 0x3f},
		{name: "e5m2 carry midpoint chooses even upper", format: "e5m2", text: "1.875", wantMag: 0x40},
		{name: "e5m2 above carry midpoint", format: "e5m2", text: "1.8751", wantMag: 0x40},
		{name: "e5m2 exact maximum finite", format: "e5m2", text: "57344", wantMag: 0x7b},
		{name: "e5m2 below overflow tie", format: "e5m2", text: "61439", wantMag: 0x7b},
		{name: "e5m2 overflow tie chooses infinity", format: "e5m2", text: "61440", wantMag: 0x7c, overflow: true},
		{name: "e5m2 overflow", format: "e5m2", text: "1e5", wantMag: 0x7c, overflow: true},
		{name: "e4m3 lower-even overflow midpoint", format: "e4m3", text: "464", wantMag: 0x7e},
		{name: "e4m3 just below midpoint", format: "e4m3", text: "463.99", wantMag: 0x7e},
		{name: "e4m3 above overflow midpoint", format: "e4m3", text: "464.01", wantMag: 0x7f, overflow: true},
		{name: "e4m3 exact maximum finite", format: "e4m3", text: "448", wantMag: 0x7e},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format := mustFormat(t, test.format)
			number := mustFiniteNumber(t, test.text)
			before := number.Rat.RatString()
			gotMag, gotOverflow := fp.EncodeMag(format, number.Rat)
			if gotMag != test.wantMag {
				t.Errorf("EncodeMag(%s, %q) magnitude = %#x, want %#x", test.format, test.text, gotMag, test.wantMag)
			}
			if gotOverflow != test.overflow {
				t.Errorf("EncodeMag(%s, %q) overflow = %t, want %t", test.format, test.text, gotOverflow, test.overflow)
			}
			if got := number.Rat.RatString(); got != before {
				t.Errorf("EncodeMag mutated its rational: got %s, want %s", got, before)
			}
		})
	}
}

func TestEncodeFiniteSignsUnderflowAndInputLimits(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		text     string
		wantBits uint64
		overflow bool
	}{
		{name: "positive zero", format: "f16", text: "0", wantBits: 0x0000},
		{name: "negative zero", format: "f16", text: "-0", wantBits: 0x8000},
		{name: "positive f16 underflow", format: "f16", text: "1e-30", wantBits: 0x0000},
		{name: "negative f16 underflow keeps sign", format: "f16", text: "-1e-30", wantBits: 0x8000},
		{name: "negative f16 finite rounding", format: "f16", text: "-0.1", wantBits: 0xae66},
		{name: "negative f16 finite overflow", format: "f16", text: "-65520", wantBits: 0xfc00, overflow: true},
		{name: "negative e4m3 tie is finite", format: "e4m3", text: "-464", wantBits: 0xfe},
		{name: "negative e4m3 overflow becomes NaN", format: "e4m3", text: "-464.01", wantBits: 0xff, overflow: true},
		{name: "inclusive minimum decimal exponent", format: "f16", text: "1e-5000", wantBits: 0x0000},
		{name: "negative inclusive minimum exponent keeps sign", format: "f16", text: "-1e-5000", wantBits: 0x8000},
		{name: "inclusive maximum decimal exponent overflows", format: "f16", text: "1e5000", wantBits: 0x7c00, overflow: true},
		{name: "negative inclusive maximum exponent overflows", format: "f16", text: "-1e5000", wantBits: 0xfc00, overflow: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format := mustFormat(t, test.format)
			number := mustFiniteNumber(t, test.text)
			before := number.Rat.RatString()
			gotBits, gotOverflow := fp.Encode(format, number)
			if gotBits != test.wantBits {
				t.Errorf("Encode(%s, %q) bits = %#x, want %#x", test.format, test.text, gotBits, test.wantBits)
			}
			if gotOverflow != test.overflow {
				t.Errorf("Encode(%s, %q) overflow = %t, want %t", test.format, test.text, gotOverflow, test.overflow)
			}
			if got := number.Rat.RatString(); got != before {
				t.Errorf("Encode mutated its rational: got %s, want %s", got, before)
			}
		})
	}
}

const fourHundredDigitOperand = "9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999" +
	"9999999999999999999999999999999999999999"

func TestEncodeHandwrittenFourHundredDigitOperand(t *testing.T) {
	if len(fourHundredDigitOperand) != 400 {
		t.Fatalf("four-hundred-digit fixture length = %d, want 400", len(fourHundredDigitOperand))
	}
	number := mustFiniteNumber(t, fourHundredDigitOperand)
	bits, overflow := fp.Encode(mustFormat(t, "f16"), number)
	if bits != 0x7c00 || !overflow {
		t.Errorf("Encode(f16, 400-digit operand) = %#x, %t, want 0x7c00, true", bits, overflow)
	}
}

func TestEncodeSpecialsAndSignedZerosAcrossFormats(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		kind     num.Kind
		neg      bool
		wantBits uint64
	}{
		{name: "f16 positive infinity", format: "f16", kind: num.Inf, wantBits: 0x7c00},
		{name: "f16 negative infinity", format: "f16", kind: num.Inf, neg: true, wantBits: 0xfc00},
		{name: "f16 positive NaN", format: "f16", kind: num.NaN, wantBits: 0x7e00},
		{name: "f16 negative NaN", format: "f16", kind: num.NaN, neg: true, wantBits: 0xfe00},
		{name: "bf16 positive infinity", format: "bf16", kind: num.Inf, wantBits: 0x7f80},
		{name: "bf16 negative infinity", format: "bf16", kind: num.Inf, neg: true, wantBits: 0xff80},
		{name: "bf16 positive NaN", format: "bf16", kind: num.NaN, wantBits: 0x7fc0},
		{name: "bf16 negative NaN", format: "bf16", kind: num.NaN, neg: true, wantBits: 0xffc0},
		{name: "f32 positive infinity", format: "f32", kind: num.Inf, wantBits: 0x7f800000},
		{name: "f32 negative infinity", format: "f32", kind: num.Inf, neg: true, wantBits: 0xff800000},
		{name: "f32 positive NaN", format: "f32", kind: num.NaN, wantBits: 0x7fc00000},
		{name: "f32 negative NaN", format: "f32", kind: num.NaN, neg: true, wantBits: 0xffc00000},
		{name: "f64 positive infinity", format: "f64", kind: num.Inf, wantBits: 0x7ff0000000000000},
		{name: "f64 negative infinity", format: "f64", kind: num.Inf, neg: true, wantBits: 0xfff0000000000000},
		{name: "f64 positive NaN", format: "f64", kind: num.NaN, wantBits: 0x7ff8000000000000},
		{name: "f64 negative NaN", format: "f64", kind: num.NaN, neg: true, wantBits: 0xfff8000000000000},
		{name: "e5m2 positive infinity", format: "e5m2", kind: num.Inf, wantBits: 0x7c},
		{name: "e5m2 negative infinity", format: "e5m2", kind: num.Inf, neg: true, wantBits: 0xfc},
		{name: "e5m2 positive NaN", format: "e5m2", kind: num.NaN, wantBits: 0x7e},
		{name: "e5m2 negative NaN", format: "e5m2", kind: num.NaN, neg: true, wantBits: 0xfe},
		{name: "e4m3 positive infinity encodes NaN", format: "e4m3", kind: num.Inf, wantBits: 0x7f},
		{name: "e4m3 negative infinity encodes NaN", format: "e4m3", kind: num.Inf, neg: true, wantBits: 0xff},
		{name: "e4m3 positive NaN", format: "e4m3", kind: num.NaN, wantBits: 0x7f},
		{name: "e4m3 negative NaN", format: "e4m3", kind: num.NaN, neg: true, wantBits: 0xff},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotBits, overflow := fp.Encode(mustFormat(t, test.format), num.Number{Kind: test.kind, Neg: test.neg})
			if gotBits != test.wantBits {
				t.Errorf("Encode(%s, kind %v, negative %t) bits = %#x, want %#x", test.format, test.kind, test.neg, gotBits, test.wantBits)
			}
			if overflow {
				t.Errorf("Encode(%s, kind %v, negative %t) overflow = true, want false for explicit special", test.format, test.kind, test.neg)
			}
		})
	}
}

func TestEncodeMagZeroAndRepeatedCalls(t *testing.T) {
	format := mustFormat(t, "f16")
	zero := mustFiniteNumber(t, "0")
	magnitude, overflow := fp.EncodeMag(format, zero.Rat)
	if magnitude != 0 || overflow {
		t.Errorf("EncodeMag(f16, zero) = %#x, %t, want 0, false", magnitude, overflow)
	}

	number := mustFiniteNumber(t, "0.1")
	before := number.Rat.RatString()
	firstBits, firstOverflow := fp.Encode(format, number)
	secondBits, secondOverflow := fp.Encode(format, number)
	if firstBits != 0x2e66 || firstOverflow {
		t.Errorf("first Encode(f16, 0.1) = %#x, %t, want 0x2e66, false", firstBits, firstOverflow)
	}
	if secondBits != firstBits || secondOverflow != firstOverflow {
		t.Errorf("repeated Encode results differ: (%#x, %t) and (%#x, %t)", firstBits, firstOverflow, secondBits, secondOverflow)
	}
	if got := number.Rat.RatString(); got != before {
		t.Errorf("repeated Encode mutated its rational: got %s, want %s", got, before)
	}

	magnitude, overflow = fp.EncodeMag(format, number.Rat)
	if magnitude != 0x2e66 || overflow {
		t.Errorf("repeated EncodeMag(f16, 0.1) = %#x, %t, want 0x2e66, false", magnitude, overflow)
	}
	if got := number.Rat.RatString(); got != before {
		t.Errorf("EncodeMag mutated its rational: got %s, want %s", got, before)
	}
}
