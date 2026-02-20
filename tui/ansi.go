package tui

import "fmt"

const (
	esc   = "\x1b["
	reset = "\x1b[0m"
)

func cursorPos(row, col int) string    { return fmt.Sprintf("%s%d;%dH", esc, row, col) }
func clearLine() string                { return esc + "2K" }
func clearScreen() string              { return esc + "2J" }
func hideCursor() string               { return esc + "?25l" }
func showCursor() string               { return esc + "?25h" }
func resetScrollRegion() string        { return esc + "r" }

func fg(c Color) string {
	if c == 0 {
		return ""
	}
	r := (c >> 16) & 0xff
	g := (c >> 8) & 0xff
	b := c & 0xff
	return fmt.Sprintf("%s38;2;%d;%d;%dm", esc, r, g, b)
}

func bg(c Color) string {
	if c == 0 {
		return ""
	}
	r := (c >> 16) & 0xff
	g := (c >> 8) & 0xff
	b := c & 0xff
	return fmt.Sprintf("%s48;2;%d;%d;%dm", esc, r, g, b)
}

func bold() string    { return esc + "1m" }
func italic() string  { return esc + "3m" }
func reverse() string { return esc + "7m" }
