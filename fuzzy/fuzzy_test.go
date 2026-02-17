package fuzzy

import (
	"testing"
)

// Test item implementation
type testItem struct {
	id   int
	name string
}

func (t testItem) GetID() int    { return t.id }
func (t testItem) GetName() string { return t.name }

func items(names ...string) []NamedItem {
	items := make([]NamedItem, len(names))
	for i, name := range names {
		items[i] = testItem{id: i + 1, name: name}
	}
	return items
}

func itemsWithIDs(idNames ...interface{}) []NamedItem {
	var items []NamedItem
	for i := 0; i < len(idNames); i += 2 {
		items = append(items, testItem{
			id:   idNames[i].(int),
			name: idNames[i+1].(string),
		})
	}
	return items
}

// Test Search function
func TestSearch_ExactMatch(t *testing.T) {
	items := items("Apple", "Banana", "Cherry")
	results := Search("apple", items, DefaultOptions())

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score != 1.0 {
		t.Errorf("expected score 1.0 for exact match, got %f", results[0].Score)
	}
	if results[0].Name != "Apple" {
		t.Errorf("expected 'Apple', got '%s'", results[0].Name)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	items := items("Apple", "Banana", "Cherry")
	results := Search("APPLE", items, DefaultOptions())

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score != 1.0 {
		t.Errorf("expected score 1.0 for case-insensitive exact match, got %f", results[0].Score)
	}
}

func TestSearch_SubstringMatch(t *testing.T) {
	items := items("Apple Pie", "Banana Split", "Pineapple")
	results := Search("apple", items, DefaultOptions())

	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	// Apple Pie should score higher (exact word) than Pineapple (substring)
	if results[0].Name != "Apple Pie" {
		t.Errorf("expected 'Apple Pie' first, got '%s'", results[0].Name)
	}
}

func TestSearch_WordBoundary(t *testing.T) {
	items := items("Customer Support", "Technical Support", "Sales")
	results := Search("supp", items, DefaultOptions())

	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	// Both should match via word boundary
	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	if !names["Customer Support"] || !names["Technical Support"] {
		t.Errorf("expected both Support items, got %v", results)
	}
}

func TestSearch_Levenshtein(t *testing.T) {
	items := items("Project Alpha", "Project Beta", "Something Else")
	results := Search("projct alpha", items, DefaultOptions())

	if len(results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(results))
	}
	// Should match Project Alpha despite typo
	if results[0].Name != "Project Alpha" {
		t.Errorf("expected 'Project Alpha', got '%s'", results[0].Name)
	}
	if results[0].Score < 0.7 {
		t.Errorf("expected score >= 0.7 for close match, got %f", results[0].Score)
	}
}

func TestSearch_NoMatch(t *testing.T) {
	items := items("Apple", "Banana", "Cherry")
	results := Search("xyz", items, DefaultOptions())

	if len(results) != 0 {
		t.Errorf("expected 0 results for no match, got %d", len(results))
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	items := items("Apple", "Banana")
	results := Search("", items, DefaultOptions())

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(results))
	}
}

func TestSearch_EmptyItems(t *testing.T) {
	results := Search("test", []NamedItem{}, DefaultOptions())

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty items, got %d", len(results))
	}
}

func TestSearch_MaxResults(t *testing.T) {
	items := items("Apple", "Apple Pie", "Apple Sauce", "Apple Juice", "Pineapple")
	opts := Options{MaxResults: 3, Threshold: 0.5}
	results := Search("apple", items, opts)

	if len(results) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(results))
	}
}

func TestSearch_Threshold(t *testing.T) {
	items := items("Project Alpha", "Something Completely Different")
	opts := Options{MaxResults: 5, Threshold: 0.1} // Very permissive
	results := Search("projct alpha", items, opts)

	// With low threshold, should still find Project Alpha
	if len(results) < 1 {
		t.Errorf("expected at least 1 result with low threshold, got %d", len(results))
	}
}

func TestSearch_Deduplication(t *testing.T) {
	// Same item shouldn't appear twice
	items := itemsWithIDs(1, "Apple", 1, "Apple")
	results := Search("apple", items, DefaultOptions())

	if len(results) != 1 {
		t.Errorf("expected 1 result (deduplicated), got %d", len(results))
	}
}

func TestSearch_SortedByScore(t *testing.T) {
	items := items("Apple", "Apple Pie", "Pineapple", "Apple Juice")
	results := Search("apple", items, DefaultOptions())

	// Results should be sorted by score descending
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("results not sorted by score: %v", results)
		}
	}
}

