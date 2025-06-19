# CLI Package

A lightweight package for building command-line tools in Go.

The library was created to allow the creation of CLI applications without the complexity and dependencies imposed by larger frameworks.

## Features

- Command and subcommand support
- Named arguments
- Flags including global flags
- Configuration file support (TOML and JSON)
- Environment variable support
- Built-in help and version commands
- Optional suggestions when command not found
- Command completions for Bash, Zsh, Fish and PowerShell

## Installation

```bash
go get github.com/paularlott/cli
```

## Usage

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
		Usage:       "An example command with subcommands",
		Description: "This is a simple example command to demonstrate the CLI package features.",
		Suggestions: true,
		ConfigFile: cli_toml.NewConfigFile(&configFile, func() []string {
			paths := []string{"."}

			home, err := os.UserHomeDir()
			if err == nil {
				paths = append(paths, home)
			}

			paths = append(paths, filepath.Join(home, ".config"))
			paths = append(paths, filepath.Join(home, ".config", "myapp"))

			return paths
		}),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				AssignTo: &configFile,
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:        "sub1",
				Usage:       "A subcommand with flags and arguments",
				Description: "This is a subcommand to demonstrate nested commands.",
				Suggestions: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "flag1",
						AssignTo: &sub1Flag,
					},
					&cli.IntFlag{Name: "number", Aliases: []string{"n"}, DefaultValue: 1, Usage: "Some number"},
				},
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:     "something",
						Usage:    "A required string argument",
						Required: true,
					},
					&cli.IntArg{
						Name:  "number",
						Usage: "An optional integer argument",
					},
				},
				Run: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Flag1:", sub1Flag)
					fmt.Println("Number:", cmd.GetInt("number"))
					fmt.Println("Arguments:", cmd.GetArgs())
					fmt.Println("Named Argument 'something':", cmd.GetStringArg("something"))
					fmt.Println("Named Argument 'number':", cmd.GetIntArg("number"))

					return nil
				},
			},
		},
		Run: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("Config File:", configFile)
			fmt.Println("Arguments:", cmd.GetArgs())

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

## Examples

### Basic Command

```bash
myapp --config config.toml hello
```

### Subcommand with Flags and Arguments

```bash
myapp sub1 -flag1 value 42 example_arg
```

### Shell Completion

#### Bash

```shell
# Generate the completion script
myapp completion bash > ~/.bash_completion.d/myapp
source ~/.bash_completion.d/myapp
```

#### Zsh

```shell
# Generate the completion script
myapp completion zsh > "${fpath[1]}/_myapp"
```

On macOS, you may need to add these lines to your `~/.zshrc`:

```shell
autoload -U compinit
compinit
```

#### Fish

```shell
myapp completion fish > ~/.config/fish/completions/myapp.fish
source ~/.config/fish/completions/myapp.fish
```
