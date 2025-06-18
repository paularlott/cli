package cli

import (
	"fmt"
	"strings"
)

// levenshteinDistance calculates the minimum number of single-character edits
// (insertions, deletions, or substitutions) required to change one string into another
func (c *Command) levenshteinDistance(s1, s2 string) int {
	// Convert strings to lowercase for case-insensitive comparison
	s1, s2 = strings.ToLower(s1), strings.ToLower(s2)

	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D slice to store the Levenshtein distances
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize the first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill in the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// FindSimilarCommands finds commands that are similar to the given command
func (c *Command) findSimilarCommands(cmdName string, commands []*Command, maxDistance int) []string {
	suggestions := make([]string, 0)

	for _, cmd := range commands {
		distance := c.levenshteinDistance(cmdName, cmd.Name)
		if distance <= maxDistance && distance > 0 {
			suggestions = append(suggestions, cmd.Name)
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
