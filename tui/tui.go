// Package tui provides a full-screen terminal UI framework for building
// interactive CLI applications, inspired by modern AI assistants.
//
// # Layout
//
// The screen is divided into three vertical regions:
//
//	┌─────────────────────────────────────────────────────────┐
//	│  Scrollable output / conversation history               │
//	│  (auto-scrolls to bottom; Page Up/Down/wheel to scroll) │
//	├─────────────────────────────────────────────────────────┤
//	│  Command palette (visible when / is typed)              │
//	├─────────────────────────────────────────────────────────┤
//	│  ┌─────────────────────────────────────────────────┐    │
//	│  │ > input goes here      Ctrl+C to exit           │    │
//	│  └─────────────────────────────────────────────────┘    │
//	│  myapp                                                  │
//	└─────────────────────────────────────────────────────────┘
//
// # Quick Start
//
//	t := tui.New(tui.Config{
//	    Theme: tui.ThemeAmber,
//	    Commands: []*tui.Command{
//	        {Name: "clear", Description: "Clear history", Handler: func(_ string) { t.ClearOutput() }},
//	    },
//	    OnSubmit: func(text string) {
//	        t.AddMessage(tui.RoleUser, text)
//	        t.AddMessage(tui.RoleAssistant, "Echo: "+text)
//	    },
//	})
//	t.Run(context.Background())
//
// # Streaming
//
// For token-by-token responses:
//
//	t.StartStreamingAs("GPT-4o")
//	for chunk := range tokenCh {
//	    t.StreamChunk(chunk)
//	}
//	t.StreamComplete()
//
// # Themes
//
// Seven built-in themes: [ThemeAmber], [ThemeBlue], [ThemeGreen], [ThemePurple],
// [ThemeLight], [ThemePlain], [ThemeDefault]. Look up by name with [ThemeByName].
// Register custom themes with [RegisterTheme] or via [Config.Themes].
package tui

import (
	"context"
	"sync"
	"time"

	"golang.org/x/term"
)

// Config holds the configuration for a TUI instance.
type Config struct {
	// Theme controls colors. Defaults to ThemeAmber if nil.
	Theme *Theme

	// Commands are the slash commands available in the palette.
	Commands []*Command

	// Themes registers additional themes into the global theme registry,
	// making them available via ThemeByName.
	Themes []*Theme

	// OnSubmit is called when the user presses Enter to submit input.
	// The TUI does NOT add a user message automatically; the caller decides.
	OnSubmit func(text string)

	// OnEscape is called when Escape is pressed and the palette is not active.
	OnEscape func()

	// UserLabel is the label shown for user messages. Defaults to "You".
	UserLabel string

	// AssistantLabel is the default label for assistant messages. Defaults to "Assistant".
	AssistantLabel string

	// SystemLabel is the label shown for system messages. Defaults to "System".
	SystemLabel string

	// HideHeaders suppresses the role header line between messages.
	HideHeaders bool

	// StatusLeft is optional text shown in the bottom-left status bar.
	StatusLeft string

	// StatusRight is optional text shown in the bottom-right status bar.
	StatusRight string

	// ShowCharCount enables the character counter below the input box. Defaults to false.
	ShowCharCount bool

	// InputEnabled controls whether the input box is shown. Defaults to true.
	// When false, the input box, char count, and palette are hidden and
	// keyboard input only handles scrolling and Ctrl+C.
	InputEnabled *bool
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// TUI is the main terminal UI instance.
type TUI struct {
	cfg           Config
	theme         *Theme
	output        *outputRegion
	input         *inputArea
	palette       *palette
	width         int
	height        int
	fd            int
	oldState      *term.State
	quit          bool
	mu            sync.RWMutex
	spinnerText   string
	spinnerFrame  int
	spinnerStop   chan struct{}
	progress      float64 // -1 = inactive
	progressLabel string
	ctx           context.Context
	menu          *menuState
}

// New creates a new TUI with the given configuration.
func New(cfg Config) *TUI {
	if cfg.Theme == nil {
		cfg.Theme = ThemeDefault
	}

	for _, th := range cfg.Themes {
		RegisterTheme(th)
	}

	t := &TUI{
		cfg:      cfg,
		theme:    cfg.Theme,
		progress: -1,
		output: &outputRegion{
			userLabel:      cfg.UserLabel,
			assistantLabel: cfg.AssistantLabel,
			systemLabel:    cfg.SystemLabel,
			hideHeaders:    cfg.HideHeaders,
		},
		input: newInputArea(),
	}
	t.palette = newPalette(cfg.Commands)
	return t
}

// OpenMenu opens a navigable menu panel, replacing the input box.
func (t *TUI) OpenMenu(m *Menu) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.menu = newMenuState(m)
	t.draw()
}

// CloseMenu closes the menu and restores the input box.
func (t *TUI) CloseMenu() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.menu = nil
	t.draw()
}

// Context returns the context that was passed to Run.
// Returns nil if Run has not been called yet.
func (t *TUI) Context() context.Context {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ctx
}

