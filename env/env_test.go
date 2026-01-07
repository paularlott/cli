package env

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a temporary .env file for testing
func createTempEnvFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return filePath
}

// Helper to clear all environment variables set during tests
func clearEnvVars(keys ...string) {
	for _, key := range keys {
		os.Unsetenv(key)
	}
}

func TestLoad_BasicKeyValue(t *testing.T) {
	content := `KEY1=value1
KEY2=value2
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value1" {
		t.Errorf("KEY1 = %q, want %q", got, "value1")
	}
	if got := os.Getenv("KEY2"); got != "value2" {
		t.Errorf("KEY2 = %q, want %q", got, "value2")
	}
}

func TestLoad_DefaultFilename(t *testing.T) {
	// Save current directory and restore after test
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	content := `KEY1=default_test
`
	if err := os.WriteFile(".env", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer clearEnvVars("KEY1")

	// Load with no arguments should use ".env" by default
	if err := Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "default_test" {
		t.Errorf("KEY1 = %q, want %q", got, "default_test")
	}
}

func TestLoad_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, ".env1")
	file2 := filepath.Join(tmpDir, ".env2")

	content1 := `KEY1=value1
KEY2=from_file1
`
	content2 := `KEY2=from_file2
KEY3=value3
`

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}
	defer clearEnvVars("KEY1", "KEY2", "KEY3")

	// Load multiple files - later files should override earlier ones
	if err := Load(file1, file2); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value1" {
		t.Errorf("KEY1 = %q, want %q", got, "value1")
	}
	if got := os.Getenv("KEY2"); got != "from_file2" {
		t.Errorf("KEY2 = %q, want %q (should be overridden by file2)", got, "from_file2")
	}
	if got := os.Getenv("KEY3"); got != "value3" {
		t.Errorf("KEY3 = %q, want %q", got, "value3")
	}
}

func TestLoad_SpacesAroundEquals(t *testing.T) {
	content := `KEY1 = value1
KEY2= value2
KEY3 =value3
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2", "KEY3")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value1" {
		t.Errorf("KEY1 = %q, want %q", got, "value1")
	}
	if got := os.Getenv("KEY2"); got != "value2" {
		t.Errorf("KEY2 = %q, want %q", got, "value2")
	}
	if got := os.Getenv("KEY3"); got != "value3" {
		t.Errorf("KEY3 = %q, want %q", got, "value3")
	}
}

func TestLoad_Comments(t *testing.T) {
	content := `# This is a comment
KEY1=value1
   # Another comment with leading spaces
KEY2=value2
#
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value1" {
		t.Errorf("KEY1 = %q, want %q", got, "value1")
	}
	if got := os.Getenv("KEY2"); got != "value2" {
		t.Errorf("KEY2 = %q, want %q", got, "value2")
	}
}

func TestLoad_InlineComments(t *testing.T) {
	content := `KEY1=value1 # this is an inline comment
KEY2=value2#this too
KEY3="value with # hash inside quotes" # and comment after
KEY4='value with # hash inside single quotes' # comment
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2", "KEY3", "KEY4")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value1" {
		t.Errorf("KEY1 = %q, want %q", got, "value1")
	}
	if got := os.Getenv("KEY2"); got != "value2" {
		t.Errorf("KEY2 = %q, want %q", got, "value2")
	}
	if got := os.Getenv("KEY3"); got != "value with # hash inside quotes" {
		t.Errorf("KEY3 = %q, want %q", got, "value with # hash inside quotes")
	}
	if got := os.Getenv("KEY4"); got != "value with # hash inside single quotes" {
		t.Errorf("KEY4 = %q, want %q", got, "value with # hash inside single quotes")
	}
}

func TestLoad_EmptyLines(t *testing.T) {
	content := `KEY1=value1

KEY2=value2


KEY3=value3
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2", "KEY3")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value1" {
		t.Errorf("KEY1 = %q, want %q", got, "value1")
	}
	if got := os.Getenv("KEY2"); got != "value2" {
		t.Errorf("KEY2 = %q, want %q", got, "value2")
	}
	if got := os.Getenv("KEY3"); got != "value3" {
		t.Errorf("KEY3 = %q, want %q", got, "value3")
	}
}

func TestLoad_DoubleQuotedValues(t *testing.T) {
	content := `KEY1="value with spaces"
KEY2="value with = equals"
KEY3="value with # hash"
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2", "KEY3")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value with spaces" {
		t.Errorf("KEY1 = %q, want %q", got, "value with spaces")
	}
	if got := os.Getenv("KEY2"); got != "value with = equals" {
		t.Errorf("KEY2 = %q, want %q", got, "value with = equals")
	}
	if got := os.Getenv("KEY3"); got != "value with # hash" {
		t.Errorf("KEY3 = %q, want %q", got, "value with # hash")
	}
}

