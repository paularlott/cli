package tui

import "strings"

type inputArea struct {
	lines   [][]rune
	row     int
	col     int
	viewOff int
	history []string
	hisIdx  int    // current position in history (-1 = not browsing)
	draft   string // saved current input while browsing
}

const inputMinHeight = 4

func newInputArea() *inputArea {
	return &inputArea{lines: [][]rune{{}}, hisIdx: -1}
}

func (a *inputArea) reset() {
	a.lines = [][]rune{{}}
	a.row = 0
	a.col = 0
	a.viewOff = 0
	a.hisIdx = -1
	a.draft = ""
}

// pushHistory saves a submitted entry to history.
func (a *inputArea) pushHistory(text string) {
	if text == "" {
		return
	}
	// Avoid consecutive duplicates.
	if len(a.history) > 0 && a.history[len(a.history)-1] == text {
		return
	}
	a.history = append(a.history, text)
}

// historyUp moves to the previous history entry when at the first input row.
// Returns true if the input was changed.
func (a *inputArea) historyUp() bool {
	if a.row != 0 || len(a.history) == 0 {
		return false
	}
	if a.hisIdx == -1 {
		a.draft = a.text()
		a.hisIdx = len(a.history) - 1
	} else if a.hisIdx > 0 {
		a.hisIdx--
	} else {
		return false
	}
	a.setLines(a.history[a.hisIdx])
	return true
}

// historyDown moves to the next history entry, or restores the draft.
// Returns true if the input was changed.
func (a *inputArea) historyDown() bool {
	if a.hisIdx == -1 {
		return false
	}
	if a.row != len(a.lines)-1 {
		return false
	}
	if a.hisIdx < len(a.history)-1 {
		a.hisIdx++
		a.setLines(a.history[a.hisIdx])
	} else {
		a.hisIdx = -1
		a.setLines(a.draft)
	}
	return true
}

func (a *inputArea) setLines(text string) {
	parts := strings.Split(text, "\n")
	a.lines = make([][]rune, len(parts))
	for i, p := range parts {
		a.lines[i] = []rune(p)
	}
	a.row = 0
	a.col = len(a.lines[0])
	a.viewOff = 0
}

func (a *inputArea) text() string {
	parts := make([]string, len(a.lines))
	for i, l := range a.lines {
		parts[i] = string(l)
	}
	return strings.Join(parts, "\n")
}

func (a *inputArea) insertRune(r rune) {
	line := a.lines[a.row]
	newLine := make([]rune, len(line)+1)
	copy(newLine, line[:a.col])
	newLine[a.col] = r
	copy(newLine[a.col+1:], line[a.col:])
	a.lines[a.row] = newLine
	a.col++
}

func (a *inputArea) insertNewline() {
	line := a.lines[a.row]
	before := make([]rune, a.col)
	copy(before, line[:a.col])
	after := make([]rune, len(line)-a.col)
	copy(after, line[a.col:])
	a.lines[a.row] = before
	rest := make([][]rune, len(a.lines)+1)
	copy(rest, a.lines[:a.row+1])
	rest[a.row+1] = after
	copy(rest[a.row+2:], a.lines[a.row+1:])
	a.lines = rest
	a.row++
	a.col = 0
}

func (a *inputArea) backspace() {
	if a.col > 0 {
		line := a.lines[a.row]
		a.lines[a.row] = append(line[:a.col-1], line[a.col:]...)
		a.col--
	} else if a.row > 0 {
		prev := a.lines[a.row-1]
		cur := a.lines[a.row]
		a.col = len(prev)
		a.lines[a.row-1] = append(prev, cur...)
		a.lines = append(a.lines[:a.row], a.lines[a.row+1:]...)
		a.row--
	}
}

func (a *inputArea) deleteForward() {
	line := a.lines[a.row]
	if a.col < len(line) {
		a.lines[a.row] = append(line[:a.col], line[a.col+1:]...)
	} else if a.row < len(a.lines)-1 {
		a.lines[a.row] = append(line, a.lines[a.row+1]...)
		a.lines = append(a.lines[:a.row+1], a.lines[a.row+2:]...)
	}
}

func (a *inputArea) moveLeft() {
	if a.col > 0 {
		a.col--
	} else if a.row > 0 {
		a.row--
		a.col = len(a.lines[a.row])
	}
}

func (a *inputArea) moveRight() {
	if a.col < len(a.lines[a.row]) {
		a.col++
	} else if a.row < len(a.lines)-1 {
		a.row++
		a.col = 0
	}
}

func (a *inputArea) moveUp() {
	if a.row > 0 {
		a.row--
		if a.col > len(a.lines[a.row]) {
			a.col = len(a.lines[a.row])
		}
	}
}

func (a *inputArea) moveDown() {
	if a.row < len(a.lines)-1 {
		a.row++
		if a.col > len(a.lines[a.row]) {
			a.col = len(a.lines[a.row])
		}
	}
}

func (a *inputArea) home() { a.col = 0 }
func (a *inputArea) end()  { a.col = len(a.lines[a.row]) }

// ctrlK clears from cursor to end of line.
func (a *inputArea) ctrlK() {
	a.lines[a.row] = a.lines[a.row][:a.col]
}

// ctrlU clears from start of line to cursor.
func (a *inputArea) ctrlU() {
	a.lines[a.row] = a.lines[a.row][a.col:]
	a.col = 0
}

// ctrlW deletes the word before the cursor.
func (a *inputArea) ctrlW() {
	line := a.lines[a.row]
	i := a.col
	for i > 0 && line[i-1] == ' ' {
		i--
	}
	for i > 0 && line[i-1] != ' ' {
		i--
	}
	a.lines[a.row] = append(line[:i], line[a.col:]...)
	a.col = i
}

