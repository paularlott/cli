package fuzzy

import (
	"fmt"
	"math/rand"
	"testing"
)

// Generate test items for benchmarking
func generateItems(count int) []NamedItem {
	adjectives := []string{"Red", "Blue", "Green", "Yellow", "Purple", "Orange", "Pink", "Black", "White", "Gray"}
	nouns := []string{"Project", "Task", "Item", "Request", "Order", "Customer", "Product", "Service", "Team", "Group"}

	items := make([]NamedItem, count)
	for i := 0; i < count; i++ {
		adj := adjectives[rand.Intn(len(adjectives))]
		noun := nouns[rand.Intn(len(nouns))]
		items[i] = testItem{
			id:   i + 1,
			name: fmt.Sprintf("%s %s %d", adj, noun, i),
		}
	}
	return items
}

func BenchmarkSearch_10Items(b *testing.B) {
	items := generateItems(10)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search("Red Project", items, opts)
	}
}

func BenchmarkSearch_100Items(b *testing.B) {
	items := generateItems(100)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search("Red Project", items, opts)
	}
}

func BenchmarkSearch_1000Items(b *testing.B) {
	items := generateItems(1000)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search("Red Project", items, opts)
	}
}

func BenchmarkSearch_10000Items(b *testing.B) {
	items := generateItems(10000)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search("Red Project", items, opts)
	}
}

func BenchmarkSearch_ExactMatch(b *testing.B) {
	items := generateItems(1000)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search("Red Project 0", items, opts)
	}
}

func BenchmarkSearch_NoMatch(b *testing.B) {
	items := generateItems(1000)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search("zzzzzzzzz", items, opts)
	}
}

func BenchmarkBest_100Items(b *testing.B) {
	items := generateItems(100)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Best("Red Project", items, "item", opts)
	}
}

func BenchmarkBest_1000Items(b *testing.B) {
	items := generateItems(1000)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Best("Red Project", items, "item", opts)
	}
}

func BenchmarkScore(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Score("hello world", "hallo world")
	}
}

func BenchmarkLevenshteinDistance_Short(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		levenshteinDistance("hello", "hallo")
	}
}

func BenchmarkLevenshteinDistance_Medium(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		levenshteinDistance("hello world test", "hallo world tast")
	}
}

func BenchmarkLevenshteinDistance_Long(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		levenshteinDistance(
			"this is a much longer string for testing performance",
			"this is a much longer string for testing performanc",
		)
	}
}

// Benchmark with different query patterns
func BenchmarkSearch_SubstringQuery(b *testing.B) {
	items := generateItems(1000)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search("Project", items, opts)
	}
}

func BenchmarkSearch_TypoQuery(b *testing.B) {
	items := generateItems(1000)
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search("Projct", items, opts)
	}
}
