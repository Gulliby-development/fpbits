package core

import (
	"strconv"
	"strings"
)

// Spec describes an option and the inclusive range and fallback for its value.
type Spec struct {
	Name             string
	Lo, Hi, Fallback int
}

// Parsed contains the final option values and all non-option arguments.
type Parsed struct {
	Values map[string]int
	Args   []string
}

// ParseOptions scans argv from left to right and validates final option values
// in the order given by specs.
func ParseOptions(argv []string, specs []Spec) (Parsed, error) {
	values := make(map[string]string)
	known := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		known[spec.Name] = struct{}{}
	}

	parsed := Parsed{
		Values: make(map[string]int, len(specs)),
		Args:   make([]string, 0, len(argv)),
	}
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if !strings.HasPrefix(arg, "--") {
			parsed.Args = append(parsed.Args, arg)
			continue
		}

		name, value, hasValue := strings.Cut(arg, "=")
		if _, ok := known[name]; !ok {
			return Parsed{}, Failf("unknown option %s", Show(arg))
		}
		if !hasValue {
			if i+1 >= len(argv) {
				return Parsed{}, Failf("%s needs a value", name)
			}
			i++
			value = argv[i]
		}
		values[name] = value
	}

	for _, spec := range specs {
		value, supplied := values[spec.Name]
		if !supplied {
			parsed.Values[spec.Name] = spec.Fallback
			continue
		}
		n, ok := parseOptionInteger(value)
		if !ok || n < spec.Lo || n > spec.Hi {
			return Parsed{}, Failf(
				"invalid %s %s (expected an integer from %d to %d)",
				spec.Name, Show(value), spec.Lo, spec.Hi,
			)
		}
		parsed.Values[spec.Name] = n
	}
	return parsed, nil
}

func parseOptionInteger(value string) (int, bool) {
	if value == "0" {
		return 0, true
	}
	if len(value) == 0 || len(value) > 6 || value[0] < '1' || value[0] > '9' {
		return 0, false
	}
	for i := 1; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return n, true
}
