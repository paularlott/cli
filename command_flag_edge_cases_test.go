package cli

import (
	"context"
	"os"
	"testing"
)

// TestFlagPositioning_GlobalFlagWithDoubleDash tests global flag followed by -- and arguments
func TestFlagPositioning_GlobalFlagWithDoubleDash(t *testing.T) {
	var globalValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "value" {
				t.Errorf("expected global to be 'value', got '%s'", globalValue)
			}
			args := cmd.GetArgs()
			expectedArgs := []string{"a", "--something"}
			if len(args) != len(expectedArgs) {
				t.Errorf("expected %d args, got %d: %v", len(expectedArgs), len(args), args)
			}
			for i, expected := range expectedArgs {
				if i < len(args) && args[i] != expected {
					t.Errorf("expected arg[%d] to be '%s', got '%s'", i, expected, args[i])
				}
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --global value sub -- a --something
	os.Args = []string{"cmd", "--global", "value", "sub", "--", "a", "--something"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_LocalFlagBeforeSubcommand tests that local flags are moved after subcommands
func TestFlagPositioning_LocalFlagBeforeSubcommand(t *testing.T) {
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if localValue != "value" {
				t.Errorf("expected local to be 'value', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --local value sub a c
	// Even though --local appears before sub, the preprocessor moves it to after sub
	// This is part of the flexible flag positioning feature
	os.Args = []string{"cmd", "--local", "value", "sub", "a", "c"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_GlobalFlagBeforeSubcommand tests global flag before its subcommand
func TestFlagPositioning_GlobalFlagBeforeSubcommand(t *testing.T) {
	var globalValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "value" {
				t.Errorf("expected global to be 'value', got '%s'", globalValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --global value sub a c
	os.Args = []string{"cmd", "--global", "value", "sub", "a", "c"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_GlobalFlagAfterSubcommand tests global flag after its subcommand
func TestFlagPositioning_GlobalFlagAfterSubcommand(t *testing.T) {
	var globalValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "value" {
				t.Errorf("expected global to be 'value', got '%s'", globalValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd sub --global value a c
	os.Args = []string{"cmd", "sub", "--global", "value", "a", "c"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_LocalFlagAfterSubcommand tests local flag after its subcommand
func TestFlagPositioning_LocalFlagAfterSubcommand(t *testing.T) {
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if localValue != "value" {
				t.Errorf("expected local to be 'value', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd sub --local value a c
	os.Args = []string{"cmd", "sub", "--local", "value", "a", "c"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_LocalFlagAtEnd tests local flag at the end
func TestFlagPositioning_LocalFlagAtEnd(t *testing.T) {
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if localValue != "value" {
				t.Errorf("expected local to be 'value', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd sub a c --local value
	os.Args = []string{"cmd", "sub", "a", "c", "--local", "value"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_MixedGlobalAndLocal tests mixing global and local flags
func TestFlagPositioning_MixedGlobalAndLocal(t *testing.T) {
	var globalValue string
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "gval" {
				t.Errorf("expected global to be 'gval', got '%s'", globalValue)
			}
			if localValue != "lval" {
				t.Errorf("expected local to be 'lval', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --global gval sub --local lval a c
	os.Args = []string{"cmd", "--global", "gval", "sub", "--local", "lval", "a", "c"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_WithNamedArguments tests flags with named arguments
func TestFlagPositioning_WithNamedArguments(t *testing.T) {
	var globalValue string
	var localValue string
	var portValue int
	var nameValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Arguments: []Argument{
			&ArgumentTyped[int]{
				Name:     "port",
				Required: true,
				AssignTo: &portValue,
			},
			&ArgumentTyped[string]{
				Name:     "name",
				Required: false,
				AssignTo: &nameValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "gval" {
				t.Errorf("expected global to be 'gval', got '%s'", globalValue)
			}
			if localValue != "lval" {
				t.Errorf("expected local to be 'lval', got '%s'", localValue)
			}
			if portValue != 8080 {
				t.Errorf("expected port to be 8080, got %d", portValue)
			}
			if nameValue != "myapp" {
				t.Errorf("expected name to be 'myapp', got '%s'", nameValue)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --global gval sub --local lval 8080 myapp
	os.Args = []string{"cmd", "--global", "gval", "sub", "--local", "lval", "8080", "myapp"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_GlobalBeforeLocalBeforeArgs tests all flags before positional args
func TestFlagPositioning_GlobalBeforeLocalBeforeArgs(t *testing.T) {
	var globalValue string
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "gval" {
				t.Errorf("expected global to be 'gval', got '%s'", globalValue)
			}
			if localValue != "lval" {
				t.Errorf("expected local to be 'lval', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			expectedArgs := []string{"a", "b", "c"}
			if len(args) != len(expectedArgs) {
				t.Errorf("expected %d args, got %d: %v", len(expectedArgs), len(args), args)
			}
			for i, expected := range expectedArgs {
				if i < len(args) && args[i] != expected {
					t.Errorf("expected arg[%d] to be '%s', got '%s'", i, expected, args[i])
				}
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --global gval sub --local lval a b c
	os.Args = []string{"cmd", "--global", "gval", "sub", "--local", "lval", "a", "b", "c"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_FlagsInterspersedWithArgs tests flags interspersed with positional args
func TestFlagPositioning_FlagsInterspersedWithArgs(t *testing.T) {
	var globalValue string
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "gval" {
				t.Errorf("expected global to be 'gval', got '%s'", globalValue)
			}
			if localValue != "lval" {
				t.Errorf("expected local to be 'lval', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			expectedArgs := []string{"a", "b", "c"}
			if len(args) != len(expectedArgs) {
				t.Errorf("expected %d args, got %d: %v", len(expectedArgs), len(args), args)
			}
			for i, expected := range expectedArgs {
				if i < len(args) && args[i] != expected {
					t.Errorf("expected arg[%d] to be '%s', got '%s'", i, expected, args[i])
				}
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --global gval sub a --local lval b c
	os.Args = []string{"cmd", "--global", "gval", "sub", "a", "--local", "lval", "b", "c"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_AllFlagsAtEndWithArgs tests all flags at the end after positional args
func TestFlagPositioning_AllFlagsAtEndWithArgs(t *testing.T) {
	var globalValue string
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "gval" {
				t.Errorf("expected global to be 'gval', got '%s'", globalValue)
			}
			if localValue != "lval" {
				t.Errorf("expected local to be 'lval', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			expectedArgs := []string{"a", "b", "c"}
			if len(args) != len(expectedArgs) {
				t.Errorf("expected %d args, got %d: %v", len(expectedArgs), len(args), args)
			}
			for i, expected := range expectedArgs {
				if i < len(args) && args[i] != expected {
					t.Errorf("expected arg[%d] to be '%s', got '%s'", i, expected, args[i])
				}
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd sub a b c --global gval --local lval
	os.Args = []string{"cmd", "sub", "a", "b", "c", "--global", "gval", "--local", "lval"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_DoubleDashWithFlags tests that -- prevents flag parsing
func TestFlagPositioning_DoubleDashWithFlags(t *testing.T) {
	var globalValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "gval" {
				t.Errorf("expected global to be 'gval', got '%s'", globalValue)
			}
			args := cmd.GetArgs()
			// Everything after -- should be treated as positional args (but -- itself is not included)
			expectedArgs := []string{"a", "--global", "other", "b"}
			if len(args) != len(expectedArgs) {
				t.Errorf("expected %d args, got %d: %v", len(expectedArgs), len(args), args)
			}
			for i, expected := range expectedArgs {
				if i < len(args) && args[i] != expected {
					t.Errorf("expected arg[%d] to be '%s', got '%s'", i, expected, args[i])
				}
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --global gval sub -- a --global other b
	os.Args = []string{"cmd", "--global", "gval", "sub", "--", "a", "--global", "other", "b"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_MultipleSubcommands tests flags with multiple levels of subcommands
func TestFlagPositioning_MultipleSubcommands(t *testing.T) {
	var global1Value string
	var global2Value string
	var localValue string
	var executed bool

	rootCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global1",
				Global:   true,
				AssignTo: &global1Value,
			},
		},
	}

	level1Cmd := &Command{
		Name: "level1",
		Flags: []Flag{
			&StringFlag{
				Name:     "global2",
				Global:   true,
				AssignTo: &global2Value,
			},
		},
	}

	level2Cmd := &Command{
		Name:    "level2",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if global1Value != "g1" {
				t.Errorf("expected global1 to be 'g1', got '%s'", global1Value)
			}
			if global2Value != "g2" {
				t.Errorf("expected global2 to be 'g2', got '%s'", global2Value)
			}
			if localValue != "l" {
				t.Errorf("expected local to be 'l', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	level1Cmd.Commands = []*Command{level2Cmd}
	rootCmd.Commands = []*Command{level1Cmd}

	// Test: cmd --global1 g1 level1 --global2 g2 level2 --local l a b
	os.Args = []string{"cmd", "--global1", "g1", "level1", "--global2", "g2", "level2", "--local", "l", "a", "b"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_ShortFlagsCombinations tests various short flag combinations
func TestFlagPositioning_ShortFlagsCombinations(t *testing.T) {
	var globalValue string
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Aliases:  []string{"g"},
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				Aliases:  []string{"l"},
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "gval" {
				t.Errorf("expected global to be 'gval', got '%s'", globalValue)
			}
			if localValue != "lval" {
				t.Errorf("expected local to be 'lval', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd -g gval sub -l lval a b
	os.Args = []string{"cmd", "-g", "gval", "sub", "-l", "lval", "a", "b"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_EmptyArguments tests handling of commands with no positional arguments
func TestFlagPositioning_EmptyArguments(t *testing.T) {
	var globalValue string
	var localValue string
	var executed bool

	parentCmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{
				Name:     "global",
				Global:   true,
				AssignTo: &globalValue,
			},
		},
	}

	subCmd := &Command{
		Name:    "sub",
		MaxArgs: 0, // No arguments allowed
		Flags: []Flag{
			&StringFlag{
				Name:     "local",
				AssignTo: &localValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if globalValue != "gval" {
				t.Errorf("expected global to be 'gval', got '%s'", globalValue)
			}
			if localValue != "lval" {
				t.Errorf("expected local to be 'lval', got '%s'", localValue)
			}
			args := cmd.GetArgs()
			if len(args) != 0 {
				t.Errorf("expected 0 args, got %d: %v", len(args), args)
			}
			return nil
		},
	}

	parentCmd.Commands = []*Command{subCmd}

	// Test: cmd --global gval sub --local lval
	os.Args = []string{"cmd", "--global", "gval", "sub", "--local", "lval"}
	err := parentCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}
