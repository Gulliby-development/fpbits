package core_test

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/core"
)

func TestFailureAndFailf(t *testing.T) {
	t.Run("Error returns the stored message", func(t *testing.T) {
		failure := core.Failure{Message: "oops"}
		if got := failure.Error(); got != "oops" {
			t.Fatalf("Error() = %q, want %q", got, "oops")
		}
	})

	t.Run("Failf formats a Failure", func(t *testing.T) {
		err := core.Failf("%s %d", "bad", 2)
		var failure core.Failure
		if !errors.As(err, &failure) {
			t.Fatalf("Failf returned %T, want core.Failure", err)
		}
		if got := failure.Message; got != "bad 2" {
			t.Fatalf("Failure.Message = %q, want %q", got, "bad 2")
		}
	})
}

func TestShow(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "\"\""},
		{name: "printable ASCII boundaries", in: "!~", want: "\"!~\""},
		{name: "printable text keeps spaces", in: "hello world", want: "\"hello?world\""},
		{name: "controls and Unicode become one question mark each", in: "a b\nλ", want: "\"a?b??\""},
		{name: "quotes and backslashes are not escaped", in: `a"b\c`, want: `"a"b\c"`},
		{name: "malformed byte", in: string([]byte{0xff}), want: "\"?\""},
		{name: "two malformed bytes", in: string([]byte{0xff, 0xfe}), want: "\"??\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := core.Show(test.in); got != test.want {
				t.Fatalf("Show(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

func TestOrList(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  string
	}{
		{name: "empty", items: []string{}, want: ""},
		{name: "one", items: []string{"A"}, want: "A"},
		{name: "two", items: []string{"A", "B"}, want: "A or B"},
		{name: "three", items: []string{"A", "B", "C"}, want: "A, B or C"},
		{name: "four", items: []string{"A", "B", "C", "D"}, want: "A, B, C or D"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := core.OrList(test.items); got != test.want {
				t.Fatalf("OrList(%q) = %q, want %q", test.items, got, test.want)
			}
		})
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		n    int
		one  string
		many string
		want string
	}{
		{n: 0, one: "hex digit", many: "hex digits", want: "0 hex digits"},
		{n: 1, one: "hex digit", many: "hex digits", want: "1 hex digit"},
		{n: 2, one: "hex digit", many: "hex digits", want: "2 hex digits"},
		{n: -1, one: "item", many: "items", want: "-1 items"},
	}
	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			if got := core.Plural(test.n, test.one, test.many); got != test.want {
				t.Fatalf("Plural(%d, %q, %q) = %q, want %q", test.n, test.one, test.many, got, test.want)
			}
		})
	}
}

func TestRenderTable(t *testing.T) {
	tests := []struct {
		name   string
		header []string
		rows   [][]string
		want   []string
	}{
		{
			name:   "rectangular example aligns header and cells",
			header: []string{"offset", "bits", "value"},
			rows:   [][]string{{"-2", "0x8002", "-1e-7"}, {"0", "0x8000", "-0e0"}},
			want:   []string{"offset  bits    value", "-2      0x8002  -1e-7", "0       0x8000  -0e0"},
		},
		{
			name:   "data widens columns",
			header: []string{"a", "b"},
			rows:   [][]string{{"long", "x"}, {"z", "longer"}},
			want:   []string{"a     b", "long  x", "z     longer"},
		},
		{
			name:   "empty cells keep column alignment",
			header: []string{"name", "value", "tail"},
			rows:   [][]string{{"", "7", ""}, {"x", "", "z"}},
			want:   []string{"name  value  tail", "      7", "x            z"},
		},
		{
			name:   "header only",
			header: []string{"a", "b"},
			rows:   [][]string{},
			want:   []string{"a  b"},
		},
		{
			name:   "empty table renders one empty header line",
			header: []string{},
			rows:   [][]string{},
			want:   []string{""},
		},
		{
			name:   "ragged rows are padded without panic",
			header: []string{"a"},
			rows:   [][]string{{"x", "y"}, {"z"}},
			want:   []string{"a", "x  y", "z"},
		},
		{
			name:   "empty row has no trailing spaces",
			header: []string{"a"},
			rows:   [][]string{{}},
			want:   []string{"a", ""},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := core.RenderTable(test.header, test.rows)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("RenderTable(%q, %q) = %#v, want %#v", test.header, test.rows, got, test.want)
			}
			for i, line := range got {
				if strings.HasSuffix(line, " ") {
					t.Errorf("line %d has trailing spaces: %q", i, line)
				}
			}
		})
	}
}

