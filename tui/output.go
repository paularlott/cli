package tui

import (
	"strings"
	"unicode/utf8"
)

// selAnchor is a point in the rendered line grid (0-based).
type selAnchor struct {
	row, col int
}

// MessageRole identifies who sent a message.
type MessageRole int

const (
	RoleAssistant MessageRole = iota
	RoleUser
	RoleSystem
)

type message struct {
	role    MessageRole
	content string
	label   string // overrides default role label if set
}

type outputRegion struct {
	messages       []*message
	streaming      *message
	scrollOff      int
	userLabel      string
	assistantLabel string
	systemLabel    string
	hideHeaders    bool
	// selection
	sel       *[2]selAnchor // nil = no selection; [0]=anchor [1]=current
	lastLines []string      // rendered lines from last render() call
	lastStart int           // index into lastLines of first visible row
}

// AddMessage appends a complete message.
func (o *outputRegion) AddMessage(role MessageRole, content string) {
	o.messages = append(o.messages, &message{role: role, content: content})
}

// AddMessageAs appends a complete message with a custom label.
func (o *outputRegion) AddMessageAs(role MessageRole, label, content string) {
	o.messages = append(o.messages, &message{role: role, label: label, content: content})
}

// StartStreaming begins a new assistant message built incrementally.
func (o *outputRegion) StartStreaming() {
	o.streaming = &message{role: RoleAssistant}
}

// StartStreamingAs begins a new assistant message with a custom label.
func (o *outputRegion) StartStreamingAs(label string) {
	o.streaming = &message{role: RoleAssistant, label: label}
}

// StreamChunk appends a chunk to the in-progress streaming message.
func (o *outputRegion) StreamChunk(chunk string) {
	if o.streaming != nil {
		o.streaming.content += chunk
	}
}

// StreamComplete finalises the streaming message.
func (o *outputRegion) StreamComplete() {
	if o.streaming != nil {
		o.messages = append(o.messages, o.streaming)
		o.streaming = nil
		// Don't reset scrollOff — keep view stable if user scrolled up.
	}
}

// Clear removes all messages.
func (o *outputRegion) Clear() {
	o.messages = nil
	o.streaming = nil
	o.scrollOff = 0
	o.sel = nil
	o.lastLines = nil
	o.lastStart = 0
}

// SetLabels updates the default role labels.
func (o *outputRegion) SetLabels(user, assistant, system string) {
	if user != "" {
		o.userLabel = user
	}
	if assistant != "" {
		o.assistantLabel = assistant
	}
	if system != "" {
		o.systemLabel = system
	}
}

func (o *outputRegion) scrollUp(n int)   { o.scrollOff += n }
func (o *outputRegion) scrollDown(n int) { o.scrollOff = max(0, o.scrollOff-n) }

// render draws the output region into buf, using height terminal rows of width w.
// startRow is the 1-based terminal row where the region begins.
func (o *outputRegion) render(buf *strings.Builder, t *Theme, w, height, startRow int) {
	o.renderAt(buf, t, w, height, startRow, 1)
}

// renderAt draws the output region at a specific column position.
// startRow and startCol are 1-based terminal positions.
func (o *outputRegion) renderAt(buf *strings.Builder, t *Theme, w, height, startRow, startCol int) {
	lineW := w
	spaces := strings.Repeat(" ", lineW)

	all := o.messages
	if o.streaming != nil {
		all = append(o.messages[:len(o.messages):len(o.messages)], o.streaming)
	}

	var lines []string
	for _, m := range all {
		lines = append(lines, renderMessage(m, t, lineW, o.userLabel, o.assistantLabel, o.systemLabel, o.hideHeaders)...)
	}

	// Only update lastLines/lastStart if there's no active selection.
	// This freezes the selection target during drag operations on dynamic content.
	if o.sel == nil {
		o.lastLines = lines
	}

	total := len(lines)
	maxOff := max(0, total-height)
	o.scrollOff = min(o.scrollOff, maxOff)

	start := max(0, total-height-o.scrollOff)
	if o.sel == nil {
		o.lastStart = start
	}
	end := min(total, start+height)

	// Normalise selection so a0 <= a1.
	var selA, selB selAnchor
	hasSelection := o.sel != nil
	if hasSelection {
		a, b := o.sel[0], o.sel[1]
		if a.row > b.row || (a.row == b.row && a.col > b.col) {
			a, b = b, a
		}
		selA, selB = a, b
	}

	for i := start; i < end; i++ {
		row := i - start
		buf.WriteString(cursorPos(startRow+row, startCol))
		buf.WriteString(linkReset)
		line := truncate(lines[i], lineW)
		if hasSelection && i >= selA.row && i <= selB.row {
			line = applyHighlight(line, selA, selB, i)
		}
		buf.WriteString(line)
		// Pad with spaces to fill the width (don't use clearLine - it clears to end of terminal)
		if pad := lineW - visibleLen(line); pad > 0 {
			buf.WriteString(spaces[:pad])
		}
	}
	for i := end - start; i < height; i++ {
		buf.WriteString(cursorPos(startRow+i, startCol))
		buf.WriteString(linkReset)
		// Pad with spaces instead of clearing the line
		buf.WriteString(spaces)
	}
}

