package fp_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
)

func mustFormat(t *testing.T, name string) core.Format {
	t.Helper()
	format, err := core.Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%q) returned error: %v", name, err)
	}
	return format
}

func TestDecodeClassesAndSignedValues(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		bits      uint64
		neg       bool
		mag       uint64
		class     fp.Class
		magRat    string
		signedRat string
	}{
		{name: "f16 positive zero", format: "f16", bits: 0x0000, mag: 0x0000, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "f16 negative zero", format: "f16", bits: 0x8000, neg: true, mag: 0x0000, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "f16 positive subnormal", format: "f16", bits: 0x0001, mag: 0x0001, class: fp.Subnormal, magRat: "1/16777216", signedRat: "1/16777216"},
		{name: "f16 negative subnormal", format: "f16", bits: 0x8001, neg: true, mag: 0x0001, class: fp.Subnormal, magRat: "1/16777216", signedRat: "-1/16777216"},
		{name: "f16 positive normal", format: "f16", bits: 0x3c00, mag: 0x3c00, class: fp.Normal, magRat: "1", signedRat: "1"},
		{name: "f16 negative normal", format: "f16", bits: 0xbc00, neg: true, mag: 0x3c00, class: fp.Normal, magRat: "1", signedRat: "-1"},
		{name: "f16 positive infinity", format: "f16", bits: 0x7c00, mag: 0x7c00, class: fp.Inf},
		{name: "f16 negative infinity", format: "f16", bits: 0xfc00, neg: true, mag: 0x7c00, class: fp.Inf},
		{name: "f16 positive NaN", format: "f16", bits: 0x7c01, mag: 0x7c01, class: fp.NaN},
		{name: "f16 negative NaN", format: "f16", bits: 0xfc01, neg: true, mag: 0x7c01, class: fp.NaN},

		{name: "bf16 positive zero", format: "bf16", bits: 0x0000, mag: 0x0000, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "bf16 negative zero", format: "bf16", bits: 0x8000, neg: true, mag: 0x0000, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "bf16 positive subnormal", format: "bf16", bits: 0x0001, mag: 0x0001, class: fp.Subnormal, magRat: "1/10889035741470030830827987437816582766592", signedRat: "1/10889035741470030830827987437816582766592"},
		{name: "bf16 negative subnormal", format: "bf16", bits: 0x8001, neg: true, mag: 0x0001, class: fp.Subnormal, magRat: "1/10889035741470030830827987437816582766592", signedRat: "-1/10889035741470030830827987437816582766592"},
		{name: "bf16 positive normal", format: "bf16", bits: 0x3f80, mag: 0x3f80, class: fp.Normal, magRat: "1", signedRat: "1"},
		{name: "bf16 negative normal", format: "bf16", bits: 0xbf80, neg: true, mag: 0x3f80, class: fp.Normal, magRat: "1", signedRat: "-1"},
		{name: "bf16 positive infinity", format: "bf16", bits: 0x7f80, mag: 0x7f80, class: fp.Inf},
		{name: "bf16 negative infinity", format: "bf16", bits: 0xff80, neg: true, mag: 0x7f80, class: fp.Inf},
		{name: "bf16 positive NaN", format: "bf16", bits: 0x7f81, mag: 0x7f81, class: fp.NaN},
		{name: "bf16 negative NaN", format: "bf16", bits: 0xff81, neg: true, mag: 0x7f81, class: fp.NaN},

		{name: "f32 positive zero", format: "f32", bits: 0x00000000, mag: 0x00000000, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "f32 negative zero", format: "f32", bits: 0x80000000, neg: true, mag: 0x00000000, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "f32 positive subnormal", format: "f32", bits: 0x00000001, mag: 0x00000001, class: fp.Subnormal, magRat: "1/713623846352979940529142984724747568191373312", signedRat: "1/713623846352979940529142984724747568191373312"},
		{name: "f32 negative subnormal", format: "f32", bits: 0x80000001, neg: true, mag: 0x00000001, class: fp.Subnormal, magRat: "1/713623846352979940529142984724747568191373312", signedRat: "-1/713623846352979940529142984724747568191373312"},
		{name: "f32 positive normal", format: "f32", bits: 0x3f800000, mag: 0x3f800000, class: fp.Normal, magRat: "1", signedRat: "1"},
		{name: "f32 negative normal", format: "f32", bits: 0xbf800000, neg: true, mag: 0x3f800000, class: fp.Normal, magRat: "1", signedRat: "-1"},
		{name: "f32 positive infinity", format: "f32", bits: 0x7f800000, mag: 0x7f800000, class: fp.Inf},
		{name: "f32 negative infinity", format: "f32", bits: 0xff800000, neg: true, mag: 0x7f800000, class: fp.Inf},
		{name: "f32 positive NaN", format: "f32", bits: 0x7f800001, mag: 0x7f800001, class: fp.NaN},
		{name: "f32 negative NaN", format: "f32", bits: 0xff800001, neg: true, mag: 0x7f800001, class: fp.NaN},

		{name: "f64 positive zero", format: "f64", bits: 0x0000000000000000, mag: 0x0000000000000000, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "f64 negative zero", format: "f64", bits: 0x8000000000000000, neg: true, mag: 0x0000000000000000, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "f64 positive subnormal", format: "f64", bits: 0x0000000000000001, mag: 0x0000000000000001, class: fp.Subnormal, magRat: "1/202402253307310618352495346718917307049556649764142118356901358027430339567995346891960383701437124495187077864316811911389808737385793476867013399940738509921517424276566361364466907742093216341239767678472745068562007483424692698618103355649159556340810056512358769552333414615230502532186327508646006263307707741093494784", signedRat: "1/202402253307310618352495346718917307049556649764142118356901358027430339567995346891960383701437124495187077864316811911389808737385793476867013399940738509921517424276566361364466907742093216341239767678472745068562007483424692698618103355649159556340810056512358769552333414615230502532186327508646006263307707741093494784"},
		{name: "f64 negative subnormal", format: "f64", bits: 0x8000000000000001, neg: true, mag: 0x0000000000000001, class: fp.Subnormal, magRat: "1/202402253307310618352495346718917307049556649764142118356901358027430339567995346891960383701437124495187077864316811911389808737385793476867013399940738509921517424276566361364466907742093216341239767678472745068562007483424692698618103355649159556340810056512358769552333414615230502532186327508646006263307707741093494784", signedRat: "-1/202402253307310618352495346718917307049556649764142118356901358027430339567995346891960383701437124495187077864316811911389808737385793476867013399940738509921517424276566361364466907742093216341239767678472745068562007483424692698618103355649159556340810056512358769552333414615230502532186327508646006263307707741093494784"},
		{name: "f64 positive normal", format: "f64", bits: 0x3ff0000000000000, mag: 0x3ff0000000000000, class: fp.Normal, magRat: "1", signedRat: "1"},
		{name: "f64 negative normal", format: "f64", bits: 0xbff0000000000000, neg: true, mag: 0x3ff0000000000000, class: fp.Normal, magRat: "1", signedRat: "-1"},
		{name: "f64 positive infinity", format: "f64", bits: 0x7ff0000000000000, mag: 0x7ff0000000000000, class: fp.Inf},
		{name: "f64 negative infinity", format: "f64", bits: 0xfff0000000000000, neg: true, mag: 0x7ff0000000000000, class: fp.Inf},
		{name: "f64 positive NaN", format: "f64", bits: 0x7ff0000000000001, mag: 0x7ff0000000000001, class: fp.NaN},
		{name: "f64 negative NaN", format: "f64", bits: 0xfff0000000000001, neg: true, mag: 0x7ff0000000000001, class: fp.NaN},

		{name: "e5m2 positive zero", format: "e5m2", bits: 0x00, mag: 0x00, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "e5m2 negative zero", format: "e5m2", bits: 0x80, neg: true, mag: 0x00, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "e5m2 positive subnormal", format: "e5m2", bits: 0x01, mag: 0x01, class: fp.Subnormal, magRat: "1/65536", signedRat: "1/65536"},
		{name: "e5m2 negative subnormal", format: "e5m2", bits: 0x81, neg: true, mag: 0x01, class: fp.Subnormal, magRat: "1/65536", signedRat: "-1/65536"},
		{name: "e5m2 positive normal", format: "e5m2", bits: 0x3c, mag: 0x3c, class: fp.Normal, magRat: "1", signedRat: "1"},
		{name: "e5m2 negative normal", format: "e5m2", bits: 0xbc, neg: true, mag: 0x3c, class: fp.Normal, magRat: "1", signedRat: "-1"},
		{name: "e5m2 positive infinity", format: "e5m2", bits: 0x7c, mag: 0x7c, class: fp.Inf},
		{name: "e5m2 negative infinity", format: "e5m2", bits: 0xfc, neg: true, mag: 0x7c, class: fp.Inf},
		{name: "e5m2 positive NaN", format: "e5m2", bits: 0x7d, mag: 0x7d, class: fp.NaN},
		{name: "e5m2 negative NaN", format: "e5m2", bits: 0xfd, neg: true, mag: 0x7d, class: fp.NaN},

		{name: "e4m3 positive zero", format: "e4m3", bits: 0x00, mag: 0x00, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "e4m3 negative zero", format: "e4m3", bits: 0x80, neg: true, mag: 0x00, class: fp.Zero, magRat: "0", signedRat: "0"},
		{name: "e4m3 positive subnormal", format: "e4m3", bits: 0x01, mag: 0x01, class: fp.Subnormal, magRat: "1/512", signedRat: "1/512"},
		{name: "e4m3 negative subnormal", format: "e4m3", bits: 0x81, neg: true, mag: 0x01, class: fp.Subnormal, magRat: "1/512", signedRat: "-1/512"},
		{name: "e4m3 positive normal", format: "e4m3", bits: 0x38, mag: 0x38, class: fp.Normal, magRat: "1", signedRat: "1"},
		{name: "e4m3 negative normal", format: "e4m3", bits: 0xb8, neg: true, mag: 0x38, class: fp.Normal, magRat: "1", signedRat: "-1"},
		{name: "e4m3 positive NaN", format: "e4m3", bits: 0x7f, mag: 0x7f, class: fp.NaN},
		{name: "e4m3 negative NaN", format: "e4m3", bits: 0xff, neg: true, mag: 0x7f, class: fp.NaN},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format := mustFormat(t, test.format)
			got := fp.Decode(format, test.bits)
			if got.Fmt != format {
				t.Errorf("Fmt = %#v, want %#v", got.Fmt, format)
			}
			if got.Bits != test.bits {
				t.Errorf("Bits = %#x, want %#x", got.Bits, test.bits)
			}
			if got.Neg != test.neg {
				t.Errorf("Neg = %t, want %t", got.Neg, test.neg)
			}
			if got.Mag != test.mag {
				t.Errorf("Mag = %#x, want %#x", got.Mag, test.mag)
			}
			if got.Class != test.class {
				t.Errorf("Class = %v, want %v", got.Class, test.class)
			}
			if got.Class.String() != test.class.String() {
				t.Errorf("Class.String() = %q, want %q", got.Class.String(), test.class.String())
			}
			wantFinite := test.class == fp.Zero || test.class == fp.Subnormal || test.class == fp.Normal
			if got.Finite() != wantFinite {
				t.Errorf("Finite() = %t, want %t", got.Finite(), wantFinite)
			}
			if wantFinite {
				if gotRat := fp.MagRat(format, test.mag).RatString(); gotRat != test.magRat {
					t.Errorf("MagRat() = %s, want %s", gotRat, test.magRat)
				}
				if signed := got.Signed(); signed == nil || signed.RatString() != test.signedRat {
					t.Errorf("Signed() = %v, want %s", signed, test.signedRat)
				}
			} else if signed := got.Signed(); signed != nil {
				t.Errorf("Signed() = %s, want nil for %s", signed.RatString(), test.class)
			}
		})
	}
}

