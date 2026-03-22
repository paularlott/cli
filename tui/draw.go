package tui

import (
	"fmt"
	"strings"
)

func (t *TUI) draw() {
	t.resize()

	inputH := t.inputBoxHeight()
	paletteH := t.paletteHeight()
	var bottomH int
	if t.menu != nil {
		// Menu replaces input: separator(1) + menu(fixed 10)
		bottomH = 1 + 10
	} else if t.inputEnabled() {
		bottomH = paletteH + inputH
	} else {
		bottomH = 1 // separator
	}
	outputH := t.height - bottomH
	if outputH < 1 {
		outputH = 1
	}

	var buf strings.Builder
	buf.WriteString(hideCursor())

	// Output region.
	t.output.render(&buf, t.theme, t.width, outputH, 1)

	row := outputH + 1

	// Separator — only rendered in output-only and menu modes.
	// In input-enabled mode the overlay goes into the input box top border instead.
	if !t.inputEnabled() || t.menu != nil {
		buf.WriteString(cursorPos(row, 1))
		buf.WriteString(clearLine())
		if t.output.scrollOff > 0 {
			scrollHint := "↑ scrolled · scroll down to follow"
			sepW := t.width - visibleLen(scrollHint) - 2
			if sepW < 0 {
				sepW = 0
			}
			buf.WriteString(fg(t.theme.Dim) + strings.Repeat("─", sepW) + " " + reset + fg(t.theme.Primary) + scrollHint + " " + reset)
		} else if !t.inputEnabled() && (t.cfg.StatusLeft != "" || t.cfg.StatusRight != "") {
			// Embed status into the separator line.
			switch {
			case t.cfg.StatusLeft != "" && t.cfg.StatusRight != "":
				left := " " + t.cfg.StatusLeft + " "
				right := " " + t.cfg.StatusRight + " "
				dashW := t.width - visibleLen(left) - visibleLen(right)
				if dashW < 0 {
					dashW = 0
				}
				buf.WriteString(fg(t.theme.Primary) + left + reset + fg(t.theme.Dim) + strings.Repeat("─", dashW) + reset + fg(t.theme.Primary) + right + reset)
			case t.cfg.StatusLeft != "":
				left := " " + t.cfg.StatusLeft + " "
				dashW := t.width - visibleLen(left)
				if dashW < 0 {
					dashW = 0
				}
				buf.WriteString(fg(t.theme.Primary) + left + reset + fg(t.theme.Dim) + strings.Repeat("─", dashW) + reset)
			case t.cfg.StatusRight != "":
				right := " " + t.cfg.StatusRight + " "
				dashW := t.width - visibleLen(right)
				if dashW < 0 {
					dashW = 0
				}
				buf.WriteString(fg(t.theme.Dim) + strings.Repeat("─", dashW) + reset + fg(t.theme.Primary) + right + reset)
			}
		} else {
			buf.WriteString(fg(t.theme.Dim) + strings.Repeat("─", t.width) + reset)
		}
		row++
	}

	// Build overlay text for input box top border (input-enabled, no menu).
	var inputOverlay string
	if t.inputEnabled() && t.menu == nil {
		switch {
		case t.output.scrollOff > 0:
			inputOverlay = "↑ scrolled · scroll down to follow"
		case t.spinnerText != "":
			inputOverlay = spinnerFrames[t.spinnerFrame] + " " + t.spinnerText
		case t.progress >= 0:
			pct := int(t.progress * 100)
			barWidth := 20
			filled := int(t.progress * float64(barWidth))
			inputOverlay = fmt.Sprintf("%s [%s%s] %3d%%", t.progressLabel, strings.Repeat("█", filled), strings.Repeat("░", barWidth-filled), pct)
		case t.cfg.StatusRight != "":
			inputOverlay = t.cfg.StatusRight
		}
	}

	// Palette.
	if t.menu != nil {
		t.menu.render(&buf, t.theme, t.width, 10, row)
		row += 10
	} else {
		if t.inputEnabled() && t.palette.active {
			t.palette.render(&buf, t.theme, t.width, 8, row)
			row += paletteH
		}

		// Input box.
		if t.inputEnabled() {
			var botLeft, botRight string
			if t.cfg.StatusLeft != "" {
				botLeft = t.cfg.StatusLeft
			}
			if t.cfg.ShowCharCount {
				botRight = fmt.Sprintf("%d chars", t.input.charCount())
			}
			t.input.render(&buf, t.theme, t.width, inputH, row, inputOverlay, botLeft, botRight)
			row += inputH
		}
	}

	fmt.Print(buf.String())
}