// applyHighlight wraps the selected column range of a rendered line with reverse video.
// row, selA, selB are all indices into lastLines (0-based absolute).
func applyHighlight(line string, selA, selB selAnchor, row int) string {
	plain := stripANSI(line)
	runes := []rune(plain)
	n := len(runes)

	colStart := 0
	if row == selA.row {
		colStart = selA.col
	}
	colEnd := n
	if row == selB.row {
		colEnd = selB.col + 1
	}
	if colStart > n {
		colStart = n
	}
	if colEnd > n {
		colEnd = n
	}
	if colStart >= colEnd {
		return line
	}

	var b strings.Builder
	b.Grow(n + 10)
	b.WriteString(string(runes[:colStart]))
	b.WriteString("\x1b[7m") // reverse video on
	b.WriteString(string(runes[colStart:colEnd]))
	b.WriteString("\x1b[27m") // reverse video off
	b.WriteString(string(runes[colEnd:]))
	return b.String()
}

// selectionText returns the plain text covered by the current selection, or "".
func (o *outputRegion) selectionText() string {
	if o.sel == nil || len(o.lastLines) == 0 {
		return ""
	}
	a, b := o.sel[0], o.sel[1]
	if a.row > b.row || (a.row == b.row && a.col > b.col) {
		a, b = b, a
	}
	var parts []string
	for r := a.row; r <= b.row && r < len(o.lastLines); r++ {
		plain := stripANSI(o.lastLines[r])
		runes := []rune(plain)
		n := len(runes)
		cs := 0
		if r == a.row {
			cs = a.col
		}
		ce := n
		if r == b.row {
			ce = b.col + 1
		}
		if cs > n {
			cs = n
		}
		if ce > n {
			ce = n
		}
		if cs < ce {
			parts = append(parts, string(runes[cs:ce]))
		}
	}
	return strings.Join(parts, "\n")
}

// renderMessage converts a message to a slice of pre-rendered lines.
func renderMessage(m *message, t *Theme, w int, userLabel, assistantLabel, systemLabel string, hideHeaders bool) []string {
	var lines []string

	if !hideHeaders {
		header := roleHeader(m, t, w, userLabel, assistantLabel, systemLabel)
		if header != "" {
			lines = append(lines, "")
			lines = append(lines, header)
			lines = append(lines, "")
		}
	}
	// Content — handle code blocks.
	content := m.content
	for len(content) > 0 {
		idx := strings.Index(content, "```")
		if idx == -1 {
			lines = append(lines, renderText(content, t, m.role, w)...)
			break
		}
		if idx > 0 {
			lines = append(lines, renderText(content[:idx], t, m.role, w)...)
		}
		content = content[idx+3:]
		end := strings.Index(content, "```")
		if end == -1 {
			// Unclosed block — treat rest as code.
			lines = append(lines, renderCodeBlock(content, t, w)...)
			break
		}
		block := content[:end]
		// Strip optional language tag on first line.
		if nl := strings.IndexByte(block, '\n'); nl != -1 {
			block = block[nl+1:]
		}
		lines = append(lines, renderCodeBlock(block, t, w)...)
		content = content[end+3:]
	}

	lines = append(lines, "")
	return lines
}

func roleHeader(m *message, t *Theme, w int, userLabel, assistantLabel, systemLabel string) string {
	var label string
	if m.label != "" {
		label = m.label
	} else {
		switch m.role {
		case RoleAssistant:
			label = assistantLabel
		case RoleUser:
			label = userLabel
		default:
			label = systemLabel
		}
	}
	if label == "" {
		return ""
	}
	label = " " + label + " "
	fill := max(0, w-utf8.RuneCountInString(label)-4)
	var b strings.Builder
	b.WriteString(fg(t.Dim))
	b.WriteString("━━")
	b.WriteString(reset)
	b.WriteString(fg(t.Primary))
	b.WriteString(bold())
	b.WriteString(label)
	b.WriteString(reset)
	b.WriteString(fg(t.Dim))
	b.WriteString(strings.Repeat("━", fill))
	b.WriteString(reset)
	return b.String()
}

func renderText(text string, t *Theme, role MessageRole, w int) []string {
	var lines []string
	c := fg(t.Text)
	if role == RoleUser {
		c = fg(t.UserText)
	}
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		for _, wrapped := range wordWrap(line, w) {
			var b strings.Builder
			b.WriteString(c)
			b.WriteString(wrapped)
			b.WriteString(reset)
			lines = append(lines, b.String())
		}
	}
	return lines
}

