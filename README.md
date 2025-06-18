# CLI Package

A lightweight package for building command-line tools in Go.

## Features

- Command and subcommand support
- Named arguments
- Flags
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
	"fmt"
	"os"

	"github.com/paularlott/cli"
)

func main() {
	cmd := &cli.Command{
		Name:        "example",
		Version:     "1.0.0",
		Usage:       "An example command",
		Description: "This is a simple example command to demonstrate the CLI package features.",
		Suggestions: true,
		ConfigFile: cli_toml.NewConfigFile(&configFile, func() []string {
			paths := []string{"."}

			home, err := os.UserHomeDir()
			if err == nil {
				paths = append(paths, home)
			}

			paths = append(paths, filepath.Join(home, ".config"))
			paths = append(paths, filepath.Join(home, ".config", "example"))

			return paths
		}),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				AssignTo: &configFile,
			},
			&cli.StringSliceFlag{
				Name:         "name",
				Usage:        "This is a string slice flag for testing the CLI library",
				Aliases:      []string{"n", "name2"},
				DefaultValue: []string{"World"},
				EnvVars:      []string{"EXAMPLE_NAME"},
				ConfigPath:   []string{"testing.name"},
				AssignTo:     &globalName,
				Required:     true,
			},
			&cli.IntFlag{Name: "count", Aliases: []string{"c"}, DefaultValue: 1, Usage: "Some number"},
			&cli.BoolFlag{Name: "verbose", DefaultValue: true, Global: true, Usage: "Enable verbose output"},
		},
		MaxArgs: cli.UnlimitedArgs,
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
			fmt.Println("Name:", cmd.GetStringSlice("name"))
			fmt.Println("Name Global:", globalName)
			fmt.Println("Count:", cmd.GetInt("count"))
			fmt.Println("Verbose:", cmd.GetBool("verbose"))
			fmt.Println("Config File:", configFile)

			fmt.Println("Arguments:", cmd.GetArgs())
			fmt.Println("Named Argument 'something':", cmd.GetStringArg("something"))
			fmt.Println("Named Argument 'number':", cmd.GetIntArg("number"))

			fmt.Println("Keys:", cmd.ConfigFile.Keys("server"))

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

## Shell Completion

The CLI supports generating shell completion scripts for Bash, Zsh, and Fish.

### Bash

```shell
# Generate the completion script
myapp completion bash > ~/.bash_completion.d/myapp
source ~/.bash_completion.d/myapp
```

### Zsh

```shell
# Generate the completion script
myapp completion zsh > "${fpath[1]}/_myapp"
```

On macOS the following 2 lines my be required in `~/.zshrc`:

```shell
autoload -U compinit
compinit
```

### Fish

```shell
myapp completion fish > ~/.config/fish/completions/myapp.fish
source ~/.config/fish/completions/myapp.fish
```
