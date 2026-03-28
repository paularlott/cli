package tui

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/term"
)

// Run enters the event loop. It blocks until the user exits (Ctrl+C, t.Exit(), or ctx cancellation).
func (t *TUI) Run(ctx context.Context) error {
	t.mu.Lock()
	t.ctx = ctx
	t.mu.Unlock()

	t.fd = int(os.Stdin.Fd())
	old, err := term.MakeRaw(t.fd)
	if err != nil {
		return err
	}
	t.oldState = old
	defer t.restore()

	// Handle terminal resize signals.
	watchResize(t)

	fmt.Print("\x1b[?1049h") // enter alternate screen
	defer fmt.Print("\x1b[?1049l") // leave alternate screen

	// Enable mouse button-event reporting (SGR extended mode) and modifyOtherKeys.
	fmt.Print("\x1b[?1002h\x1b[?1006h\x1b[>4;1m")
	defer fmt.Print("\x1b[?1006l\x1b[?1002l\x1b[>4;0m")

	t.resize()
	t.draw()

	go func() {
		<-ctx.Done()
		t.mu.Lock()
		t.quit = true
		t.mu.Unlock()
		// Write a NUL byte to unblock the stdin Read so the loop can exit.
		os.Stdin.Write([]byte{0})
	}()

	buf := make([]byte, 128)
	for {
		t.mu.Lock()
		quit := t.quit
		t.mu.Unlock()
		if quit {
			break
		}
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}
		t.mu.Lock()
		cb := t.handleInput(buf[:n])
		t.draw()
		t.mu.Unlock()
		if cb != nil {
			cb()
		}
	}
	return nil
}

func (t *TUI) restore() {
	if t.oldState != nil {
		term.Restore(t.fd, t.oldState)
	}
	fmt.Print(resetScrollRegion(), showCursor(), reset)
}

func (t *TUI) resize() bool {
	w, h, err := term.GetSize(t.fd)
	if err != nil || w < 10 || h < 5 {
		w, h = 80, 24
	}
	changed := t.width != w || t.height != h
	t.width = w
	t.height = h
	return changed
}

// outputHeight returns the number of rows the output region occupies.
func (t *TUI) outputHeight() int {
	var bottomH int
	if t.menu != nil {
		bottomH = 1 + 10
	} else if t.inputEnabled() {
		bottomH = t.paletteHeight() + t.inputBoxHeight()
	} else {
		bottomH = 1
	}
	h := t.height - bottomH
	if h < 1 {
		h = 1
	}
	return h
}

// focusedPanelInfo holds layout info for a panel needed for mouse selection
type focusedPanelInfo struct {
	region  *outputRegion
	xOffset int // 0-based column where panel content starts
	yOffset int // 0-based row where panel content starts
	height  int // content height in rows
}

