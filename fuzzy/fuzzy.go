// Package fuzzy provides fuzzy string matching utilities using a multi-tier
// matching algorithm (exact → substring → word boundary → Levenshtein distance).
package fuzzy

import (
	"fmt"
	"strings"
	"sync"
)

// NamedItem represents an item with an ID and Name for fuzzy matching.
// Items must implement this interface to be searchable.
type NamedItem interface {
	GetID() int
	GetName() string
}

// NamedItemString is a simple implementation of NamedItem for string-based items.
type NamedItemString struct {
	ID   int
	Name string
}

func (n NamedItemString) GetID() int    { return n.ID }
func (n NamedItemString) GetName() string { return n.Name }

// Result represents a single fuzzy match result.
type Result struct {
	ID    int     // The matched item's ID
	Name  string  // The matched item's name
	Score float64 // Match score (0.0 to 1.0, higher is better)
}

// BestResult represents the result of a Best() call.
type BestResult struct {
	Found bool   // True if a match was found
	ID    int    // The matched item's ID (0 if not found)
	Name  string // The matched item's name (empty if not found)
	Score float64 // Match score (0 if not found)
	Error string // Error message with suggestions (empty if found)
}

// Options configures the fuzzy search behavior.
type Options struct {
	MaxResults int     // Maximum number of results to return (default: 5)
	Threshold  float64 // Minimum score threshold 0.0-1.0 (default: 0.5)
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxResults: 5,
		Threshold:  0.5,
	}
}

// Search performs a fuzzy search and returns multiple matches sorted by score.
// Returns an empty slice if no matches are found.
func Search(query string, items []NamedItem, opts Options) []Result {
	if len(items) == 0 {
		return nil
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	if opts.MaxResults <= 0 {
		opts.MaxResults = 5
	}

	results := make([]Result, 0, opts.MaxResults)
	seen := make(map[int]bool, len(items))
	queryWords := strings.Fields(query)

	// Single pass through items with tier-based scoring
	for _, item := range items {
		if seen[item.GetID()] {
			continue
		}

		nameLower := strings.ToLower(item.GetName())
		var score float64
		var matched bool

		// Tier 1: Exact match
		if nameLower == query {
			score = 1.0
			matched = true
		} else if strings.Contains(nameLower, query) {
			// Tier 2: Substring match
			score = 0.9 * (float64(len(query)) / float64(len(nameLower)))
			matched = true
		} else {
			// Tier 3: Word boundary match
			nameWords := strings.Fields(nameLower)
			for _, word := range nameWords {
				if strings.HasPrefix(word, query) {
					score = 0.85
					matched = true
					break
				}
			}

			// Tier 4: Levenshtein distance
			if !matched && len(results) < opts.MaxResults {
				minDistance := 999999.0
				for _, searchWord := range queryWords {
					for _, nameWord := range nameWords {
						searchLen := len(searchWord)
						nameLen := len(nameWord)
						maxLen := searchLen
						if nameLen > searchLen {
							maxLen = nameLen
						}
						if maxLen == 0 {
							continue
						}
						dist := levenshteinDistance(searchWord, nameWord)
						normalizedDist := float64(dist) / float64(maxLen)
						if normalizedDist < minDistance {
							minDistance = normalizedDist
						}
					}
				}

				if minDistance <= opts.Threshold {
					score = 1.0 - minDistance
					matched = true
				}
			}
		}

		if matched {
			results = append(results, Result{
				ID:    item.GetID(),
				Name:  item.GetName(),
				Score: score,
			})
			seen[item.GetID()] = true
		}
	}

	// Sort results by score using insertion sort (efficient for small slices)
	if len(results) > 1 {
		for i := 1; i < len(results); i++ {
			key := results[i]
			j := i - 1
			for j >= 0 && results[j].Score < key.Score {
				results[j+1] = results[j]
				j--
			}
			results[j+1] = key
		}
	}

	if len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}

	return results
}

// Best finds the best match for a query and returns a formatted result.
// If no match is found, returns a result with Error containing suggestions.
func Best(query string, items []NamedItem, entityType string, opts Options) BestResult {
	if len(items) == 0 {
		return BestResult{
			Found: false,
			Error: fmt.Sprintf("no %s available", entityType),
		}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return BestResult{
			Found: false,
			Error: fmt.Sprintf("%s name is required", entityType),
		}
	}

	// First try exact match (case-insensitive)
	for _, item := range items {
		if strings.ToLower(item.GetName()) == query {
			return BestResult{
				Found: true,
				ID:    item.GetID(),
				Name:  item.GetName(),
				Score: 1.0,
			}
		}
	}

	// No exact match - find similar matches for suggestions
	suggestions := Search(query, items, Options{MaxResults: 3, Threshold: 0.5})

	if len(suggestions) == 0 {
		return BestResult{
			Found: false,
			Error: fmt.Sprintf("%s '%s' is unknown. No similar matches found", entityType, query),
		}
	}

	// Return best match
	best := suggestions[0]
	return BestResult{
		Found: true,
		ID:    best.ID,
		Name:  best.Name,
		Score: best.Score,
	}
}

// Score calculates the similarity score between two strings.
// Returns a value between 0.0 (completely different) and 1.0 (identical).
func Score(s1, s2 string) float64 {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	dist := levenshteinDistance(s1, s2)
	maxLen := len1
	if len2 > len1 {
		maxLen = len2
	}

	return 1.0 - (float64(dist) / float64(maxLen))
}

// FormatSuggestions formats a list of suggestion names into a readable string.
func FormatSuggestions(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}
	if len(suggestions) == 1 {
		return fmt.Sprintf("'%s'", suggestions[0])
	}
	if len(suggestions) == 2 {
		return fmt.Sprintf("'%s' or '%s'", suggestions[0], suggestions[1])
	}

	// For 3+ suggestions
	quoted := make([]string, len(suggestions))
	for i, s := range suggestions {
		quoted[i] = fmt.Sprintf("'%s'", s)
	}

	lastIdx := len(quoted) - 1
	return strings.Join(quoted[:lastIdx], ", ") + ", or " + quoted[lastIdx]
}

// Pool for reusing Levenshtein distance buffers
var levPool = sync.Pool{
	New: func() interface{} {
		return make([]int, 0)
	},
}

func getLevBuffer(size int) []int {
	buf := levPool.Get().([]int)
	if cap(buf) < size {
		buf = make([]int, size)
	} else {
		buf = buf[:size]
	}
	return buf
}

func putLevBuffer(buf []int) {
	levPool.Put(buf)
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)

	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	// Ensure s2 is the shorter string for memory efficiency
	if len1 < len2 {
		s1, s2 = s2, s1
		len1, len2 = len2, len1
	}

	// Use pooled buffers
	prev := getLevBuffer(len2 + 1)
	curr := getLevBuffer(len2 + 1)
	defer putLevBuffer(prev)
	defer putLevBuffer(curr)

	// Initialize first row
	for j := 0; j <= len2; j++ {
		prev[j] = j
	}

	// Fill matrix
	for i := 1; i <= len1; i++ {
		curr[0] = i
		for j := 1; j <= len2; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			// Inline min3 for performance
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			if del < ins {
				if del < sub {
					curr[j] = del
				} else {
					curr[j] = sub
				}
			} else {
				if ins < sub {
					curr[j] = ins
				} else {
					curr[j] = sub
				}
			}
		}
		prev, curr = curr, prev
	}

	return prev[len2]
}
