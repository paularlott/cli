// Example TUI application demonstrating OnInterrupt with quit confirmation.
//
// Run with:
//
//	go run ./tui/example-interrupt
//
// Press Ctrl+C once to be asked to confirm, then Ctrl+C again within 3 seconds
// to quit. Any other key, or waiting 3 seconds, cancels the quit.
package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/paularlott/cli/tui"
)

func main() {
	var (
		t       *tui.TUI
		mu      sync.Mutex
		pending bool
		cancel  func()
	)

	confirmQuit := func() {
		mu.Lock()
		defer mu.Unlock()

		if pending {
			// Second Ctrl+C — confirmed, exit.
			if cancel != nil {
				cancel()
			}
			t.Exit()
			return
		}

		// First Ctrl+C — ask for confirmation.
		pending = true
		t.AddMessage(tui.RoleSystem, "Press Ctrl+C again within 3 seconds to quit, or keep typing to cancel.")

		ctx, c := context.WithCancel(context.Background())
		cancel = c

		go func() {
			select {
			case <-time.After(3 * time.Second):
				mu.Lock()
				if pending {
					pending = false
					t.AddMessage(tui.RoleSystem, "Quit cancelled.")
				}
				mu.Unlock()
			case <-ctx.Done():
			}
		}()
	}

	t = tui.New(tui.Config{
		StatusLeft:  "example-interrupt",
		StatusRight: "Ctrl+C to quit",
		OnInterrupt: confirmQuit,
		OnSubmit: func(text string) {
			// Any submission resets the pending quit.
			mu.Lock()
			if pending {
				pending = false
				if cancel != nil {
					cancel()
				}
			}
			mu.Unlock()
			t.AddMessage(tui.RoleUser, text)
			t.AddMessage(tui.RoleAssistant, "Echo: "+text)
		},
	})

	t.AddMessage(tui.RoleSystem, "Welcome! Press Ctrl+C once to be asked to confirm quit.")

	if err := t.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