// scanEscape returns the end index of the escape sequence starting at s[i]
// where s[i] == '\x1b'. i must be a valid index into s.
func scanEscape(s string, i int) int {
	i++ // skip ESC
	if i >= len(s) {
		return i
	}
	switch s[i] {
	case '[': // CSI: ESC [ ... final-byte (letter)
		i++
		for i < len(s) {
			c := s[i]
			i++
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				break
			}
		}
	case ']': // OSC: ESC ] ... BEL  or  ESC ] ... ST (ESC \)
		i++
		for i < len(s) {
			if s[i] == '\x07' {
				i++
				break
			}
			if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '\\' {
				i += 2
				break
			}
			i++
		}
	default: // two-char sequence (SS3, etc.)
		i++
	}
	return i
}

// wordWrap splits a string into lines of at most w visible runes, breaking on spaces.
// Escape sequences (CSI and OSC, including hyperlinks) are treated as zero-width.
func wordWrap(s string, w int) []string {
	if w <= 0 || visibleLen(s) <= w {
		return []string{s}
	}
	tokens := splitTokens(s)
	var lines []string
	var current strings.Builder
	currentW := 0
	for _, tok := range tokens {
		if tok == " " {
			if current.Len() > 0 {
				current.WriteByte(' ')
				currentW++
			}
			continue
		}
		tokW := visibleLen(tok)
		if current.Len() == 0 {
			if tokW > w {
				lines = append(lines, splitHardWrap(tok, w)...)
				continue
			}
			current.WriteString(tok)
			currentW = tokW
		} else if currentW+1+tokW <= w {
			current.WriteString(tok)
			currentW += tokW
		} else {
			lines = append(lines, strings.TrimRight(current.String(), " "))
			// Token doesn't fit on current line - check if it needs hard-wrapping
			if tokW > w {
				lines = append(lines, splitHardWrap(tok, w)...)
				current.Reset()
				currentW = 0
			} else {
				current.Reset()
				current.WriteString(tok)
				currentW = tokW
			}
		}
	}
	if current.Len() > 0 {
		lines = append(lines, strings.TrimRight(current.String(), " "))
	}
	return lines
}

// splitTokens splits s into space tokens (" ") and non-space tokens, keeping
// escape sequences attached to adjacent visible text.
func splitTokens(s string) []string {
	var tokens []string
	var cur strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == ' ' {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
			tokens = append(tokens, " ")
			i++
			continue
		}
		if s[i] == '\x1b' {
			end := scanEscape(s, i)
			cur.WriteString(s[i:end])
			i = end
			continue
		}
		cur.WriteByte(s[i])
		i++
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

// splitHardWrap hard-breaks s at w visible runes per line.
func splitHardWrap(s string, w int) []string {
	var lines []string
	for visibleLen(s) > w {
		lines = append(lines, truncate(s, w))
		s = skipVisible(s, w)
	}
	if s != "" {
		lines = append(lines, s)
	}
	return lines
}

// skipVisible advances past n visible runes in s, skipping escape sequences.
func skipVisible(s string, n int) string {
	count := 0
	i := 0
	for i < len(s) && count < n {
		if s[i] == '\x1b' {
			i = scanEscape(s, i)
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		count++
	}
	return s[i:]
}

func renderCodeBlock(code string, t *Theme, w int) []string {
	var lines []string
	border := bg(t.CodeBG) + strings.Repeat(" ", w) + reset
	lines = append(lines, border)
	for _, line := range strings.Split(strings.TrimRight(code, "\n"), "\n") {
		var b strings.Builder
		b.WriteString(bg(t.CodeBG))
		b.WriteString(fg(t.CodeText))
		b.WriteString("  ")
		b.WriteString(line)
		// Pad to full width so background fills the row.
		if pad := w - 2 - utf8.RuneCountInString(line); pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
		b.WriteString(reset)
		lines = append(lines, b.String())
	}
	lines = append(lines, border)
	return lines
}

// truncate trims s to at most n visible runes, preserving escape sequences.
func truncate(s string, n int) string {
	if visibleLen(s) <= n {
		return s
	}
	var b strings.Builder
	count := 0
	i := 0
	for i < len(s) && count < n {
		if s[i] == '\x1b' {
			end := scanEscape(s, i)
			b.WriteString(s[i:end])
			i = end
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		b.WriteString(s[i : i+size])
		i += size
		count++
	}
	return b.String()
}

func visibleLen(s string) int {
	count := 0
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			i = scanEscape(s, i)
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		count++
	}
	return count
}

func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			i = scanEscape(s, i)
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
