<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# fpbits

[![CI](https://github.com/Gulliby-development/fpbits/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/fpbits/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

fpbits is an offline command-line tool for inspecting floating-point values. It uses exact integer and rational arithmetic to show how bit patterns represent numbers, how decimal inputs are rounded, and how rounding affects calculations. It is intended for people debugging numeric code, machine-learning engineers working across low-precision formats, and teachers explaining floating-point behavior.

## Install

A Go toolchain is required. From the project directory, build the executable with `go build -o fpbits .`, or run a command directly with `go run . COMMAND`. On Windows, use the executable name produced by Go (typically `fpbits.exe`) when invoking the built program.

## Usage

The command-line usage text is:

```text
usage: fpbits decode FORMAT HEX...
       fpbits encode FORMAT NUMBER...
       fpbits convert FROM TO HEX...
       fpbits ulp FORMAT NUMBER
       fpbits next [--count N] FORMAT HEX
       fpbits range FORMAT
       fpbits sum FORMAT NUMBER...
```

Commands accept values as command-line arguments. For `decode`, `encode`, `convert`, and `sum`, a single `-` in place of the values reads whitespace-separated words from standard input. `next` also accepts a single `-` in place of HEX; standard input must then hold exactly one word. The input must be valid UTF-8; a leading byte-order mark is accepted.

Examples of commands include `fpbits decode f16 3c00`, `fpbits encode f16 0.1`, `fpbits convert f32 f16 3f800000`, `fpbits ulp f16 1`, `fpbits next --count 3 f16 3c00`, `fpbits range e4m3`, and `fpbits sum f16 2048 1 1`.

## Formats

The six supported formats and binary layouts are:

| Format | Exponent bits | Fraction bits | Bias | Special-value convention |
| --- | ---: | ---: | ---: | --- |
| `f16` | 5 | 10 | 15 | IEEE |
| `bf16` | 8 | 7 | 127 | IEEE |
| `f32` | 8 | 23 | 127 | IEEE |
| `f64` | 11 | 52 | 1023 | IEEE |
| `e5m2` | 5 | 2 | 15 | IEEE |
| `e4m3` | 4 | 3 | 7 | No infinity; the all-ones exponent and fraction encode NaN |

Every pattern consists of one sign bit, the format's exponent field, and its fraction field. The IEEE formats encode infinities and NaNs; `e4m3` reserves only its all-ones exponent-and-fraction pattern for NaN. For each format, `range` reports the limits and layout.

## Commands

- `decode FORMAT HEX...` describes each hexadecimal bit pattern, including its fields, class, shortest round-tripping value, and exact value when finite.
- `encode FORMAT NUMBER...` rounds each decimal input to the selected format and reports the rounded description and rounding error.
- `convert FROM TO HEX...` decodes each source pattern and re-encodes its value in the destination format, reporting the error when the source is finite.
- `ulp FORMAT NUMBER` rounds a number to the format and reports the spacing above that value and the gap toward zero. It requires a finite rounded value.
- `next [--count N] FORMAT HEX` lists nearby representable values around a bit pattern. The optional count is from 1 through 64 and defaults to 2; NaN values cannot be stepped from.
- `range FORMAT` reports a format's layout, smallest subnormal and normal values, largest finite value, epsilon, and counts of finite and special encodings.
- `sum FORMAT NUMBER...` rounds each finite decimal term, adds terms sequentially in the selected format, and compares the result with the exact sum of the inputs. It accepts at most 50 terms.

For example, running `fpbits range f16` prints:

```text
format: f16
layout: 1 sign bit, 5 exponent bits, 10 fraction bits
bias: 15
smallest subnormal: 6e-8 (0x0001)
smallest normal: 6.104e-5 (0x0400)
largest: 6.55e4 (0x7bff)
epsilon: 9.77e-4
infinities: 2
nan encodings: 2046
finite encodings: 63488
```

## Exit codes

The program exits with status 0 on success and status 2 on any command or input error. Errors are written to standard error; successful command output is written to standard output.

## Development

The project uses only the Go standard library. From the project directory, format and check the code with `gofmt -l .`, `go vet ./...`, and `go test ./...`.

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED
