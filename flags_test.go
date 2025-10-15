package cli

import (
	"context"
	"os"
	"testing"
)

func TestStringFlag(t *testing.T) {
	var value string

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&StringFlag{
				Name:         "string",
				Aliases:      []string{"s"},
				DefaultValue: "default",
				AssignTo:     &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if value != "test-value" {
				t.Fatalf("expected value to be 'test-value', got %s", value)
			}
			return nil
		},
	}

	os.Args = []string{"test", "--string", "test-value"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStringFlagAlias(t *testing.T) {
	var value string

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&StringFlag{
				Name:     "string",
				Aliases:  []string{"s"},
				AssignTo: &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if value != "alias-value" {
				t.Fatalf("expected value to be 'alias-value', got %s", value)
			}
			return nil
		},
	}

	os.Args = []string{"test", "-s", "alias-value"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestIntFlag(t *testing.T) {
	var value int

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&IntFlag{
				Name:         "number",
				DefaultValue: 0,
				AssignTo:     &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if value != 42 {
				t.Fatalf("expected value to be 42, got %d", value)
			}
			return nil
		},
	}

	os.Args = []string{"test", "--number", "42"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBoolFlag(t *testing.T) {
	var value bool

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&BoolFlag{
				Name:     "bool",
				AssignTo: &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if !value {
				t.Fatal("expected value to be true")
			}
			return nil
		},
	}

	os.Args = []string{"test", "--bool"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBoolFlagWithValue(t *testing.T) {
	var value bool

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&BoolFlag{
				Name:     "bool",
				AssignTo: &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if value {
				t.Fatal("expected value to be false")
			}
			return nil
		},
	}

	os.Args = []string{"test", "--bool=false"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFloat64Flag(t *testing.T) {
	var value float64

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&Float64Flag{
				Name:     "float",
				AssignTo: &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if value != 3.14 {
				t.Fatalf("expected value to be 3.14, got %f", value)
			}
			return nil
		},
	}

	os.Args = []string{"test", "--float", "3.14"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStringSliceFlag(t *testing.T) {
	var values []string

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&StringSliceFlag{
				Name:     "items",
				AssignTo: &values,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if len(values) != 3 {
				t.Fatalf("expected 3 values, got %d", len(values))
			}
			expected := []string{"one", "two", "three"}
			for i, v := range values {
				if v != expected[i] {
					t.Fatalf("expected values[%d] to be %s, got %s", i, expected[i], v)
				}
			}
			return nil
		},
	}

	os.Args = []string{"test", "--items", "one", "--items", "two", "--items", "three"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestIntSliceFlag(t *testing.T) {
	var values []int

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&IntSliceFlag{
				Name:     "numbers",
				AssignTo: &values,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if len(values) != 3 {
				t.Fatalf("expected 3 values, got %d", len(values))
			}
			expected := []int{1, 2, 3}
			for i, v := range values {
				if v != expected[i] {
					t.Fatalf("expected values[%d] to be %d, got %d", i, expected[i], v)
				}
			}
			return nil
		},
	}

	os.Args = []string{"test", "--numbers", "1", "--numbers", "2", "--numbers", "3"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFlagDefaultValue(t *testing.T) {
	var value string

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&StringFlag{
				Name:         "flag",
				DefaultValue: "default-value",
				AssignTo:     &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if value != "default-value" {
				t.Fatalf("expected default value to be used, got %s", value)
			}
			return nil
		},
	}

	os.Args = []string{"test"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFlagValidation(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&IntFlag{
				Name: "port",
				ValidateFlag: func(c *Command) error {
					port := c.GetInt("port")
					if port < 1 || port > 65535 {
						t.Fatal("validation should have been called")
					}
					return nil
				},
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	os.Args = []string{"test", "--port", "8080"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFlagEnvironmentVariable(t *testing.T) {
	var value string

	// Set environment variable
	os.Setenv("TEST_FLAG", "env-value")
	defer os.Unsetenv("TEST_FLAG")

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&StringFlag{
				Name:     "flag",
				EnvVars:  []string{"TEST_FLAG"},
				AssignTo: &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if value != "env-value" {
				t.Fatalf("expected value from env var, got %s", value)
			}
			return nil
		},
	}

	os.Args = []string{"test"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFlagEnvironmentVariableOverride(t *testing.T) {
	var value string

	// Set environment variable
	os.Setenv("TEST_FLAG", "env-value")
	defer os.Unsetenv("TEST_FLAG")

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&StringFlag{
				Name:     "flag",
				EnvVars:  []string{"TEST_FLAG"},
				AssignTo: &value,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if value != "cli-value" {
				t.Fatalf("expected CLI value to override env var, got %s", value)
			}
			return nil
		},
	}

	os.Args = []string{"test", "--flag", "cli-value"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUnknownFlag(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Run: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	os.Args = []string{"test", "--unknown"}
	err := cmd.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestBundledShortFlags(t *testing.T) {
	var verbose bool
	var all bool
	var long bool

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&BoolFlag{
				Name:     "verbose",
				Aliases:  []string{"v"},
				AssignTo: &verbose,
			},
			&BoolFlag{
				Name:     "all",
				Aliases:  []string{"a"},
				AssignTo: &all,
			},
			&BoolFlag{
				Name:     "long",
				Aliases:  []string{"l"},
				AssignTo: &long,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if !verbose || !all || !long {
				t.Fatal("expected all bundled flags to be true")
			}
			return nil
		},
	}

	os.Args = []string{"test", "-val"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