func TestParseOptions(t *testing.T) {
	count := core.Spec{Name: "--count", Lo: 1, Hi: 64, Fallback: 2}
	zero := core.Spec{Name: "--zero", Lo: 0, Hi: 5, Fallback: 3}

	tests := []struct {
		name      string
		argv      []string
		specs     []core.Spec
		wantArgs  []string
		wantValue map[string]int
		wantError string
	}{
		{name: "equals form and ordinary args", argv: []string{"before", "--count=3", "after"}, specs: []core.Spec{count}, wantArgs: []string{"before", "after"}, wantValue: map[string]int{"--count": 3}},
		{name: "separate value", argv: []string{"--count", "3", "word"}, specs: []core.Spec{count}, wantArgs: []string{"word"}, wantValue: map[string]int{"--count": 3}},
		{name: "fallback is inserted", argv: []string{"word"}, specs: []core.Spec{count, zero}, wantArgs: []string{"word"}, wantValue: map[string]int{"--count": 2, "--zero": 3}},
		{name: "last repeated value wins", argv: []string{"--count=4", "--count", "8"}, specs: []core.Spec{count}, wantArgs: []string{}, wantValue: map[string]int{"--count": 8}},
		{name: "dash arguments stay ordinary and ordered", argv: []string{"-", "-5", "x"}, specs: []core.Spec{count}, wantArgs: []string{"-", "-5", "x"}, wantValue: map[string]int{"--count": 2}},
		{name: "a dashed next argument is consumed as a value", argv: []string{"--count", "--help"}, specs: []core.Spec{count}, wantError: "invalid --count \"--help\" (expected an integer from 1 to 64)"},
		{name: "only the first equals is structural", argv: []string{"--count=3=4"}, specs: []core.Spec{count}, wantError: "invalid --count \"3=4\" (expected an integer from 1 to 64)"},
		{name: "unknown option reports whole argument", argv: []string{"--other=value"}, specs: []core.Spec{count}, wantError: "unknown option \"--other=value\""},
		{name: "bare double dash is unknown", argv: []string{"--"}, specs: []core.Spec{count}, wantError: "unknown option \"--\""},
		{name: "help is unknown when not specified", argv: []string{"--help"}, specs: []core.Spec{count}, wantError: "unknown option \"--help\""},
		{name: "missing value", argv: []string{"--count"}, specs: []core.Spec{count}, wantError: "--count needs a value"},
		{name: "unknown option with no specs", argv: []string{"--anything"}, specs: []core.Spec{}, wantError: "unknown option \"--anything\""},
		{name: "scan error takes precedence over invalid value", argv: []string{"--count=0", "--bad"}, specs: []core.Spec{count}, wantError: "unknown option \"--bad\""},
		{name: "invalid earlier value is ignored when overwritten", argv: []string{"--count=0", "--count=1"}, specs: []core.Spec{count}, wantArgs: []string{}, wantValue: map[string]int{"--count": 1}},
		{name: "validation follows spec order", argv: []string{"--zero=6", "--count=65"}, specs: []core.Spec{zero, count}, wantError: "invalid --zero \"6\" (expected an integer from 0 to 5)"},
		{name: "zero can be valid for an inclusive range", argv: []string{"--zero=0"}, specs: []core.Spec{zero}, wantArgs: []string{}, wantValue: map[string]int{"--zero": 0}},
		{name: "minimum count", argv: []string{"--count=1"}, specs: []core.Spec{count}, wantArgs: []string{}, wantValue: map[string]int{"--count": 1}},
		{name: "maximum count", argv: []string{"--count=64"}, specs: []core.Spec{count}, wantArgs: []string{}, wantValue: map[string]int{"--count": 64}},
		{name: "lower bound is enforced", argv: []string{"--count=0"}, specs: []core.Spec{count}, wantError: "invalid --count \"0\" (expected an integer from 1 to 64)"},
		{name: "upper bound is enforced", argv: []string{"--count=65"}, specs: []core.Spec{count}, wantError: "invalid --count \"65\" (expected an integer from 1 to 64)"},
		{name: "leading zero is rejected", argv: []string{"--count=01"}, specs: []core.Spec{count}, wantError: "invalid --count \"01\" (expected an integer from 1 to 64)"},
		{name: "empty value is rejected", argv: []string{"--count="}, specs: []core.Spec{count}, wantError: "invalid --count \"\" (expected an integer from 1 to 64)"},
		{name: "plus sign is rejected", argv: []string{"--count=+1"}, specs: []core.Spec{count}, wantError: "invalid --count \"+1\" (expected an integer from 1 to 64)"},
		{name: "minus sign is rejected", argv: []string{"--count=-1"}, specs: []core.Spec{count}, wantError: "invalid --count \"-1\" (expected an integer from 1 to 64)"},
		{name: "decimal notation is rejected", argv: []string{"--count=1.0"}, specs: []core.Spec{count}, wantError: "invalid --count \"1.0\" (expected an integer from 1 to 64)"},
		{name: "exponent notation is rejected", argv: []string{"--count=1e0"}, specs: []core.Spec{count}, wantError: "invalid --count \"1e0\" (expected an integer from 1 to 64)"},
		{name: "unicode digit is rejected", argv: []string{"--count=１"}, specs: []core.Spec{count}, wantError: "invalid --count \"?\" (expected an integer from 1 to 64)"},
		{name: "seven digit value is rejected", argv: []string{"--count=1000000"}, specs: []core.Spec{count}, wantError: "invalid --count \"1000000\" (expected an integer from 1 to 64)"},
		{name: "invalid value is sanitized", argv: []string{"--count=1\nλ"}, specs: []core.Spec{count}, wantError: "invalid --count \"1??\" (expected an integer from 1 to 64)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := core.ParseOptions(test.argv, test.specs)
			if test.wantError != "" {
				if err == nil || err.Error() != test.wantError {
					t.Fatalf("ParseOptions error = %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseOptions returned error: %v", err)
			}
			if !reflect.DeepEqual(got.Args, test.wantArgs) {
				t.Errorf("Args = %#v, want %#v", got.Args, test.wantArgs)
			}
			if !reflect.DeepEqual(got.Values, test.wantValue) {
				t.Errorf("Values = %#v, want %#v", got.Values, test.wantValue)
			}
		})
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("bad\nλ")
}

type panicReader struct{}

func (panicReader) Read([]byte) (int, error) {
	panic("Words unexpectedly read stdin")
}

func TestWords(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		args      []string
		wantWords []string
		wantError string
	}{
		{name: "BOM and CRLF are accepted", input: []byte("\ufeff3c00\r\n0x7BFF\n"), args: []string{"-"}, wantWords: []string{"3c00", "0x7BFF"}},
		{name: "only the first leading BOM is stripped", input: []byte("\ufeff\ufeffx"), args: []string{"-"}, wantWords: []string{"\ufeffx"}},
		{name: "BOM inside a word remains content", input: []byte("a\ufeffb"), args: []string{"-"}, wantWords: []string{"a\ufeffb"}},
		{name: "only named separators split", input: []byte("a\vb\fc\u00a0d"), args: []string{"-"}, wantWords: []string{"a\vb\fc\u00a0d"}},
		{name: "NUL remains word content", input: []byte("a\x00b c"), args: []string{"-"}, wantWords: []string{"a\x00b", "c"}},
		{name: "empty stdin has no input", input: []byte{}, args: []string{"-"}, wantError: "<stdin>: no input"},
		{name: "separator-only stdin has no input", input: []byte(" \t\r\n"), args: []string{"-"}, wantError: "<stdin>: no input"},
		{name: "BOM-only stdin has no input", input: []byte("\ufeff"), args: []string{"-"}, wantError: "<stdin>: no input"},
		{name: "malformed byte is rejected", input: []byte{0x33, 0x63, 0x30, 0x30, 0x20, 0xff, 0x0a}, args: []string{"-"}, wantError: "<stdin>: input is not valid UTF-8"},
		{name: "overlong UTF-8 is rejected", input: []byte{0xc0, 0xaf}, args: []string{"-"}, wantError: "<stdin>: input is not valid UTF-8"},
		{name: "truncated UTF-8 is rejected", input: []byte{0xe2, 0x82}, args: []string{"-"}, wantError: "<stdin>: input is not valid UTF-8"},
		{name: "surrogate UTF-8 is rejected", input: []byte{0xed, 0xa0, 0x80}, args: []string{"-"}, wantError: "<stdin>: input is not valid UTF-8"},
		{name: "out of range UTF-8 is rejected", input: []byte{0xf4, 0x90, 0x80, 0x80}, args: []string{"-"}, wantError: "<stdin>: input is not valid UTF-8"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := core.Words(bytes.NewReader(test.input), test.args)
			if test.wantError != "" {
				if err == nil || err.Error() != test.wantError {
					t.Fatalf("Words error = %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("Words returned error: %v", err)
			}
			if !reflect.DeepEqual(got, test.wantWords) {
				t.Fatalf("Words = %#v, want %#v", got, test.wantWords)
			}
		})
	}

	t.Run("reader error is sanitized", func(t *testing.T) {
		_, err := core.Words(failingReader{}, []string{"-"})
		if err == nil || err.Error() != "<stdin>: \"bad??\"" {
			t.Fatalf("Words error = %v, want %q", err, "<stdin>: \"bad??\"")
		}
	})

	t.Run("non-stdin args pass through without reading", func(t *testing.T) {
		for _, args := range [][]string{{}, {"word"}, {"-", "word"}} {
			got, err := core.Words(panicReader{}, args)
			if err != nil {
				t.Fatalf("Words(%q) returned error: %v", args, err)
			}
			if !reflect.DeepEqual(got, args) {
				t.Fatalf("Words(%q) = %#v, want %#v", args, got, args)
			}
		}
	})
}

func TestFormatsAndLookup(t *testing.T) {
	tests := []struct {
		name      string
		expBits   int
		fracBits  int
		bias      int
		ieee      bool
		totalBits int
		hexDigits int
		first     uint64
		max       uint64
	}{
		{name: "f16", expBits: 5, fracBits: 10, bias: 15, ieee: true, totalBits: 16, hexDigits: 4, first: 0x7c00, max: 0x7bff},
		{name: "bf16", expBits: 8, fracBits: 7, bias: 127, ieee: true, totalBits: 16, hexDigits: 4, first: 0x7f80, max: 0x7f7f},
		{name: "f32", expBits: 8, fracBits: 23, bias: 127, ieee: true, totalBits: 32, hexDigits: 8, first: 0x7f800000, max: 0x7f7fffff},
		{name: "f64", expBits: 11, fracBits: 52, bias: 1023, ieee: true, totalBits: 64, hexDigits: 16, first: 0x7ff0000000000000, max: 0x7fefffffffffffff},
		{name: "e5m2", expBits: 5, fracBits: 2, bias: 15, ieee: true, totalBits: 8, hexDigits: 2, first: 0x7c, max: 0x7b},
		{name: "e4m3", expBits: 4, fracBits: 3, bias: 7, ieee: false, totalBits: 8, hexDigits: 2, first: 0x7f, max: 0x7e},
	}
	if len(core.Formats) != len(tests) {
		t.Fatalf("Formats has %d entries, want %d", len(core.Formats), len(tests))
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format := core.Formats[i]
			if format.Name != test.name || format.ExpBits != test.expBits || format.FracBits != test.fracBits || format.Bias != test.bias || format.IEEE != test.ieee {
				t.Fatalf("Formats[%d] = %#v, want name=%s exp=%d frac=%d bias=%d IEEE=%t", i, format, test.name, test.expBits, test.fracBits, test.bias, test.ieee)
			}
			if got := format.TotalBits(); got != test.totalBits {
				t.Errorf("TotalBits() = %d, want %d", got, test.totalBits)
			}
			if got := format.HexDigits(); got != test.hexDigits {
				t.Errorf("HexDigits() = %d, want %d", got, test.hexDigits)
			}
			if got := format.FirstNonFinite(); got != test.first {
				t.Errorf("FirstNonFinite() = %#x, want %#x", got, test.first)
			}
			if got := format.MaxFinite(); got != test.max {
				t.Errorf("MaxFinite() = %#x, want %#x", got, test.max)
			}
			lookedUp, err := core.Lookup(test.name)
			if err != nil {
				t.Fatalf("Lookup(%q) returned error: %v", test.name, err)
			}
			if lookedUp != format {
				t.Errorf("Lookup(%q) = %#v, want %#v", test.name, lookedUp, format)
			}
		})
	}
}

func TestLookupUnknownFormat(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "unknown name", want: "unknown format \"f17\" (expected f16, bf16, f32, f64, e5m2 or e4m3)"},
		{name: "case alias", want: "unknown format \"F16\" (expected f16, bf16, f32, f64, e5m2 or e4m3)"},
		{name: "empty name", want: "unknown format \"\" (expected f16, bf16, f32, f64, e5m2 or e4m3)"},
		{name: "control and Unicode are sanitized", want: "unknown format \"bad??\" (expected f16, bf16, f32, f64, e5m2 or e4m3)"},
		{name: "option-looking name", want: "unknown format \"--help\" (expected f16, bf16, f32, f64, e5m2 or e4m3)"},
	}
	inputs := []string{"f17", "F16", "", "bad\nλ", "--help"}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := core.Lookup(inputs[i])
			if err == nil || err.Error() != test.want {
				t.Fatalf("Lookup(%q) error = %v, want %q", inputs[i], err, test.want)
			}
		})
	}
}
