package tui

// Theme color names for use with Panel.StyledWith.
const (
	ThemePrimary   = "primary"
	ThemeSecondary = "secondary"
	ThemeError     = "error"
	ThemeDim       = "dim"
	ThemeText      = "text"
	ThemeUser      = "user"
)

// Styled returns text wrapped in the given color, reset after.
// Use the theme color fields directly: t.Theme().Primary, t.Theme().Secondary, etc.
func Styled(color Color, text string) string {
	if color == 0 {
		return text
	}
	return fg(color) + text + reset
}

// Theme returns the active theme, allowing callers to access color values for use with Styled.
func (t *TUI) Theme() *Theme {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.theme
}
