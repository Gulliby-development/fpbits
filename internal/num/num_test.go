package num_test

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/num"
)

func TestParseFiniteDecimals(t *testing.T) {
	tests := []struct {
		name string
		text string
		neg  bool
		rat  string
	}{
		{name: "integer", text: "12", rat: "12"},
		{name: "explicit plus sign", text: "+12", rat: "12"},
		{name: "ordinary decimal", text: "+1.5", rat: "3/2"},
		{name: "leading decimal point", text: ".5", rat: "1/2"},
		{name: "trailing decimal point", text: "5.", rat: "5"},
		{name: "negative exponent", text: "1e-3", rat: "1/1000"},
		{name: "uppercase exponent and plus", text: "2E+2", rat: "200"},
		{name: "negative zero", text: "-0", neg: true, rat: "0"},
		{name: "negative zero with exponent", text: "-0.00e+5", neg: true, rat: "0"},
		{name: "negative decimal", text: "-12.50", neg: true, rat: "25/2"},
		{name: "leading zeroes", text: "00012.5", rat: "25/2"},
		{name: "fractional trailing zeroes", text: "0.1250", rat: "1/8"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, finiteOnly := range []bool{false, true} {
				got, err := num.Parse(test.text, finiteOnly)
				if err != nil {
					t.Fatalf("Parse(%q, %t) returned error: %v", test.text, finiteOnly, err)
				}
				if got.Kind != num.Finite {
					t.Errorf("Kind = %v, want Finite", got.Kind)
				}
				if got.Neg != test.neg {
					t.Errorf("Neg = %t, want %t", got.Neg, test.neg)
				}
				if got.Rat == nil {
					t.Fatal("finite Rat is nil")
				}
				if got.Rat.Sign() < 0 {
					t.Errorf("finite Rat is negative: %s", got.Rat.RatString())
				}
				if got.Rat.RatString() != test.rat {
					t.Errorf("Rat = %s, want %s", got.Rat.RatString(), test.rat)
				}
			}
		})
	}
}

func TestParseSpecialValuesAndFiniteOnly(t *testing.T) {
	tests := []struct {
		text        string
		kind        num.Kind
		neg         bool
		finiteError string
	}{
		{text: "inf", kind: num.Inf, finiteError: "invalid number \"inf\" (sum takes finite numbers)"},
		{text: "+inf", kind: num.Inf, finiteError: "invalid number \"+inf\" (sum takes finite numbers)"},
		{text: "-inf", kind: num.Inf, neg: true, finiteError: "invalid number \"-inf\" (sum takes finite numbers)"},
		{text: "nan", kind: num.NaN, finiteError: "invalid number \"nan\" (sum takes finite numbers)"},
		{text: "+nan", kind: num.NaN, finiteError: "invalid number \"+nan\" (sum takes finite numbers)"},
		{text: "-nan", kind: num.NaN, neg: true, finiteError: "invalid number \"-nan\" (sum takes finite numbers)"},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			got, err := num.Parse(test.text, false)
			if err != nil {
				t.Fatalf("Parse(%q, false) returned error: %v", test.text, err)
			}
			if got.Kind != test.kind || got.Neg != test.neg || got.Rat != nil {
				t.Errorf("Parse(%q, false) = %#v, want kind %v, Neg %t and nil Rat", test.text, got, test.kind, test.neg)
			}

			_, err = num.Parse(test.text, true)
			assertFailure(t, err, test.finiteError)
		})
	}
}

func TestParseRejectsMalformedText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "empty", text: "", want: "invalid number \"\""},
		{name: "lone plus", text: "+", want: "invalid number \"+\""},
		{name: "lone minus", text: "-", want: "invalid number \"-\""},
		{name: "point only", text: ".", want: "invalid number \".\""},
		{name: "signed point only", text: "-.", want: "invalid number \"-.\""},
		{name: "repeated point", text: "1..0", want: "invalid number \"1..0\""},
		{name: "exponent without digits", text: "1e", want: "invalid number \"1e\""},
		{name: "exponent sign without digits", text: "1e+", want: "invalid number \"1e+\""},
		{name: "exponent minus without digits", text: "1E-", want: "invalid number \"1E-\""},
		{name: "point after exponent", text: "1e2.0", want: "invalid number \"1e2.0\""},
		{name: "trailing text", text: "1x", want: "invalid number \"1x\""},
		{name: "leading whitespace", text: " 1", want: "invalid number \"?1\""},
		{name: "trailing whitespace", text: "1 ", want: "invalid number \"1?\""},
		{name: "line end", text: "1\n", want: "invalid number \"1?\""},
		{name: "uppercase infinity", text: "INF", want: "invalid number \"INF\""},
		{name: "uppercase nan", text: "NaN", want: "invalid number \"NaN\""},
		{name: "hexadecimal", text: "0x10", want: "invalid number \"0x10\""},
		{name: "unicode digit", text: "１", want: "invalid number \"?\""},
		{name: "malformed UTF-8", text: string([]byte{0xff}), want: "invalid number \"?\""},
		{name: "malformed over-range syntax takes precedence", text: "1e5001x", want: "invalid number \"1e5001x\""},
		{name: "over-range exponent missing digits is syntax error", text: "1e999999999999999999999999999999999999+", want: "invalid number \"1e999999999999999999999999999999999999+\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, finiteOnly := range []bool{false, true} {
				_, err := num.Parse(test.text, finiteOnly)
				assertFailure(t, err, test.want)
			}
		})
	}
}

