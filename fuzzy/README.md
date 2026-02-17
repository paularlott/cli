# Fuzzy String Matching Package

A high-performance fuzzy string matching library for Go, optimized for CLI command suggestions and similar use cases.

## Features

- **Multi-tier matching algorithm**: Exact → Substring → Word boundary → Levenshtein distance
- **High performance**: 3-8x faster than naive implementations
- **Memory efficient**: Zero allocations for distance calculations via buffer pooling
- **Flexible API**: Search for multiple matches or find the best match
- **Type-safe**: Generic interface for matching any named items
- **Well-tested**: Comprehensive test suite with 37 test cases

## Installation

```bash
go get github.com/paularlott/cli/fuzzy
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/paularlott/cli/fuzzy"
)

func main() {
    // Create items to search
    items := []fuzzy.NamedItem{
        fuzzy.NamedItemString{ID: 1, Name: "server"},
        fuzzy.NamedItemString{ID: 2, Name: "start"},
        fuzzy.NamedItemString{ID: 3, Name: "stop"},
        fuzzy.NamedItemString{ID: 4, Name: "status"},
    }

    // Search for matches
    results := fuzzy.Search("srvr", items, fuzzy.DefaultOptions())

    for _, r := range results {
        fmt.Printf("%s (score: %.2f)\n", r.Name, r.Score)
    }
    // Output: server (score: 0.80)
}
```

## API

### Search Function

Find multiple matches sorted by relevance:

```go
results := fuzzy.Search(query string, items []NamedItem, opts Options) []Result
```

### Best Function

Find the single best match:

```go
result := fuzzy.Best(query string, items []NamedItem, entityType string, opts Options) BestResult
```

### Score Function

Calculate similarity between two strings:

```go
score := fuzzy.Score(s1, s2 string) float64  // Returns 0.0 to 1.0
```

### Options

```go
type Options struct {
    MaxResults int     // Maximum results to return (default: 5)
    Threshold  float64 // Minimum score threshold (default: 0.5)
}
```

## Custom Items

Implement the `NamedItem` interface for your types:

```go
type Command struct {
    ID   int
    Name string
}

func (c Command) GetID() int    { return c.ID }
func (c Command) GetName() string { return c.Name }

// Now use with fuzzy.Search
commands := []fuzzy.NamedItem{
    Command{ID: 1, Name: "deploy"},
    Command{ID: 2, Name: "build"},
}
```

## Algorithm

The fuzzy matcher uses a four-tier approach:

1. **Exact match** (score: 1.0) - Case-insensitive exact match
2. **Substring match** (score: 0.9) - Query is substring of name
3. **Word boundary match** (score: 0.85) - Query matches word prefix
4. **Levenshtein distance** (score: 0.0-1.0) - Edit distance based similarity

This ensures fast matches for common cases while falling back to fuzzy matching when needed.

## Use Cases

- CLI command suggestions ("Did you mean...?")
- Autocomplete and search
- Typo correction
- Tool/function name matching
- Any scenario requiring fuzzy string matching

## License

MIT License - See LICENSE.txt for details
