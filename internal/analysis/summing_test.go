package analysis_test

import (
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/analysis"
	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/num"
)

func mustSummingFormat(t *testing.T, name string) core.Format {
	t.Helper()
	format, err := core.Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%q) returned error: %v", name, err)
	}
	return format
}

func mustSummingNumbers(t *testing.T, texts []string) ([]num.Number, []*big.Rat) {
	t.Helper()
	numbers := make([]num.Number, 0, len(texts))
	originals := make([]*big.Rat, 0, len(texts))
	for _, text := range texts {
		number, err := num.Parse(text, true)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", text, err)
		}
		if number.Kind != num.Finite || number.Rat == nil {
			t.Fatalf("Parse(%q) = %#v, want finite number", text, number)
		}
		numbers = append(numbers, number)
		originals = append(originals, new(big.Rat).Set(number.Rat))
	}
	return numbers, originals
}

func TestSummingCompleteLiteralReports(t *testing.T) {
	tests := []struct {
		name   string
		format string
		texts  []string
		want   []string
	}{
		{
			name:   "f64 decimal inputs preserve exact and sequential totals",
			format: "f64",
			texts:  []string{"0.1", "0.2"},
			want: []string{
				"step  term  sum",
				"1     1e-1  1e-1",
				"2     2e-1  3.0000000000000004e-1",
				"exact: 3e-1",
				"result: 3.0000000000000004e-1 (0x3fd3333333333334)",
				"error: 4.44089209850062616169452667236328125e-17",
			},
		},
		{
			name:   "f16 sequential rounding loses three unit increments",
			format: "f16",
			texts:  []string{"2048", "1", "1", "1"},
			want: []string{
				"step  term     sum",
				"1     2.048e3  2.048e3",
				"2     1e0      2.048e3",
				"3     1e0      2.048e3",
				"4     1e0      2.048e3",
				"exact: 2.051e3",
				"result: 2.048e3 (0x6800)",
				"error: -3e0",
			},
		},
		{
			name:   "two negative zeros add to positive zero",
			format: "f16",
			texts:  []string{"-0", "-0"},
			want: []string{
				"step  term  sum",
				"1     -0e0  0e0",
				"2     -0e0  0e0",
				"exact: 0e0",
				"result: 0e0 (0x0000)",
				"error: 0e0",
			},
		},
		{
			name:   "exact single term",
			format: "f16",
			texts:  []string{"1"},
			want: []string{
				"step  term  sum",
				"1     1e0   1e0",
				"exact: 1e0",
				"result: 1e0 (0x3c00)",
				"error: 0e0",
			},
		},
		{
			name:   "exact finite addition",
			format: "f16",
			texts:  []string{"0.5", "0.5"},
			want: []string{
				"step  term  sum",
				"1     5e-1  5e-1",
				"2     5e-1  1e0",
				"exact: 1e0",
				"result: 1e0 (0x3c00)",
				"error: 0e0",
			},
		},
		{
			name:   "finite cancellation returns positive zero",
			format: "f16",
			texts:  []string{"1", "-1"},
			want: []string{
				"step  term  sum",
				"1     1e0   1e0",
				"2     -1e0  0e0",
				"exact: 0e0",
				"result: 0e0 (0x0000)",
				"error: 0e0",
			},
		},
		{
			name:   "negative input underflow keeps term sign",
			format: "f16",
			texts:  []string{"-1e-30"},
			want: []string{
				"step  term  sum",
				"1     -0e0  0e0",
				"exact: -1e-30",
				"result: 0e0 (0x0000)",
				"error: 1e-30",
			},
		},
		{
			name:   "positive IEEE term overflow still accumulates later exact input",
			format: "f16",
			texts:  []string{"1e10", "1"},
			want: []string{
				"step  term  sum",
				"1     inf   inf",
				"2     1e0   inf",
				"exact: 1.0000000001e10",
				"result: inf (0x7c00)",
			},
		},
		{
			name:   "negative IEEE term overflow still accumulates later exact input",
			format: "f16",
			texts:  []string{"-1e10", "-1"},
			want: []string{
				"step  term  sum",
				"1     -inf  -inf",
				"2     -1e0  -inf",
				"exact: -1.0000000001e10",
				"result: -inf (0xfc00)",
			},
		},
		{
			name:   "finite accumulator overflows on second maximum term",
			format: "f16",
			texts:  []string{"65504", "65504"},
			want: []string{
				"step  term    sum",
				"1     6.55e4  6.55e4",
				"2     6.55e4  inf",
				"exact: 1.31008e5",
				"result: inf (0x7c00)",
			},
		},
		{
			name:   "opposite overflow infinities yield NaN and later terms count exactly",
			format: "f16",
			texts:  []string{"1e10", "-1e10", "1"},
			want: []string{
				"step  term  sum",
				"1     inf   inf",
				"2     -inf  nan",
				"3     1e0   nan",
				"exact: 1e0",
				"result: nan (0x7e00)",
			},
		},
		{
			name:   "e4m3 overflow NaN propagates while later input counts exactly",
			format: "e4m3",
			texts:  []string{"464.01", "1"},
			want: []string{
				"step  term  sum",
				"1     nan   nan",
				"2     1e0   nan",
				"exact: 4.6501e2",
				"result: nan (0x7f)",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			numbers, originals := mustSummingNumbers(t, test.texts)
			got := analysis.Summing(mustSummingFormat(t, test.format), test.texts, numbers)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Summing(%s, %q) = %#v, want %#v", test.format, test.texts, got, test.want)
			}
			for _, line := range got {
				if strings.Contains(line, "\n") {
					t.Errorf("Summing(%s, %q) returned a line containing LF: %q", test.format, test.texts, line)
				}
			}
			for i, number := range numbers {
				if number.Rat.Cmp(originals[i]) != 0 {
					t.Errorf("Summing(%s, %q) mutated number %d: got %s, want %s", test.format, test.texts, i, number.Rat, originals[i])
				}
			}
		})
	}
}
