package tui

import (
	"strings"
	"unicode/utf8"
)

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
}

func (o *outputRegion) scrollUp(n int)   { o.scrollOff += n }
func (o *outputRegion) scrollDown(n int) {
	o.scrollOff -= n
	if o.scrollOff < 0 {
		o.scrollOff = 0
	}
}

// render draws the output region into buf, using height terminal rows of width w.
// startRow is the 1-based terminal row where the region begins.
func (o *outputRegion) render(buf *strings.Builder, t *Theme, w, height, startRow int) {
	lineW := w

	all := o.messages
	if o.streaming != nil {
		all = append(o.messages[:len(o.messages):len(o.messages)], o.streaming)
	}

	var lines []string
	for _, m := range all {
		lines = append(lines, renderMessage(m, t, lineW, o.userLabel, o.assistantLabel, o.systemLabel, o.hideHeaders)...)
	}

	total := len(lines)
	maxOff := total - height
	if maxOff < 0 {
		maxOff = 0
	}
	if o.scrollOff > maxOff {
		o.scrollOff = maxOff
	}

	start := total - height - o.scrollOff
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
	}

	for i := start; i < end; i++ {
		row := i - start
		buf.WriteString(cursorPos(startRow+row, 1))
		buf.WriteString(clearLine())
		buf.WriteString(truncate(lines[i], lineW))
	}
	for i := end - start; i < height; i++ {
		buf.WriteString(cursorPos(startRow+i, 1))
		buf.WriteString(clearLine())
	}
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
	fill := w - utf8.RuneCountInString(label) - 4
	if fill < 0 {
		fill = 0
	}
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

// wordWrap splits a plain string into lines of at most w runes, breaking on spaces.
func wordWrap(s string, w int) []string {
	if w <= 0 || utf8.RuneCountInString(s) <= w {
		return []string{s}
	}
	var lines []string
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{s}
	}
	current := ""
	for _, word := range words {
		wl := utf8.RuneCountInString(word)
		if current == "" {
			if wl > w {
				// Single word longer than width — hard break it.
				runes := []rune(word)
				for len(runes) > 0 {
					chunk := w
					if chunk > len(runes) {
						chunk = len(runes)
					}
					lines = append(lines, string(runes[:chunk]))
					runes = runes[chunk:]
				}
				continue
			}
			current = word
		} else if utf8.RuneCountInString(current)+1+wl <= w {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
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

// truncate trims a potentially ANSI-escaped string to at most n visible runes.
func truncate(s string, n int) string {
	// Simple approach: strip ANSI then measure; if short enough return as-is.
	if visibleLen(s) <= n {
		return s
	}
	return truncatePlain(stripANSI(s), n)
}

func truncatePlain(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

func visibleLen(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
}

func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			inEsc = true
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
