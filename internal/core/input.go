package core

import (
	"io"
	"strings"
	"unicode/utf8"
)

// Words returns command arguments unchanged unless the sole argument is "-",
// in which case it decodes and splits standard input.
func Words(stdin io.Reader, args []string) ([]string, error) {
	if len(args) != 1 || args[0] != "-" {
		return args, nil
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return nil, Failf("<stdin>: %s", Show(err.Error()))
	}
	if !utf8.Valid(data) {
		return nil, Failf("<stdin>: input is not valid UTF-8")
	}

	text := strings.TrimPrefix(string(data), "\uFEFF")
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n'
	})
	if len(words) == 0 {
		return nil, Failf("<stdin>: no input")
	}
	return words, nil
}
