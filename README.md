# CLI Package

A simple and lightweight package for building command-line tools in Go.

This library was developed to address the need for creating CLI applications without the added complexity and dependencies of larger frameworks. It is designed to have a minimal footprint while maintaining functionality.

## Features

- Command and subcommand support
- Named arguments
- Flags including global flags
- Configuration file support (TOML and JSON)
- Environment variable support
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
