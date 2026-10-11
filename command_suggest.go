package cli

import (
	"fmt"
	"strings"

	"github.com/paularlott/cli/fuzzy"
)

// suggestItem is a command name or alias to match against, implementing
// fuzzy.NamedItem. The fuzzy search skips an item whose ID it has already
// matched, so every item needs its own.
type suggestItem struct {
	cmd  *Command
	name string
	id   int
}

func (s suggestItem) GetID() int      { return s.id }
func (s suggestItem) GetName() string { return s.name }

// findSimilarCommands finds commands that are similar to the given command
// name, matching their names and aliases. Each command is suggested once,
// best match first, as its name followed by any aliases, e.g. "ls, list".
func (c *Command) findSimilarCommands(cmdName string, commands []*Command, maxDistance int) []string {
	var items []fuzzy.NamedItem
	for _, cmd := range commands {
		items = append(items, suggestItem{cmd: cmd, name: cmd.Name, id: len(items)})
		for _, alias := range cmd.Aliases {
			items = append(items, suggestItem{cmd: cmd, name: alias, id: len(items)})
		}
	}

	// Use fuzzy search with threshold based on maxDistance
	// Convert maxDistance to a threshold (normalized)
	// A maxDistance of 2-3 is typical for command suggestions
	threshold := 0.5
	if maxDistance > 0 && len(cmdName) > 0 {
		// Calculate threshold: if maxDistance is 2 and name length is 5,
		// that means 40% difference is acceptable, so threshold is 0.6
		threshold = 1.0 - (float64(maxDistance) / float64(len(cmdName)))
		// Short names give a low threshold, so floor it: below half
		// alike, a word isn't a typo of the command.
		if threshold < 0.5 {
			threshold = 0.5
		}
	}

	opts := fuzzy.Options{
		MaxResults: len(items),
		Threshold:  threshold,
	}

	results := fuzzy.Search(cmdName, items, opts)

	suggestions := make([]string, 0, len(results))
	seen := make(map[*Command]bool, len(results))
	for _, r := range results {
		// Don't include exact matches (distance 0)
		if r.Score >= 1.0 {
			continue
		}
		cmd := items[r.ID].(suggestItem).cmd
		if seen[cmd] {
			continue
		}
		seen[cmd] = true
		suggestions = append(suggestions, strings.Join(append([]string{cmd.Name}, cmd.Aliases...), ", "))
		if len(suggestions) == 5 {
			break
		}
	}

	return suggestions
}

// unknownCommandError reports a mistyped subcommand of cmd, with any
// suggestions, as one error for the caller to print.
func unknownCommandError(word string, suggestions []string, cmd *Command) error {
	var b strings.Builder
	fmt.Fprintf(&b, "unknown command %q", word)
	if len(suggestions) > 0 {
		b.WriteString("\n\nDid you mean this?\n")
		for _, suggestion := range suggestions {
			fmt.Fprintf(&b, "   - %s\n", suggestion)
		}
	} else {
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "\nRun '%s --help' for usage", cmd.commandPath())
	return fmt.Errorf("%s", b.String())
}

// commandPath is the command's full name from the root, e.g. "app file ls".
func (c *Command) commandPath() string {
	if len(c.commandChain) == 0 {
		return c.Name
	}
	names := make([]string, 0, len(c.commandChain))
	for _, cmd := range c.commandChain {
		names = append(names, cmd.Name)
	}
	return strings.Join(names, " ")
}