func TestLoad_SingleQuotedValues(t *testing.T) {
	content := `KEY1='value with spaces'
KEY2='value with = equals'
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value with spaces" {
		t.Errorf("KEY1 = %q, want %q", got, "value with spaces")
	}
	if got := os.Getenv("KEY2"); got != "value with = equals" {
		t.Errorf("KEY2 = %q, want %q", got, "value with = equals")
	}
}

func TestLoad_EscapeSequences(t *testing.T) {
	content := `KEY1="line1\nline2"
KEY2="tab\there"
KEY3="back\\slash"
KEY4="quote\"inside"
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2", "KEY3", "KEY4")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "line1\nline2" {
		t.Errorf("KEY1 = %q, want newline", got)
	}
	if got := os.Getenv("KEY2"); got != "tab\there" {
		t.Errorf("KEY2 = %q, want tab", got)
	}
	if got := os.Getenv("KEY3"); got != `back\slash` {
		t.Errorf("KEY3 = %q, want backslash", got)
	}
	if got := os.Getenv("KEY4"); got != `quote"inside` {
		t.Errorf("KEY4 = %q, want quoted quote", got)
	}
}

func TestLoad_VariableExpansion_Braced(t *testing.T) {
	// Set up environment variables first
	os.Setenv("BASE_DIR", "/usr/local")
	os.Setenv("APP_NAME", "myapp")
	defer clearEnvVars("BASE_DIR", "APP_NAME", "DATA_PATH", "LOG_PATH")

	content := `DATA_PATH=${BASE_DIR}/data
LOG_PATH=${BASE_DIR}/${APP_NAME}/logs
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("DATA_PATH"); got != "/usr/local/data" {
		t.Errorf("DATA_PATH = %q, want %q", got, "/usr/local/data")
	}
	if got := os.Getenv("LOG_PATH"); got != "/usr/local/myapp/logs" {
		t.Errorf("LOG_PATH = %q, want %q", got, "/usr/local/myapp/logs")
	}
}

func TestLoad_VariableExpansion_Simple(t *testing.T) {
	os.Setenv("BASE_DIR", "/usr/local")
	os.Setenv("APP_NAME", "myapp")
	defer clearEnvVars("BASE_DIR", "APP_NAME", "DATA_PATH", "LOG_PATH")

	content := `DATA_PATH=$BASE_DIR/data
LOG_PATH=$BASE_DIR/$APP_NAME/logs
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("DATA_PATH"); got != "/usr/local/data" {
		t.Errorf("DATA_PATH = %q, want %q", got, "/usr/local/data")
	}
	if got := os.Getenv("LOG_PATH"); got != "/usr/local/myapp/logs" {
		t.Errorf("LOG_PATH = %q, want %q", got, "/usr/local/myapp/logs")
	}
}

func TestLoad_VariableExpansion_Mixed(t *testing.T) {
	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "8080")
	defer clearEnvVars("HOST", "PORT", "DATABASE_URL")

	content := `DATABASE_URL=postgresql://${HOST}:${PORT}/mydb?sslmode=disable
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("DATABASE_URL"); got != "postgresql://localhost:8080/mydb?sslmode=disable" {
		t.Errorf("DATABASE_URL = %q, want %q", got, "postgresql://localhost:8080/mydb?sslmode=disable")
	}
}

func TestLoad_VariableExpansion_ExistingInFile(t *testing.T) {
	defer clearEnvVars("FIRST", "SECOND", "COMBINED")

	content := `FIRST=hello
SECOND=world
COMBINED=${FIRST}-${SECOND}
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("FIRST"); got != "hello" {
		t.Errorf("FIRST = %q, want %q", got, "hello")
	}
	if got := os.Getenv("SECOND"); got != "world" {
		t.Errorf("SECOND = %q, want %q", got, "world")
	}
	if got := os.Getenv("COMBINED"); got != "hello-world" {
		t.Errorf("COMBINED = %q, want %q", got, "hello-world")
	}
}

