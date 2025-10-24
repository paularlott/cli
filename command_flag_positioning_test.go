package cli

import (
	"context"
	"os"
	"testing"
)

// TestFlagPositioning_PersistentFlagAfterParent tests that persistent flags can be placed after their parent command
func TestFlagPositioning_PersistentFlagAfterParent(t *testing.T) {
	var secretValue string
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "123" {
				t.Errorf("expected secret to be '123', got '%s'", secretValue)
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin --secret=123 filesystem create test
	os.Args = []string{"granitefs", "admin", "--secret=123", "filesystem", "create", "test"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_MultiplePersistentFlags tests multiple persistent flags after different commands
func TestFlagPositioning_MultiplePersistentFlags(t *testing.T) {
	var secretValue string
	var replicationValue int
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
		Flags: []Flag{
			&IntFlag{
				Name:     "replication",
				Global:   true,
				AssignTo: &replicationValue,
			},
		},
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "123" {
				t.Errorf("expected secret to be '123', got '%s'", secretValue)
			}
			if replicationValue != 3 {
				t.Errorf("expected replication to be 3, got %d", replicationValue)
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin --secret=123 filesystem --replication=3 create test
	os.Args = []string{"granitefs", "admin", "--secret=123", "filesystem", "--replication=3", "create", "test"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_AllFlagsAtEnd tests backward compatibility - all flags at the end
func TestFlagPositioning_AllFlagsAtEnd(t *testing.T) {
	var secretValue string
	var forceValue bool
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&BoolFlag{
				Name:     "force",
				AssignTo: &forceValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "123" {
				t.Errorf("expected secret to be '123', got '%s'", secretValue)
			}
			if !forceValue {
				t.Errorf("expected force to be true")
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin filesystem create test --secret=123 --force
	os.Args = []string{"granitefs", "admin", "filesystem", "create", "test", "--secret=123", "--force"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_MixedPositioning tests the most readable form with mixed positioning
func TestFlagPositioning_MixedPositioning(t *testing.T) {
	var secretValue string
	var replicationValue int
	var forceValue bool
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
		Flags: []Flag{
			&IntFlag{
				Name:     "replication",
				Global:   true,
				AssignTo: &replicationValue,
			},
		},
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&BoolFlag{
				Name:     "force",
				AssignTo: &forceValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "123" {
				t.Errorf("expected secret to be '123', got '%s'", secretValue)
			}
			if replicationValue != 3 {
				t.Errorf("expected replication to be 3, got %d", replicationValue)
			}
			if !forceValue {
				t.Errorf("expected force to be true")
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin --secret=123 filesystem --replication=3 create test --force
	os.Args = []string{"granitefs", "admin", "--secret=123", "filesystem", "--replication=3", "create", "test", "--force"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_Precedence tests that later flag values override earlier ones
func TestFlagPositioning_Precedence(t *testing.T) {
	var secretValue string
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "new" {
				t.Errorf("expected secret to be 'new' (last value), got '%s'", secretValue)
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin --secret=old filesystem --secret=new create test
	os.Args = []string{"granitefs", "admin", "--secret=old", "filesystem", "--secret=new", "create", "test"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_FlagTerminator tests that -- marks the end of flags
func TestFlagPositioning_FlagTerminator(t *testing.T) {
	var secretValue string
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "123" {
				t.Errorf("expected secret to be '123', got '%s'", secretValue)
			}
			args := cmd.GetArgs()
			if len(args) != 2 {
				t.Errorf("expected 2 args, got %d", len(args))
			}
			if len(args) > 0 && args[0] != "test" {
				t.Errorf("expected first arg to be 'test', got '%s'", args[0])
			}
			if len(args) > 1 && args[1] != "--not-a-flag" {
				t.Errorf("expected second arg to be '--not-a-flag', got '%s'", args[1])
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin --secret=123 filesystem create -- test --not-a-flag
	os.Args = []string{"granitefs", "admin", "--secret=123", "filesystem", "create", "--", "test", "--not-a-flag"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_ShortFlagForm tests short flag forms with positioning
func TestFlagPositioning_ShortFlagFormAfterParent(t *testing.T) {
	var secretValue string
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Aliases:  []string{"s"},
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "123" {
				t.Errorf("expected secret to be '123', got '%s'", secretValue)
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin -s 123 filesystem create test
	os.Args = []string{"granitefs", "admin", "-s", "123", "filesystem", "create", "test"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_WithArguments tests that positional arguments work correctly
func TestFlagPositioning_WithArguments(t *testing.T) {
	var secretValue string
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
	}

	createCmd := &Command{
		Name:    "create",
		MinArgs: 1,
		MaxArgs: 1,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "123" {
				t.Errorf("expected secret to be '123', got '%s'", secretValue)
			}
			args := cmd.GetArgs()
			if len(args) != 1 {
				t.Errorf("expected 1 arg, got %d", len(args))
			}
			if len(args) > 0 && args[0] != "myfs" {
				t.Errorf("expected arg to be 'myfs', got '%s'", args[0])
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin --secret=123 filesystem create myfs
	os.Args = []string{"granitefs", "admin", "--secret=123", "filesystem", "create", "myfs"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_LocalFlagOnly tests local flag positioned correctly
func TestFlagPositioning_LocalFlagOnly(t *testing.T) {
	var forceValue bool
	var executed bool

	adminCmd := &Command{
		Name: "admin",
	}

	filesystemCmd := &Command{
		Name: "filesystem",
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Flags: []Flag{
			&BoolFlag{
				Name:     "force",
				AssignTo: &forceValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if !forceValue {
				t.Errorf("expected force to be true")
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin filesystem create --force test
	os.Args = []string{"granitefs", "admin", "filesystem", "create", "--force", "test"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}

// TestFlagPositioning_EmptyStringValue tests flags with empty string values
func TestFlagPositioning_EmptyStringValue(t *testing.T) {
	var secretValue string
	var executed bool

	adminCmd := &Command{
		Name: "admin",
		Flags: []Flag{
			&StringFlag{
				Name:     "secret",
				Global:   true,
				AssignTo: &secretValue,
			},
		},
	}

	filesystemCmd := &Command{
		Name: "filesystem",
	}

	createCmd := &Command{
		Name:    "create",
		MaxArgs: UnlimitedArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			executed = true
			if secretValue != "" {
				t.Errorf("expected secret to be empty, got '%s'", secretValue)
			}
			return nil
		},
	}

	filesystemCmd.Commands = []*Command{createCmd}
	adminCmd.Commands = []*Command{filesystemCmd}

	rootCmd := &Command{
		Name:     "granitefs",
		Commands: []*Command{adminCmd},
	}

	// Test: granitefs admin --secret= filesystem create test
	os.Args = []string{"granitefs", "admin", "--secret=", "filesystem", "create", "test"}
	err := rootCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !executed {
		t.Fatal("expected command to be executed")
	}
}
