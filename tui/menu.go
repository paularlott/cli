package tui

import (
	"strings"
	"unicode/utf8"
)

// MenuItem is a single entry in a Menu.
type MenuItem struct {
	Label    string
	Value    string      // optional value passed to OnSelect
	Prompt   string      // if set, selecting this item opens a text-entry prompt with this label
	Children []*MenuItem // non-nil → sub-menu
	OnSelect func(item *MenuItem, input string) // input is non-empty only for Prompt items
}

// Menu is a navigable panel that replaces the input box.
type Menu struct {
	Title string
	Items []*MenuItem
}

// menuLevel is one level of the navigation stack.
type menuLevel struct {
	menu       *Menu
	selected   int
	viewOff    int
	promptItem *MenuItem // non-nil when in text-entry mode
	promptBuf  []rune
}

// menuState holds the navigation stack while a menu is open.
type menuState struct {
	stack []*menuLevel
}

func newMenuState(m *Menu) *menuState {
	return &menuState{stack: []*menuLevel{{menu: m}}}
}

func (ms *menuState) current() *menuLevel { return ms.stack[len(ms.stack)-1] }

func (ms *menuState) push(m *Menu) {
	ms.stack = append(ms.stack, &menuLevel{menu: m})
}

// pop removes the top level. Returns false if already at root.
func (ms *menuState) pop() bool {
	if len(ms.stack) <= 1 {
		return false
	}
	ms.stack = ms.stack[:len(ms.stack)-1]
	return true
}

func (ms *menuState) moveUp(maxRows int) {
	lv := ms.current()
	if lv.promptItem != nil {
		return
	}
	if lv.selected > 0 {
		lv.selected--
		if lv.selected < lv.viewOff {
			lv.viewOff = lv.selected
		}
	}
}

func (ms *menuState) moveDown(maxRows int) {
	lv := ms.current()
	if lv.promptItem != nil {
		return
	}
	n := len(lv.menu.Items)
	if lv.selected < n-1 {
		lv.selected++
		if lv.selected >= lv.viewOff+maxRows {
			lv.viewOff = lv.selected - maxRows + 1
		}
	}
}