func TestLoad_VariableExpansion_UndefiniedVariable(t *testing.T) {
	defer clearEnvVars("KEY1", "KEY2")

	content := `KEY1=${UNDEFINED_VAR}/value
KEY2=$ANOTHER_UNDEFINED/value
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Undefined variables should remain unchanged
	if got := os.Getenv("KEY1"); got != "${UNDEFINED_VAR}/value" {
		t.Errorf("KEY1 = %q, want %q", got, "${UNDEFINED_VAR}/value")
	}
	if got := os.Getenv("KEY2"); got != "$ANOTHER_UNDEFINED/value" {
		t.Errorf("KEY2 = %q, want %q", got, "$ANOTHER_UNDEFINED/value")
	}
}

func TestLoad_KeyWithHyphen(t *testing.T) {
	content := `API-KEY=test123
DB-NAME=mydb
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("API-KEY", "DB-NAME")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("API-KEY"); got != "test123" {
		t.Errorf("API-KEY = %q, want %q", got, "test123")
	}
	if got := os.Getenv("DB-NAME"); got != "mydb" {
		t.Errorf("DB-NAME = %q, want %q", got, "mydb")
	}
}

func TestLoad_InvalidKey(t *testing.T) {
	content := `123INVALID=value
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err == nil {
		t.Error("Load should have failed with invalid key")
	}
}

func TestLoad_EmptyKey(t *testing.T) {
	content := `=value
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err == nil {
		t.Error("Load should have failed with empty key")
	}
}

func TestLoad_NoEqualsSign(t *testing.T) {
	content := `INVALID_LINE
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err == nil {
		t.Error("Load should have failed with no equals sign")
	}
}

func TestLoadFile_FileNotFound(t *testing.T) {
	if err := Load("/nonexistent/path/.env"); err == nil {
		t.Error("Load should have failed with non-existent file")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	if err := Load("/nonexistent/path/.env"); err == nil {
		t.Error("Load should have failed with non-existent file")
	}
}

func TestLoad_EqualsInValue(t *testing.T) {
	content := `DATABASE_URL=postgresql://localhost:5432/db?sslmode=disable
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("DATABASE_URL")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expected := "postgresql://localhost:5432/db?sslmode=disable"
	if got := os.Getenv("DATABASE_URL"); got != expected {
		t.Errorf("DATABASE_URL = %q, want %q", got, expected)
	}
}

func TestLoad_EqualsInQuotedValue(t *testing.T) {
	content := `EQUATION="a=b+c=d"
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("EQUATION")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("EQUATION"); got != "a=b+c=d" {
		t.Errorf("EQUATION = %q, want %q", got, "a=b+c=d")
	}
}

func TestLoad_EmptyValue(t *testing.T) {
	content := `KEY1=
KEY2=""
KEY3=''
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2", "KEY3")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "" {
		t.Errorf("KEY1 = %q, want empty string", got)
	}
	if got := os.Getenv("KEY2"); got != "" {
		t.Errorf("KEY2 = %q, want empty string", got)
	}
	if got := os.Getenv("KEY3"); got != "" {
		t.Errorf("KEY3 = %q, want empty string", got)
	}
}

