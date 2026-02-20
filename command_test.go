package cli

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestCommand_Execute_BasicCommand(t *testing.T) {
	executed := false
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Usage:   "Test command",
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			return nil
		},
	}

	os.Args = []string{"test"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

func TestCommand_Execute_Help(t *testing.T) {
	executed := false
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Usage:   "Test command",
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			return nil
		},
	}

	os.Args = []string{"test", "--help"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if executed {
		t.Fatal("expected command not to be executed when help is shown")
	}
}

func TestCommand_Execute_Version(t *testing.T) {
	executed := false
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Usage:   "Test command",
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			return nil
		},
	}

	os.Args = []string{"test", "--version"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if executed {
		t.Fatal("expected command not to be executed when version is shown")
	}
}

func TestCommand_Execute_RequiredFlagMissing(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Usage:   "Test command",
		Flags: []Flag{
			&StringFlag{
				Name:     "required",
				Required: true,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	os.Args = []string{"test"}
	err := cmd.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error for missing required flag")
	}
}

func TestCommand_Execute_RequiredFlagWithHelp(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Usage:   "Test command",
		Flags: []Flag{
			&StringFlag{
				Name:     "required",
				Required: true,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	os.Args = []string{"test", "--help"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error when showing help with required flag, got %v", err)
	}
}

func TestCommand_Execute_RequiredFlagWithVersion(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Usage:   "Test command",
		Flags: []Flag{
			&StringFlag{
				Name:     "required",
				Required: true,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	os.Args = []string{"test", "--version"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error when showing version with required flag, got %v", err)
	}
}

func TestCommand_Execute_Subcommands(t *testing.T) {
	parentExecuted := false
	childExecuted := false

	cmd := &Command{
		Name:    "parent",
		Version: "1.0.0",
		Usage:   "Parent command",
		Run: func(ctx context.Context, cmd *Command) error {
			parentExecuted = true
			return nil
		},
		Commands: []*Command{
			{
				Name:  "child",
				Usage: "Child command",
				Run: func(ctx context.Context, cmd *Command) error {
					childExecuted = true
					return nil
				},
			},
		},
	}

	os.Args = []string{"parent", "child"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parentExecuted {
		t.Fatal("expected parent command not to be executed")
	}
	if !childExecuted {
		t.Fatal("expected child command to be executed")
	}
}

func TestCommand_Execute_SubcommandWithHelp(t *testing.T) {
	childExecuted := false

	cmd := &Command{
		Name:    "parent",
		Version: "1.0.0",
		Usage:   "Parent command",
		Commands: []*Command{
			{
				Name:  "child",
				Usage: "Child command",
				Flags: []Flag{
					&StringFlag{
						Name:     "required",
						Required: true,
					},
				},
				Run: func(ctx context.Context, cmd *Command) error {
					childExecuted = true
					return nil
				},
			},
		},
	}

	os.Args = []string{"parent", "child", "--help"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error when showing help for subcommand with required flag, got %v", err)
	}
	if childExecuted {
		t.Fatal("expected child command not to be executed when help is shown")
	}
}

func TestCommand_Execute_PreRun(t *testing.T) {
	preRunCalled := false
	runCalled := false

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		PreRun: func(ctx context.Context, cmd *Command) (context.Context, error) {
			preRunCalled = true
			return ctx, nil
		},
		Run: func(ctx context.Context, cmd *Command) error {
			runCalled = true
			return nil
		},
	}

	os.Args = []string{"test"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !preRunCalled {
		t.Fatal("expected PreRun to be called")
	}
	if !runCalled {
		t.Fatal("expected Run to be called")
	}
}

func TestCommand_Execute_PostRun(t *testing.T) {
	runCalled := false
	postRunCalled := false

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Run: func(ctx context.Context, cmd *Command) error {
			runCalled = true
			return nil
		},
		PostRun: func(ctx context.Context, cmd *Command) error {
			postRunCalled = true
			return nil
		},
	}

	os.Args = []string{"test"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !runCalled {
		t.Fatal("expected Run to be called")
	}
	if !postRunCalled {
		t.Fatal("expected PostRun to be called")
	}
}

func TestCommand_Execute_MinMaxArgs(t *testing.T) {
	tests := []struct {
		name    string
		minArgs int
		maxArgs int
		args    []string
		wantErr bool
	}{
		{
			name:    "exact args",
			minArgs: 2,
			maxArgs: 2,
			args:    []string{"test", "arg1", "arg2"},
			wantErr: false,
		},
		{
			name:    "too few args",
			minArgs: 2,
			maxArgs: 3,
			args:    []string{"test", "arg1"},
			wantErr: true,
		},
		{
			name:    "too many args",
			minArgs: 1,
			maxArgs: 2,
			args:    []string{"test", "arg1", "arg2", "arg3"},
			wantErr: true,
		},
		{
			name:    "unlimited args",
			minArgs: 0,
			maxArgs: UnlimitedArgs,
			args:    []string{"test", "arg1", "arg2", "arg3", "arg4"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:    "test",
				Version: "1.0.0",
				MinArgs: tt.minArgs,
				MaxArgs: tt.maxArgs,
				Run: func(ctx context.Context, cmd *Command) error {
					return nil
				},
			}

			os.Args = tt.args
			err := cmd.Execute(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCommand_HasFlag(t *testing.T) {
	var flagValue string

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&StringFlag{
				Name:     "flag",
				AssignTo: &flagValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if !cmd.HasFlag("flag") {
				t.Fatal("expected HasFlag to return true for provided flag")
			}
			return nil
		},
	}

	os.Args = []string{"test", "--flag", "value"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCommand_GetArgs(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		MinArgs: 0,
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Fatalf("expected 2 args, got %d", len(args))
			}
			if args[0] != "arg1" || args[1] != "arg2" {
				t.Fatalf("unexpected args: %v", args)
			}
			return nil
		},
	}

	os.Args = []string{"test", "arg1", "arg2"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCommand_GlobalFlags(t *testing.T) {
	var globalFlag string

	parent := &Command{
		Name:    "parent",
		Version: "1.0.0",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalFlag,
			},
		},
		Commands: []*Command{
			{
				Name: "child",
				Run: func(ctx context.Context, cmd *Command) error {
					if globalFlag != "value" {
						t.Fatalf("expected global flag to be set, got %s", globalFlag)
					}
					return nil
				},
			},
		},
	}

	os.Args = []string{"parent", "child", "--global", "value"}
	err := parent.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCommand_Execute_UnknownSubcommand(t *testing.T) {
	tests := []struct {
		name       string
		cmd        *Command
		args       []string
		wantErr    bool
		errContains string
	}{
		{
			name: "unknown subcommand returns error",
			cmd: &Command{
				Name: "parent",
				Commands: []*Command{
					{Name: "child1", Run: func(ctx context.Context, cmd *Command) error { return nil }},
					{Name: "child2", Run: func(ctx context.Context, cmd *Command) error { return nil }},
				},
			},
			args:       []string{"parent", "unknown"},
			wantErr:    true,
			errContains: "unknown command",
		},
		{
			name: "valid subcommand with MaxArgs=0 works",
			cmd: &Command{
				Name: "parent",
				Commands: []*Command{
					{
						Name:    "child",
						MaxArgs: 0,
						Run:     func(ctx context.Context, cmd *Command) error { return nil },
					},
				},
			},
			args:    []string{"parent", "child"},
			wantErr: false,
		},
		{
			name: "command with subcommands but no args shows help",
			cmd: &Command{
				Name: "parent",
				Commands: []*Command{
					{Name: "child", Run: func(ctx context.Context, cmd *Command) error { return nil }},
				},
			},
			args:       []string{"parent"},
			wantErr:    false, // Shows help instead of error
		},
		{
			name: "command without subcommands validates MaxArgs",
			cmd: &Command{
				Name:     "parent",
				MaxArgs:  0,
				Run:      func(ctx context.Context, cmd *Command) error { return nil },
			},
			args:       []string{"parent", "extra"},
			wantErr:    true,
			errContains: "too many arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			err := tt.cmd.Execute(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			}
		})
	}
}

func TestCommand_Execute_MinMaxArgs_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		cmd        *Command
		args       []string
		wantErr    bool
		errContains string
	}{
		{
			name: "MinArgs=1 no args given",
			cmd: &Command{
				Name:     "test",
				MinArgs:  1,
				Run:      func(ctx context.Context, cmd *Command) error { return nil },
			},
			args:       []string{"test"},
			wantErr:    true,
			errContains: "too few arguments",
		},
		{
			name: "MinArgs=2 MaxArgs=3 with exactly 2 args",
			cmd: &Command{
				Name:     "test",
				MinArgs:  2,
				MaxArgs:  3,
				Run:      func(ctx context.Context, cmd *Command) error { return nil },
			},
			args:    []string{"test", "arg1", "arg2"},
			wantErr: false,
		},
		{
			name: "MinArgs=2 MaxArgs=3 with exactly 3 args",
			cmd: &Command{
				Name:     "test",
				MinArgs:  2,
				MaxArgs:  3,
				Run:      func(ctx context.Context, cmd *Command) error { return nil },
			},
			args:    []string{"test", "arg1", "arg2", "arg3"},
			wantErr: false,
		},
		{
			name: "MinArgs=1 with 0 args but has subcommands",
			cmd: &Command{
				Name: "parent",
				MinArgs: 1,
				Commands: []*Command{
					{Name: "child", Run: func(ctx context.Context, cmd *Command) error { return nil }},
				},
			},
			args:    []string{"parent"},
			wantErr: false, // Shows help instead of error
		},
		{
			name: "MinArgs=1 with unknown subcommand",
			cmd: &Command{
				Name: "parent",
				MinArgs: 1,
				Commands: []*Command{
					{Name: "child", Run: func(ctx context.Context, cmd *Command) error { return nil }},
				},
			},
			args:       []string{"parent", "unknown"},
			wantErr:    true,
			errContains: "unknown command",
		},
		{
			name: "MinArgs=0 MaxArgs=0 with no args",
			cmd: &Command{
				Name:     "test",
				MinArgs:  0,
				MaxArgs:  0,
				Run:      func(ctx context.Context, cmd *Command) error { return nil },
			},
			args:    []string{"test"},
			wantErr: false,
		},
		{
			name: "MinArgs=0 MaxArgs=1 with 1 arg",
			cmd: &Command{
				Name:     "test",
				MinArgs:  0,
				MaxArgs:  1,
				Run:      func(ctx context.Context, cmd *Command) error { return nil },
			},
			args:    []string{"test", "arg1"},
			wantErr: false,
		},
		{
			name: "MinArgs=0 MaxArgs=1 with 2 args",
			cmd: &Command{
				Name:     "test",
				MinArgs:  0,
				MaxArgs:  1,
				Run:      func(ctx context.Context, cmd *Command) error { return nil },
			},
			args:       []string{"test", "arg1", "arg2"},
			wantErr:    true,
			errContains: "too many arguments",
		},
		{
			name: "command with subcommands and valid subcommand call",
			cmd: &Command{
				Name: "parent",
				Commands: []*Command{
					{
						Name:     "child",
						MinArgs:  1,
						MaxArgs:  UnlimitedArgs,
						Run:      func(ctx context.Context, cmd *Command) error { return nil },
					},
				},
			},
			args:    []string{"parent", "child", "arg1"},
			wantErr: false,
		},
		{
			name: "command with subcommands and subcommand with too few args",
			cmd: &Command{
				Name: "parent",
				Commands: []*Command{
					{
						Name:    "child",
						MinArgs: 1,
						Run:     func(ctx context.Context, cmd *Command) error { return nil },
					},
				},
			},
			args:       []string{"parent", "child"},
			wantErr:    true,
			errContains: "too few arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			err := tt.cmd.Execute(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			}
		})
	}
}