func TestFiniteMagnitudeBoundaries(t *testing.T) {
	tests := []struct {
		name   string
		format string
		mag    uint64
		rat    string
	}{
		{name: "f16 zero", format: "f16", mag: 0x0000, rat: "0"},
		{name: "f16 minimum subnormal", format: "f16", mag: 0x0001, rat: "1/16777216"},
		{name: "f16 maximum subnormal", format: "f16", mag: 0x03ff, rat: "1023/16777216"},
		{name: "f16 minimum normal", format: "f16", mag: 0x0400, rat: "1/16384"},
		{name: "f16 maximum finite", format: "f16", mag: 0x7bff, rat: "65504"},
		{name: "bf16 zero", format: "bf16", mag: 0x0000, rat: "0"},
		{name: "bf16 minimum subnormal", format: "bf16", mag: 0x0001, rat: "1/10889035741470030830827987437816582766592"},
		{name: "bf16 maximum subnormal", format: "bf16", mag: 0x007f, rat: "127/10889035741470030830827987437816582766592"},
		{name: "bf16 minimum normal", format: "bf16", mag: 0x0080, rat: "1/85070591730234615865843651857942052864"},
		{name: "bf16 maximum finite", format: "bf16", mag: 0x7f7f, rat: "338953138925153547590470800371487866880"},
		{name: "f32 zero", format: "f32", mag: 0x00000000, rat: "0"},
		{name: "f32 minimum subnormal", format: "f32", mag: 0x00000001, rat: "1/713623846352979940529142984724747568191373312"},
		{name: "f32 maximum subnormal", format: "f32", mag: 0x007fffff, rat: "8388607/713623846352979940529142984724747568191373312"},
		{name: "f32 minimum normal", format: "f32", mag: 0x00800000, rat: "1/85070591730234615865843651857942052864"},
		{name: "f32 maximum finite", format: "f32", mag: 0x7f7fffff, rat: "340282346638528859811704183484516925440"},
		{name: "f64 zero", format: "f64", mag: 0x0000000000000000, rat: "0"},
		{name: "f64 minimum subnormal", format: "f64", mag: 0x0000000000000001, rat: "1/202402253307310618352495346718917307049556649764142118356901358027430339567995346891960383701437124495187077864316811911389808737385793476867013399940738509921517424276566361364466907742093216341239767678472745068562007483424692698618103355649159556340810056512358769552333414615230502532186327508646006263307707741093494784"},
		{name: "f64 maximum subnormal", format: "f64", mag: 0x000fffffffffffff, rat: "4503599627370495/202402253307310618352495346718917307049556649764142118356901358027430339567995346891960383701437124495187077864316811911389808737385793476867013399940738509921517424276566361364466907742093216341239767678472745068562007483424692698618103355649159556340810056512358769552333414615230502532186327508646006263307707741093494784"},
		{name: "f64 minimum normal", format: "f64", mag: 0x0010000000000000, rat: "1/44942328371557897693232629769725618340449424473557664318357520289433168951375240783177119330601884005280028469967848339414697442203604155623211857659868531094441973356216371319075554900311523529863270738021251442209537670585615720368478277635206809290837627671146574559986811484619929076208839082406056034304"},
		{name: "f64 maximum finite", format: "f64", mag: 0x7fefffffffffffff, rat: "179769313486231570814527423731704356798070567525844996598917476803157260780028538760589558632766878171540458953514382464234321326889464182768467546703537516986049910576551282076245490090389328944075868508455133942304583236903222948165808559332123348274797826204144723168738177180919299881250404026184124858368"},
		{name: "e5m2 zero", format: "e5m2", mag: 0x00, rat: "0"},
		{name: "e5m2 minimum subnormal", format: "e5m2", mag: 0x01, rat: "1/65536"},
		{name: "e5m2 maximum subnormal", format: "e5m2", mag: 0x03, rat: "3/65536"},
		{name: "e5m2 minimum normal", format: "e5m2", mag: 0x04, rat: "1/16384"},
		{name: "e5m2 maximum finite", format: "e5m2", mag: 0x7b, rat: "57344"},
		{name: "e4m3 zero", format: "e4m3", mag: 0x00, rat: "0"},
		{name: "e4m3 minimum subnormal", format: "e4m3", mag: 0x01, rat: "1/512"},
		{name: "e4m3 maximum subnormal", format: "e4m3", mag: 0x07, rat: "7/512"},
		{name: "e4m3 minimum normal", format: "e4m3", mag: 0x08, rat: "1/64"},
		{name: "e4m3 maximum finite", format: "e4m3", mag: 0x7e, rat: "448"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format := mustFormat(t, test.format)
			want, ok := new(big.Rat).SetString(test.rat)
			if !ok {
				t.Fatalf("invalid literal rational %q", test.rat)
			}
			got := fp.MagRat(format, test.mag)
			if got.Cmp(want) != 0 {
				t.Errorf("MagRat(%s, %#x) = %s, want %s", test.format, test.mag, got.RatString(), test.rat)
			}
			value := fp.Decode(format, test.mag)
			if signed := value.Signed(); signed == nil || signed.Cmp(want) != 0 {
				t.Errorf("Signed() = %v, want %s", signed, test.rat)
			}
		})
	}
}