// Test Best function
func TestBest_ExactMatch(t *testing.T) {
	items := items("Active", "Pending", "Complete")
	result := Best("active", items, "status", DefaultOptions())

	if !result.Found {
		t.Errorf("expected Found=true for exact match, got false")
	}
	if result.ID != 1 {
		t.Errorf("expected ID=1, got %d", result.ID)
	}
	if result.Name != "Active" {
		t.Errorf("expected 'Active', got '%s'", result.Name)
	}
}

func TestBest_FuzzyMatch(t *testing.T) {
	items := items("Website Redesign", "Mobile App", "Server Migration")
	result := Best("web design", items, "project", DefaultOptions())

	if !result.Found {
		t.Errorf("expected Found=true for fuzzy match, got false")
	}
	if result.Name != "Website Redesign" {
		t.Errorf("expected 'Website Redesign', got '%s'", result.Name)
	}
}

func TestBest_NoMatch(t *testing.T) {
	items := items("Active", "Pending", "Complete")
	result := Best("xyz", items, "status", DefaultOptions())

	if result.Found {
		t.Errorf("expected Found=false for no match")
	}
	if result.Error == "" {
		t.Error("expected error message for no match")
	}
}

func TestBest_EmptyItems(t *testing.T) {
	result := Best("test", []NamedItem{}, "item", DefaultOptions())

	if result.Found {
		t.Error("expected Found=false for empty items")
	}
	if result.Error == "" {
		t.Error("expected error message for empty items")
	}
}

func TestBest_EmptyQuery(t *testing.T) {
	items := items("Active", "Pending")
	result := Best("", items, "status", DefaultOptions())

	if result.Found {
		t.Error("expected Found=false for empty query")
	}
}

// Test Score function
func TestScore_Identical(t *testing.T) {
	score := Score("hello", "hello")
	if score != 1.0 {
		t.Errorf("expected score 1.0 for identical strings, got %f", score)
	}
}

func TestScore_CaseInsensitive(t *testing.T) {
	score := Score("HELLO", "hello")
	if score != 1.0 {
		t.Errorf("expected score 1.0 for case-insensitive match, got %f", score)
	}
}

func TestScore_CompletelyDifferent(t *testing.T) {
	score := Score("abc", "xyz")
	if score >= 0.5 {
		t.Errorf("expected low score for different strings, got %f", score)
	}
}

func TestScore_SingleEdit(t *testing.T) {
	// One character difference
	score := Score("hello", "hallo")
	if score < 0.7 || score > 0.9 {
		t.Errorf("expected score around 0.8 for single edit, got %f", score)
	}
}

func TestScore_EmptyStrings(t *testing.T) {
	score := Score("", "hello")
	if score != 0.0 {
		t.Errorf("expected score 0.0 for empty string, got %f", score)
	}

	score = Score("hello", "")
	if score != 0.0 {
		t.Errorf("expected score 0.0 for empty string, got %f", score)
	}
}

func TestScore_BothEmpty(t *testing.T) {
	score := Score("", "")
	if score != 1.0 {
		t.Errorf("expected score 1.0 for both empty, got %f", score)
	}
}

