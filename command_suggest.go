package cli

import (
	"fmt"

	"github.com/paularlott/cli/fuzzy"
)

// commandItem wraps Command to implement fuzzy.NamedItem
type commandItem struct {
	cmd *Command
}

func (c commandItem) GetID() int    { return 0 } // ID not needed for command suggestions
func (c commandItem) GetName() string { return c.cmd.Name }

// findSimilarCommands finds commands that are similar to the given command name
func (c *Command) findSimilarCommands(cmdName string, commands []*Command, maxDistance int) []string {
	// Convert commands to fuzzy items
	items := make([]fuzzy.NamedItem, len(commands))
	for i, cmd := range commands {
		items[i] = commandItem{cmd: cmd}
	}

	// Use fuzzy search with threshold based on maxDistance
	// Convert maxDistance to a threshold (normalized)
	// A maxDistance of 2-3 is typical for command suggestions
	threshold := 0.5
	if maxDistance > 0 && len(cmdName) > 0 {
		// Calculate threshold: if maxDistance is 2 and name length is 5,
		// that means 40% difference is acceptable, so threshold is 0.6
		threshold = 1.0 - (float64(maxDistance) / float64(len(cmdName)))
		if threshold < 0.3 {
			threshold = 0.3 // Minimum threshold
		}
	}

	opts := fuzzy.Options{
		MaxResults: 5,
		Threshold:  threshold,
	}

	results := fuzzy.Search(cmdName, items, opts)

	// Extract command names
	suggestions := make([]string, 0, len(results))
	for _, r := range results {
		// Don't include exact matches (distance 0)
		if r.Score < 1.0 {
			suggestions = append(suggestions, r.Name)
		}
	}

	return suggestions
}

func (c *Command) displaySuggestions(suggestions []string, remainingArgs []string) {
	fmt.Printf("Unknown command: %s\n\nDid you mean this?\n", remainingArgs[0])
	if len(suggestions) == 1 {
		fmt.Printf("   - %s\n", suggestions[0])
	} else {
		for _, suggestion := range suggestions {
			fmt.Printf("   - %s\n", suggestion)
		}
	}
	fmt.Println()
	fmt.Printf("Run '%s --help' for usage.\n", c.Name)
	fmt.Println()
}
