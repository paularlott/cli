# tui

A full-screen terminal UI framework for building interactive CLI applications — chat interfaces, log viewers, AI assistants, and more.

## Layout

```
  Scrollable output / conversation history
  (auto-scrolls to bottom; Page Up/Down/wheel to scroll)

  Command palette (visible when / is typed)

┌─────────────────────────────────────── spinner / status ─┐
│                                                          │
│ > input goes here                                        │
│                                                          │
└─ myapp ──────────────────────────────────────────────────┘
```

The input box top border carries the spinner, progress bar, scroll hint, or `StatusRight` text (in that priority order). `StatusLeft` is embedded in the bottom border.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/paularlott/cli/tui"
)

func main() {
    var t *tui.TUI

    t = tui.New(tui.Config{
        Theme:       tui.ThemeAmber,
        StatusLeft:  "myapp",
        StatusRight: "Ctrl+C to exit",
        Commands: []*tui.Command{
            {
                Name:        "exit",
                Description: "Exit the application",
                Handler:     func(_ string) { t.Exit() },
            },
            {
                Name:        "clear",
                Description: "Clear conversation history",
                Handler:     func(_ string) { t.ClearOutput() },
            },
        },
        OnSubmit: func(text string) {
            t.AddMessage(tui.RoleUser, text)
            t.AddMessage(tui.RoleAssistant, "Echo: "+text)
        },
    })

    t.AddMessage(tui.RoleSystem, "Welcome! Type / for commands.")

    if err := t.Run(context.Background()); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

## Configuration

```go
type Config struct {
    Theme          *Theme      // Active theme. Defaults to ThemeDefault.
    Themes         []*Theme    // Additional themes registered into the global registry.
    Commands       []*Command  // Slash commands shown in the palette.
    OnSubmit       func(text string) // Called when the user submits input.
    OnEscape       func()            // Called when Escape is pressed outside the palette.
    UserLabel      string      // Label for user messages. Default: "You".
    AssistantLabel string      // Label for assistant messages. Default: "Assistant".
    SystemLabel    string      // Label for system messages. Default: "System".
    StatusLeft     string      // Text shown bottom-left.
    StatusRight    string      // Text shown bottom-right (overridden by spinner/progress/scroll hint).
    ShowCharCount  bool        // Show character counter below the input box. Default: false.
    HideHeaders    bool        // Suppress the role header line between messages. Default: false.
    InputEnabled   *bool       // Set to false for output-only / log-viewer mode.
}
```

## Messages

```go
// Append a message from a standard role.
t.AddMessage(tui.RoleUser, "Hello!")
t.AddMessage(tui.RoleAssistant, "Hi there.")
t.AddMessage(tui.RoleSystem, "Connected.")

// Append a message with a custom label (overrides the role label).
t.AddMessageAs(tui.RoleAssistant, "GPT-4o", "Here is my answer…")
```

Message content supports fenced code blocks:

````
t.AddMessage(tui.RoleAssistant, "Example:\n\n```go\nfmt.Println(\"hello\")\n```")
````

## Streaming

For token-by-token responses:

```go
t.StartStreamingAs("Claude 3.5") // or t.StartStreaming() for the default label
for chunk := range tokenCh {
    if !t.IsStreaming() {
        break // stopped externally via StopStreaming()
    }
    t.StreamChunk(chunk)
}
t.StreamComplete()

// To stop early and finalise the message (e.g. from OnEscape):
t.StopStreaming()
```

## Commands

All commands are supplied by the caller — there are no built-ins. Register them in `Config.Commands`:

```go
Commands: []*tui.Command{
    {
        Name:        "exit",
        Description: "Exit the application",
        Handler:     func(_ string) { t.Exit() },
    },
    {
        Name:        "theme",
        Description: "Switch theme",
        Args:        []string{"amber", "blue", "green"}, // shown as sub-options
        Handler: func(args string) {
            if th, ok := tui.ThemeByName(args); ok {
                t.SetTheme(th)
            }
        },
    },
},
```

Type `/` to open the palette. Use `↑`/`↓` to navigate, `Tab` to complete, `Enter` to execute, `Esc` to close.

## Spinner

Displays an animated braille spinner in the input box top border:

```go
t.StartSpinner("Thinking…")
// … do work …
t.StopSpinner()
```

Calling `StartSpinner` while one is already running replaces the text. Starting a progress bar stops the spinner automatically.

## Progress Bar

Displays a labelled progress bar in the input box top border:

```go
for i := 0; i <= 100; i++ {
    t.SetProgress("Uploading…", float64(i)/100)
}
t.ClearProgress()
```

The border renders: `┌──────── Uploading… [████████████░░░░░░░░]  60% ┐`

Spinner and progress bar are mutually exclusive. Both are overridden by the scroll hint when the user has scrolled up.

## Status Bar

```go
t.SetStatus("myapp", "v1.2.3")   // set both at once
t.SetStatusLeft("myapp")
t.SetStatusRight("v1.2.3")
```

## Output

```go
t.ClearOutput() // remove all messages from the output region
```

## Styled Text

`Styled` wraps a string in a theme color for use in message content:

```go
t.AddMessage(tui.RoleSystem,
    tui.Styled(t.Theme().Text, "myapp") + "\n" +
    tui.Styled(t.Theme().Primary, "v1.2.3"),
)
```

`t.Theme()` returns the active theme so you can reference its color fields (`Primary`, `Secondary`, `Text`, `Dim`, `Error`, etc.) at call time, picking up any theme changes automatically.

## Menus

`t.OpenMenu(m)` replaces the input box with a navigable bordered panel. `t.CloseMenu()` dismisses it programmatically.

```go
t.OpenMenu(&tui.Menu{
    Title: "Settings",
    Items: []*tui.MenuItem{
        {
            Label: "Theme",
            Children: []*tui.MenuItem{          // sub-menu — shows › indicator
                {Label: "Amber", Value: "amber", OnSelect: func(item *tui.MenuItem, _ string) {
                    if th, ok := tui.ThemeByName(item.Value); ok { t.SetTheme(th) }
                }},
            },
        },
        {
            Label:  "API Key",
            Prompt: "Enter API key:",            // text-entry mode
            OnSelect: func(_ *tui.MenuItem, input string) {
                // input holds what the user typed
            },
        },
        {
            Label: "About",
            OnSelect: func(_ *tui.MenuItem, _ string) {
                t.AddMessage(tui.RoleSystem, "v1.0.0")
            },
        },
    },
})
```

### MenuItem fields

| Field      | Type                                 | Description                                                                 |
| ---------- | ------------------------------------ | --------------------------------------------------------------------------- |
| `Label`    | `string`                             | Display text                                                                |
| `Value`    | `string`                             | Optional data passed to `OnSelect`                                          |
| `Prompt`   | `string`                             | If set, selecting this item opens a text-entry prompt (shows `›` indicator) |
| `Children` | `[]*MenuItem`                        | If set, selecting pushes a sub-menu (shows `›` indicator)                   |
| `OnSelect` | `func(item *MenuItem, input string)` | Called when item is confirmed; `input` is non-empty for Prompt items        |

### Navigation

| Key     | Action                                                     |
| ------- | ---------------------------------------------------------- |
| `↑`/`↓` | Move selection                                             |
| `Enter` | Select item / confirm prompt                               |
| `Esc`   | Cancel prompt → back to list; pop sub-menu → close at root |

The title bar shows a breadcrumb when inside a sub-menu: `Settings › Theme`.

## Output-Only Mode

Set `InputEnabled` to `false` to hide the input box, char count, and palette. Only scrolling and `Ctrl+C` remain active. Useful for log viewers or progress displays driven entirely by the application.

```go
enabled := false
t = tui.New(tui.Config{
    InputEnabled: &enabled,
    // …
})
```

## Context & Shutdown

`Run` accepts a `context.Context`. Cancelling the context shuts down the event loop cleanly:

```go
ctx, cancel := context.WithCancel(context.Background())
// cancel() from anywhere to exit

t.Run(ctx)

// Retrieve the context later (e.g. inside a command handler):
ctx := t.Context()
```

`t.Exit()` is a convenience method that sets the quit flag directly — wire it to a `/exit` command.

## Themes

### Built-in themes

| Name      | Description                                             |
| --------- | ------------------------------------------------------- |
| `default` | Dark navy, cyan primary, crimson secondary (default)    |
| `amber`   | Dark warm background, amber primary, teal secondary     |
| `blue`    | Deep dark background, periwinkle primary, sky secondary |
| `green`   | Dark terminal, mint primary, gold secondary             |
| `purple`  | Dark background, lavender primary, rose secondary       |
| `light`   | Light background, blue primary, green secondary         |
| `plain`   | No colors (monochrome)                                  |

### Selecting by name

```go
if th, ok := tui.ThemeByName("amber"); ok {
    t.SetTheme(th)
}

// List all available theme names:
names := tui.ThemeNames() // ["amber", "blue", "default", "green", "light", "plain", "purple"]
```

### Custom themes

Define a `Theme` struct and either pass it in `Config.Themes` (registered automatically) or call `RegisterTheme` directly:

```go
myTheme := &tui.Theme{
    Name:      "solarized",
    Primary:   0x268BD2,
    Secondary: 0x2AA198,
    Text:      0x839496,
    UserText:  0x268BD2,
    Dim:       0x586E75,
    CodeBG:    0x073642,
    CodeText:  0x839496,
    Error:     0xDC322F,
}

// Option A — via Config (registered before Run):
t = tui.New(tui.Config{
    Themes: []*tui.Theme{myTheme},
    Theme:  myTheme,
})

// Option B — global registry at any time:
tui.RegisterTheme(myTheme)
```

`Color` values are 24-bit RGB packed as `0xRRGGBB`. A zero value means the terminal's default color.

## Keyboard Reference

| Key            | Action                                                                                       |
| -------------- | -------------------------------------------------------------------------------------------- |
| `Enter`        | Submit input / execute selected command                                                      |
| `Shift+Enter`  | Insert newline                                                                               |
| `↑` / `↓`      | Move cursor up/down in multi-line input, navigate history (single-line), or navigate palette |
| `←` / `→`      | Move cursor left/right                                                                       |
| `Home` / `End` | Jump to start/end of line                                                                    |
| `Backspace`    | Delete character before cursor                                                               |
| `Delete`       | Delete character at cursor                                                                   |
| `Ctrl+A`       | Move to start of line                                                                        |
| `Ctrl+K`       | Delete to end of line                                                                        |
| `Ctrl+U`       | Delete to start of line                                                                      |
| `Ctrl+W`       | Delete word before cursor                                                                    |
| `Page Up/Down` | Scroll output half a page                                                                    |
| Mouse wheel    | Scroll output 3 lines                                                                        |
| `Tab`          | Complete selected palette command/arg                                                        |
| `Esc`          | Close palette / fire `OnEscape`                                                              |
| `Ctrl+C`       | Exit                                                                                         |
