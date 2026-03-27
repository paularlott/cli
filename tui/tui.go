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

	// OnFocusChange is called when panel focus changes via Tab cycling.
	// The callback receives the newly focused panel.
	OnFocusChange func(panel *Panel)
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
	prevWidth     int // Previous width for resize detection
	prevHeight    int // Previous height for resize detection
	prevOutputH   int // Previous output height for layout change detection
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

	// Panel support
	panels        map[string]*Panel
	mainPanel     *Panel // Cached main panel for legacy methods
	layout        LayoutConfig
	focusIdx      int // Index of focused panel (0=left, 1=main, 2=right)
	panelColorIdx int // For auto-assigning panel colors
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
		panels:   make(map[string]*Panel),
		output: &outputRegion{
			userLabel:      cfg.UserLabel,
			assistantLabel: cfg.AssistantLabel,
			systemLabel:    cfg.SystemLabel,
			hideHeaders:    cfg.HideHeaders,
		},
		input: newInputArea(),
	}

	// Create main panel wrapping the existing output region
	t.mainPanel = &Panel{
		name:   "main",
		tui:    t,
		region: t.output,
		color:  cfg.Theme.Primary,
	}
	t.panels["main"] = t.mainPanel

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

// TerminalSize returns the full terminal dimensions (width, height).
// Returns 0, 0 if the terminal size has not been determined yet.
func (t *TUI) TerminalSize() (width, height int) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.width, t.height
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
	t.mainPanel.AddMessageAs(role, label, content)
}

// AddMessage appends a complete message to the output region.
func (t *TUI) AddMessage(role MessageRole, content string) {
	t.mainPanel.AddMessage(role, content)
}

// IsStreaming returns true if a streaming message is in progress.
func (t *TUI) IsStreaming() bool {
	return t.mainPanel.IsStreaming()
}

// StartStreaming begins a new streaming assistant message.
func (t *TUI) StartStreaming() {
	t.mainPanel.StartStreaming()
}

// StartStreamingAs begins a new streaming assistant message with a custom label.
func (t *TUI) StartStreamingAs(label string) {
	t.mainPanel.StartStreamingAs(label)
}

// StreamChunk appends a chunk to the current streaming message.
func (t *TUI) StreamChunk(chunk string) {
	t.mainPanel.StreamChunk(chunk)
}

// StopStreaming finalises any in-progress streaming message.
func (t *TUI) StopStreaming() {
	t.mainPanel.StopStreaming()
}

// StreamComplete finalises the streaming message.
func (t *TUI) StreamComplete() {
	t.mainPanel.StreamComplete()
}

func (t *TUI) inputEnabled() bool {
	return t.cfg.InputEnabled == nil || *t.cfg.InputEnabled
}

// ClearOutput removes all messages from the output region.
func (t *TUI) ClearOutput() {
	t.mainPanel.Clear()
}

func (t *TUI) WriteString(s string) {
	t.mainPanel.WriteString(s)
}

// SetLabels updates the default role labels shown in message headers.
// Empty strings leave the corresponding label unchanged.
func (t *TUI) SetLabels(user, assistant, system string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output.SetLabels(user, assistant, system)
	t.draw()
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
	t.progress = min(1, max(0, value))
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
// If region is provided, clears that region's selection; otherwise clears main output.
func (t *TUI) flashCopied(region *outputRegion) {
	t.spinnerText = "✓ Copied"
	go func() {
		time.Sleep(1500 * time.Millisecond)
		t.mu.Lock()
		if t.spinnerText == "✓ Copied" {
			t.spinnerText = ""
			if region != nil {
				region.sel = nil
			}
			t.draw()
		}
		t.mu.Unlock()
	}()
}

// redraw acquires the lock and redraws the screen.
func (t *TUI) redraw() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.draw()
}

// Panel returns the panel with the given name, creating it if it doesn't exist.
// The special name "main" returns the main panel.
func (t *TUI) Panel(name string) *Panel {
	t.mu.Lock()
	defer t.mu.Unlock()

	if p, ok := t.panels[name]; ok {
		return p
	}

	// Create new panel with auto-assigned color
	p := t.createPanel(PanelConfig{Name: name})
	t.panels[name] = p
	return p
}

// createPanel creates a new panel with auto-assigned color.
// Must be called with t.mu held.
func (t *TUI) createPanel(cfg PanelConfig) *Panel {
	colors := []Color{t.theme.Primary, t.theme.Secondary}
	color := colors[t.panelColorIdx%len(colors)]
	t.panelColorIdx++
	return newPanel(cfg, t, color)
}

// SetLayout configures the panel layout.
func (t *TUI) SetLayout(cfg LayoutConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.layout = cfg

	// Create panels defined in layout (if not already created)
	// Skip split parents - they're containers, not renderable panels
	if cfg.Left != nil && cfg.Left.Top == nil && cfg.Left.Bottom == nil {
		if _, ok := t.panels[cfg.Left.Name]; !ok {
			t.panels[cfg.Left.Name] = t.createPanel(*cfg.Left)
		}
	}

	if cfg.Right != nil && cfg.Right.Top == nil && cfg.Right.Bottom == nil {
		if _, ok := t.panels[cfg.Right.Name]; !ok {
			t.panels[cfg.Right.Name] = t.createPanel(*cfg.Right)
		}
	}

	// Set initial focus to main panel (after all left panels)
	t.focusIdx = t.countPanelChildren(t.layout.Left)

	t.draw()
}

