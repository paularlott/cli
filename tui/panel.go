package tui

import (
	"io"
	"strings"
	"sync"
	"unicode/utf8"
)

// maxScrollOff is a sentinel value meaning "scroll to top".
// It will be clamped to the actual max offset during render.
const maxScrollOff = 1 << 30

// PanelConfig configures a panel in the layout.
type PanelConfig struct {
	Name       string        // Panel identifier
	Width      int           // Positive = columns, negative = percentage (e.g., -30 = 30%)
	Height     int           // Positive = rows, negative = percentage (e.g., -50 = 50%). Used when panel is split vertically
	MinWidth   int           // Minimum width to render; 0 = always show
	Scrollable bool          // True = content scrolls, false = fixed viewport
	Title      string        // Optional title for border; empty = no title
	Color      *Color        // Optional border/accent color; nil = auto-assign
	NoBorder   bool          // True = hide border for this panel
	SkipFocus  bool          // True = exclude from Tab focus cycle
	Top        *PanelConfig  // Optional top child panel (horizontal split)
	Bottom     *PanelConfig  // Optional bottom child panel (horizontal split)
}

// LayoutConfig configures the panel layout.
type LayoutConfig struct {
	Left   *PanelConfig // nil = no left panel
	Right  *PanelConfig // nil = no right panel
}

// Panel represents a content area within the TUI.
// It implements io.Writer for easy integration with other systems.
type Panel struct {
	name       string
	tui        *TUI
	region     *outputRegion
	title      string
	color      Color
	noBorder   bool
	scrollable bool
	skipFocus  bool
	mu         sync.Mutex

	// Last rendered dimensions (updated during render)
	width  int
	height int

	// Low-level content mode (alternative to message mode)
	rawLines []string
	rawMode  bool
}

// newPanel creates a new panel with the given configuration.
func newPanel(cfg PanelConfig, t *TUI, color Color) *Panel {
	p := &Panel{
		name:       cfg.Name,
		tui:        t,
		title:      cfg.Title,
		color:      color,
		noBorder:   cfg.NoBorder,
		scrollable: cfg.Scrollable,
		skipFocus:  cfg.SkipFocus,
		region: &outputRegion{
			userLabel:      "You",
			assistantLabel: "Assistant",
			systemLabel:    "System",
		},
	}
	if cfg.Color != nil {
		p.color = *cfg.Color
	}
	return p
}

// Name returns the panel's identifier.
func (p *Panel) Name() string {
	return p.name
}

// --- io.Writer interface ---

// Write implements io.Writer. It appends bytes to the panel content.
func (p *Panel) Write(b []byte) (int, error) {
	p.WriteString(string(b))
	return len(b), nil
}

// WriteString appends a string to the panel content.
func (p *Panel) WriteString(s string) {
	p.mu.Lock()
	// Use raw mode for direct writes
	p.rawMode = true
	for _, line := range strings.Split(s, "\n") {
		p.rawLines = append(p.rawLines, line)
	}
	// Remove trailing empty line if present (from trailing newline)
	if len(p.rawLines) > 0 && p.rawLines[len(p.rawLines)-1] == "" {
		p.rawLines = p.rawLines[:len(p.rawLines)-1]
	}
	p.mu.Unlock()

	if p.tui != nil {
		p.tui.redraw()
	}
}

// --- Message API (same as TUI output) ---