// render draws the input box into buf using absolute cursor positioning.
// overlay is optional text embedded right-aligned into the top border (replaces ─ chars).
// statusLeft/statusRight are embedded into the bottom border; empty strings are hidden.
func (a *inputArea) render(buf *strings.Builder, t *Theme, w, maxHeight, startRow int, overlay, statusLeft, statusRight string) int {
	height := len(a.lines) + 4 // 2 borders + 2 padding rows
	if height < inputMinHeight {
		height = inputMinHeight
	}
	if height > maxHeight {
		height = maxHeight
	}
	innerH := height - 4 // content rows excluding padding

	// Keep cursor row in the visible window.
	if a.row < a.viewOff {
		a.viewOff = a.row
	}
	if a.row >= a.viewOff+innerH {
		a.viewOff = a.row - innerH + 1
	}

	contentW := w - 5
	if contentW < 1 {
		contentW = 1
	}

	// top border — embed overlay right-aligned if provided
	buf.WriteString(cursorPos(startRow, 1))
	buf.WriteString(clearLine())
	if overlay == "" {
		buf.WriteString(fg(t.Dim) + "┌" + strings.Repeat("─", w-2) + "┐" + reset)
	} else {
		ovl := " " + overlay + " "
		ovlLen := visibleLen(ovl)
		dashW := w - 2 - ovlLen
		if dashW < 0 {
			dashW = 0
		}
		buf.WriteString(fg(t.Dim) + "┌" + strings.Repeat("─", dashW) + reset + fg(t.Primary) + ovl + reset + fg(t.Dim) + "┐" + reset)
	}

	// blank padding row
	buf.WriteString(cursorPos(startRow+1, 1))
	buf.WriteString(clearLine())
	buf.WriteString(fg(t.Dim) + "│" + reset + strings.Repeat(" ", w-2) + fg(t.Dim) + "│" + reset)

	// content rows
	for i := 0; i < innerH; i++ {
		lineIdx := a.viewOff + i
		buf.WriteString(cursorPos(startRow+2+i, 1))
		buf.WriteString(clearLine())
		buf.WriteString(fg(t.Dim) + "│" + reset + " ")
		var rendered string
		if lineIdx < len(a.lines) {
			line := a.lines[lineIdx]
			if lineIdx == 0 {
				buf.WriteString(fg(t.Primary) + "> " + reset + fg(t.Text))
			} else {
				buf.WriteString("  " + fg(t.Text))
			}
			rendered = renderLineWithCursor(line, a.col, lineIdx == a.row, contentW)
			buf.WriteString(rendered + reset)
		} else {
			buf.WriteString("  ")
		}
		if pad := contentW - visibleLen(rendered); pad > 0 {
			buf.WriteString(strings.Repeat(" ", pad))
		}
		buf.WriteString(fg(t.Dim) + "│" + reset)
	}

	// blank padding row
	buf.WriteString(cursorPos(startRow+2+innerH, 1))
	buf.WriteString(clearLine())
	buf.WriteString(fg(t.Dim) + "│" + reset + strings.Repeat(" ", w-2) + fg(t.Dim) + "│" + reset)

	// bottom border — embed statusLeft and statusRight if provided
	buf.WriteString(cursorPos(startRow+3+innerH, 1))
	buf.WriteString(clearLine())
	switch {
	case statusLeft != "" && statusRight != "":
		left := " " + statusLeft + " "
		right := " " + statusRight + " "
		leftLen := visibleLen(left)
		rightLen := visibleLen(right)
		dashW := w - 2 - leftLen - rightLen
		if dashW < 0 {
			dashW = 0
		}
		buf.WriteString(fg(t.Dim) + "└" + reset + fg(t.Dim) + left + reset + fg(t.Dim) + strings.Repeat("─", dashW) + reset + fg(t.Dim) + right + reset + fg(t.Dim) + "┘" + reset)
	case statusLeft != "":
		left := " " + statusLeft + " "
		leftLen := visibleLen(left)
		dashW := w - 2 - leftLen
		if dashW < 0 {
			dashW = 0
		}
		buf.WriteString(fg(t.Dim) + "└" + reset + fg(t.Dim) + left + reset + fg(t.Dim) + strings.Repeat("─", dashW) + "┘" + reset)
	case statusRight != "":
		right := " " + statusRight + " "
		rightLen := visibleLen(right)
		dashW := w - 2 - rightLen
		if dashW < 0 {
			dashW = 0
		}
		buf.WriteString(fg(t.Dim) + "└" + strings.Repeat("─", dashW) + reset + fg(t.Dim) + right + reset + fg(t.Dim) + "┘" + reset)
	default:
		buf.WriteString(fg(t.Dim) + "└" + strings.Repeat("─", w-2) + "┘" + reset)
	}

	return height
}

// renderLineWithCursor renders a line of runes clipped to maxW visible chars,
// inserting a cursor marker if active.
func renderLineWithCursor(line []rune, col int, active bool, maxW int) string {
	// Determine visible window: scroll so cursor is always in view.
	start := 0
	if active && col >= maxW {
		start = col - maxW + 1
	}
	var b strings.Builder
	visible := 0
	for i := start; i < len(line) && visible < maxW; i++ {
		r := line[i]
		if active && i == col {
			b.WriteString(reverse())
			b.WriteRune(r)
			b.WriteString(reset)
		} else {
			b.WriteRune(r)
		}
		visible++
	}
	if active && col == len(line) && visible < maxW {
		b.WriteString(reverse())
		b.WriteString(" ")
		b.WriteString(reset)
		visible++
	}
	return b.String()
}

// charCount returns total characters across all lines.
func (a *inputArea) charCount() int {
	n := 0
	for _, l := range a.lines {
		n += len(l)
	}
	return n
}
