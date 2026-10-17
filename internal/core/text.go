package core

import (
	"strconv"
	"strings"
)

// Show returns s quoted, replacing every non-printable ASCII rune with '?'.
func Show(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		if r < '!' || r > '~' {
			b.WriteByte('?')
			continue
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}

// OrList joins items with commas and a final "or".
func OrList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " or " + items[len(items)-1]
	}
}

// Plural formats n with one when n is 1 and many otherwise.
func Plural(n int, one, many string) string {
	unit := many
	if n == 1 {
		unit = one
	}
	return strconv.Itoa(n) + " " + unit
}

// RenderTable renders a left-aligned table with two spaces between columns.
func RenderTable(header []string, rows [][]string) []string {
	columnCount := len(header)
	for _, row := range rows {
		if len(row) > columnCount {
			columnCount = len(row)
		}
	}

	widths := make([]int, columnCount)
	for i, cell := range header {
		widths[i] = len(cell)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	lines := make([]string, 1, len(rows)+1)
	lines[0] = renderTableRow(header, widths)
	for _, row := range rows {
		lines = append(lines, renderTableRow(row, widths))
	}
	return lines
}

func renderTableRow(cells []string, widths []int) string {
	var b strings.Builder
	for i, width := range widths {
		if i > 0 {
			b.WriteString("  ")
		}
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteString(cell)
		if padding := width - len(cell); padding > 0 {
			b.WriteString(strings.Repeat(" ", padding))
		}
	}
	return strings.TrimRight(b.String(), " ")
}
