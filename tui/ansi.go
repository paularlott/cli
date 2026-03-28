package tui

import (
	"strconv"
	"strings"
)

const (
	esc   = "\x1b["
	reset = "\x1b[0m"
)

func cursorPos(row, col int) string {
	var b strings.Builder
	b.Grow(8)
	b.WriteString(esc)
	b.WriteString(strconv.Itoa(row))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(col))
	b.WriteByte('H')
	return b.String()
}

func clearLine() string         { return esc + "2K" }
func clearScreen() string       { return esc + "2J" }
func hideCursor() string        { return esc + "?25l" }
func showCursor() string        { return esc + "?25h" }
func resetScrollRegion() string { return esc + "r" }

func fg(c Color) string {
	if c == 0 {
		return ""
	}
	r := (c >> 16) & 0xff
	g := (c >> 8) & 0xff
	b := c & 0xff
	var buf strings.Builder
	buf.Grow(20)
	buf.WriteString(esc)
	buf.WriteString("38;2;")
	buf.WriteString(strconv.Itoa(int(r)))
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(int(g)))
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(int(b)))
	buf.WriteByte('m')
	return buf.String()
}

func bg(c Color) string {
	if c == 0 {
		return ""
	}
	r := (c >> 16) & 0xff
	g := (c >> 8) & 0xff
	b := c & 0xff
	var buf strings.Builder
	buf.Grow(20)
	buf.WriteString(esc)
	buf.WriteString("48;2;")
	buf.WriteString(strconv.Itoa(int(r)))
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(int(g)))
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(int(b)))
	buf.WriteByte('m')
	return buf.String()
}

func bold() string    { return esc + "1m" }
func italic() string  { return esc + "3m" }
func reverse() string { return esc + "7m" }