// FocusPanel sets the focused panel by index.
func (t *TUI) FocusPanel(idx int) {
	var cb func(*Panel)
	var panel *Panel

	t.mu.Lock()
	visible := t.countFocusablePanels()
	if idx >= 0 && idx < visible && idx != t.focusIdx {
		t.focusIdx = idx
		t.draw()
		if t.cfg.OnFocusChange != nil {
			cb = t.cfg.OnFocusChange
			panel = t.focusedPanelPtr()
		}
	}
	t.mu.Unlock()

	if cb != nil && panel != nil {
		cb(panel)
	}
}

// FocusedPanel returns the currently focused panel.
// Returns nil if no panel is focused (shouldn't happen in normal use).
func (t *TUI) FocusedPanel() *Panel {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.focusedPanelPtr()
}

// focusedPanelPtr returns the focused panel's *Panel. Must be called with t.mu held.
func (t *TUI) focusedPanelPtr() *Panel {
	idx := 0

	// Helper to check focus match and return panel
	checkFocus := func(cfg *PanelConfig) *Panel {
		if cfg.SkipFocus {
			return nil
		}
		if t.focusIdx == idx {
			return t.panels[cfg.Name]
		}
		idx++
		return nil
	}

	// Check left panel and its children
	if t.layout.Left != nil {
		if t.layout.Left.Top != nil {
			if p := checkFocus(t.layout.Left.Top); p != nil {
				return p
			}
		}
		if t.layout.Left.Bottom != nil {
			if p := checkFocus(t.layout.Left.Bottom); p != nil {
				return p
			}
		}
		if t.layout.Left.Top == nil && t.layout.Left.Bottom == nil {
			if p := checkFocus(t.layout.Left); p != nil {
				return p
			}
		}
	}

	// Main panel (always focusable)
	if t.focusIdx == idx {
		return t.mainPanel
	}
	idx++

	// Check right panel and its children
	if t.layout.Right != nil {
		if t.layout.Right.Top != nil {
			if p := checkFocus(t.layout.Right.Top); p != nil {
				return p
			}
		}
		if t.layout.Right.Bottom != nil {
			if p := checkFocus(t.layout.Right.Bottom); p != nil {
				return p
			}
		}
		if t.layout.Right.Top == nil && t.layout.Right.Bottom == nil {
			if p := checkFocus(t.layout.Right); p != nil {
				return p
			}
		}
	}

	// Default to main panel
	return t.mainPanel
}

// CycleFocus moves focus to the next panel.
func (t *TUI) CycleFocus() {
	t.mu.Lock()
	cb := t.cycleFocusLocked()
	t.mu.Unlock()
	if cb != nil {
		cb()
	}
}

// countFocusablePanels counts all focusable panels including children.
func (t *TUI) countFocusablePanels() int {
	count := 1 // main always focusable

	if t.layout.Left != nil {
		count += t.countPanelChildren(t.layout.Left)
	}
	if t.layout.Right != nil {
		count += t.countPanelChildren(t.layout.Right)
	}

	return count
}

// countPanelChildren counts focusable panels (excluding SkipFocus ones).
func (t *TUI) countPanelChildren(cfg *PanelConfig) int {
	if cfg == nil {
		return 0
	}
	if cfg.Top != nil || cfg.Bottom != nil {
		count := 0
		if cfg.Top != nil && !cfg.Top.SkipFocus {
			count++
		}
		if cfg.Bottom != nil && !cfg.Bottom.SkipFocus {
			count++
		}
		return count
	}
	if cfg.SkipFocus {
		return 0
	}
	return 1
}

// cycleFocusLocked moves focus to the next panel. Must be called with t.mu held.
// Returns a callback to invoke after releasing the lock, or nil.
func (t *TUI) cycleFocusLocked() func() {
	visible := t.countFocusablePanels()
	t.focusIdx = (t.focusIdx + 1) % visible
	t.draw()

	if t.cfg.OnFocusChange != nil {
		panel := t.focusedPanelPtr()
		return func() { t.cfg.OnFocusChange(panel) }
	}
	return nil
}

// hasMultiplePanels returns true if there's more than one visible panel.
func (t *TUI) hasMultiplePanels() bool {
	return t.layout.Left != nil || t.layout.Right != nil
}

// HasMultiplePanels returns true if there's more than one visible panel.
func (t *TUI) HasMultiplePanels() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.hasMultiplePanels()
}

// focusedPanel returns the currently focused panel's output region.
// Must be called with t.mu held.
func (t *TUI) focusedPanel() *outputRegion {
	idx := 0

	// Helper to advance index and check focus match
	checkFocus := func(cfg *PanelConfig) *outputRegion {
		if cfg.SkipFocus {
			return nil
		}
		if t.focusIdx == idx {
			if p := t.panels[cfg.Name]; p != nil {
				return p.region
			}
		}
		idx++
		return nil
	}

	// Check left panel and its children
	if t.layout.Left != nil {
		if t.layout.Left.Top != nil {
			if r := checkFocus(t.layout.Left.Top); r != nil {
				return r
			}
		}
		if t.layout.Left.Bottom != nil {
			if r := checkFocus(t.layout.Left.Bottom); r != nil {
				return r
			}
		}
		if t.layout.Left.Top == nil && t.layout.Left.Bottom == nil {
			if r := checkFocus(t.layout.Left); r != nil {
				return r
			}
		}
	}

	// Main panel (always focusable)
	if t.focusIdx == idx {
		return t.output
	}
	idx++

	// Check right panel and its children
	if t.layout.Right != nil {
		if t.layout.Right.Top != nil {
			if r := checkFocus(t.layout.Right.Top); r != nil {
				return r
			}
		}
		if t.layout.Right.Bottom != nil {
			if r := checkFocus(t.layout.Right.Bottom); r != nil {
				return r
			}
		}
		if t.layout.Right.Top == nil && t.layout.Right.Bottom == nil {
			if r := checkFocus(t.layout.Right); r != nil {
				return r
			}
		}
	}

	// Default to main output
	return t.output
}