// Exit cleanly shuts down the TUI event loop.
// Useful as a /exit command handler: func(_ string) { t.Exit() }
func (t *TUI) Exit() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.quit = true
}

// AddMessageAs appends a complete message with a custom label.
func (t *TUI) AddMessageAs(role MessageRole, label, content string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output.AddMessageAs(role, label, content)
	t.draw()
}

// AddMessage appends a complete message to the output region.
func (t *TUI) AddMessage(role MessageRole, content string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output.AddMessage(role, content)
	t.draw()
}

// IsStreaming returns true if a streaming message is in progress.
func (t *TUI) IsStreaming() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.output.streaming != nil
}

// StartStreaming begins a new streaming assistant message.
func (t *TUI) StartStreaming() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output.StartStreaming()
	t.draw()
}

// StartStreamingAs begins a new streaming assistant message with a custom label.
func (t *TUI) StartStreamingAs(label string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output.StartStreamingAs(label)
	t.draw()
}

// StreamChunk appends a chunk to the current streaming message.
func (t *TUI) StreamChunk(chunk string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output.StreamChunk(chunk)
	t.draw()
}

// StopStreaming finalises any in-progress streaming message.
func (t *TUI) StopStreaming() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.output.streaming != nil {
		t.output.StreamComplete()
		t.draw()
	}
}

// StreamComplete finalises the streaming message.
func (t *TUI) StreamComplete() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output.StreamComplete()
	t.draw()
}

func (t *TUI) inputEnabled() bool {
	return t.cfg.InputEnabled == nil || *t.cfg.InputEnabled
}

// ClearOutput removes all messages from the output region.
func (t *TUI) ClearOutput() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output.Clear()
	t.draw()
}

// refresh redraws the screen.
func (t *TUI) refresh() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.draw()
}

// SetLabels updates the default role labels shown in message headers.
// Empty strings leave the corresponding label unchanged.
func (t *TUI) SetLabels(user, assistant, system string) {
	t.mu.Lock()
	t.output.SetLabels(user, assistant, system)
	t.mu.Unlock()
	t.refresh()
}

// SetStatus updates both status bar texts.
func (t *TUI) SetStatus(left, right string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cfg.StatusLeft = left
	t.cfg.StatusRight = right
	t.draw()
}

// SetStatusLeft updates the left status bar text.
func (t *TUI) SetStatusLeft(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cfg.StatusLeft = s
	t.draw()
}

// SetStatusRight updates the right status bar text.
func (t *TUI) SetStatusRight(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cfg.StatusRight = s
	t.draw()
}

// StartSpinner shows an animated spinner in the separator with the given text.
// Calling StartSpinner while one is running replaces the text.
func (t *TUI) StartSpinner(text string) {
	t.mu.Lock()
	if t.spinnerStop != nil {
		close(t.spinnerStop)
	}
	t.spinnerText = text
	t.spinnerFrame = 0
	t.progress = -1
	stop := make(chan struct{})
	t.spinnerStop = stop
	t.mu.Unlock()

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				t.mu.Lock()
				t.spinnerFrame = (t.spinnerFrame + 1) % len(spinnerFrames)
				t.draw()
				t.mu.Unlock()
			}
		}
	}()
}

// StopSpinner stops the spinner and clears it from the separator.
func (t *TUI) StopSpinner() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.spinnerStop != nil {
		close(t.spinnerStop)
		t.spinnerStop = nil
	}
	t.spinnerText = ""
	t.draw()
}

// SetProgress shows a labelled progress bar in the separator (0.0–1.0).
// Stops any active spinner. Call ClearProgress to remove it.
func (t *TUI) SetProgress(label string, value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.spinnerStop != nil {
		close(t.spinnerStop)
		t.spinnerStop = nil
		t.spinnerText = ""
	}
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	t.progress = value
	t.progressLabel = label
	t.draw()
}

// ClearProgress removes the progress bar from the separator.
func (t *TUI) ClearProgress() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.progress = -1
	t.draw()
}

// SetInputEnabled toggles the input box at runtime.
func (t *TUI) SetInputEnabled(enabled bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cfg.InputEnabled = &enabled
	t.draw()
}

// AddCommand registers a new slash command at runtime.
func (t *TUI) AddCommand(cmd *Command) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.palette.commands = append(t.palette.commands, cmd)
}

// RemoveCommand removes a slash command by name.
func (t *TUI) RemoveCommand(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cmds := t.palette.commands[:0]
	for _, c := range t.palette.commands {
		if c.Name != name {
			cmds = append(cmds, c)
		}
	}
	t.palette.commands = cmds
}

// SetTheme changes the active theme.
func (t *TUI) SetTheme(theme *Theme) {
	if theme == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.theme = theme
	t.draw()
}

// flashCopied briefly shows "✓ Copied" in the input overlay then clears selection.
func (t *TUI) flashCopied() {
	t.spinnerText = "✓ Copied"
	go func() {
		time.Sleep(1500 * time.Millisecond)
		t.mu.Lock()
		if t.spinnerText == "✓ Copied" {
			t.spinnerText = ""
			t.output.sel = nil
			t.draw()
		}
		t.mu.Unlock()
	}()
}
