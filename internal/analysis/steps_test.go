package analysis_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Gulliby-development/fpbits/internal/analysis"
	"github.com/Gulliby-development/fpbits/internal/core"
	"github.com/Gulliby-development/fpbits/internal/fp"
)

func mustStepsFormat(t *testing.T, name string) core.Format {
	t.Helper()
	format, err := core.Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%q) returned error: %v", name, err)
	}
	return format
}

func TestStepsCompleteLiteralTables(t *testing.T) {
	tests := []struct {
		name   string
		format string
		bits   uint64
		count  int
		want   []string
	}{
		{
			name:   "f16 negative-zero start preserves its encoding",
			format: "f16",
			bits:   0x8000,
			count:  2,
			want: []string{
				"offset  bits    value",
				"-2      0x8002  -1e-7",
				"-1      0x8001  -6e-8",
				"0       0x8000  -0e0",
				"+1      0x0001  6e-8",
				"+2      0x0002  1e-7",
			},
		},
		{
			name:   "bf16 signed-zero crossing",
			format: "bf16",
			bits:   0x0000,
			count:  1,
			want: []string{
				"offset  bits    value",
				"-1      0x8001  -9e-41",
				"0       0x0000  0e0",
				"+1      0x0001  9e-41",
			},
		},
		{
			name:   "f32 signed-zero crossing",
			format: "f32",
			bits:   0x00000000,
			count:  1,
			want: []string{
				"offset  bits        value",
				"-1      0x80000001  -1e-45",
				"0       0x00000000  0e0",
				"+1      0x00000001  1e-45",
			},
		},
		{
			name:   "e5m2 signed-zero crossing",
			format: "e5m2",
			bits:   0x00,
			count:  1,
			want: []string{
				"offset  bits  value",
				"-1      0x81  -2e-5",
				"0       0x00  0e0",
				"+1      0x01  2e-5",
			},
		},
		{
			name:   "positive subnormal steps through positive zero to negative values",
			format: "f16",
			bits:   0x0001,
			count:  2,
			want: []string{
				"offset  bits    value",
				"-2      0x8001  -6e-8",
				"-1      0x0000  0e0",
				"0       0x0001  6e-8",
				"+1      0x0002  1e-7",
				"+2      0x0003  2e-7",
			},
		},
		{
			name:   "positive largest finite steps to positive infinity",
			format: "f16",
			bits:   0x7bff,
			count:  1,
			want: []string{
				"offset  bits    value",
				"-1      0x7bfe  6.547e4",
				"0       0x7bff  6.55e4",
				"+1      0x7c00  inf",
			},
		},
		{
			name:   "negative largest finite steps from negative infinity",
			format: "f16",
			bits:   0xfbff,
			count:  1,
			want: []string{
				"offset  bits    value",
				"-1      0xfc00  -inf",
				"0       0xfbff  -6.55e4",
				"+1      0xfbfe  -6.547e4",
			},
		},
		{
			name:   "e4m3 positive maximum clips above the finite endpoint",
			format: "e4m3",
			bits:   0x7e,
			count:  3,
			want: []string{
				"offset  bits  value",
				"-3      0x7b  3.5e2",
				"-2      0x7c  4e2",
				"-1      0x7d  4.2e2",
				"0       0x7e  4.5e2",
			},
		},
		{
			name:   "e4m3 negative maximum clips below the finite endpoint",
			format: "e4m3",
			bits:   0xfe,
			count:  3,
			want: []string{
				"offset  bits  value",
				"0       0xfe  -4.5e2",
				"+1      0xfd  -4.2e2",
				"+2      0xfc  -4e2",
				"+3      0xfb  -3.5e2",
			},
		},
		{
			name:   "f16 negative infinity clips below and steps toward finite values",
			format: "f16",
			bits:   0xfc00,
			count:  1,
			want: []string{
				"offset  bits    value",
				"0       0xfc00  -inf",
				"+1      0xfbff  -6.55e4",
			},
		},
		{
			name:   "f16 positive infinity clips above and includes largest finite",
			format: "f16",
			bits:   0x7c00,
			count:  1,
			want: []string{
				"offset  bits    value",
				"-1      0x7bff  6.55e4",
				"0       0x7c00  inf",
			},
		},
		{
			name:   "f64 negative infinity endpoint",
			format: "f64",
			bits:   0xfff0000000000000,
			count:  1,
			want: []string{
				"offset  bits                value",
				"0       0xfff0000000000000  -inf",
				"+1      0xffefffffffffffff  -1.7976931348623157e308",
			},
		},
		{
			name:   "f64 positive infinity endpoint",
			format: "f64",
			bits:   0x7ff0000000000000,
			count:  1,
			want: []string{
				"offset  bits                value",
				"-1      0x7fefffffffffffff  1.7976931348623157e308",
				"0       0x7ff0000000000000  inf",
			},
		},
		{
			name:   "positive-zero start uses the positive zero encoding",
			format: "f16",
			bits:   0x0000,
			count:  2,
			want: []string{
				"offset  bits    value",
				"-2      0x8002  -1e-7",
				"-1      0x8001  -6e-8",
				"0       0x0000  0e0",
				"+1      0x0001  6e-8",
				"+2      0x0002  1e-7",
			},
		},
		{
			name:   "crossing zero from a negative minimum subnormal",
			format: "f16",
			bits:   0x8001,
			count:  2,
			want: []string{
				"offset  bits    value",
				"-2      0x8003  -2e-7",
				"-1      0x8002  -1e-7",
				"0       0x8001  -6e-8",
				"+1      0x0000  0e0",
				"+2      0x0001  6e-8",
			},
		},
		{
			name:   "count one crosses a spacing boundary around one",
			format: "f16",
			bits:   0x3c00,
			count:  1,
			want: []string{
				"offset  bits    value",
				"-1      0x3bff  9.995e-1",
				"0       0x3c00  1e0",
				"+1      0x3c01  1.001e0",
			},
		},
		{
			name:   "count sixty-four returns a complete table ending at e4m3 maximum",
			format: "e4m3",
			bits:   0x7e,
			count:  64,
			want: []string{
				"offset  bits  value",
				"-64     0x3e  1.8e0",
				"-63     0x3f  1.9e0",
				"-62     0x40  2e0",
				"-61     0x41  2.2e0",
				"-60     0x42  2.5e0",
				"-59     0x43  2.8e0",
				"-58     0x44  3e0",
				"-57     0x45  3.2e0",
				"-56     0x46  3.5e0",
				"-55     0x47  3.8e0",
				"-54     0x48  4e0",
				"-53     0x49  4.5e0",
				"-52     0x4a  5e0",
				"-51     0x4b  5.5e0",
				"-50     0x4c  6e0",
				"-49     0x4d  6.5e0",
				"-48     0x4e  7e0",
				"-47     0x4f  7.5e0",
				"-46     0x50  8e0",
				"-45     0x51  9e0",
				"-44     0x52  1e1",
				"-43     0x53  1.1e1",
				"-42     0x54  1.2e1",
				"-41     0x55  1.3e1",
				"-40     0x56  1.4e1",
				"-39     0x57  1.5e1",
				"-38     0x58  1.6e1",
				"-37     0x59  1.8e1",
				"-36     0x5a  2e1",
				"-35     0x5b  2.2e1",
				"-34     0x5c  2.4e1",
				"-33     0x5d  2.6e1",
				"-32     0x5e  2.8e1",
				"-31     0x5f  3e1",
				"-30     0x60  3.2e1",
				"-29     0x61  3.6e1",
				"-28     0x62  4e1",
				"-27     0x63  4.4e1",
				"-26     0x64  5e1",
				"-25     0x65  5.2e1",
				"-24     0x66  5.6e1",
				"-23     0x67  6e1",
				"-22     0x68  6.4e1",
				"-21     0x69  7e1",
				"-20     0x6a  8e1",
				"-19     0x6b  9e1",
				"-18     0x6c  1e2",
				"-17     0x6d  1.04e2",
				"-16     0x6e  1.1e2",
				"-15     0x6f  1.2e2",
				"-14     0x70  1.3e2",
				"-13     0x71  1.4e2",
				"-12     0x72  1.6e2",
				"-11     0x73  1.8e2",
				"-10     0x74  2e2",
				"-9      0x75  2.1e2",
				"-8      0x76  2.2e2",
				"-7      0x77  2.4e2",
				"-6      0x78  2.6e2",
				"-5      0x79  3e2",
				"-4      0x7a  3.2e2",
				"-3      0x7b  3.5e2",
				"-2      0x7c  4e2",
				"-1      0x7d  4.2e2",
				"0       0x7e  4.5e2",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			format := mustStepsFormat(t, test.format)
			value := fp.Decode(format, test.bits)
			before := value

			got, err := analysis.Steps(value, test.count)
			if err != nil {
				t.Fatalf("Steps(%s, %#x, %d) returned error: %v", test.format, test.bits, test.count, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Steps(%s, %#x, %d) = %#v, want %#v", test.format, test.bits, test.count, got, test.want)
			}
			for _, line := range got {
				if strings.Contains(line, "\n") {
					t.Errorf("Steps returned a line containing LF: %q", line)
				}
			}
			if value != before {
				t.Errorf("Steps mutated its input: got %#v, want %#v", value, before)
			}
		})
	}
}

func TestStepsRejectsEitherSignOfNaN(t *testing.T) {
	tests := []struct {
		name string
		bits uint64
	}{
		{name: "positive f16 NaN", bits: 0x7c01},
		{name: "negative f16 NaN", bits: 0xfc01},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := fp.Decode(mustStepsFormat(t, "f16"), test.bits)
			lines, err := analysis.Steps(value, 1)
			if err == nil {
				t.Fatalf("Steps(%#x, 1) returned lines %#v, want an error", test.bits, lines)
			}
			failure, ok := err.(core.Failure)
			if !ok {
				t.Fatalf("Steps(%#x, 1) error has type %T, want core.Failure", test.bits, err)
			}
			if failure.Message != "cannot step from nan" {
				t.Errorf("Steps(%#x, 1) error = %q, want %q", test.bits, failure.Message, "cannot step from nan")
			}
			if len(lines) != 0 {
				t.Errorf("Steps(%#x, 1) returned lines on failure: %#v", test.bits, lines)
			}
		})
	}
}
