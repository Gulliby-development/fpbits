package core

// Format describes the binary layout and special encodings of a floating-point format.
type Format struct {
	Name                    string
	ExpBits, FracBits, Bias int
	IEEE                    bool
}

// Formats lists the supported formats in their canonical order.
var Formats = []Format{
	{Name: "f16", ExpBits: 5, FracBits: 10, Bias: 15, IEEE: true},
	{Name: "bf16", ExpBits: 8, FracBits: 7, Bias: 127, IEEE: true},
	{Name: "f32", ExpBits: 8, FracBits: 23, Bias: 127, IEEE: true},
	{Name: "f64", ExpBits: 11, FracBits: 52, Bias: 1023, IEEE: true},
	{Name: "e5m2", ExpBits: 5, FracBits: 2, Bias: 15, IEEE: true},
	{Name: "e4m3", ExpBits: 4, FracBits: 3, Bias: 7, IEEE: false},
}

// Lookup returns the format with the exact given name.
func Lookup(name string) (Format, error) {
	names := make([]string, len(Formats))
	for i, format := range Formats {
		names[i] = format.Name
		if name == format.Name {
			return format, nil
		}
	}
	return Format{}, Failf("unknown format %s (expected %s)", Show(name), OrList(names))
}

// TotalBits returns the number of bits in a complete pattern.
func (f Format) TotalBits() int {
	return 1 + f.ExpBits + f.FracBits
}

// HexDigits returns the number of hexadecimal digits needed for a complete pattern.
func (f Format) HexDigits() int {
	return (f.TotalBits() + 3) / 4
}

// FirstNonFinite returns the first magnitude pattern that is not finite.
func (f Format) FirstNonFinite() uint64 {
	exponent := (uint64(1) << uint(f.ExpBits)) - 1
	magnitude := exponent << uint(f.FracBits)
	if !f.IEEE {
		magnitude |= (uint64(1) << uint(f.FracBits)) - 1
	}
	return magnitude
}

// MaxFinite returns the largest finite magnitude pattern.
func (f Format) MaxFinite() uint64 {
	return f.FirstNonFinite() - 1
}
