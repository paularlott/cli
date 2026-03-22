package tui

// OSC (Operating System Command) escape sequence generators.
// These functions return strings containing OSC sequences that can be
// printed directly or embedded in TUI messages.
//
// OSC sequences use the format: ESC ] <command> ; <params> ST
// where ST is the string terminator (ESC \).
//
// Terminals that don't support a specific OSC sequence will typically
// ignore it or render the content without the special behavior.

const (
	osc       = "\x1b]"        // OSC prefix: ESC ]
	st        = "\x1b\\"      // String terminator: ESC \
	linkReset = "\x1b]8;;\x1b\\" // OSC 8 close: terminates any open hyperlink
)

// Hyperlink wraps text in an OSC 8 hyperlink.
// Supported by: iTerm2, Kitty, WezTerm, Alacritty, Windows Terminal,
// VTE (GNOME Terminal), Ghostty, Terminal.app (macOS 10.15+).
// If url is empty, returns text unchanged.
func Hyperlink(url, text string) string {
	if url == "" {
		return text
	}
	return osc + "8;;" + url + st + text + osc + "8;;" + st
}

// SetWindowTitle returns an OSC 0 sequence to set the window title and icon.
// Supported by virtually all terminal emulators.
func SetWindowTitle(title string) string {
	return osc + "0;" + title + st
}

// Notify returns an OSC 9 sequence for a desktop notification.
// Supported by: iTerm2, Ghostty, ConEmu.
func Notify(message string) string {
	return osc + "9;" + message + st
}
