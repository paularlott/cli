// Example TUI log viewer — output-only mode.
//
// Run with:
//
//	go run ./tui/example2
//
// Pumps coloured log lines into the output region. Ctrl+C exits.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/paularlott/cli/tui"
)

type level struct {
	label string
	color tui.Color
	role  tui.MessageRole
}

var levels = []level{
	{"INFO ", 0x7EC87A, tui.RoleAssistant}, // green
	{"WARN ", 0xE8A87C, tui.RoleSystem},    // amber
	{"ERROR", 0xE05050, tui.RoleSystem},    // red
	{"DEBUG", 0x7BA7E8, tui.RoleAssistant}, // blue
}

var messages = []string{
	"Starting service on :8080",
	"Connected to database",
	"Request received: GET /api/health",
	"Cache miss for key user:42",
	"Retrying connection attempt 1/3",
	"TLS handshake completed",
	"Worker pool initialised with 8 goroutines",
	"Config reloaded from disk",
	"Slow query detected: 320ms",
	"Graceful shutdown initiated",
	"Flushing write buffer",
	"Metrics exported to Prometheus",
	"Rate limit exceeded for client 10.0.0.5",
	"Disk usage at 78%",
	"Heartbeat OK",
}

func main() {
	disabled := false
	t := tui.New(tui.Config{
		Theme:        tui.ThemeDefault,
		InputEnabled: &disabled,
		HideHeaders:  true,
		StatusLeft:   "tui/example2 — log viewer",
		StatusRight:  "Ctrl+C to exit",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pump(ctx, t)

	if err := t.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func pump(ctx context.Context, t *tui.TUI) {
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case ts := <-ticker.C:
			lvl := levels[rand.Intn(len(levels))]
			msg := messages[rand.Intn(len(messages))]
			line := tui.Styled(lvl.color, lvl.label) + "  " +
				tui.Styled(0x6C6F85, ts.Format("15:04:05")) + "  " + msg
			t.WriteString(line + "\n")
		}
	}
}
