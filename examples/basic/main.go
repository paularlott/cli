package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paularlott/cli"
	cli_toml "github.com/paularlott/cli/toml"
)

var (
	globalName []string
	configFile = "test.toml" // Example configuration file
)

func main() {
	cmd := &cli.Command{
		Name:        "example",
		Version:     "1.0.0",
		Usage:       "An example command",
		Description: "This is a simple example command to demonstrate the CLI package features.",
		Suggestions: true,
		ConfigFile: cli_toml.NewConfigFile(&configFile, func() []string {
			// Look for the config file in:
			//   - The current directory
			//   - The user's home directory
			//   - The user's .config directory

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
				Global:   true,
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
			&cli.IntFlag{
				Name:         "count",
				Aliases:      []string{"c"},
				DefaultValue: 1,
				Usage:        "Some number",
				ValidateFlag: func(c *cli.Command) error {
					if c.GetInt("count") > 10 {
						return fmt.Errorf("count must be 10 or less")
					}
					return nil
				},
			},
			&cli.BoolFlag{Name: "verbose", DefaultValue: true, Global: true, Usage: "Enable verbose output"},
		},
		MaxArgs: cli.UnlimitedArgs,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:     "something",
				Usage:    "A required string argument",
				Required: true,
				ValidateArg: func(c *cli.Command) error {
					if len(c.GetStringArg("something")) > 2 {
						return fmt.Errorf("argument 'something' is must be less than 2 characters long")
					}
					return nil
				},
			},
			&cli.IntArg{
				Name:  "number",
				Usage: "An optional integer argument",
			},
		},
		PreRun: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			ctx = context.WithValue(ctx, "exampleKey", "exampleValue")
			return ctx, nil
		},
		PostRun: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("Running after run")
			return nil
		},
		Run: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("Context Value:", ctx.Value("exampleKey"))

			fmt.Println("Name:", cmd.GetStringSlice("name"))
			fmt.Println("Name Global:", globalName)
			fmt.Println("Count:", cmd.GetInt("count"))
			fmt.Println("Verbose:", cmd.GetBool("verbose"))
			fmt.Println("Config File:", configFile)

			fmt.Println("Arguments:", cmd.GetArgs())
			fmt.Println("Named Argument 'something':", cmd.GetStringArg("something"))
			fmt.Println("Named Argument 'number':", cmd.GetIntArg("number"))

			fmt.Println("Keys:", cmd.ConfigFile.GetKeys("testing"))

			return nil
		},
		Commands: []*cli.Command{
			cli.GenerateCompletionCommand(),
			{
				Name:  "greet",
				Usage: "Greet someone",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "greeting", Aliases: []string{"g"}, DefaultValue: "Hello", DefaultText: "Test Default"},
				},
				Run: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Greetings")
					fmt.Println("Greeting:", cmd.GetString("greeting"))
					return nil
				},
			},
		},
	}

	err := cmd.Execute(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	os.Exit(0)
}
