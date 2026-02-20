// Package env provides functionality for loading .env files and setting environment variables.
// It supports variable expansion, comments, and quoted values.
package env

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var bracedVarRe = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// Load loads .env files. If no filenames are provided, it defaults to ".env".
// Each file is loaded in turn, with later files overriding values from earlier ones.
// It parses key=value pairs, expands variables, and sets them as environment variables.
// Variables can reference other environment variables using ${VAR} or $VAR syntax.
func Load(filenames ...string) error {
	if len(filenames) == 0 {
		filenames = []string{".env"}
	}

	for _, filename := range filenames {
		if err := loadFile(filename); err != nil {
			return err
		}
	}

	return nil
}

// loadFile loads a specific .env file path.
// It parses key=value pairs, expands variables, and sets them as environment variables.
func loadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines and full-line comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Parse key=value pair (also handles inline comment stripping)
		key, value, err := parseLine(line)
		if err != nil {
			return fmt.Errorf("error on line %d: %w", lineNum, err)
		}

		// Expand variables in the value
		value = expandVariables(value)

		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// stripInlineComment removes comments from the end of a line.
// It respects quoted strings and only removes # that are outside of quotes.
func stripInlineComment(line string) string {
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for i, r := range line {
		if escaped {
			escaped = false
			continue
		}

		switch r {
		case '\\':
			escaped = true
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case '#':
			// Only treat as comment if we're not in quotes
			if !inSingleQuote && !inDoubleQuote {
				return strings.TrimRight(line[:i], " \t")
			}
		}
	}

	return line
}

// parseLine parses a line in KEY=VALUE format.
// It supports quoted values (both single and double quotes) and handles escaping.
// Inline comments are properly handled by only stripping # that are outside of quotes.
func parseLine(line string) (string, string, error) {
	// Strip inline comments from the entire line first (respects quotes)
	line = stripInlineComment(line)

	// Find the first unescaped '='
	eqIdx := findUnescapedEquals(line)
	if eqIdx == -1 {
		return "", "", fmt.Errorf("invalid line format, expected KEY=VALUE")
	}

	key := strings.TrimSpace(line[:eqIdx])
	value := line[eqIdx+1:]

	// Trim leading space from value
	value = strings.TrimLeft(value, " \t")

	// Validate key
	if key == "" {
		return "", "", errors.New("empty key")
	}

	// Check for invalid characters in key
	if !isValidKey(key) {
		return "", "", fmt.Errorf("invalid key format: %s", key)
	}

	// Unquote value if necessary
	value = unquoteValue(value)

	return key, value, nil
}

// findUnescapedEquals finds the first '=' that is not escaped or inside quotes.
func findUnescapedEquals(line string) int {
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for i, r := range line {
		if escaped {
			escaped = false
			continue
		}

		switch r {
		case '\\':
			escaped = true
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case '=':
			if !inSingleQuote && !inDoubleQuote {
				return i
			}
		}
	}

	return -1
}

// isValidKey checks if a key is valid according to standard .env file conventions.
func isValidKey(key string) bool {
	if key == "" {
		return false
	}

	// Key must start with a letter or underscore
	firstChar := key[0]
	if !isAlpha(firstChar) && firstChar != '_' {
		return false
	}

	// Rest can be alphanumeric, underscore, or hyphen
	for _, r := range key[1:] {
		if !isAlphaNumeric(r) && r != '_' && r != '-' {
			return false
		}
	}

	return true
}

// isAlpha checks if a byte is an ASCII letter.
func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// isAlphaNumeric checks if a rune is alphanumeric.
func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// unquoteValue removes quotes from a value if present, handling escape sequences.
func unquoteValue(value string) string {
	if len(value) < 2 {
		return value
	}

	firstChar := value[0]
	lastChar := value[len(value)-1]

	// Check for double quotes
	if firstChar == '"' && lastChar == '"' {
		return unescapeString(value[1 : len(value)-1])
	}

	// Check for single quotes (no escape processing)
	if firstChar == '\'' && lastChar == '\'' {
		return value[1 : len(value)-1]
	}

	return value
}

// unescapeString processes escape sequences in a string.
func unescapeString(s string) string {
	var result strings.Builder
	escaped := false

	for _, r := range s {
		if escaped {
			switch r {
			case 'n':
				result.WriteRune('\n')
			case 'r':
				result.WriteRune('\r')
			case 't':
				result.WriteRune('\t')
			case '\\':
				result.WriteRune('\\')
			case '"':
				result.WriteRune('"')
			default:
				// If we don't recognize the escape, keep the backslash and the character
				result.WriteRune('\\')
				result.WriteRune(r)
			}
			escaped = false
		} else if r == '\\' {
			escaped = true
		} else {
			result.WriteRune(r)
		}
	}

	// If we end with an escaped backslash, add it
	if escaped {
		result.WriteRune('\\')
	}

	return result.String()
}

// expandVariables expands variable references in the form ${VAR} or $VAR.
func expandVariables(value string) string {
	// First, handle ${VAR} syntax
	result := expandBracedVariables(value)

	// Then, handle $VAR syntax (but be careful not to match ${VAR} again)
	result = expandSimpleVariables(result)

	return result
}

// expandBracedVariables expands ${VAR} style variables.
func expandBracedVariables(value string) string {
	return bracedVarRe.ReplaceAllStringFunc(value, func(match string) string {
		// Extract the variable name from ${VAR}
		varName := match[2 : len(match)-1]
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match
	})
}

// expandSimpleVariables expands $VAR style variables.
// This is more conservative and only matches valid variable names.
func expandSimpleVariables(value string) string {
	var result strings.Builder
	i := 0

	for i < len(value) {
		// Look for $ that's not escaped
		if value[i] == '$' && (i == 0 || value[i-1] != '\\') {
			// Start of variable reference
			j := i + 1

			// Variable must start with letter or underscore
			if j < len(value) && (isAlpha(value[j]) || value[j] == '_') {
				// Consume the rest of the variable name
				for j < len(value) && (isAlphaNumeric(rune(value[j])) || value[j] == '_') {
					j++
				}

				varName := value[i+1 : j]
				if val := os.Getenv(varName); val != "" {
					result.WriteString(val)
				} else {
					// Keep the original if variable is not set
					result.WriteString(value[i:j])
				}
				i = j
				continue
			}
		}

		result.WriteByte(value[i])
		i++
	}

	return result.String()
}
