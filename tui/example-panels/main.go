// Example TUI with multiple panels — demonstrates the builder panel API.
//
// Run with:
//
//	go run ./tui/example-panels
//
// Shows a four-panel layout: logs (left), main chat (center), CPU stats (right top), Memory stats (right bottom).
// Type messages in the input to chat. Tab cycles panel focus. Ctrl+C exits.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/paularlott/cli/tui"
)

var logMessages = []string{
	"Starting service on :8080",
	"Connected to database",
	"Request received: GET /api/health",
	"Cache miss for key user:42",
	"Retrying connection attempt 1/3",
	"TLS handshake completed",
}

func main() {
	var t *tui.TUI

	cfg := tui.Config{
		Theme:       tui.ThemeAmber,
		StatusRight: "Tab: cycle panels · Ctrl+C: exit",
		OnSubmit: func(text string) {
			mainPanel := t.Panel("main")
			mainPanel.AddMessage(tui.RoleUser, text)
			mainPanel.StartStreamingAs("Demo Bot")
			mainPanel.StreamChunk("You said: \"" + text + "\"")
			mainPanel.StreamComplete()
		},
	}

	t = tui.New(cfg)

	// Create panels using the builder API
	logs := t.CreatePanel(tui.PanelConfig{Name: "logs", Width: -25, MinWidth: 15, Scrollable: true, Title: "Logs"})
	right := t.CreatePanel(tui.PanelConfig{Width: -30, MinWidth: 16})
	statsCpu := t.CreatePanel(tui.PanelConfig{Name: "stats-cpu", Height: -50, Title: "CPU"})
	statsMem := t.CreatePanel(tui.PanelConfig{Name: "stats-mem", Height: -50, Title: "Memory"})

	// Build the right panel as a vertical split
	right.AddRow(statsCpu)
	right.AddRow(statsMem)

	// Attach panels to layout
	t.AddLeft(logs)
	t.AddRight(right)

	// Customize panels
	logs.SetColor(tui.Color(0x7EC87A))     // green
	statsCpu.SetColor(tui.Color(0x7BA7E8)) // blue
	statsMem.SetColor(tui.Color(0xE87BA7)) // pink

	mainPanel := t.Panel("main")

	// Add slash commands
	t.AddCommand(&tui.Command{
		Name:        "clear",
		Description: "Clear main output",
		Handler:     func(_ string) { mainPanel.Clear() },
	})
	t.AddCommand(&tui.Command{
		Name:        "layout",
		Description: "Toggle panel layout",
		Handler: func(_ string) {
			if t.HasMultiplePanels() {
				t.ClearLayout()
			} else {
				t.AddLeft(logs)
				t.AddRight(right)
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background log pump
	go pumpLogs(ctx, logs)

	// Start stats updaters
	go updateCPUStats(ctx, statsCpu)
	go updateMemStats(ctx, statsMem)

	// Add welcome message
	mainPanel.AddMessage(tui.RoleSystem, "Welcome to the panels demo!")
	mainPanel.AddMessage(tui.RoleAssistant, "This example shows a four-panel layout:\n\n"+
		"  Left: Live log stream\n"+
		"  Center: Chat (main panel)\n"+
		"  Right top: CPU stats\n"+
		"  Right bottom: Memory stats\n\n"+
		"Type a message and press Enter to chat.\n"+
		"Press Tab to cycle focus between panels.\n"+
		"Try /clear or /layout commands.")

	if err := t.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func pumpLogs(ctx context.Context, logs *tui.Panel) {
	ticker := time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ts := <-ticker.C:
			msg := logMessages[rand.Intn(len(logMessages))]
			line := logs.StyledWith(tui.ThemeDim, ts.Format("15:04:05")) + " " + msg
			logs.WriteString(line + "\n")
		}
	}
}

func updateCPUStats(ctx context.Context, panel *tui.Panel) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	cpu := 25.0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpu += (rand.Float64() - 0.5) * 10
			if cpu < 5 {
				cpu = 5
			}
			if cpu > 95 {
				cpu = 95
			}

			barWidth := 16
			filled := int(cpu / 100 * float64(barWidth))
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

			content := panel.StyledWith(tui.ThemePrimary, "CPU Usage") + "\n\n" +
				fmt.Sprintf("  %s\n\n", bar) +
				fmt.Sprintf("  %.0f%%", cpu)
			panel.SetContent(content)
		}
	}
}

func updateMemStats(ctx context.Context, panel *tui.Panel) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	mem := 40.0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mem += (rand.Float64() - 0.5) * 5
			if mem < 20 {
				mem = 20
			}
			if mem > 90 {
				mem = 90
			}

			barWidth := 16
			filled := int(mem / 100 * float64(barWidth))
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

			content := panel.StyledWith(tui.ThemePrimary, "Memory Usage") + "\n\n" +
				fmt.Sprintf("  %s\n\n", bar) +
				fmt.Sprintf("  %.0f%%", mem)
			panel.SetContent(content)
		}
	}
}
