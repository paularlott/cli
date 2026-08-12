package cli

import (
	"context"
	"os"
	"testing"
)

func TestCommand_AliasExecution(t *testing.T) {
	// Test that a command can be invoked by its alias
	executed := false
	var executedCmd string

	cmd := &Command{
		Name:    "app",
		Version: "1.0.0",
		Commands: []*Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List items",
				Run: func(ctx context.Context, cmd *Command) error {
					executed = true
					executedCmd = "list"
					return nil
				},
			},
		},
	}

	os.Args = []string{"app", "ls"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed via alias")
	}
	if executedCmd != "list" {
		t.Fatalf("expected 'list' command to run, got %q", executedCmd)
	}
}

func TestCommand_AliasAndPrimaryName(t *testing.T) {
	// Test that both the primary name and alias work
	tests := []struct {
		name string
		arg  string
	}{
		{"primary name", "copy"},
		{"alias", "cp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executed := false

			cmd := &Command{
				Name: "app",
				Commands: []*Command{
					{
						Name:    "copy",
						Aliases: []string{"cp"},
						Usage:   "Copy items",
						Run: func(ctx context.Context, cmd *Command) error {
							executed = true
							return nil
						},
					},
				},
			}

			os.Args = []string{"app", tt.arg}
			err := cmd.Execute(context.Background())
			if err != nil {
				t.Fatalf("expected no error via %s, got %v", tt.name, err)
			}
			if !executed {
				t.Fatalf("expected command to execute via %s", tt.name)
			}
		})
	}
}

func TestCommand_MultipleAliases(t *testing.T) {
	// Test that multiple aliases all work
	for _, alias := range []string{"remove", "rm", "del", "delete"} {
		t.Run(alias, func(t *testing.T) {
			executed := false

			cmd := &Command{
				Name: "app",
				Commands: []*Command{
					{
						Name:    "remove",
						Aliases: []string{"rm", "del", "delete"},
						Usage:   "Remove items",
						Run: func(ctx context.Context, cmd *Command) error {
							executed = true
							return nil
						},
					},
				},
			}

			os.Args = []string{"app", alias}
			err := cmd.Execute(context.Background())
			if err != nil {
				t.Fatalf("expected no error for %q, got %v", alias, err)
			}
			if !executed {
				t.Fatalf("expected command to execute via %q", alias)
			}
		})
	}
}

func TestCommand_AliasWithFlags(t *testing.T) {
	// Test that flags work correctly when invoked via alias
	var flagValue string

	cmd := &Command{
		Name: "app",
		Commands: []*Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Flags: []Flag{
					&StringFlag{
						Name:     "format",
						Aliases:  []string{"f"},
						AssignTo: &flagValue,
					},
				},
				Run: func(ctx context.Context, cmd *Command) error {
					return nil
				},
			},
		},
	}

	os.Args = []string{"app", "ls", "--format", "json"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if flagValue != "json" {
		t.Fatalf("expected flag value 'json', got %q", flagValue)
	}
}

func TestCommand_AliasSubcommand(t *testing.T) {
	// Test that aliases work for nested subcommands
	executed := false

	cmd := &Command{
		Name: "app",
		Commands: []*Command{
			{
				Name:    "model",
				Aliases: []string{"m"},
				Commands: []*Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Run: func(ctx context.Context, cmd *Command) error {
							executed = true
							return nil
						},
					},
				},
			},
		},
	}

	os.Args = []string{"app", "m", "ls"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected nested command to execute via aliases")
	}
}
