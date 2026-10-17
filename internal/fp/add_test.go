package fp_test

import (
	"testing"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
)

func mustAddFormat(t *testing.T, name string) core.Format {
	t.Helper()
	format, err := core.Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%q) returned error: %v", name, err)
	}
	return format
}

func TestAddSpecialCasesAndFiniteRounding(t *testing.T) {
	tests := []struct {
		name       string
		formatName string
		a          uint64
		b          uint64
		wantBits   uint64
		wantClass  fp.Class
		wantNeg    bool
	}{
		{name: "f16 negative zero plus negative zero", formatName: "f16", a: 0x8000, b: 0x8000, wantBits: 0x8000, wantClass: fp.Zero, wantNeg: true},
		{name: "f16 positive zero plus negative zero", formatName: "f16", a: 0x0000, b: 0x8000, wantBits: 0x0000, wantClass: fp.Zero},
		{name: "f16 negative zero plus positive zero", formatName: "f16", a: 0x8000, b: 0x0000, wantBits: 0x0000, wantClass: fp.Zero},
		{name: "f16 positive zero plus positive zero", formatName: "f16", a: 0x0000, b: 0x0000, wantBits: 0x0000, wantClass: fp.Zero},
		{name: "f16 positive and negative one cancel", formatName: "f16", a: 0x3c00, b: 0xbc00, wantBits: 0x0000, wantClass: fp.Zero},
		{name: "f16 two subnormal units", formatName: "f16", a: 0x0001, b: 0x0001, wantBits: 0x0002, wantClass: fp.Subnormal},
		{name: "f16 subnormal sum carries to smallest normal", formatName: "f16", a: 0x03ff, b: 0x0001, wantBits: 0x0400, wantClass: fp.Normal},
		{name: "f16 midpoint sum rounds across exponent boundary", formatName: "f16", a: 0x3fff, b: 0x1000, wantBits: 0x4000, wantClass: fp.Normal},
		{name: "f16 rounded sum chooses even lower value", formatName: "f16", a: 0x6800, b: 0x3c00, wantBits: 0x6800, wantClass: fp.Normal},
		{name: "f16 positive finite overflow", formatName: "f16", a: 0x7bff, b: 0x7bff, wantBits: 0x7c00, wantClass: fp.Inf},
		{name: "f16 negative finite overflow", formatName: "f16", a: 0xfbff, b: 0xfbff, wantBits: 0xfc00, wantClass: fp.Inf, wantNeg: true},
		{name: "f16 positive infinity plus positive infinity", formatName: "f16", a: 0x7c00, b: 0x7c00, wantBits: 0x7c00, wantClass: fp.Inf},
		{name: "f16 negative infinity plus negative infinity", formatName: "f16", a: 0xfc00, b: 0xfc00, wantBits: 0xfc00, wantClass: fp.Inf, wantNeg: true},
		{name: "f16 opposite infinities produce canonical NaN", formatName: "f16", a: 0x7c00, b: 0xfc00, wantBits: 0x7e00, wantClass: fp.NaN},
		{name: "f16 reversed opposite infinities produce canonical NaN", formatName: "f16", a: 0xfc00, b: 0x7c00, wantBits: 0x7e00, wantClass: fp.NaN},
		{name: "f16 positive infinity plus finite", formatName: "f16", a: 0x7c00, b: 0x3c00, wantBits: 0x7c00, wantClass: fp.Inf},
		{name: "f16 finite plus negative infinity", formatName: "f16", a: 0x3c00, b: 0xfc00, wantBits: 0xfc00, wantClass: fp.Inf, wantNeg: true},
		{name: "f16 NaN on left propagates canonical NaN", formatName: "f16", a: 0xfe01, b: 0x3c00, wantBits: 0x7e00, wantClass: fp.NaN},
		{name: "f16 NaN on right propagates canonical NaN", formatName: "f16", a: 0x3c00, b: 0xfe01, wantBits: 0x7e00, wantClass: fp.NaN},
		{name: "f16 two NaNs produce canonical NaN", formatName: "f16", a: 0xfe01, b: 0x7fff, wantBits: 0x7e00, wantClass: fp.NaN},
		{name: "bf16 one plus one", formatName: "bf16", a: 0x3f80, b: 0x3f80, wantBits: 0x4000, wantClass: fp.Normal},
		{name: "f32 one plus one", formatName: "f32", a: 0x3f800000, b: 0x3f800000, wantBits: 0x40000000, wantClass: fp.Normal},
		{name: "f64 one plus one", formatName: "f64", a: 0x3ff0000000000000, b: 0x3ff0000000000000, wantBits: 0x4000000000000000, wantClass: fp.Normal},
		{name: "e5m2 one plus one", formatName: "e5m2", a: 0x3c, b: 0x3c, wantBits: 0x40, wantClass: fp.Normal},
		{name: "e4m3 one plus one", formatName: "e4m3", a: 0x38, b: 0x38, wantBits: 0x40, wantClass: fp.Normal},
		{name: "e4m3 positive finite overflow becomes NaN", formatName: "e4m3", a: 0x7e, b: 0x7e, wantBits: 0x7f, wantClass: fp.NaN},
		{name: "e4m3 negative finite overflow becomes signed NaN", formatName: "e4m3", a: 0xfe, b: 0xfe, wantBits: 0xff, wantClass: fp.NaN, wantNeg: true},
		{name: "e4m3 NaN propagation is canonical and positive", formatName: "e4m3", a: 0xff, b: 0x38, wantBits: 0x7f, wantClass: fp.NaN},
		{name: "e4m3 NaN on right propagates canonical NaN", formatName: "e4m3", a: 0x38, b: 0x7f, wantBits: 0x7f, wantClass: fp.NaN},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format := mustAddFormat(t, test.formatName)
			a := fp.Decode(format, test.a)
			b := fp.Decode(format, test.b)
			originalA, originalB := a, b

			got := fp.Add(a, b)
			if got.Bits != test.wantBits {
				t.Errorf("Add(%s, %#x, %#x).Bits = %#x, want %#x", test.formatName, test.a, test.b, got.Bits, test.wantBits)
			}
			if got.Class != test.wantClass {
				t.Errorf("Add(%s, %#x, %#x).Class = %s, want %s", test.formatName, test.a, test.b, got.Class, test.wantClass)
			}
			if got.Fmt != format || got.Fmt.Name != test.formatName {
				t.Errorf("Add result format = %#v, want %s", got.Fmt, test.formatName)
			}
			if got.Neg != test.wantNeg {
				t.Errorf("Add result Neg = %t, want %t", got.Neg, test.wantNeg)
			}
			if a != originalA || b != originalB {
				t.Errorf("Add mutated its operands: a = %#v, b = %#v", a, b)
			}

			repeated := fp.Add(a, b)
			if repeated != got {
				t.Errorf("repeated Add(%s, %#x, %#x) = %#v, want identical result %#v", test.formatName, test.a, test.b, repeated, got)
			}
		})
	}
}