func (t *TUI) handleInput(b []byte) func() {
	// Ctrl+C — quit (no selection mode in this build).
	if len(b) == 1 && b[0] == 3 {
		t.quit = true
		return nil
	}

	// Clear selection on any non-mouse keypress.
	isMouse := len(b) >= 3 && b[0] == 0x1b && b[1] == '[' && (b[2] == '<' || b[2] == 'M')
	if !isMouse && t.output.sel != nil {
		t.output.sel = nil
	}

	// Menu navigation takes priority.
	if t.menu != nil {
		lv := t.menu.current()
		// Prompt mode: capture text input.
		if lv.promptItem != nil {
			if len(b) >= 3 && b[0] == 0x1b && b[1] == '[' {
				switch b[2] {
				case '5':
					if len(b) >= 4 && b[3] == '~' {
						t.output.scrollUp(t.height / 2)
					}
				case '6':
					if len(b) >= 4 && b[3] == '~' {
						t.output.scrollDown(t.height / 2)
					}
				}
				return nil
			}
			if len(b) == 1 && b[0] == 0x1b {
				lv.promptItem = nil
				lv.promptBuf = nil
				return nil
			}
			if len(b) == 1 && (b[0] == '\r' || b[0] == '\n') {
				item := lv.promptItem
				input := string(lv.promptBuf)
				lv.promptItem = nil
				lv.promptBuf = nil
				t.menu = nil
				if item.OnSelect != nil {
					return func() { item.OnSelect(item, input) }
				}
				return nil
			}
			if len(b) == 1 && (b[0] == 0x7f || b[0] == 0x08) {
				if len(lv.promptBuf) > 0 {
					lv.promptBuf = lv.promptBuf[:len(lv.promptBuf)-1]
				}
				return nil
			}
			for _, r := range string(b) {
				if r >= 0x20 {
					lv.promptBuf = append(lv.promptBuf, r)
				}
			}
			return nil
		}
		// List navigation mode.
		if len(b) == 3 && b[0] == 0x1b && b[1] == 'O' {
			switch b[2] {
			case 'A':
				t.menu.moveUp(6)
			case 'B':
				t.menu.moveDown(6)
			}
			return nil
		}
		if len(b) >= 3 && b[0] == 0x1b && b[1] == '[' {
			switch b[2] {
			case 'A':
				t.menu.moveUp(6)
			case 'B':
				t.menu.moveDown(6)
			case '5':
				if len(b) >= 4 && b[3] == '~' {
					t.output.scrollUp(t.height / 2)
				}
			case '6':
				if len(b) >= 4 && b[3] == '~' {
					t.output.scrollDown(t.height / 2)
				}
			case 'M':
				if len(b) >= 6 {
					switch b[3] & 0x7f {
					case 64:
						t.output.scrollUp(3)
					case 65:
						t.output.scrollDown(3)
					}
				}
			}
			return nil
		}
		if len(b) == 1 && b[0] == 0x1b {
			if !t.menu.pop() {
				t.menu = nil
			}
			return nil
		}
		if len(b) == 1 && (b[0] == '\r' || b[0] == '\n') {
			if lv.selected < len(lv.menu.Items) {
				item := lv.menu.Items[lv.selected]
				if item.Children != nil {
					t.menu.push(&Menu{Title: item.Label, Items: item.Children})
				} else if item.Prompt != "" {
					lv.promptItem = item
					lv.promptBuf = nil
				} else {
					t.menu = nil
					if item.OnSelect != nil {
						return func() { item.OnSelect(item, "") }
					}
				}
			}
			return nil
		}
		return nil
	}

	// Escape sequences.
	// SS3 arrow keys: ESC O A/B/C/D; ESC O M = Shift+Enter (macOS Terminal.app)
	if len(b) == 3 && b[0] == 0x1b && b[1] == 'O' {
		switch b[2] {
		case 'A':
			if t.inputEnabled() && t.palette.active {
				t.palette.moveUp()
			} else if !t.inputEnabled() || !t.input.historyUp() {
				t.input.moveUp()
			}
		case 'B':
			if t.inputEnabled() && t.palette.active {
				t.palette.moveDown(8)
			} else if !t.inputEnabled() || !t.input.historyDown() {
				t.input.moveDown()
			}
		case 'C':
			t.input.moveRight()
		case 'D':
			t.input.moveLeft()
		case 'M':
			if t.inputEnabled() {
				t.input.insertNewline()
			}
		}
		return nil
	}

	if len(b) >= 3 && b[0] == 0x1b && b[1] == '[' {
		switch b[2] {
		case 'A': // Up
			if t.inputEnabled() && t.palette.active {
				t.palette.moveUp()
			} else if !t.inputEnabled() || !t.input.historyUp() {
				t.input.moveUp()
			}
			return nil
		case 'B': // Down
			if t.inputEnabled() && t.palette.active {
				t.palette.moveDown(8)
			} else if !t.inputEnabled() || !t.input.historyDown() {
				t.input.moveDown()
			}
			return nil
		case 'C': // Right
			t.input.moveRight()
			return nil
		case 'D': // Left
			t.input.moveLeft()
			return nil
		case 'H': // Home
			t.input.home()
			return nil
		case 'F': // End
			t.input.end()
			return nil
		case '2': // Shift+Enter (modifyOtherKeys): ESC [ 2 7 ; 2 ; 1 3 ~
			if len(b) == 10 && string(b) == "\x1b[27;2;13~" {
				t.input.insertNewline()
			}
			return nil
		case '3': // Delete (ESC [ 3 ~)
			if len(b) >= 4 && b[3] == '~' {
				t.input.deleteForward()
			}
			return nil
		case '5': // Page Up
			if len(b) >= 4 && b[3] == '~' {
				t.output.scrollUp(t.height / 2)
			}
			return nil
		case '6': // Page Down
			if len(b) >= 4 && b[3] == '~' {
				t.output.scrollDown(t.height / 2)
			}
			return nil
		case 'M': // X10 mouse event: ESC [ M b x y
			if len(b) >= 6 {
				switch b[3] & 0x7f {
				case 64: // wheel up
					t.output.scrollUp(3)
				case 65: // wheel down
					t.output.scrollDown(3)
				}
			}
			return nil
		}
		// SGR mouse: ESC [ < btn ; col ; row M/m
		if b[2] == '<' {
			s := string(b[3:])
			if len(s) > 0 && (s[len(s)-1] == 'M' || s[len(s)-1] == 'm') {
				release := s[len(s)-1] == 'm'
				parts := strings.SplitN(s[:len(s)-1], ";", 3)
				if len(parts) == 3 {
					btn := parts[0]
					col := atoi(parts[1]) - 1      // 0-based
					screenRow := atoi(parts[2]) - 1 // 0-based screen row
					switch btn {
					case "64": // wheel up
						t.output.scrollUp(3)
					case "65": // wheel down
						t.output.scrollDown(3)
					case "0": // left button press/release
						lineIdx := t.output.lastStart + screenRow
						if screenRow < t.outputHeight() && lineIdx < len(t.output.lastLines) {
							pt := selAnchor{row: lineIdx, col: col}
							if !release {
								t.output.sel = &[2]selAnchor{pt, pt}
							} else if t.output.sel != nil {
								t.output.sel[1] = pt
								if text := t.output.selectionText(); text != "" {
									fmt.Print(CopyToClipboard(text))
									t.flashCopied()
								} else {
									t.output.sel = nil
								}
							}
						} else if release {
							t.output.sel = nil
						}
					case "32": // left button drag (motion)
						lineIdx := t.output.lastStart + screenRow
						if t.output.sel != nil && screenRow < t.outputHeight() && lineIdx < len(t.output.lastLines) {
							t.output.sel[1] = selAnchor{row: lineIdx, col: col}
						}
					}
				}
			}
			return nil
		}
		return nil
	}

	// Escape alone — close palette or fire OnEscape.
	if len(b) == 1 && b[0] == 0x1b {
		if t.inputEnabled() && t.palette.active {
			t.palette.close()
			t.input.reset()
		} else if t.cfg.OnEscape != nil {
			cb := t.cfg.OnEscape
			return func() { cb() }
		}
		return nil
	}

	if !t.inputEnabled() {
		return nil
	}

	// Tab — complete from palette.
	if len(b) == 1 && b[0] == '\t' {
		if t.palette.active {
			if t.palette.argMode {
				if arg := t.palette.selectedArg(); arg != "" {
					current := t.input.text()
					if idx := strings.Index(current, " "); idx != -1 {
						t.input.reset()
						for _, r := range current[:idx+1] + arg {
							t.input.insertRune(r)
						}
						t.palette.filter(t.input.text()[1:])
					}
				}
			} else if cmd := t.palette.selectedCommand(); cmd != nil {
				t.input.reset()
				for _, r := range "/" + cmd.Name + " " {
					t.input.insertRune(r)
				}
				t.palette.filter(cmd.Name + " ")
			}
		}
		return nil
	}

	// Enter.
	if len(b) == 1 && (b[0] == '\r' || b[0] == '\n') {
		if t.palette.active {
			if t.palette.argMode {
				if arg := t.palette.selectedArg(); arg != "" {
					cmd := t.palette.argCmd
					t.palette.close()
					t.input.reset()
					return func() { cmd.Handler(arg) }
				}
			} else if cmd := t.palette.selectedCommand(); cmd != nil {
				// If the command has args and no arg has been provided yet, enter arg mode.
				if len(cmd.Args) > 0 && !t.palette.argMode {
					t.input.reset()
					for _, r := range "/" + cmd.Name + " " {
						t.input.insertRune(r)
					}
					t.palette.filter(cmd.Name + " ")
					return nil
				}
				text := t.input.text()
				args := ""
				if parts := strings.SplitN(text[1:], " ", 2); len(parts) == 2 {
					args = parts[1]
				}
				t.palette.close()
				t.input.reset()
				return func() { cmd.Handler(args) }
			}
			// No palette match — close and fall through to slash-command execution.
			t.palette.close()
		}
		text := strings.TrimSpace(t.input.text())
		t.input.pushHistory(text)
		t.input.reset()
		if text == "" {
			return nil
		}
		// Execute slash command if present.
		if strings.HasPrefix(text, "/") {
			parts := strings.SplitN(text[1:], " ", 2)
			name := parts[0]
			args := ""
			if len(parts) > 1 {
				args = parts[1]
			}
			for _, cmd := range t.palette.commands {
				if cmd.Name == name {
					return func() { cmd.Handler(args) }
				}
			}
			unknown := name
			return func() {
				t.mu.Lock()
				t.output.AddMessage(RoleSystem, "Unknown command: /"+unknown)
				t.draw()
				t.mu.Unlock()
			}
		}
		if t.cfg.OnSubmit != nil {
			cb := t.cfg.OnSubmit
			return func() { cb(text) }
		}
		return nil
	}

	// Shift+Enter: ESC \r (sent by xterm.js DOM interceptor) or kitty ESC [ 1 3 ; 2 u
	if (len(b) == 2 && b[0] == 0x1b && b[1] == '\r') ||
		(len(b) == 7 && string(b) == "\x1b[13;2u") {
		t.input.insertNewline()
		return nil
	}

	// Backspace.
	if len(b) == 1 && (b[0] == 0x7f || b[0] == 0x08) {
		if t.palette.active {
			if t.input.col == 1 && t.input.row == 0 {
				t.palette.close()
			}
			t.input.backspace()
			if t.palette.active {
				t.palette.filter(t.input.text()[1:])
			}
		} else {
			t.input.backspace()
		}
		return nil
	}

	// Ctrl keys.
	if len(b) == 1 {
		switch b[0] {
		case 1: // Ctrl+A — move to start
			t.input.home()
			return nil
		case 11: // Ctrl+K
			t.input.ctrlK()
			return nil
		case 21: // Ctrl+U
			t.input.ctrlU()
			return nil
		case 23: // Ctrl+W
			t.input.ctrlW()
			return nil
		}
	}

	// Printable runes (including pasted newlines).
	s := string(b)
	for _, r := range s {
		if r == '\r' || r == '\n' {
			t.input.insertNewline()
			continue
		}
		if r < 0x20 {
			continue
		}
		t.input.insertRune(r)
	}

	// Check for palette activation: / at start of otherwise empty input.
	text := t.input.text()
	if strings.HasPrefix(text, "/") {
		query := text[1:]
		if !t.palette.active {
			t.palette.open(query)
		} else {
			t.palette.filter(query)
		}
	} else if t.palette.active {
		t.palette.close()
	}
	return nil
}