const (
	zero10  = "0000000000"
	zero40  = zero10 + zero10 + zero10 + zero10
	zero390 = zero40 + zero40 + zero40 + zero40 + zero40 + zero40 + zero40 + zero40 + zero40 +
		zero10 + zero10 + zero10
	zero399 = zero390 + "000000000"
	zero400 = zero390 + zero10

	fourHundredLeadingDigits     = zero399 + "1"
	fourHundredOneLeadingDigits  = zero400 + "1"
	fourHundredFractionDigits    = "0." + zero390 + "00000000" + "1"
	fourHundredOneFractionDigits = "0." + zero390 + "000000000" + "1"
)

func TestParseDigitCountBoundaries(t *testing.T) {
	if len(fourHundredLeadingDigits) != 400 {
		t.Fatalf("400-digit leading-zero fixture has %d digits", len(fourHundredLeadingDigits))
	}
	if len(fourHundredOneLeadingDigits) != 401 {
		t.Fatalf("401-digit leading-zero fixture has %d digits", len(fourHundredOneLeadingDigits))
	}
	if len(strings.ReplaceAll(fourHundredFractionDigits, ".", "")) != 400 {
		t.Fatalf("400-digit fractional fixture has the wrong digit count")
	}
	if len(strings.ReplaceAll(fourHundredOneFractionDigits, ".", "")) != 401 {
		t.Fatalf("401-digit fractional fixture has the wrong digit count")
	}

	for _, test := range []struct {
		name string
		text string
		want string
	}{
		{name: "400 digits with leading zeroes", text: fourHundredLeadingDigits, want: "1e0"},
		{name: "400 digits across decimal point", text: fourHundredFractionDigits, want: "1e-399"},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, finiteOnly := range []bool{false, true} {
				got, err := num.Parse(test.text, finiteOnly)
				if err != nil {
					t.Fatalf("Parse returned error at the 400-digit boundary (finiteOnly=%t): %v", finiteOnly, err)
				}
				if got.Kind != num.Finite || got.Rat == nil {
					t.Fatalf("Parse returned %#v, want a finite value with Rat", got)
				}
				if rendered := num.Exact(got.Neg, got.Rat); rendered != test.want {
					t.Errorf("Exact(Parse(%q)) = %q, want %q", test.name, rendered, test.want)
				}
			}
		})
	}

	for _, text := range []string{fourHundredOneLeadingDigits, fourHundredOneFractionDigits} {
		for _, finiteOnly := range []bool{false, true} {
			_, err := num.Parse(text, finiteOnly)
			var failure core.Failure
			if !errors.As(err, &failure) {
				t.Errorf("Parse of a 401-digit significand returned %T, want core.Failure", err)
			}
			if err == nil || !strings.HasSuffix(err.Error(), "is out of range (at most 400 digits, exponent from -5000 to 5000)") {
				t.Errorf("Parse of a 401-digit significand error = %v, want the out-of-range diagnostic", err)
			}
		}
	}
}

