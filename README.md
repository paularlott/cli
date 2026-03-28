# CLI Package

A simple and lightweight package for building command-line tools in Go.

This library was developed to address the need for creating CLI applications without the added complexity and dependencies of larger frameworks. It is designed to have a minimal footprint while maintaining functionality.

## Features

- Command and subcommand support
- Named arguments
- Flags including global flags
- Configuration file support (TOML and JSON)
- Environment variable support
- .env file support with variable expansion
- Built-in help and version commands
- Optional suggestions when command not found
- Automatic help generation
- Shell completions for Bash, Zsh, Fish, and PowerShell
- Storing of flag values into variables
- Type safe
- Full-screen TUI framework (themes, streaming, menus, panels)

## Installation

```bash
go get github.com/paularlott/cli
```

Requires Go 1.26.1 or later

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/paularlott/cli"
)

func main() {
	cmd := &cli.Command{
		Name:        "myapp",
		Version:     "1.0.0",
		Usage:       "Simple Example",
		Description: "This is a simple example command to demonstrate the CLI package features.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Your name",
			},
		},
		Run: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("Hello:", cmd.GetString("name"))

			return nil
		},
	}

	err := cmd.Execute(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	os.Exit(0)
}
```

## .env File Support

The `env` sub-package loads `.env` files with variable expansion, comments, and quoted values:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/paularlott/cli"
    "github.com/paularlott/cli/env"
)

func main() {
    // Load .env file BEFORE executing the command
    if err := env.Load(); err != nil {
        fmt.Printf("Warning: .env file not found: %v\n", err)
    }

    cmd := &cli.Command{
        Name:  "myapp",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "database-url",
                EnvVars: []string{"DATABASE_URL"}, // Read from environment
            },
        },
        Run: func(ctx context.Context, cmd *cli.Command) error {
            dbURL := cmd.GetString("database-url")
            fmt.Println("Database URL:", dbURL)
            return nil
        },
    }

    cmd.Execute(context.Background())
}
```

### .env File Example

```bash
# .env file
DATABASE_URL=postgresql://user:pass@localhost:5432/mydb
API_KEY=secret-key
DEBUG=true

# Variable expansion
BASE_DIR=/usr/local
LOG_PATH=${BASE_DIR}/logs
```

### Features

- Variable expansion with `${VAR}` or `$VAR`
- Full-line and inline comments
- Single/double quoted values with escape sequences
- Multiple files loaded in order

For more details, see the [dotenv example](examples/dotenv/) or the [env package documentation](env/).

## Help Command

When the help is enabled `-h` or `--help` will show the usage information for the current command.

### Help Syntax Notation

| Syntax        | Description                   |
| ------------- | ----------------------------- |
| `<required>`  | A required argument           |
| `[optional]`  | An optional argument          |
| `[args...]`   | Additional optional arguments |
| `<args...>`   | Additional arguments          |
| `[flags]`     | Command flags                 |
| `[command]`   | Subcommands available         |

## TUI Package

The `tui` sub-package provides a full-screen terminal UI framework for building interactive CLI applications — chat interfaces, log viewers, AI assistants, and more. It supports themes, slash-command palettes, streaming output, menus, spinner/progress indicators, and output-only mode.

```bash
go get github.com/paularlott/cli/tui
```

See the [tui package README](tui/README.md) for full documentation.

## Documentation

- [Arguments](docs/arguments.md)
- [Commands](docs/commands.md)
- [Configuration Files](docs/configuration_files.md)
- [Flags](docs/flags.md)
- [Shell Completion](docs/shell_completion.md)

## License

This project is licensed under the MIT License - see [LICENSE.txt](LICENSE.txt) file for details.
