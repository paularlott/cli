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
	"fmt"
	"os"
	"strings"
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
	mu            sync.Mutex
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
	if cfg.UserLabel == "" {
		cfg.UserLabel = "You"
	}
	if cfg.AssistantLabel == "" {
		cfg.AssistantLabel = "Assistant"
	}
	if cfg.SystemLabel == "" {
		cfg.SystemLabel = "System"
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
	t.mu.Lock()
	defer t.mu.Unlock()
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
	t.mu.Lock()
	defer t.mu.Unlock()
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

	t.resize()
	t.draw()

	// Enable mouse wheel reporting (SGR extended mode).
	fmt.Print("\x1b[?1000h\x1b[?1006h")
	defer fmt.Print("\x1b[?1006l\x1b[?1000l")

	go func() {
		<-ctx.Done()
		t.mu.Lock()
		t.quit = true
		t.mu.Unlock()
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

func (t *TUI) draw() {
	t.resize()

	inputH := t.inputBoxHeight()
	paletteH := t.paletteHeight()
	// Fixed bottom rows: palette + input + charcount + status(1)
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
	buf.WriteString(clearScreen())

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
	// Ctrl+C
	if len(b) == 1 && b[0] == 3 {
		t.quit = true
		return nil
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
		case '2': // Shift+Enter: ESC [ 2 7 ; 2 ; 1 3 ~
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
		// SGR mouse: ESC [ < ... M/m
		if b[2] == '<' {
			s := string(b[3:])
			if len(s) > 0 && (s[len(s)-1] == 'M' || s[len(s)-1] == 'm') {
				parts := strings.SplitN(s[:len(s)-1], ";", 3)
				if len(parts) == 3 {
					switch parts[0] {
					case "64": // wheel up
						t.output.scrollUp(3)
					case "65": // wheel down
						t.output.scrollDown(3)
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
			// Unknown command — add message directly (no lock needed, called unlocked via cb).
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

	// Shift+Enter: ESC \r or kitty ESC [ 1 3 ; 2 u
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