// panelAtPosition returns layout info for whichever panel contains the given
// screen coordinates. Falls back to main panel if nothing matches.
func (t *TUI) panelAtPosition(screenCol, screenRow int) focusedPanelInfo {
	outputH := t.outputHeight()

	// Single panel mode — always main
	if !t.hasMultiplePanels() {
		return focusedPanelInfo{
			region:  t.output,
			xOffset: 0,
			yOffset: 0,
			height:  outputH,
		}
	}

	// Helper to check visibility
	isVisible := func(p *Panel, availW int) (int, bool) {
		w := t.calculatePanelWidth(p.layoutWidth, availW)
		minW := p.minWidth
		if minW == 0 {
			minW = 2
		}
		return w, w >= minW
	}

	// Helper to compute content bounds for a panel
	panelBounds := func(panel *Panel, px, py, h int, hasBorder bool) focusedPanelInfo {
		contentX := px
		contentY := py
		if hasBorder {
			contentX += 1
			contentY += 1
			h -= 2
		}
		if h < 1 {
			h = 1
		}
		return focusedPanelInfo{
			region:  panel.region,
			xOffset: contentX,
			yOffset: contentY,
			height:  h,
		}
	}

	// Helper to check if a point is inside a content rect
	contains := func(col, row, rx, ry, rw, rh int) bool {
		return col >= rx && col < rx+rw && row >= ry && row < ry+rh
	}

	// Helper to check a panel (possibly split) and return info if hit
	checkPanel := func(p *Panel, px int, availW int) (focusedPanelInfo, int, bool) {
		w, visible := isVisible(p, availW)
		if !visible {
			return focusedPanelInfo{}, 0, false
		}
		hasBorder := !p.noBorder
		totalW := w
		if hasBorder {
			totalW += 2
		}

		if len(p.rows) > 0 {
			y := 0
			availH := outputH
			rowHeights := distributeRowHeights(p, availH)
			for i, child := range p.rows {
				h := rowHeights[i]
				b := panelBounds(child, px, y, h, !child.noBorder)
				if contains(screenCol, screenRow, b.xOffset, b.yOffset, w, b.height) {
					return b, totalW, true
				}
				y += h
			}
		} else {
			b := panelBounds(p, px, 0, outputH, hasBorder)
			if contains(screenCol, screenRow, b.xOffset, b.yOffset, w, b.height) {
				return b, totalW, true
			}
		}
		return focusedPanelInfo{}, totalW, false
	}

	// Calculate widths in the SAME order as drawPanels:
	// left panel takes its share, right panel takes its share from remainder,
	// then main panel gets what's left. Layout on screen is: left | main | right.

	remainingWidth := t.width

	// Step 1: Calculate left panel width
	leftTotalW := 0
	if t.leftRoot != nil {
		p := t.leftRoot
		w, visible := isVisible(p, remainingWidth)
		if visible {
			leftTotalW = w
			if !p.noBorder {
				leftTotalW += 2
			}
			remainingWidth -= leftTotalW
		}
	}

	// Step 2: Calculate right panel width (from remainder after left, before main)
	var rightTotalW int
	rightVisible := false
	if t.rightRoot != nil {
		p := t.rightRoot
		w, vis := isVisible(p, remainingWidth)
		if vis {
			rightTotalW = w
			if !p.noBorder {
				rightTotalW += 2
			}
			rightVisible = true
		}
	}

	// Step 3: Main panel gets the rest
	mainWidth := remainingWidth - rightTotalW
	mainBorderW := 2 // always has border in multi-panel
	mainContentW := mainWidth - mainBorderW
	if mainContentW < 2 {
		mainContentW = 2
	}

	// Now walk panels in display order (left -> main -> right), checking hits
	x := 0

	// Left panel hit test
	if t.leftRoot != nil && leftTotalW > 0 {
		p := t.leftRoot
		if b, _, hit := checkPanel(p, x, t.width); hit {
			return b
		}
		x += leftTotalW
	}

	// Main panel hit test
	{
		b := focusedPanelInfo{
			region:  t.output,
			xOffset: x + 1,
			yOffset: 1,
			height:  outputH - 2,
		}
		if contains(screenCol, screenRow, b.xOffset, b.yOffset, mainContentW, b.height) {
			return b
		}
		x += mainWidth
	}

	// Right panel hit test
	if t.rightRoot != nil && rightVisible {
		p := t.rightRoot
		// Right panel's available width was remainingWidth after left (before main took its share)
		rightAvailW := remainingWidth + rightTotalW // restore: this is what was available after left
		if b, _, hit := checkPanel(p, x, rightAvailW); hit {
			return b
		}
	}

	// Fallback: main panel
	return focusedPanelInfo{
		region:  t.output,
		xOffset: x + 1,
		yOffset: 1,
		height:  outputH - 2,
	}
}

// inputBoxHeight returns the number of rows the input box occupies (including borders).
func (t *TUI) inputBoxHeight() int {
	h := len(t.input.lines) + 4
	if h < inputMinHeight {
		h = inputMinHeight
	}
	max := t.height / 2
	if max < 3 {
		max = 3
	}
	if h > max {
		h = max
	}
	return h
}

// paletteHeight returns the number of rows the palette occupies (0 if inactive).
func (t *TUI) paletteHeight() int {
	if !t.palette.active {
		return 0
	}
	var n int
	if t.palette.argMode {
		n = len(t.palette.argFiltered)
	} else {
		n = len(t.palette.filtered)
	}
	if n == 0 {
		return 0
	}
	if n > 8 {
		n = 8
	}
	return n + 1 // +1 for hint row
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