func TestLoad_ValueWithNewline(t *testing.T) {
	content := `MULTI_LINE="line1\nline2\nline3"
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("MULTI_LINE")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expected := "line1\nline2\nline3"
	if got := os.Getenv("MULTI_LINE"); got != expected {
		t.Errorf("MULTI_LINE = %q, want %q", got, expected)
	}
}

func TestParseLine_ValidKeys(t *testing.T) {
	tests := []struct {
		name string
		line string
		want struct {
			key   string
			value string
		}
	}{
		{
			name: "simple key",
			line: "KEY=value",
			want: struct{ key, value string }{"KEY", "value"},
		},
		{
			name: "key with underscore",
			line: "KEY_NAME=value",
			want: struct{ key, value string }{"KEY_NAME", "value"},
		},
		{
			name: "key with hyphen",
			line: "KEY-NAME=value",
			want: struct{ key, value string }{"KEY-NAME", "value"},
		},
		{
			name: "key starting with underscore",
			line: "_PRIVATE_KEY=value",
			want: struct{ key, value string }{"_PRIVATE_KEY", "value"},
		},
		{
			name: "key with numbers",
			line: "KEY123=value",
			want: struct{ key, value string }{"KEY123", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, err := parseLine(tt.line)
			if err != nil {
				t.Fatalf("parseLine() error = %v", err)
			}
			if key != tt.want.key {
				t.Errorf("key = %q, want %q", key, tt.want.key)
			}
			if value != tt.want.value {
				t.Errorf("value = %q, want %q", value, tt.want.value)
			}
		})
	}
}

func TestParseLine_InvalidKeys(t *testing.T) {
	tests := []string{
		"123KEY=value",      // starts with number
		"=value",            // empty key
		"KEY@NAME=value",    // invalid character
		"KEY NAME=value",    // space in key
		"KEY.NAME=value",    // dot in key
	}

	for _, line := range tests {
		t.Run(line, func(t *testing.T) {
			_, _, err := parseLine(line)
			if err == nil {
				t.Error("parseLine() should have returned an error")
			}
		})
	}
}

func TestLoad_RealWorldExample(t *testing.T) {
	os.Setenv("HOME", "/home/user")
	defer clearEnvVars("HOME", "APP_ENV", "APP_PORT", "APP_HOST", "APP_DB_URL", "APP_LOG_LEVEL")

	content := `# Application Configuration
APP_ENV=production
APP_PORT=8080
APP_HOST=${HOST:-localhost}
APP_DB_URL=postgresql://user:pass@localhost:5432/mydb?sslmode=disable
APP_LOG_LEVEL=info
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("APP_ENV"); got != "production" {
		t.Errorf("APP_ENV = %q, want %q", got, "production")
	}
	if got := os.Getenv("APP_PORT"); got != "8080" {
		t.Errorf("APP_PORT = %q, want %q", got, "8080")
	}
}

