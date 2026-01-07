# CLI Package

A simple and lightweight package for building command-line tools in Go.

This library was developed to address the need for creating CLI applications without the added complexity and dependencies of larger frameworks. It is designed to have a minimal footprint while maintaining functionality.

## Features

- Command and subcommand support
- Named arguments
- Flags including global flags
- Configuration file support (TOML and JSON)
- Environment variable support
- **.env file support** - Load environment variables from .env files with variable expansion
- Built-in help and version commands
- Optional suggestions when command not found
- Automatic help generation
- Command completions for Bash, Zsh, Fish and PowerShell
- Storing of flag values into variables
- Type safe

## Installation

```bash
go get github.com/paularlott/cli
```

Requires Go version 1.24.4 or later

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

The package includes a sub-package for loading `.env` files. This allows you to define environment variables in a file and have them automatically loaded into your application.

### Installation

```bash
go get github.com/paularlott/cli/env
```

### Quick Start

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

- **Variable Expansion**: Use `${VAR}` or `$VAR` syntax to reference other environment variables
- **Comments**: Full-line (`# comment`) and inline (`KEY=value # comment`) comments supported
- **Quoted Values**: Both single and double quotes with escape sequence support
- **Whitespace Handling**: Spaces around the `=` sign are permitted
- **Multiple Files**: Load multiple `.env` files in order
- **No Dependencies**: Uses only Go standard library

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

## Documentation

- [Arguments](docs/arguments.md)
- [Commands](docs/commands.md)
- [Configuration Files](docs/configuration_files.md)
- [Flags](docs/flags.md)
- [Shell Completion](docs/shell_completion.md)

## License

This project is licensed under the MIT License - see [LICENSE.txt](LICENSE.txt) file for details.
