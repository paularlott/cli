package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func (t *TUI) draw() {
	// Clear screen on resize or output height change to avoid artifacts
	terminalResized := t.resize()

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

	// Clear screen if terminal resized or output area height changed
	if terminalResized || outputH != t.prevOutputH {
		fmt.Print("\x1b[2J")
		t.prevOutputH = outputH
	}

	var buf strings.Builder
	buf.WriteString(hideCursor())

	// Check if we have a multi-panel layout
	hasLeft := t.layout.Left != nil
	hasRight := t.layout.Right != nil
	multiPanel := hasLeft || hasRight

	if multiPanel {
		t.drawPanels(&buf, outputH)
	} else {
		// Single panel mode - render through main panel to support raw mode
		if t.mainPanel.rawMode {
			t.mainPanel.render(&buf, t.theme, t.width, outputH, 1, 1, true)
		} else {
			t.output.render(&buf, t.theme, t.width, outputH, 1)
		}
	}

	row := outputH + 1

	// Separator — only rendered in output-only and menu modes.
	// In input-enabled mode the overlay goes into the input box top border instead.
	if !t.inputEnabled() || t.menu != nil {
		buf.WriteString(cursorPos(row, 1))
		buf.WriteString(clearLine())
		if t.output.scrollOff > 0 {
			scrollHint := "↑ scrolled · scroll down to follow"
			sepW := max(0, t.width-visibleLen(scrollHint)-2)
			buf.WriteString(fg(t.theme.Dim) + strings.Repeat("─", sepW) + " " + reset + fg(t.theme.Primary) + scrollHint + " " + reset)
		} else if !t.inputEnabled() && (t.cfg.StatusLeft != "" || t.cfg.StatusRight != "") {
			// Embed status into the separator line.
			switch {
			case t.cfg.StatusLeft != "" && t.cfg.StatusRight != "":
				left := " " + t.cfg.StatusLeft + " "
				right := " " + t.cfg.StatusRight + " "
				dashW := max(0, t.width-visibleLen(left)-visibleLen(right))
				buf.WriteString(fg(t.theme.Primary) + left + reset + fg(t.theme.Dim) + strings.Repeat("─", dashW) + reset + fg(t.theme.Primary) + right + reset)
			case t.cfg.StatusLeft != "":
				left := " " + t.cfg.StatusLeft + " "
				dashW := max(0, t.width-visibleLen(left))
				buf.WriteString(fg(t.theme.Primary) + left + reset + fg(t.theme.Dim) + strings.Repeat("─", dashW) + reset)
			case t.cfg.StatusRight != "":
				right := " " + t.cfg.StatusRight + " "
				dashW := max(0, t.width-visibleLen(right))
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

// panelLayout holds computed panel dimensions
type panelLayout struct {
	name      string
	panel     *Panel
	x         int // starting column (0-based)
	y         int // starting row (0-based, relative to panel area)
	width     int // content width (excluding borders)
	height    int // content height (excluding borders)
	hasBorder bool
	focused   bool
	children  []panelLayout // for horizontal splits (top/bottom)
}

// drawPanels renders multiple panels side by side with borders
func (t *TUI) drawPanels(buf *strings.Builder, height int) {
	// Build list of visible panels
	var panels []panelLayout
	remainingWidth := t.width
	currentFocusIdx := 0 // Track position in focus cycle

	// Calculate widths and determine visibility
	if t.layout.Left != nil {
		cfg := t.layout.Left
		w := t.calculatePanelWidth(cfg.Width, remainingWidth)
		minW := cfg.MinWidth
		if minW == 0 {
			minW = 2 // default minimum
		}
		if w >= minW {
			p := t.panels[cfg.Name]
			if p == nil {
				p = t.createPanel(*cfg)
				t.panels[cfg.Name] = p
			}
			hasBorder := !cfg.NoBorder
			pl := panelLayout{
				name: cfg.Name, panel: p, width: w, height: height, hasBorder: hasBorder,
			}
			// Handle horizontal split (top/bottom children)
			if cfg.Top != nil || cfg.Bottom != nil {
				pl.children = t.buildChildPanels(cfg, w, height, &currentFocusIdx)
			} else {
				pl.focused = t.focusIdx == currentFocusIdx
				currentFocusIdx++
			}
			panels = append(panels, pl)
			remainingWidth -= w
			if hasBorder {
				remainingWidth -= 2 // left and right border columns
			}
		}
	}

	// Main panel always visible
	mainFocusIdx := currentFocusIdx
	panels = append(panels, panelLayout{
		name: "main", panel: t.mainPanel, width: 0, height: height, hasBorder: false,
		focused: t.focusIdx == currentFocusIdx,
	})
	currentFocusIdx++

	// Right panel
	if t.layout.Right != nil {
		cfg := t.layout.Right
		w := t.calculatePanelWidth(cfg.Width, remainingWidth)
		minW := cfg.MinWidth
		if minW == 0 {
			minW = 2 // default minimum
		}
		if w >= minW {
			p := t.panels[cfg.Name]
			if p == nil {
				p = t.createPanel(*cfg)
				t.panels[cfg.Name] = p
			}
			hasBorder := !cfg.NoBorder
			pl := panelLayout{
				name: cfg.Name, panel: p, width: w, height: height, hasBorder: hasBorder,
			}
			// Handle horizontal split (top/bottom children)
			if cfg.Top != nil || cfg.Bottom != nil {
				pl.children = t.buildChildPanels(cfg, w, height, &currentFocusIdx)
			} else {
				pl.focused = t.focusIdx == currentFocusIdx
				currentFocusIdx++
			}
			panels = append(panels, pl)
			remainingWidth -= w
			if hasBorder {
				remainingWidth -= 2 // left and right border columns
			}
		}
	}

	// Assign remaining width to main panel
	panels[mainFocusIdx].width = remainingWidth
	if panels[mainFocusIdx].width < 2 {
		panels[mainFocusIdx].width = 2
	}

	// Calculate x positions
	x := 0
	for i := range panels {
		panels[i].x = x
		if panels[i].hasBorder {
			x += panels[i].width + 2 // content width + left border + right border
		} else {
			x += panels[i].width
		}
	}

	// Draw panel borders (top line with titles)
	// Skip drawing top border for panels with children - children have their own borders
	for _, pl := range panels {
		if pl.hasBorder && len(pl.children) == 0 {
			t.drawPanelTopBorder(buf, pl)
		}
	}

	// Draw panel content
	for _, pl := range panels {
		if len(pl.children) > 0 {
			// Render child panels with horizontal split
			t.drawChildPanels(buf, pl, height)
		} else {
			t.drawSinglePanel(buf, pl, height)
		}
	}
}

// buildChildPanels creates child panel layouts for horizontal splits
func (t *TUI) buildChildPanels(cfg *PanelConfig, width, totalHeight int, currentFocusIdx *int) []panelLayout {
	var children []panelLayout

	// Children use full height - no parent border to account for
	availableHeight := totalHeight
	if availableHeight < 2 {
		availableHeight = 2
	}

	// Calculate heights for top and bottom
	var topH, bottomH int
	if cfg.Top != nil && cfg.Bottom != nil {
		// Both specified - split based on Height config
		if cfg.Top.Height < 0 {
			// Percentage
			topH = (availableHeight * (-cfg.Top.Height)) / 100
		} else if cfg.Top.Height > 0 {
			topH = cfg.Top.Height
		} else {
			topH = availableHeight / 2
		}
		// No gap between panels
		bottomH = availableHeight - topH
		if bottomH < 1 {
			bottomH = 1
			topH = availableHeight - 1
		}
	} else if cfg.Top != nil {
		topH = availableHeight
		bottomH = 0
	} else if cfg.Bottom != nil {
		topH = 0
		bottomH = availableHeight
	}

	// Build top child - each child has its own border
	if cfg.Top != nil && topH > 0 {
		p := t.panels[cfg.Top.Name]
		if p == nil {
			p = t.createPanel(*cfg.Top)
			t.panels[cfg.Top.Name] = p
		}
		children = append(children, panelLayout{
			name:      cfg.Top.Name,
			panel:     p,
			width:     width,
			height:    topH,
			hasBorder: !cfg.Top.NoBorder, // Children have their own borders
			focused:   t.focusIdx == *currentFocusIdx,
		})
		*currentFocusIdx++
	}

	// Build bottom child - each child has its own border
	if cfg.Bottom != nil && bottomH > 0 {
		p := t.panels[cfg.Bottom.Name]
		if p == nil {
			p = t.createPanel(*cfg.Bottom)
			t.panels[cfg.Bottom.Name] = p
		}
		children = append(children, panelLayout{
			name:      cfg.Bottom.Name,
			panel:     p,
			width:     width,
			height:    bottomH,
			hasBorder: !cfg.Bottom.NoBorder, // Children have their own borders
			focused:   t.focusIdx == *currentFocusIdx,
		})
		*currentFocusIdx++
	}

	return children
}

// drawChildPanels renders horizontally split child panels within a parent panel
func (t *TUI) drawChildPanels(buf *strings.Builder, pl panelLayout, height int) {
	if len(pl.children) == 0 {
		return
	}

	// Calculate starting positions - children start at row 1 (no parent top border)
	currentRow := 1

	// Render each child as an independent panel with its own border
	for _, child := range pl.children {
		// Child panels inherit parent's x position
		childX := pl.x

		if child.hasBorder {
			borderColor := t.theme.Dim
			if child.focused && child.panel != nil {
				borderColor = child.panel.color
			}

			// Draw top border with title
			t.drawTopBorder(buf, borderColor, childTitle(child), childX, currentRow, child.width)

			// Draw left and right borders (between top and bottom)
			t.drawVerticalBorders(buf, borderColor, childX, child.width, currentRow+1, currentRow+child.height-2)

			// Draw bottom border
			t.drawBottomBorder(buf, borderColor, childX, currentRow+child.height-1, child.width)

			// Render content inside the child's border
			contentH := child.height - 2
			if contentH > 0 && child.panel != nil {
				child.panel.render(buf, t.theme, child.width, contentH, currentRow+1, childX+2, child.focused)
			}
		} else {
			// No border - just render content
			if child.panel != nil && child.height > 0 {
				child.panel.render(buf, t.theme, child.width, child.height, currentRow, childX+1, child.focused)
			}
		}

		currentRow += child.height
	}

	// No parent border when there are children - the children have their own borders
}

// childTitle returns the title for a child panel
func childTitle(child panelLayout) string {
	if child.panel != nil && child.panel.title != "" {
		return child.panel.title
	}
	return child.name
}

// drawTopBorder draws a top border with optional title at the specified position
func (t *TUI) drawTopBorder(buf *strings.Builder, borderColor Color, title string, x, row, width int) {
	buf.WriteString(cursorPos(row, x+1))

	// Total width includes left border + content + right border
	w := width + 2
	if w < 3 {
		buf.WriteString(fg(borderColor) + "│" + reset)
		return
	}

	var line strings.Builder
	line.WriteString(fg(borderColor) + "┌" + reset)

	if title != "" && w > 6 {
		titlePart := " " + title + " "
		afterW := (w - 2 - utf8.RuneCountInString(titlePart)) / 2
		beforeW := w - 2 - utf8.RuneCountInString(titlePart) - afterW
		beforeW = max(0, beforeW)
		afterW = max(0, afterW)
		line.WriteString(fg(borderColor) + strings.Repeat("─", beforeW) + reset)
		line.WriteString(fg(borderColor) + bold() + titlePart + reset)
		line.WriteString(fg(borderColor) + strings.Repeat("─", afterW) + reset)
	} else {
		line.WriteString(fg(borderColor) + strings.Repeat("─", w-2) + reset)
	}

	line.WriteString(fg(borderColor) + "┐" + reset)
	buf.WriteString(line.String())
}

// drawVerticalBorders draws left and right borders for a panel
func (t *TUI) drawVerticalBorders(buf *strings.Builder, borderColor Color, x, width, startRow, endRow int) {
	leftCol := x + 1
	rightCol := x + width + 2
	for row := startRow; row <= endRow; row++ {
		buf.WriteString(cursorPos(row, leftCol))
		buf.WriteString(fg(borderColor) + "│" + reset)
		buf.WriteString(cursorPos(row, rightCol))
		buf.WriteString(fg(borderColor) + "│" + reset)
	}
}

// drawBottomBorder draws a bottom border at the specified position
func (t *TUI) drawBottomBorder(buf *strings.Builder, borderColor Color, x, row, width int) {
	buf.WriteString(cursorPos(row, x+1))
	buf.WriteString(fg(borderColor) + "└" + reset)
	if width > 0 {
		buf.WriteString(fg(borderColor) + strings.Repeat("─", width) + reset)
	}
	buf.WriteString(fg(borderColor) + "┘" + reset)
}

// drawPanelTopBorder draws the top border line with title for a panel
func (t *TUI) drawPanelTopBorder(buf *strings.Builder, pl panelLayout) {
	borderColor := t.theme.Dim
	if pl.focused && pl.panel != nil {
		borderColor = pl.panel.color
	}
	title := pl.name
	if pl.panel != nil && pl.panel.title != "" {
		title = pl.panel.title
	}
	t.drawTopBorder(buf, borderColor, title, pl.x, 1, pl.width)
}

// drawSinglePanel renders a single panel without children
func (t *TUI) drawSinglePanel(buf *strings.Builder, pl panelLayout, height int) {
	contentW := pl.width
	startCol := pl.x + 1 // 1-based column
	contentStartRow := 1
	contentHeight := height

	if pl.hasBorder {
		borderColor := t.theme.Dim
		if pl.focused && pl.panel != nil {
			borderColor = pl.panel.color
		}

		// Draw left and right borders (skip row 1 - it has the top border corner)
		t.drawVerticalBorders(buf, borderColor, pl.x, pl.width, 2, height)

		// Draw bottom border
		t.drawBottomBorder(buf, borderColor, pl.x, height, pl.width)

		startCol = pl.x + 2                 // Content starts after left border
		contentW = pl.width                 // Content width is the panel width
		contentStartRow = 2                 // Content starts after top border
		contentHeight = max(1, height-2)    // Reduce height for top and bottom borders
	}

	// Render panel content
	if pl.panel != nil && contentHeight > 0 {
		pl.panel.render(buf, t.theme, contentW, contentHeight, contentStartRow, startCol, pl.focused)
	}
}

// calculatePanelWidth converts config width to actual columns
func (t *TUI) calculatePanelWidth(configWidth, available int) int {
	if configWidth < 0 {
		// Percentage: -30 means 30%
		pct := -configWidth
		return (available * pct) / 100
	}
	return configWidth
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
					t.focusedPanel().scrollUp(t.height / 2)
				}
			case '6':
				if len(b) >= 4 && b[3] == '~' {
					t.focusedPanel().scrollDown(t.height / 2)
				}
			case 'M':
				if len(b) >= 6 {
					switch b[3] & 0x7f {
					case 64:
						t.focusedPanel().scrollUp(3)
					case 65:
						t.focusedPanel().scrollDown(3)
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
				t.focusedPanel().scrollUp(t.height / 2)
			}
			return nil
		case '6': // Page Down
			if len(b) >= 4 && b[3] == '~' {
				t.focusedPanel().scrollDown(t.height / 2)
			}
			return nil
		case 'M': // X10 mouse event: ESC [ M b x y
			if len(b) >= 6 {
				switch b[3] & 0x7f {
				case 64: // wheel up
					t.focusedPanel().scrollUp(3)
				case 65: // wheel down
					t.focusedPanel().scrollDown(3)
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
					col := atoi(parts[1]) - 1       // 0-based
					screenRow := atoi(parts[2]) - 1 // 0-based screen row
					switch btn {
					case "64": // wheel up
						t.focusedPanel().scrollUp(3)
					case "65": // wheel down
						t.focusedPanel().scrollDown(3)
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

	// Tab — complete from palette or cycle panel focus.
	if len(b) == 1 && b[0] == '\t' {
		// If palette is active, handle completion
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
		} else if t.hasMultiplePanels() {
			// Cycle panel focus
			t.cycleFocusLocked()
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