// Test FormatSuggestions
func TestFormatSuggestions_Empty(t *testing.T) {
	result := FormatSuggestions([]string{})
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestFormatSuggestions_Single(t *testing.T) {
	result := FormatSuggestions([]string{"Apple"})
	if result != "'Apple'" {
		t.Errorf("expected \"'Apple'\", got '%s'", result)
	}
}

func TestFormatSuggestions_Two(t *testing.T) {
	result := FormatSuggestions([]string{"Apple", "Banana"})
	if result != "'Apple' or 'Banana'" {
		t.Errorf("expected \"'Apple' or 'Banana'\", got '%s'", result)
	}
}

func TestFormatSuggestions_Multiple(t *testing.T) {
	result := FormatSuggestions([]string{"Apple", "Banana", "Cherry"})
	if result != "'Apple', 'Banana', or 'Cherry'" {
		t.Errorf("expected proper formatting, got '%s'", result)
	}
}

// Test NamedItemString
func TestNamedItemString(t *testing.T) {
	item := NamedItemString{ID: 42, Name: "Test Item"}
	if item.GetID() != 42 {
		t.Errorf("expected ID 42, got %d", item.GetID())
	}
	if item.GetName() != "Test Item" {
		t.Errorf("expected 'Test Item', got '%s'", item.GetName())
	}
}

// Test levenshteinDistance directly
func TestLevenshteinDistance_Identical(t *testing.T) {
	dist := levenshteinDistance("hello", "hello")
	if dist != 0 {
		t.Errorf("expected distance 0 for identical strings, got %d", dist)
	}
}

func TestLevenshteinDistance_Empty(t *testing.T) {
	dist := levenshteinDistance("", "hello")
	if dist != 5 {
		t.Errorf("expected distance 5 for empty to hello, got %d", dist)
	}
}

func TestLevenshteinDistance_SingleSubstitution(t *testing.T) {
	dist := levenshteinDistance("hello", "hallo")
	if dist != 1 {
		t.Errorf("expected distance 1 for single substitution, got %d", dist)
	}
}

func TestLevenshteinDistance_Insertion(t *testing.T) {
	dist := levenshteinDistance("hello", "helllo")
	if dist != 1 {
		t.Errorf("expected distance 1 for insertion, got %d", dist)
	}
}

func TestLevenshteinDistance_Deletion(t *testing.T) {
	dist := levenshteinDistance("hello", "helo")
	if dist != 1 {
		t.Errorf("expected distance 1 for deletion, got %d", dist)
	}
}

// Edge cases
func TestSearch_Unicode(t *testing.T) {
	items := items("日本語", "中文", "한국어")
	results := Search("日本", items, DefaultOptions())

	if len(results) < 1 {
		t.Errorf("expected at least 1 result for unicode, got %d", len(results))
	}
}

func TestSearch_SpecialCharacters(t *testing.T) {
	items := items("test@example.com", "user.name", "some-text")
	results := Search("test@", items, DefaultOptions())

	if len(results) < 1 {
		t.Errorf("expected at least 1 result for special chars, got %d", len(results))
	}
}

func TestSearch_Whitespace(t *testing.T) {
	items := items("  spaced  ", "normal", "trimmed")
	results := Search("spaced", items, DefaultOptions())

	// Query is trimmed, so should match
	if len(results) < 1 {
		t.Errorf("expected at least 1 result, got %d", len(results))
	}
}

// Concurrency tests
func TestSearch_Concurrent(t *testing.T) {
	items := items("Apple", "Banana", "Cherry", "Date", "Elderberry", "Fig", "Grape")
	queries := []string{"apple", "ban", "cher", "dat", "elder", "fig", "grap"}

	const goroutines = 100
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			query := queries[idx%len(queries)]
			results := Search(query, items, DefaultOptions())
			if len(results) == 0 {
				t.Errorf("expected results for query '%s'", query)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestBest_Concurrent(t *testing.T) {
	items := items("Active", "Pending", "Complete", "Cancelled", "Failed")
	queries := []string{"active", "pend", "comp", "cancel", "fail"}

	const goroutines = 100
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			query := queries[idx%len(queries)]
			result := Best(query, items, "status", DefaultOptions())
			if !result.Found {
				t.Errorf("expected to find match for query '%s'", query)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestScore_Concurrent(t *testing.T) {
	pairs := [][2]string{
		{"hello", "hallo"},
		{"world", "word"},
		{"test", "text"},
		{"fuzzy", "fuzz"},
	}

	const goroutines = 100
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			pair := pairs[idx%len(pairs)]
			score := Score(pair[0], pair[1])
			if score < 0.0 || score > 1.0 {
				t.Errorf("invalid score %f for pair %v", score, pair)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestLevenshteinDistance_Concurrent(t *testing.T) {
	pairs := [][2]string{
		{"hello", "hallo"},
		{"world", "word"},
		{"test", "text"},
		{"fuzzy", "fuzz"},
		{"concurrent", "concurent"},
	}

	const goroutines = 100
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			pair := pairs[idx%len(pairs)]
			dist := levenshteinDistance(pair[0], pair[1])
			if dist < 0 {
				t.Errorf("invalid distance %d for pair %v", dist, pair)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// Test buffer pool under concurrent load
func TestBufferPool_Concurrent(t *testing.T) {
	const goroutines = 200
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			// Stress test the buffer pool
			for j := 0; j < 10; j++ {
				_ = levenshteinDistance("concurrent test string", "concurent test strng")
			}
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// Test race conditions with -race flag
func TestSearch_RaceConditions(t *testing.T) {
	items := items("Item1", "Item2", "Item3", "Item4", "Item5")
	
	const goroutines = 50
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			// Mix of different operations
			Search("item", items, DefaultOptions())
			Best("item", items, "item", DefaultOptions())
			Score("item1", "item2")
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}
