# .env File Support Example

This example demonstrates how to use the `github.com/paularlott/cli/env` package to load `.env` files for your CLI application.

## Features

- Load environment variables from `.env` files
- Supports variable expansion (`${VAR}` and `$VAR` syntax)
- Handles inline and full-line comments
- Supports quoted values (both single and double quotes)
- Handles escape sequences in double-quoted values
- No external dependencies (uses only Go standard library)

## Usage

1. Copy `.env.example` to `.env`:

   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration:

   ```bash
   # .env file
   APP_NAME=myapp
   APP_PORT=8080
   DATABASE_URL=postgresql://user:pass@localhost:5432/mydb
   API_KEY=your-api-key-here
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

## Command-Line Flags vs Environment Variables

The application supports both command-line flags and environment variables:

```bash
# Using flags
go run main.go --db-url="postgresql://user:pass@localhost:5432/mydb" --api-key="your-key"

# Using environment variables (from .env)
DATABASE_URL=postgresql://user:pass@localhost:5432/mydb
API_KEY=your-key
go run main.go

# Environment variables take precedence over flags
# Flags can override environment variables
go run main.go --port=9000  # Overrides APP_PORT from .env
```

## .env File Syntax

### Basic Key-Value Pairs

```bash
APP_NAME=myapp
APP_PORT=8080
```

### Comments

```bash
# This is a full-line comment
APP_NAME=myapp  # This is an inline comment
```

### Quoted Values

```bash
# Double quotes - supports escape sequences
MESSAGE="Hello\nWorld"  # Will contain an actual newline

# Single quotes - no escape processing
MESSAGE='Hello\nWorld'  # Literal \n characters
```

### Variable Expansion

```bash
BASE_DIR=/usr/local
APP_NAME=myapp

# Variables can reference other variables
DATA_PATH=${BASE_DIR}/data
LOG_PATH=${BASE_DIR}/${APP_NAME}/logs

# Also supports simple $VAR syntax
URL=https://${HOST}:${PORT}/path
```

### Whitespace

Spaces around the `=` are permitted:

```bash
KEY=value
KEY = value
KEY= value
KEY  =  value
```

### Special Characters

```bash
# Use quotes for values with special characters
EMAIL="My App <noreply@example.com>"

# Hash symbols inside quotes are preserved
MESSAGE="This is #not a comment"
```

## Integration with CLI Flags

The `env` package sets environment variables which can then be read by CLI flags using the `EnvVars` option:

```go
&cli.StringFlag{
    Name:         "api-key",
    Usage:        "API key for external services",
    EnvVars:      []string{"API_KEY"},  // Read from API_KEY environment variable
    Required:     true,
    AssignTo:     &apiKey,
}
```

## Loading Multiple Files

You can load multiple `.env` files in order:

```go
// Load multiple files (later files override earlier ones)
if err := env.Load(".env.local", ".env"); err != nil {
    log.Fatal(err)
}
```

## Default Behavior

Calling `Load()` with no arguments defaults to loading `.env` from the current directory:

```go
// Equivalent to Load(".env")
if err := env.Load(); err != nil {
    log.Fatal(err)
}
```

## Error Handling

The `.env` file is optional in this example. If the file doesn't exist, the application continues with defaults:

```go
if err := env.Load(); err != nil {
    // Log but don't fail - .env is optional
    log.Printf("Note: .env file not found: %v", err)
}
```

If you want to require the `.env` file:

```go
if err := env.Load(); err != nil {
    log.Fatalf("Failed to load .env file: %v", err)
}
```