func TestClassStrings(t *testing.T) {
	tests := []struct {
		class fp.Class
		want  string
	}{
		{class: fp.Zero, want: "zero"},
		{class: fp.Subnormal, want: "subnormal"},
		{class: fp.Normal, want: "normal"},
		{class: fp.Inf, want: "inf"},
		{class: fp.NaN, want: "nan"},
		{class: fp.Class(99), want: ""},
	}
	for _, test := range tests {
		if got := test.class.String(); got != test.want {
			t.Errorf("Class(%d).String() = %q, want %q", test.class, got, test.want)
		}
	}
}

func TestParseHexAcceptsPrefixesAndCase(t *testing.T) {
	tests := []struct {
		name   string
		format string
		text   string
		want   uint64
	}{
		{name: "f16 lowercase digits", format: "f16", text: "3c00", want: 0x3c00},
		{name: "f16 lowercase prefix and uppercase digits", format: "f16", text: "0x3C00", want: 0x3c00},
		{name: "f16 uppercase prefix and mixed digits", format: "f16", text: "0X3c00", want: 0x3c00},
		{name: "bf16", format: "bf16", text: "3F80", want: 0x3f80},
		{name: "f32", format: "f32", text: "3f800000", want: 0x3f800000},
		{name: "f64 uppercase prefix", format: "f64", text: "0X3FF0000000000000", want: 0x3ff0000000000000},
		{name: "e5m2", format: "e5m2", text: "3C", want: 0x3c},
		{name: "e4m3 lowercase prefix", format: "e4m3", text: "0x38", want: 0x38},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := fp.ParseHex(mustFormat(t, test.format), test.text)
			if err != nil {
				t.Fatalf("ParseHex(%q) returned error: %v", test.text, err)
			}
			if got != test.want {
				t.Errorf("ParseHex(%q) = %#x, want %#x", test.text, got, test.want)
			}
		})
	}
}

