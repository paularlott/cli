package tui

import (
	"strings"
)

// Command is a slash command that can be registered with the TUI.
type Command struct {
	Name        string
	Description string
	Args        []string         // Optional sub-options shown in palette after a space
	Handler     func(args string)
}

type palette struct {
	commands    []*Command
	filtered    []*Command
	selected    int
	viewOff     int // index of first visible item
	active      bool
	query       string
	argMode     bool
	argCmd      *Command
	argFiltered []string
}

func newPalette(cmds []*Command) *palette {
	return &palette{commands: cmds}
}

func (p *palette) open(query string) {
	p.active = true
	p.filter(query)
}

func (p *palette) close() {
	p.active = false
	p.query = ""
	p.filtered = nil
	p.selected = 0
	p.viewOff = 0
	p.argMode = false
	p.argCmd = nil
	p.argFiltered = nil
}

func (p *palette) filter(query string) {
	p.query = query
	// Check if query is "cmdname args" — switch to arg mode.
	if idx := strings.Index(query, " "); idx != -1 {
		name := query[:idx]
		argQuery := query[idx+1:]
		for _, c := range p.commands {
			if c.Name == name && len(c.Args) > 0 {
				p.argMode = true
				p.argCmd = c
				if p.argFiltered == nil {
					p.argFiltered = make([]string, 0, len(c.Args))
				} else {
					p.argFiltered = p.argFiltered[:0]
				}
			for _, a := range c.Args {
					if argQuery == "" || strings.HasPrefix(a, argQuery) {
						p.argFiltered = append(p.argFiltered, a)
					}
				}
				if p.selected >= len(p.argFiltered) {
					p.selected = max(0, len(p.argFiltered)-1)
					p.viewOff = 0
				}
				return
			}
		}
		// Command has no args — close arg mode, let normal flow handle it.
		p.argMode = false
		p.argCmd = nil
		p.argFiltered = nil
		return
	}
	p.argMode = false
	p.argCmd = nil
	p.argFiltered = nil
	if p.filtered == nil {
		p.filtered = make([]*Command, 0, len(p.commands))
	} else {
		p.filtered = p.filtered[:0]
	}
	for _, c := range p.commands {
		if query == "" || strings.HasPrefix(c.Name, query) {
			p.filtered = append(p.filtered, c)
		}
	}
	if p.selected >= len(p.filtered) {
		p.selected = max(0, len(p.filtered)-1)
		p.viewOff = 0
	}
}

func (p *palette) moveUp() {
	if p.selected > 0 {
		p.selected--
		if p.selected < p.viewOff {
			p.viewOff = p.selected
		}
	}
}

func (p *palette) moveDown(maxRows int) {
	n := len(p.filtered)
	if p.argMode {
		n = len(p.argFiltered)
	}
	if p.selected < n-1 {
		p.selected++
		if p.selected >= p.viewOff+maxRows {
			p.viewOff = p.selected - maxRows + 1
		}
	}
}

// selectedCommand returns the currently highlighted command, or nil.
// In arg mode it returns the parent command (args are handled separately).
func (p *palette) selectedCommand() *Command {
	if p.argMode {
		if len(p.argFiltered) == 0 {
			return nil
		}
		return p.argCmd
	}
	if len(p.filtered) == 0 {
		return nil
	}
	return p.filtered[p.selected]
}

// selectedArg returns the currently highlighted arg value in arg mode, or "".
func (p *palette) selectedArg() string {
	if p.argMode && p.selected < len(p.argFiltered) {
		return p.argFiltered[p.selected]
	}
	return ""
}

// render writes the palette into buf using absolute cursor positioning.
// startRow is the 1-based terminal row where the palette begins.
func (p *palette) render(buf *strings.Builder, t *Theme, w, maxRows, startRow int) int {
	if !p.active {
		return 0
	}
	if p.argMode {
		if len(p.argFiltered) == 0 {
			return 0
		}
		visible := p.argFiltered[p.viewOff:]
		if len(visible) > maxRows {
			visible = visible[:maxRows]
		}
		for i, a := range visible {
			buf.WriteString(cursorPos(startRow+i, 1))
			buf.WriteString(clearLine())
			if p.viewOff+i == p.selected {
				buf.WriteString(fg(t.Primary) + bold() + "  " + a + reset)
			} else {
				buf.WriteString("  " + fg(t.Secondary) + a + reset)
			}
		}
		buf.WriteString(cursorPos(startRow+len(visible), 1))
		buf.WriteString(clearLine())
		buf.WriteString(fg(t.Dim) + "  ↑↓ navigate · Tab select · Esc close" + reset)
		return len(visible) + 1
	}
	if len(p.filtered) == 0 {
		return 0
	}
	visible := p.filtered[p.viewOff:]
	if len(visible) > maxRows {
		visible = visible[:maxRows]
	}
	for i, cmd := range visible {
		buf.WriteString(cursorPos(startRow+i, 1))
		buf.WriteString(clearLine())
		if p.viewOff+i == p.selected {
			buf.WriteString(fg(t.Primary) + bold() + "  /" + cmd.Name + reset)
			buf.WriteString("  " + italic() + fg(t.Secondary) + cmd.Description + reset)
		} else {
			buf.WriteString("  " + fg(t.Secondary) + "/" + cmd.Name + reset)
			buf.WriteString("  " + fg(t.Dim) + cmd.Description + reset)
		}
	}
	buf.WriteString(cursorPos(startRow+len(visible), 1))
	buf.WriteString(clearLine())
	buf.WriteString(fg(t.Dim) + "  ↑↓ navigate · Tab select · Esc close" + reset)
	return len(visible) + 1
}

