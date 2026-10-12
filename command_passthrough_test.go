package cli

import (
	"context"
	"os"
	"reflect"
	"testing"
)

func TestPassThroughAfterArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		timeout int
		space   string
		command string
		rest    []string
	}{
		{"flags before args", []string{"app", "run", "-t", "60", "box", "claude", "-p", "fix it", "--yes"}, 60, "box", "claude", []string{"-p", "fix it", "--yes"}},
		{"flag between args", []string{"app", "run", "box", "-t", "5", "ls", "-la"}, 5, "box", "ls", []string{"-la"}},
		{"own flag after command passes through", []string{"app", "run", "box", "ls", "-t", "9"}, 30, "box", "ls", []string{"-t", "9"}},
		{"double dash passes through", []string{"app", "run", "box", "sh", "--", "-c"}, 30, "box", "sh", []string{"--", "-c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var timeout int
			var space, command string
			var rest []string
			app := &Command{
				Name: "app",
				Commands: []*Command{{
					Name:        "run",
					PassThrough: true,
					Flags:       []Flag{&IntFlag{Name: "timeout", Aliases: []string{"t"}, DefaultValue: 30}},
					Arguments: []Argument{
						&StringArg{Name: "space", Required: true},
						&StringArg{Name: "command", Required: true},
					},
					MaxArgs: UnlimitedArgs,
					Run: func(ctx context.Context, cmd *Command) error {
						timeout = cmd.GetInt("timeout")
						space = cmd.GetStringArg("space")
						command = cmd.GetStringArg("command")
						rest = cmd.GetArgs()
						return nil
					},
				}},
			}
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = tt.args
			if err := app.Execute(context.Background()); err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if timeout != tt.timeout || space != tt.space || command != tt.command || !reflect.DeepEqual(rest, tt.rest) {
				t.Fatalf("got timeout=%d space=%q command=%q rest=%q", timeout, space, command, rest)
			}
		})
	}
}

// Without PassThrough an unknown flag after the arguments is still an error.
func TestNoPassThroughRejectsUnknownFlag(t *testing.T) {
	app := &Command{
		Name: "app",
		Commands: []*Command{{
			Name:      "run",
			Arguments: []Argument{&StringArg{Name: "space", Required: true}},
			MaxArgs:   UnlimitedArgs,
			Run:       func(ctx context.Context, cmd *Command) error { return nil },
		}},
	}
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"app", "run", "box", "-p"}
	if err := app.Execute(context.Background()); err == nil {
		t.Fatal("expected an unknown flag error")
	}
}

// A command's own flag wins over an inherited global flag with the same alias.
func TestOwnFlagAliasBeatsGlobal(t *testing.T) {
	var timeout int
	var token string
	app := &Command{
		Name:  "app",
		Flags: []Flag{&StringFlag{Name: "token", Aliases: []string{"t"}, Global: true}},
		Commands: []*Command{{
			Name:  "run",
			Flags: []Flag{&IntFlag{Name: "timeout", Aliases: []string{"t"}, DefaultValue: 30}},
			Run: func(ctx context.Context, cmd *Command) error {
				timeout = cmd.GetInt("timeout")
				token = cmd.GetString("token")
				return nil
			},
		}},
	}
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"app", "run", "-t", "60", "--token", "abc"}
	if err := app.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if timeout != 60 || token != "abc" {
		t.Fatalf("got timeout=%d token=%q, want 60 and abc", timeout, token)
	}
}
