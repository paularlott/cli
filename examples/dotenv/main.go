package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/paularlott/cli"
	"github.com/paularlott/cli/env"
)

func main() {
	// Load .env file BEFORE executing the command
	// This sets environment variables that can be accessed by the CLI
	if err := env.Load(); err != nil {
		// .env file is optional - log but don't fail if it doesn't exist
		log.Printf("Note: .env file not found or could not be loaded: %v", err)
	}

	// Define flags that can read from environment variables
	var (
		appName   string
		appPort   int
		appDebug  bool
		dbURL     string
		apiKey    string
		logLevel  string
	)

	cmd := &cli.Command{
		Name:        "dotenv-example",
		Version:     "1.0.0",
		Usage:       "Example CLI with .env file support",
		Description: "This example demonstrates how to use the env package to load .env files for your CLI application.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:         "name",
				Usage:        "Application name",
				EnvVars:      []string{"APP_NAME"},
				DefaultValue: "myapp",
				AssignTo:     &appName,
			},
			&cli.IntFlag{
				Name:         "port",
				Usage:        "Application port",
				EnvVars:      []string{"APP_PORT"},
				DefaultValue: 8080,
				AssignTo:     &appPort,
			},
			&cli.BoolFlag{
				Name:         "debug",
				Usage:        "Enable debug mode",
				EnvVars:      []string{"APP_DEBUG"},
				AssignTo:     &appDebug,
			},
			&cli.StringFlag{
				Name:         "db-url",
				Usage:        "Database connection URL",
				EnvVars:      []string{"DATABASE_URL"},
				Required:     true,
				AssignTo:     &dbURL,
			},
			&cli.StringFlag{
				Name:         "api-key",
				Usage:        "API key for external services",
				EnvVars:      []string{"API_KEY"},
				Required:     true,
				AssignTo:     &apiKey,
			},
			&cli.StringFlag{
				Name:         "log-level",
				Usage:        "Logging level (debug, info, warn, error)",
				EnvVars:      []string{"LOG_LEVEL"},
				DefaultValue: "info",
				AssignTo:     &logLevel,
			},
		},
		Run: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("=== Configuration ===")
			fmt.Printf("Application Name: %s\n", appName)
			fmt.Printf("Port: %d\n", appPort)
			fmt.Printf("Debug Mode: %t\n", appDebug)
			fmt.Printf("Database URL: %s\n", dbURL)
			fmt.Printf("API Key: %s\n", maskAPIKey(apiKey))
			fmt.Printf("Log Level: %s\n", logLevel)

			// Show how to access environment variables directly
			fmt.Println("\n=== Direct Environment Access ===")
			if homeDir, ok := os.LookupEnv("HOME"); ok {
				fmt.Printf("HOME: %s\n", homeDir)
			}
			if path, ok := os.LookupEnv("PATH"); ok {
				fmt.Printf("PATH: %s\n", path)
			}

			return nil
		},
		Commands: []*cli.Command{
			cli.GenerateCompletionCommand(),
		},
	}

	// Execute the command
	err := cmd.Execute(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// maskAPIKey masks all but the first and last 4 characters of an API key
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
