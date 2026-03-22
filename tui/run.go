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

func (t *TUI) resize() {
	w, h, err := term.GetSize(t.fd)
	if err != nil || w < 10 || h < 5 {
		w, h = 80, 24
	}
	t.width = w
	t.height = h
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
