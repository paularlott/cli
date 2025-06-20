package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/paularlott/cli"
	cli_toml "github.com/paularlott/cli/toml"
)

var (
	globalName []string
	configFile = "test.toml" // Example configuration file
)

func main() {
	cmd := &cli.Command{
		Name:    "file_watch",
		Version: "1.0.0",
		Usage:   "File Watch Example",
		Description: `This is an example command that watches the configuration file for changes.

If a change is detected then the flags are reloaded and the changed values are displayed.

This example must be built with -tags cli_watch to enable file watch support.`,
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
			},
			&cli.StringSliceFlag{
				Name:         "name",
				Usage:        "This is a string slice flag for testing the CLI library",
				Aliases:      []string{"n", "name2"},
				DefaultValue: []string{"World"},
				EnvVars:      []string{"EXAMPLE_NAME"},
				ConfigPath:   []string{"testing.name"},
				AssignTo:     &globalName,
			},
		},
		MaxArgs: cli.NoArgs,
		Run: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("Name:", cmd.GetStringSlice("name"))
			fmt.Println("Name Global:", globalName)

			// Watch for changes in the config file and reload the flags
			cmd.ConfigFile.OnChange(func() {
				fmt.Println("Config file changed:", cmd.ConfigFile.FileUsed())
				cmd.ReloadFlags()

				fmt.Println("Name:", cmd.GetStringSlice("name"))
				fmt.Println("Name Global:", globalName)
			})

			fmt.Println("\nWatching for changes, press ctrl+c to quit...")

			// Wait for ctrl+c
			shutdownChan := make(chan os.Signal, 1)
			signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

			<-shutdownChan

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
