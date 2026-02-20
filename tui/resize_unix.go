//go:build !windows

package tui

import (
	"os"
	"os/signal"
	"syscall"
)

func watchResize(t *TUI) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		for range sigCh {
			t.mu.Lock()
			t.draw()
			t.mu.Unlock()
		}
	}()
}