func TestParseHexFailuresAreExactAndSanitized(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "empty", text: "", want: "invalid hex word \"\" for f16 (expected 4 hex digits)"},
		{name: "wrong length", text: "3c0", want: "invalid hex word \"3c0\" for f16 (expected 4 hex digits)"},
		{name: "signed", text: "-3c00", want: "invalid hex word \"-3c00\" for f16 (expected 4 hex digits)"},
		{name: "leading whitespace", text: " 3c00", want: "invalid hex word \"?3c00\" for f16 (expected 4 hex digits)"},
		{name: "control character", text: "3c00\n", want: "invalid hex word \"3c00?\" for f16 (expected 4 hex digits)"},
		{name: "Unicode lookalike", text: "３c00", want: "invalid hex word \"?c00\" for f16 (expected 4 hex digits)"},
		{name: "nonhex digit", text: "3c0g", want: "invalid hex word \"3c0g\" for f16 (expected 4 hex digits)"},
		{name: "short prefixed input", text: "0x3c0", want: "invalid hex word \"0x3c0\" for f16 (expected 4 hex digits)"},
		{name: "handwritten overlong input", text: "3c00000000000000", want: "invalid hex word \"3c00000000000000\" for f16 (expected 4 hex digits)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := fp.ParseHex(mustFormat(t, "f16"), test.text)
			if err == nil {
				t.Fatalf("ParseHex(%q) returned nil error, want %q", test.text, test.want)
			}
			var failure core.Failure
			if !errors.As(err, &failure) {
				t.Fatalf("ParseHex(%q) error has type %T, want core.Failure", test.text, err)
			}
			if got := failure.Message; got != test.want {
				t.Errorf("ParseHex(%q) error = %q, want %q", test.text, got, test.want)
			}
		})
	}
}