func TestParseExponentBoundariesAndLongExponents(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "positive inclusive endpoint", text: "1e5000", want: "1e5000"},
		{name: "negative inclusive endpoint", text: "1e-5000", want: "1e-5000"},
		{name: "fractional adjusted scale below limit", text: "1.0e-5000", want: "1e-5000"},
		{name: "long leading exponent zeroes", text: "1e00000000000000000000000000000000000000000000000002", want: "1e2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, finiteOnly := range []bool{false, true} {
				got, err := num.Parse(test.text, finiteOnly)
				if err != nil {
					t.Fatalf("Parse(%q, %t) returned error: %v", test.text, finiteOnly, err)
				}
				if got.Rat == nil {
					t.Fatal("finite Rat is nil")
				}
				if rendered := num.Exact(got.Neg, got.Rat); rendered != test.want {
					t.Errorf("Exact(Parse(%q)) = %q, want %q", test.text, rendered, test.want)
				}
			}
		})
	}

	for _, test := range []struct {
		text string
		want string
	}{
		{text: "1e5001", want: "number \"1e5001\" is out of range (at most 400 digits, exponent from -5000 to 5000)"},
		{text: "1e-5001", want: "number \"1e-5001\" is out of range (at most 400 digits, exponent from -5000 to 5000)"},
		{text: "0e5001", want: "number \"0e5001\" is out of range (at most 400 digits, exponent from -5000 to 5000)"},
		{text: "1e999999999999999999999999999999999999999999999999999999999999999999", want: "number \"1e999999999999999999999999999999999999999999999999999999999999999999\" is out of range (at most 400 digits, exponent from -5000 to 5000)"},
		{text: "0e999999999999999999999999999999999999999999999999999999999999999999", want: "number \"0e999999999999999999999999999999999999999999999999999999999999999999\" is out of range (at most 400 digits, exponent from -5000 to 5000)"},
	} {
		for _, finiteOnly := range []bool{false, true} {
			_, err := num.Parse(test.text, finiteOnly)
			assertFailure(t, err, test.want)
		}
	}
}

func TestSci(t *testing.T) {
	tests := []struct {
		name   string
		neg    bool
		digits string
		exp10  int
		want   string
	}{
		{name: "one", digits: "1", want: "1e0"},
		{name: "fraction", digits: "15", exp10: -1, want: "1.5e-1"},
		{name: "large exponent", digits: "6550", exp10: 4, want: "6.55e4"},
		{name: "trailing zeroes do not change first digit exponent", digits: "100", exp10: 2, want: "1e2"},
		{name: "negative and trailing zeroes", neg: true, digits: "1250", exp10: -2, want: "-1.25e-2"},
		{name: "zero", digits: "0", exp10: 99, want: "0e0"},
		{name: "negative all-zero digits", neg: true, digits: "000", exp10: -99, want: "-0e0"},
		{name: "positive exponent boundary", digits: "1", exp10: 5000, want: "1e5000"},
		{name: "negative exponent boundary", digits: "1", exp10: -5000, want: "1e-5000"},
		{name: "beyond positive parser range", digits: "1", exp10: 5001, want: "1e5001"},
		{name: "beyond negative parser range", digits: "1", exp10: -5001, want: "1e-5001"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := num.Sci(test.neg, test.digits, test.exp10); got != test.want {
				t.Errorf("Sci(%t, %q, %d) = %q, want %q", test.neg, test.digits, test.exp10, got, test.want)
			}
		})
	}
}

func TestExact(t *testing.T) {
	tests := []struct {
		name string
		neg  bool
		text string
		want string
	}{
		{name: "positive zero", text: "0", want: "0e0"},
		{name: "negative zero", neg: true, text: "0", want: "-0e0"},
		{name: "one", text: "1", want: "1e0"},
		{name: "integer with trailing decimal zeroes", text: "100", want: "1e2"},
		{name: "mixed denominator", text: "3/20", want: "1.5e-1"},
		{name: "negative eighth", neg: true, text: "1/8", want: "-1.25e-1"},
		{name: "power of five denominator", text: "1/125", want: "8e-3"},
		{name: "power of two denominator", text: "65504", want: "6.5504e4"},
		{name: "smallest f16 subnormal", text: "1/1048576", want: "9.5367431640625e-7"},
		{name: "mixed powers of two and five", text: "7/40", want: "1.75e-1"},
		{name: "integer beyond parser exponent range", text: "1e5001", want: "1e5001"},
		{name: "fraction beyond parser exponent range", text: "1e-5001", want: "1e-5001"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, ok := new(big.Rat).SetString(test.text)
			if !ok {
				t.Fatalf("could not construct literal rational %q", test.text)
			}
			before := r.RatString()
			if got := num.Exact(test.neg, r); got != test.want {
				t.Errorf("Exact(%t, %s) = %q, want %q", test.neg, before, got, test.want)
			}
			if got := r.RatString(); got != before {
				t.Errorf("Exact mutated its rational operand: RatString() = %s, want %s", got, before)
			}
		})
	}
}

func assertFailure(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want core.Failure %q", want)
	}
	var failure core.Failure
	if !errors.As(err, &failure) {
		t.Fatalf("error type = %T, want core.Failure", err)
	}
	if failure.Message != want {
		t.Errorf("Failure.Message = %q, want %q", failure.Message, want)
	}
	if err.Error() != want {
		t.Errorf("error string = %q, want %q", err.Error(), want)
	}
}
