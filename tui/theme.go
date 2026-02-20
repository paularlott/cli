package tui

import "sort"

// Color is a 24-bit RGB color packed as 0xRRGGBB. Zero means "default terminal color".
type Color uint32

// Theme defines the visual identity of the TUI.
type Theme struct {
	Name      string
	Primary   Color // Accents, prompt >
	Secondary Color // Muted text, hints
	Text      Color // Normal content
	UserText  Color // User message text
	Dim       Color // Very muted (scrollbar, borders)
	CodeBG    Color // Code block background
	CodeText  Color // Code block text
	Error     Color // Error messages
}

// Built-in themes.
var (
	// ThemeDefault is the default theme — dark navy background, cyan primary, crimson secondary.
	ThemeDefault = &Theme{
		Name:      "default",
		Primary:   0x4EB8C8,
		Secondary: 0xC0395A,
		Text:      0xE8EAF0,
		UserText:  0x4EB8C8,
		Dim:       0x7A8492,
		CodeBG:    0x111A26,
		CodeText:  0xE8EAF0,
		Error:     0xC0395A,
	}

	// ThemeAmber — warm dark background, amber primary, teal secondary.
	ThemeAmber = &Theme{
		Name:      "amber",
		Primary:   0xE8A87C,
		Secondary: 0x7EC8A4,
		Text:      0xCDD6F4,
		UserText:  0xE8A87C,
		Dim:       0x6C6F85,
		CodeBG:    0x1E1A14,
		CodeText:  0xCDD6F4,
		Error:     0xF38BA8,
	}

	// ThemeBlue — deep dark background, periwinkle primary, sky secondary.
	ThemeBlue = &Theme{
		Name:      "blue",
		Primary:   0x7BA7E8,
		Secondary: 0x5BC8D8,
		Text:      0xD0D8F0,
		UserText:  0x7BA7E8,
		Dim:       0x5A6070,
		CodeBG:    0x0D1117,
		CodeText:  0xD0D8F0,
		Error:     0xE06C75,
	}

	// ThemeGreen — dark terminal, mint primary, gold secondary.
	ThemeGreen = &Theme{
		Name:      "green",
		Primary:   0x7EC87A,
		Secondary: 0xD4A843,
		Text:      0xD8E0D0,
		UserText:  0x7EC87A,
		Dim:       0x5A6650,
		CodeBG:    0x0D1A0F,
		CodeText:  0xD8E0D0,
		Error:     0xE05050,
	}

	// ThemePurple — dark background, lavender primary, rose secondary.
	ThemePurple = &Theme{
		Name:      "purple",
		Primary:   0xB48EE8,
		Secondary: 0xE87EB4,
		Text:      0xE0D8F0,
		UserText:  0xB48EE8,
		Dim:       0x6A6080,
		CodeBG:    0x130D1E,
		CodeText:  0xE0D8F0,
		Error:     0xF07070,
	}

	// ThemeLight — light background, blue primary, green secondary.
	ThemeLight = &Theme{
		Name:      "light",
		Primary:   0x1A56CC,
		Secondary: 0x0A7A50,
		Text:      0x1A1A2E,
		UserText:  0x1A56CC,
		Dim:       0x666677,
		CodeBG:    0xE8EAF0,
		CodeText:  0x1A1A2E,
		Error:     0xCC2020,
	}

	// ThemePlain uses no colors (monochrome).
	ThemePlain = &Theme{
		Name: "plain",
	}
)

var themeRegistry = map[string]*Theme{
	"default": ThemeDefault,
	"amber":   ThemeAmber,
	"blue":    ThemeBlue,
	"green":   ThemeGreen,
	"purple":  ThemePurple,
	"light":   ThemeLight,
	"plain":   ThemePlain,
}

// ThemeByName returns the theme registered under name, or (nil, false).
func ThemeByName(name string) (*Theme, bool) {
	t, ok := themeRegistry[name]
	return t, ok
}

// ThemeNames returns a sorted slice of all registered theme names.
func ThemeNames() []string {
	names := make([]string, 0, len(themeRegistry))
	for name := range themeRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterTheme adds a custom theme to the global registry, keyed by Theme.Name.
func RegisterTheme(t *Theme) {
	themeRegistry[t.Name] = t
}
