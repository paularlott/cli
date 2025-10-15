package cli

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestFlagValidationWithHelp(t *testing.T) {
	var size int64

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&Int64Flag{
				Name:     "superblock-size",
				Aliases:  []string{"s"},
				Usage:    "Maximum superblock size in megabytes",
				AssignTo: &size,
				ValidateFlag: func(c *Command) error {
					if c.GetInt64("superblock-size") < 4 {
						return fmt.Errorf("superblock size should be at least 4MB")
					}
					return nil
				},
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	// Test that --help doesn't trigger validation
	os.Args = []string{"test", "--help"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error when showing help, got %v", err)
	}
}

func TestFlagValidationWithVersion(t *testing.T) {
	var size int64

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&Int64Flag{
				Name:     "superblock-size",
				Aliases:  []string{"s"},
				Usage:    "Maximum superblock size in megabytes",
				AssignTo: &size,
				ValidateFlag: func(c *Command) error {
					if c.GetInt64("superblock-size") < 4 {
						return fmt.Errorf("superblock size should be at least 4MB")
					}
					return nil
				},
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	// Test that --version doesn't trigger validation
	os.Args = []string{"test", "--version"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error when showing version, got %v", err)
	}
}

func TestFlagValidationStillWorksNormally(t *testing.T) {
	var size int64

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&Int64Flag{
				Name:     "superblock-size",
				Aliases:  []string{"s"},
				Usage:    "Maximum superblock size in megabytes",
				AssignTo: &size,
				ValidateFlag: func(c *Command) error {
					if c.GetInt64("superblock-size") < 4 {
						return fmt.Errorf("superblock size should be at least 4MB")
					}
					return nil
				},
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	// Test that validation still runs normally
	os.Args = []string{"test", "--superblock-size", "2"}
	err := cmd.Execute(context.Background())
	if err == nil {
		t.Fatal("expected validation error for invalid superblock size")
	}
	if err.Error() != "superblock size should be at least 4MB" {
		t.Fatalf("expected specific validation error, got %v", err)
	}
}

func TestFlagValidationPassesWithValidValue(t *testing.T) {
	var size int64

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&Int64Flag{
				Name:     "superblock-size",
				Aliases:  []string{"s"},
				Usage:    "Maximum superblock size in megabytes",
				AssignTo: &size,
				ValidateFlag: func(c *Command) error {
					if c.GetInt64("superblock-size") < 4 {
						return fmt.Errorf("superblock size should be at least 4MB")
					}
					return nil
				},
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if size != 8 {
				t.Fatalf("expected size to be 8, got %d", size)
			}
			return nil
		},
	}

	// Test that validation passes with valid value
	os.Args = []string{"test", "--superblock-size", "8"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error with valid value, got %v", err)
	}
}

func TestRequiredFlagWithValidation(t *testing.T) {
	var size int64

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&Int64Flag{
				Name:     "superblock-size",
				Aliases:  []string{"s"},
				Usage:    "Maximum superblock size in megabytes",
				Required: true,
				AssignTo: &size,
				ValidateFlag: func(c *Command) error {
					if c.GetInt64("superblock-size") < 4 {
						return fmt.Errorf("superblock size should be at least 4MB")
					}
					return nil
				},
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	// Test that --help works even with required flag that has validation
	os.Args = []string{"test", "--help"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error when showing help with required flag, got %v", err)
	}
}

func TestArgumentValidationWithHelp(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&IntArg{
				Name:     "port",
				Required: true,
				ValidateArg: func(c *Command) error {
					port := c.GetIntArg("port")
					if port < 1024 {
						return fmt.Errorf("port must be at least 1024")
					}
					return nil
				},
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	// Test that --help doesn't require arguments or trigger validation
	os.Args = []string{"test", "--help"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error when showing help with required argument, got %v", err)
	}
}