func TestLoad_MultipleEqualsInValue(t *testing.T) {
	content := `QUERY=SELECT * FROM users WHERE active=true AND role='admin'
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("QUERY")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expected := "SELECT * FROM users WHERE active=true AND role='admin'"
	if got := os.Getenv("QUERY"); got != expected {
		t.Errorf("QUERY = %q, want %q", got, expected)
	}
}

func TestExpandVariables_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		setup    func()
		expected string
	}{
		{
			name:     "empty value",
			value:    "",
			setup:    func() {},
			expected: "",
		},
		{
			name:     "no variables",
			value:    "plain text",
			setup:    func() {},
			expected: "plain text",
		},
		{
			name:     "variable at start",
			value:    "${VAR}/path",
			setup:    func() { os.Setenv("VAR", "home") },
			expected: "home/path",
		},
		{
			name:     "variable at end",
			value:    "/path/${VAR}",
			setup:    func() { os.Setenv("VAR", "end") },
			expected: "/path/end",
		},
		{
			name: "multiple variables",
			value: "${A}/${B}/${C}",
			setup: func() {
				os.Setenv("A", "1")
				os.Setenv("B", "2")
				os.Setenv("C", "3")
			},
			expected: "1/2/3",
		},
		{
			name:     "simple dollar",
			value:    "$VAR/path",
			setup:    func() { os.Setenv("VAR", "simple") },
			expected: "simple/path",
		},
		{
			name:     "escaped dollar",
			value:    "\\$VAR/notexpended",
			setup:    func() { os.Setenv("VAR", "shouldnotexpand") },
			expected: "\\$VAR/notexpended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := expandVariables(tt.value)
			if result != tt.expected {
				t.Errorf("expandVariables(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestUnquoteValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"no quotes", "value", "value"},
		{"double quotes", `"value"`, "value"},
		{"single quotes", `'value'`, "value"},
		{"double quotes with spaces", `"value with spaces"`, "value with spaces"},
		{"single quotes with spaces", `'value with spaces'`, "value with spaces"},
		{"unmatched double quotes", `"value`, `"value`},
		{"unmatched single quotes", `'value`, `'value`},
		{"empty double quotes", `""`, ""},
		{"empty single quotes", `''`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unquoteValue(tt.value)
			if result != tt.expected {
				t.Errorf("unquoteValue(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestUnescapeString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"no escape", "value", "value"},
		{"newline", `line\nline`, "line\nline"},
		{"carriage return", `line\rline`, "line\rline"},
		{"tab", `val\tue`, "val\tue"},
		{"backslash", `path\\to\\file`, `path\to\file`},
		{"quote", `quote\"inside`, `quote"inside`},
		{"mixed", `a\nb\tc\\d"e`, "a\nb\tc\\d\"e"},
		{"escape at end", `value\\`, "value\\"},
		{"unknown escape", `\x`, "\\x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unescapeString(tt.value)
			if result != tt.expected {
				t.Errorf("unescapeString(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestLoad_UTF8Values(t *testing.T) {
	content := `KEY1=hello世界
KEY2="emoji 🎉"
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "hello世界" {
		t.Errorf("KEY1 = %q, want %q", got, "hello世界")
	}
	if got := os.Getenv("KEY2"); got != "emoji 🎉" {
		t.Errorf("KEY2 = %q, want %q", got, "emoji 🎉")
	}
}

func TestLoad_UnicodeEscape(t *testing.T) {
	content := `JSON_VALUE="{\"key\": \"value\"}"
`
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("JSON_VALUE")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expected := `{"key": "value"}`
	if got := os.Getenv("JSON_VALUE"); got != expected {
		t.Errorf("JSON_VALUE = %q, want %q", got, expected)
	}
}

func TestFindUnescapedEquals(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"simple", "KEY=value", 3},
		{"with spaces", "KEY = value", 4},
		{"in quotes", `KEY="value=test"`, 3},
		{"multiple", "A=B=C", 1},
		{"escaped", `KEY=value\=test`, 3},
		{"in single quotes", `KEY='value=test'`, 3},
		{"no equals", "KEYvalue", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findUnescapedEquals(tt.line)
			if result != tt.expected {
				t.Errorf("findUnescapedEquals(%q) = %d, want %d", tt.line, result, tt.expected)
			}
		})
	}
}

func TestIsValidKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"simple", "KEY", true},
		{"with underscore", "KEY_NAME", true},
		{"with hyphen", "KEY-NAME", true},
		{"with numbers", "KEY123", true},
		{"starting with underscore", "_PRIVATE", true},
		{"single letter", "K", true},
		{"empty", "", false},
		{"starts with number", "1KEY", false},
		{"with dot", "KEY.NAME", false},
		{"with space", "KEY NAME", false},
		{"with special char", "KEY@NAME", false},
		{"only numbers", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidKey(tt.key)
			if result != tt.expected {
				t.Errorf("isValidKey(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestStripInlineComment(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"no comment", "KEY=value", "KEY=value"},
		{"comment at end", "KEY=value # comment", "KEY=value"},
		{"comment with no space", "KEY=value#comment", "KEY=value"},
		{"hash in quotes", `KEY="val#ue" # comment`, `KEY="val#ue"`},
		{"hash in single quotes", `KEY='val#ue' # comment`, `KEY='val#ue'`},
		{"escaped hash", `KEY=value\#notcomment`, `KEY=value\#notcomment`},
		{"multiple hashes", `KEY=#value`, `KEY=`},
		{"only comment", `# comment`, ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripInlineComment(tt.line)
			if result != tt.expected {
				t.Errorf("stripInlineComment(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

// Benchmark for performance testing
func BenchmarkLoadFile(b *testing.B) {
	content := `KEY1=value1
KEY2=value2
KEY3=value3
KEY4=value4
KEY5=value5
`
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load(filePath)
	}
}

func BenchmarkExpandVariables(b *testing.B) {
	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	defer clearEnvVars("VAR1", "VAR2")

	value := "prefix-${VAR1}-${VAR2}-suffix"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expandVariables(value)
	}
}

// Test that we properly handle Windows-style line endings
func TestLoad_WindowsLineEndings(t *testing.T) {
	content := "KEY1=value1\r\nKEY2=value2\r\n"
	filePath := createTempEnvFile(t, content)
	defer clearEnvVars("KEY1", "KEY2")

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value1" {
		t.Errorf("KEY1 = %q, want %q", got, "value1")
	}
	if got := os.Getenv("KEY2"); got != "value2" {
		t.Errorf("KEY2 = %q, want %q", got, "value2")
	}
}

// Test variable expansion within quoted strings
func TestLoad_VariableExpansionInQuotes(t *testing.T) {
	os.Setenv("USER", "john")
	defer clearEnvVars("USER", "KEY")

	content := `KEY="hello ${USER}"
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY"); got != "hello john" {
		t.Errorf("KEY = %q, want %q", got, "hello john")
	}
}

// Test complex real-world scenario
func TestLoad_ComplexRealWorld(t *testing.T) {
	os.Setenv("HOME", "/home/user")
	os.Setenv("USER", "testuser")
	defer clearEnvVars("HOME", "USER", "APP_NAME", "APP_PORT", "APP_HOST", "APP_DB_HOST", "APP_DB_PORT", "APP_DB_NAME", "APP_REDIS_URL")

	content := `# Application Configuration
APP_NAME=myapp
APP_PORT=${APP_PORT:-8080}
APP_HOST=0.0.0.0

# Database Configuration
APP_DB_HOST=${APP_DB_HOST:-localhost}
APP_DB_PORT=${APP_DB_PORT:-5432}
APP_DB_NAME=mydb

# Redis URL
APP_REDIS_URL=redis://${APP_REDIS_HOST:-localhost}:${APP_REDIS_PORT:-6379}

# Mixed expansion
MIXED_VAR=${HOME}/apps/${APP_NAME}
`
	filePath := createTempEnvFile(t, content)

	if err := Load(filePath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("APP_NAME"); got != "myapp" {
		t.Errorf("APP_NAME = %q, want %q", got, "myapp")
	}
	if got := os.Getenv("MIXED_VAR"); got != "/home/user/apps/myapp" {
		t.Errorf("MIXED_VAR = %q, want %q", got, "/home/user/apps/myapp")
	}
}

// Integration tests using testdata files

func TestLoadIntegration_Basic(t *testing.T) {
	// Set HOME for variable expansion tests
	os.Setenv("HOME", "/home/testuser")
	defer os.Unsetenv("HOME")
	defer clearEnvVars("APP_NAME", "APP_VERSION", "APP_PORT")

	if err := Load("testdata/basic.env"); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("APP_NAME"); got != "myapp" {
		t.Errorf("APP_NAME = %q, want %q", got, "myapp")
	}
	if got := os.Getenv("APP_VERSION"); got != "1.0.0" {
		t.Errorf("APP_VERSION = %q, want %q", got, "1.0.0")
	}
	if got := os.Getenv("APP_PORT"); got != "8080" {
		t.Errorf("APP_PORT = %q, want %q", got, "8080")
	}
}

func TestLoadIntegration_Comments(t *testing.T) {
	defer clearEnvVars("KEY1", "KEY2", "KEY3", "KEY4", "KEY5")

	if err := Load("testdata/comments.env"); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value1" {
		t.Errorf("KEY1 = %q, want %q", got, "value1")
	}
	if got := os.Getenv("KEY2"); got != "value2" {
		t.Errorf("KEY2 = %q, want %q", got, "value2")
	}
	if got := os.Getenv("KEY3"); got != "value3" {
		t.Errorf("KEY3 = %q, want %q", got, "value3")
	}
	if got := os.Getenv("KEY4"); got != "value with # hash inside" {
		t.Errorf("KEY4 = %q, want %q", got, "value with # hash inside")
	}
	if got := os.Getenv("KEY5"); got != "value with # hash inside" {
		t.Errorf("KEY5 = %q, want %q", got, "value with # hash inside")
	}
}

func TestLoadIntegration_Quotes(t *testing.T) {
	defer clearEnvVars("DOUBLE_QUOTED", "DOUBLE_EQUALS", "DOUBLE_HASH", "SINGLE_QUOTED", "SINGLE_EQUALS", "ESCAPED_NEWLINE", "EMPTY_UNQUOTED", "EMPTY_DOUBLE", "EMPTY_SINGLE")

	if err := Load("testdata/quotes.env"); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("DOUBLE_QUOTED"); got != "value with spaces" {
		t.Errorf("DOUBLE_QUOTED = %q, want %q", got, "value with spaces")
	}
	if got := os.Getenv("DOUBLE_EQUALS"); got != "value with = equals" {
		t.Errorf("DOUBLE_EQUALS = %q, want %q", got, "value with = equals")
	}
	if got := os.Getenv("DOUBLE_HASH"); got != "value with # hash" {
		t.Errorf("DOUBLE_HASH = %q, want %q", got, "value with # hash")
	}
	if got := os.Getenv("SINGLE_QUOTED"); got != "value with spaces" {
		t.Errorf("SINGLE_QUOTED = %q, want %q", got, "value with spaces")
	}
	if got := os.Getenv("ESCAPED_NEWLINE"); got != "line1\nline2" {
		t.Errorf("ESCAPED_NEWLINE = %q, want newline", got)
	}
	if got := os.Getenv("EMPTY_UNQUOTED"); got != "" {
		t.Errorf("EMPTY_UNQUOTED = %q, want empty string", got)
	}
	if got := os.Getenv("EMPTY_DOUBLE"); got != "" {
		t.Errorf("EMPTY_DOUBLE = %q, want empty string", got)
	}
}

func TestLoadIntegration_Variables(t *testing.T) {
	defer clearEnvVars("BASE_DIR", "APP_NAME", "DB_HOST", "DB_PORT", "DATA_PATH", "LOG_PATH", "DATABASE_URL")

	if err := Load("testdata/variables.env"); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("BASE_DIR"); got != "/usr/local" {
		t.Errorf("BASE_DIR = %q, want %q", got, "/usr/local")
	}
	if got := os.Getenv("DATA_PATH"); got != "/usr/local/data" {
		t.Errorf("DATA_PATH = %q, want %q", got, "/usr/local/data")
	}
	if got := os.Getenv("LOG_PATH"); got != "/usr/local/myapp/logs" {
		t.Errorf("LOG_PATH = %q, want %q", got, "/usr/local/myapp/logs")
	}
	if got := os.Getenv("DATABASE_URL"); got != "postgresql://localhost:5432/mydb?sslmode=disable" {
		t.Errorf("DATABASE_URL = %q, want %q", got, "postgresql://localhost:5432/mydb?sslmode=disable")
	}
}

func TestLoadIntegration_Whitespace(t *testing.T) {
	defer clearEnvVars("KEY1", "KEY2", "KEY3", "KEY4", "KEY5", "KEY6", "KEY7", "KEY8", "KEY9")

	if err := Load("testdata/whitespace.env"); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("KEY1"); got != "value" {
		t.Errorf("KEY1 = %q, want %q", got, "value")
	}
	if got := os.Getenv("KEY2"); got != "value" {
		t.Errorf("KEY2 = %q, want %q", got, "value")
	}
	if got := os.Getenv("KEY3"); got != "value" {
		t.Errorf("KEY3 = %q, want %q", got, "value")
	}
	if got := os.Getenv("KEY4"); got != "value" {
		t.Errorf("KEY4 = %q, want %q", got, "value")
	}
	if got := os.Getenv("KEY9"); got != "value  with  spaces" {
		t.Errorf("KEY9 = %q, want %q", got, "value  with  spaces")
	}
}

func TestLoadIntegration_Complex(t *testing.T) {
	defer clearEnvVars("APP_ENV", "APP_DEBUG", "APP_PORT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSL_MODE", "DATABASE_URL", "REDIS_HOST", "REDIS_PORT", "REDIS_URL", "EMAIL_FROM", "API_RATE_LIMIT", "API_TIMEOUT", "BASE_URL", "API_ENDPOINT")

	if err := Load("testdata/complex.env"); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := os.Getenv("APP_ENV"); got != "production" {
		t.Errorf("APP_ENV = %q, want %q", got, "production")
	}
	if got := os.Getenv("APP_PORT"); got != "8080" {
		t.Errorf("APP_PORT = %q, want %q", got, "8080")
	}
	if got := os.Getenv("DATABASE_URL"); got != "postgresql://admin:secret123@localhost:5432/myapp_db?sslmode=disable" {
		t.Errorf("DATABASE_URL = %q, want %q", got, "postgresql://admin:secret123@localhost:5432/myapp_db?sslmode=disable")
	}
	if got := os.Getenv("REDIS_URL"); got != "redis://localhost:6379/0" {
		t.Errorf("REDIS_URL = %q, want %q", got, "redis://localhost:6379/0")
	}
	if got := os.Getenv("EMAIL_FROM"); got != "My App <noreply@myapp.com>" {
		t.Errorf("EMAIL_FROM = %q, want %q", got, "My App <noreply@myapp.com>")
	}
	if got := os.Getenv("API_TIMEOUT"); got != "30" {
		t.Errorf("API_TIMEOUT = %q, want %q", got, "30")
	}
	if got := os.Getenv("API_ENDPOINT"); got != "https://api.example.com/v1/users" {
		t.Errorf("API_ENDPOINT = %q, want %q", got, "https://api.example.com/v1/users")
	}
}