// render draws the menu panel into buf, occupying exactly height rows starting at startRow.
func (ms *menuState) render(buf *strings.Builder, t *Theme, w, height, startRow int) {
	lv := ms.current()

	// In prompt mode the panel shows: border + title + prompt-label + input + hint + border = 6 rows.
	// In list mode: border + title + items… + hint + border.
	maxItems := height - 4 // top border + title + hint + bottom border
	if maxItems < 1 {
		maxItems = 1
	}

	// Breadcrumb title.
	title := lv.menu.Title
	if len(ms.stack) > 1 {
		parts := make([]string, len(ms.stack))
		for i, l := range ms.stack {
			parts[i] = l.menu.Title
		}
		title = strings.Join(parts, " › ")
	}

	innerW := w - 2
	row := startRow

	// Top border.
	buf.WriteString(cursorPos(row, 1))
	buf.WriteString(clearLine())
	buf.WriteString(fg(t.Dim) + "┌" + strings.Repeat("─", innerW) + "┐" + reset)
	row++

	// Title row.
	buf.WriteString(cursorPos(row, 1))
	buf.WriteString(clearLine())
	titleStr := " " + title + " "
	pad := innerW - utf8.RuneCountInString(titleStr)
	if pad < 0 {
		pad = 0
	}
	buf.WriteString(fg(t.Dim) + "│" + reset)
	buf.WriteString(fg(t.Primary) + bold() + titleStr + reset)
	buf.WriteString(strings.Repeat(" ", pad))
	buf.WriteString(fg(t.Dim) + "│" + reset)
	row++

	if lv.promptItem != nil {
		// Prompt label row.
		buf.WriteString(cursorPos(row, 1))
		buf.WriteString(clearLine())
		labelStr := "  " + lv.promptItem.Prompt
		labelPad := innerW - utf8.RuneCountInString(labelStr)
		if labelPad < 0 {
			labelPad = 0
		}
		buf.WriteString(fg(t.Dim) + "│" + reset)
		buf.WriteString(fg(t.Secondary) + labelStr + strings.Repeat(" ", labelPad) + reset)
		buf.WriteString(fg(t.Dim) + "│" + reset)
		row++

		// Input row.
		buf.WriteString(cursorPos(row, 1))
		buf.WriteString(clearLine())
		inputVal := string(lv.promptBuf)
		inputStr := "  " + fg(t.Primary) + "> " + reset + fg(t.Text) + inputVal + reverse() + " " + reset
		inputPad := innerW - 4 - utf8.RuneCountInString(inputVal)
		if inputPad < 0 {
			inputPad = 0
		}
		buf.WriteString(fg(t.Dim) + "│" + reset)
		buf.WriteString(inputStr + strings.Repeat(" ", inputPad))
		buf.WriteString(fg(t.Dim) + "│" + reset)
		row++

		// Fill remaining item rows.
		for row < startRow+2+maxItems {
			buf.WriteString(cursorPos(row, 1))
			buf.WriteString(clearLine())
			buf.WriteString(fg(t.Dim) + "│" + reset + strings.Repeat(" ", innerW) + fg(t.Dim) + "│" + reset)
			row++
		}

		// Hint.
		buf.WriteString(cursorPos(row, 1))
		buf.WriteString(clearLine())
		hint := "  Enter confirm · Esc cancel"
		hintPad := innerW - utf8.RuneCountInString(hint)
		if hintPad < 0 {
			hintPad = 0
		}
		buf.WriteString(fg(t.Dim) + "│" + reset)
		buf.WriteString(fg(t.Dim) + hint + strings.Repeat(" ", hintPad) + reset)
		buf.WriteString(fg(t.Dim) + "│" + reset)
		row++
	} else {
		// Items.
		items := lv.menu.Items
		end := lv.viewOff + maxItems
		if end > len(items) {
			end = len(items)
		}
		for i := lv.viewOff; i < end; i++ {
			item := items[i]
			buf.WriteString(cursorPos(row, 1))
			buf.WriteString(clearLine())
			buf.WriteString(fg(t.Dim) + "│" + reset)
			var line strings.Builder
			if i == lv.selected {
				line.WriteString(fg(t.Primary) + bold() + " › " + item.Label)
				if item.Children != nil || item.Prompt != "" {
					line.WriteString(" ›")
				}
				line.WriteString(reset)
			} else {
				line.WriteString(fg(t.Secondary) + "   " + item.Label)
				if item.Children != nil || item.Prompt != "" {
					line.WriteString(" ›")
				}
				line.WriteString(reset)
			}
			content := line.String()
			vl := visibleLen(content)
			buf.WriteString(content)
			if pad := innerW - vl; pad > 0 {
				buf.WriteString(strings.Repeat(" ", pad))
			}
			buf.WriteString(fg(t.Dim) + "│" + reset)
			row++
		}

		// Fill empty rows.
		for row < startRow+2+maxItems {
			buf.WriteString(cursorPos(row, 1))
			buf.WriteString(clearLine())
			buf.WriteString(fg(t.Dim) + "│" + reset + strings.Repeat(" ", innerW) + fg(t.Dim) + "│" + reset)
			row++
		}

		// Hint.
		buf.WriteString(cursorPos(row, 1))
		buf.WriteString(clearLine())
		hint := "  ↑↓ navigate · Enter select"
		if len(ms.stack) > 1 {
			hint += " · Esc back"
		} else {
			hint += " · Esc close"
		}
		hintPad := innerW - utf8.RuneCountInString(hint)
		if hintPad < 0 {
			hintPad = 0
		}
		buf.WriteString(fg(t.Dim) + "│" + reset)
		buf.WriteString(fg(t.Dim) + hint + strings.Repeat(" ", hintPad) + reset)
		buf.WriteString(fg(t.Dim) + "│" + reset)
		row++
	}

	// Bottom border.
	buf.WriteString(cursorPos(row, 1))
	buf.WriteString(clearLine())
	buf.WriteString(fg(t.Dim) + "└" + strings.Repeat("─", innerW) + "┘" + reset)
}