// AddMessage appends a complete message with role.
func (p *Panel) AddMessage(role MessageRole, content string) {
	p.mu.Lock()
	p.rawMode = false
	p.region.AddMessage(role, content)
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// AddMessageAs appends a complete message with a custom label.
func (p *Panel) AddMessageAs(role MessageRole, label, content string) {
	p.mu.Lock()
	p.rawMode = false
	p.region.AddMessageAs(role, label, content)
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// StartStreaming begins a new assistant message built incrementally.
func (p *Panel) StartStreaming() {
	p.mu.Lock()
	p.rawMode = false
	p.region.StartStreaming()
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// StartStreamingAs begins a new streaming message with a custom label.
func (p *Panel) StartStreamingAs(label string) {
	p.mu.Lock()
	p.rawMode = false
	p.region.StartStreamingAs(label)
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// StreamChunk appends a chunk to the in-progress streaming message.
func (p *Panel) StreamChunk(chunk string) {
	p.mu.Lock()
	p.region.StreamChunk(chunk)
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// StreamComplete finalises the streaming message.
func (p *Panel) StreamComplete() {
	p.mu.Lock()
	p.region.StreamComplete()
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// IsStreaming returns true if a streaming message is in progress.
func (p *Panel) IsStreaming() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.region.streaming != nil
}

// StopStreaming finalises any in-progress streaming message.
func (p *Panel) StopStreaming() {
	p.mu.Lock()
	if p.region.streaming != nil {
		p.region.StreamComplete()
	}
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// Clear removes all content from the panel.
func (p *Panel) Clear() {
	p.mu.Lock()
	p.region.Clear()
	p.rawLines = nil
	p.rawMode = false
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// --- Scroll Control ---

// ScrollToTop scrolls to the top of the content.
func (p *Panel) ScrollToTop() {
	p.mu.Lock()
	// Set a large value that will be clamped to maxOff during render.
	// scrollOff = 0 means bottom, higher values scroll toward top.
	p.region.scrollOff = maxScrollOff
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// ScrollToBottom scrolls to the bottom of the content.
func (p *Panel) ScrollToBottom() {
	p.mu.Lock()
	p.region.scrollOff = 0
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// ScrollUp scrolls up by n lines.
func (p *Panel) ScrollUp(n int) {
	p.mu.Lock()
	p.region.scrollUp(n)
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// ScrollDown scrolls down by n lines.
func (p *Panel) ScrollDown(n int) {
	p.mu.Lock()
	p.region.scrollDown(n)
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// ScrollOffset returns the current scroll offset.
func (p *Panel) ScrollOffset() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.region.scrollOff
}

// --- Size Information ---

// Size returns the panel's content dimensions (width, height).
// Returns 0, 0 if the panel has not been rendered yet.
func (p *Panel) Size() (width, height int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.width, p.height
}

// ContentLines returns the total number of content lines.
func (p *Panel) ContentLines() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.rawMode {
		return len(p.rawLines)
	}
	return len(p.region.lastLines)
}

// --- Low-Level Output ---

// WriteAt writes text at a specific position (0-indexed, relative to panel).
// If the position is outside current content, content is extended with blank lines.
func (p *Panel) WriteAt(row, col int, s string) {
	p.mu.Lock()
	p.rawMode = true

	// Extend rawLines if needed
	for len(p.rawLines) <= row {
		p.rawLines = append(p.rawLines, "")
	}

	line := p.rawLines[row]
	runes := []rune(line)

	// Extend line if needed
	for len(runes) < col {
		runes = append(runes, ' ')
	}

	// Write the string at position
	sRunes := []rune(s)
	if col+len(sRunes) > len(runes) {
		runes = append(runes[:col], sRunes...)
	} else {
		copy(runes[col:], sRunes)
	}

	p.rawLines[row] = string(runes)
	p.mu.Unlock()

	if p.tui != nil {
		p.tui.redraw()
	}
}

// ClearLine clears a specific line.
func (p *Panel) ClearLine(row int) {
	p.mu.Lock()
	p.rawMode = true

	if row >= 0 && row < len(p.rawLines) {
		p.rawLines[row] = ""
	}
	p.mu.Unlock()

	if p.tui != nil {
		p.tui.redraw()
	}
}

// ClearRegion clears a rectangular region.
func (p *Panel) ClearRegion(startRow, startCol, endRow, endCol int) {
	p.mu.Lock()
	p.rawMode = true

	for row := startRow; row <= endRow && row < len(p.rawLines); row++ {
		line := p.rawLines[row]
		runes := []rune(line)

		cs := startCol
		ce := endCol
		if row > startRow {
			cs = 0
		}
		if row < endRow {
			ce = len(runes) - 1
		}

		for i := cs; i <= ce && i < len(runes); i++ {
			runes[i] = ' '
		}

		p.rawLines[row] = string(runes)
	}
	p.mu.Unlock()

	if p.tui != nil {
		p.tui.redraw()
	}
}

// SetContent replaces all content with the given string.
func (p *Panel) SetContent(s string) {
	p.mu.Lock()
	p.rawMode = true
	p.rawLines = strings.Split(s, "\n")
	p.mu.Unlock()

	if p.tui != nil {
		p.tui.redraw()
	}
}

// --- Title & Appearance ---

// SetTitle sets the title shown in the border. Empty string = no title.
func (p *Panel) SetTitle(title string) {
	p.mu.Lock()
	p.title = title
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// Title returns the current title (empty string if unset).
func (p *Panel) Title() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.title
}

// SetBorder enables or disables the border for this panel.
func (p *Panel) SetBorder(enabled bool) {
	p.mu.Lock()
	p.noBorder = !enabled
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// HasBorder returns true if the border is enabled.
func (p *Panel) HasBorder() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return !p.noBorder
}

// SetColor sets the panel's border/accent color.
func (p *Panel) SetColor(color Color) {
	p.mu.Lock()
	p.color = color
	p.mu.Unlock()
	if p.tui != nil {
		p.tui.redraw()
	}
}

// Color returns the panel's current color.
func (p *Panel) Color() Color {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.color
}

// SetScrollable sets whether the panel content scrolls.
func (p *Panel) SetScrollable(scrollable bool) {
	p.mu.Lock()
	p.scrollable = scrollable
	p.mu.Unlock()
}

// Scrollable returns whether the panel is scrollable.
func (p *Panel) Scrollable() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.scrollable
}

// SetSkipFocus sets whether the panel is excluded from Tab focus cycling.
func (p *Panel) SetSkipFocus(skip bool) {
	p.mu.Lock()
	p.skipFocus = skip
	p.mu.Unlock()
}

// SkipFocus returns whether the panel is excluded from focus cycling.
func (p *Panel) SkipFocus() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.skipFocus
}

// --- Styling ---

// Styled returns text wrapped in the given color.
func (p *Panel) Styled(color Color, text string) string {
	return Styled(color, text)
}

// StyledWith returns text wrapped in a named theme color.
// Name can be: "primary", "secondary", "error", "dim", "text", "user"
func (p *Panel) StyledWith(name string, text string) string {
	if p.tui == nil {
		return text
	}
	theme := p.tui.Theme()
	if theme == nil {
		return text
	}

	var c Color
	switch name {
	case "primary":
		c = theme.Primary
	case "secondary":
		c = theme.Secondary
	case "error":
		c = theme.Error
	case "dim":
		c = theme.Dim
	case "text":
		c = theme.Text
	case "user":
		c = theme.UserText
	default:
		return text
	}
	return Styled(c, text)
}

// --- Internal rendering ---

// render draws the panel content into buf.
// Called with TUI lock held; acquires Panel.mu for data access.
func (p *Panel) render(buf *strings.Builder, theme *Theme, width, height, startRow, startCol int, focused bool) {
	if width <= 0 || height <= 0 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Store the rendered dimensions
	p.width = width
	p.height = height

	if p.rawMode {
		p.renderRawLocked(buf, theme, width, height, startRow, startCol)
	} else {
		p.region.renderAt(buf, theme, width, height, startRow, startCol)
	}
}

// renderRawLocked renders raw lines content.
// Must be called with both TUI lock and Panel.mu held.
func (p *Panel) renderRawLocked(buf *strings.Builder, theme *Theme, width, height, startRow, startCol int) {
	total := len(p.rawLines)

	// Calculate scroll offset for raw mode
	maxOff := max(0, total-height)
	scrollOff := 0
	if !p.scrollable && total > height {
		// Fixed mode: show last N lines
		scrollOff = 0
	} else if p.region.scrollOff > 0 {
		scrollOff = min(p.region.scrollOff, maxOff)
	}

	start := max(0, total-height-scrollOff)
	end := start + height
	if end > total {
		end = total
	}

	spaces := strings.Repeat(" ", width)
	for i := start; i < end; i++ {
		row := i - start
		buf.WriteString(cursorPos(startRow+row, startCol))
		if i < len(p.rawLines) {
			line := p.rawLines[i]
			if utf8.RuneCountInString(line) > width {
				line = truncate(line, width)
			}
			buf.WriteString(line)
			// Pad with spaces to fill width
			if pad := width - utf8.RuneCountInString(line); pad > 0 {
				buf.WriteString(spaces[:pad])
			}
		} else {
			buf.WriteString(spaces)
		}
	}

	// Clear remaining lines with spaces
	for i := end - start; i < height; i++ {
		buf.WriteString(cursorPos(startRow+i, startCol))
		buf.WriteString(spaces)
	}
}

// Ensure Panel implements io.Writer
var _ io.Writer = (*Panel)(nil)