func TestHexAndBinaryWidthsForEveryFormat(t *testing.T) {
	tests := []struct {
		format     string
		hexBits    uint64
		hex        string
		binaryBits uint64
		binary     string
	}{
		{format: "f16", hexBits: 0x0001, hex: "0x0001", binaryBits: 0x3c00, binary: "0 01111 0000000000"},
		{format: "bf16", hexBits: 0x0001, hex: "0x0001", binaryBits: 0x3f80, binary: "0 01111111 0000000"},
		{format: "f32", hexBits: 0x00000001, hex: "0x00000001", binaryBits: 0x3f800000, binary: "0 01111111 00000000000000000000000"},
		{format: "f64", hexBits: 0x0000000000000001, hex: "0x0000000000000001", binaryBits: 0x3ff0000000000000, binary: "0 01111111111 0000000000000000000000000000000000000000000000000000"},
		{format: "e5m2", hexBits: 0x01, hex: "0x01", binaryBits: 0x3c, binary: "0 01111 00"},
		{format: "e4m3", hexBits: 0x01, hex: "0x01", binaryBits: 0x38, binary: "0 0111 000"},
	}
	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			format := mustFormat(t, test.format)
			if got := fp.Hex(format, test.hexBits); got != test.hex {
				t.Errorf("Hex(%#x) = %q, want %q", test.hexBits, got, test.hex)
			}
			if got := fp.Binary(format, test.binaryBits); got != test.binary {
				t.Errorf("Binary(%#x) = %q, want %q", test.binaryBits, got, test.binary)
			}
		})
	}
}

func TestRationalResultsAreFreshAndDecodeIsRepeatable(t *testing.T) {
	format := mustFormat(t, "f16")

	firstMagnitude := fp.MagRat(format, 0x3c00)
	firstMagnitude.SetInt64(17)
	if got := fp.MagRat(format, 0x3c00).RatString(); got != "1" {
		t.Errorf("a later MagRat() = %s after mutating its predecessor, want 1", got)
	}

	firstSigned := fp.Decode(format, 0xbc00).Signed()
	firstSigned.SetInt64(17)
	if got := fp.Decode(format, 0xbc00).Signed().RatString(); got != "-1" {
		t.Errorf("a later Signed() = %s after mutating its predecessor, want -1", got)
	}
	if first := fp.Decode(format, 0xbc00); first != fp.Decode(format, 0xbc00) {
		t.Errorf("repeated Decode() values differ: %#v and %#v", first, fp.Decode(format, 0xbc00))
	}
}
