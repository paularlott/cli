package cli

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestStringArgument(t *testing.T) {
	var argValue string

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&StringArg{
				Name:     "name",
				Required: true,
				AssignTo: &argValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if argValue != "test-name" {
				t.Fatalf("expected argument value to be 'test-name', got %s", argValue)
			}
			return nil
		},
	}

	os.Args = []string{"test", "test-name"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestIntArgument(t *testing.T) {
	var argValue int

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&IntArg{
				Name:     "number",
				Required: true,
				AssignTo: &argValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if argValue != 42 {
				t.Fatalf("expected argument value to be 42, got %d", argValue)
			}
			return nil
		},
	}

	os.Args = []string{"test", "42"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBoolArgument(t *testing.T) {
	var argValue bool

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&BoolArg{
				Name:     "bool",
				Required: true,
				AssignTo: &argValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if !argValue {
				t.Fatal("expected argument value to be true")
			}
			return nil
		},
	}

	os.Args = []string{"test", "true"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOptionalArgument(t *testing.T) {
	var argValue string

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&StringArg{
				Name:     "name",
				Required: false,
				AssignTo: &argValue,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if argValue != "" {
				t.Fatalf("expected argument value to be empty, got %s", argValue)
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

func TestRequiredArgumentMissing(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&StringArg{
				Name:     "name",
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
		t.Fatal("expected error for missing required argument")
	}
}

func TestMultipleArguments(t *testing.T) {
	var name string
	var age int

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&StringArg{
				Name:     "name",
				Required: true,
				AssignTo: &name,
			},
			&IntArg{
				Name:     "age",
				Required: true,
				AssignTo: &age,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if name != "John" || age != 30 {
				t.Fatalf("expected name=John age=30, got name=%s age=%d", name, age)
			}
			return nil
		},
	}

	os.Args = []string{"test", "John", "30"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestArgumentsWithFlags(t *testing.T) {
	var name string
	var verbose bool

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Flags: []Flag{
			&BoolFlag{
				Name:     "verbose",
				AssignTo: &verbose,
			},
		},
		Arguments: []Argument{
			&StringArg{
				Name:     "name",
				Required: true,
				AssignTo: &name,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if name != "test" || !verbose {
				t.Fatalf("expected name=test verbose=true, got name=%s verbose=%v", name, verbose)
			}
			return nil
		},
	}

	os.Args = []string{"test", "--verbose", "test"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestArgumentValidation(t *testing.T) {
	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&IntArg{
				Name:     "port",
				Required: true,
				ValidateArg: func(c *Command) error {
					port := c.GetIntArg("port")
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

	os.Args = []string{"test", "8080"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHasArg(t *testing.T) {
	var name string

	cmd := &Command{
		Name:    "test",
		Version: "1.0.0",
		Arguments: []Argument{
			&StringArg{
				Name:     "name",
				Required: true,
				AssignTo: &name,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if !cmd.HasArg("name") {
				t.Fatal("expected HasArg to return true for provided argument")
			}
			return nil
		},
	}

	os.Args = []string{"test", "value"}
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestArgumentsWithSubcommands_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		cmd        *Command
		args       []string
		wantErr    bool
		errContains string
	}{
		{
			name: "parent with arguments and unknown subcommand",
			cmd: &Command{
				Name: "parent",
				Arguments: []Argument{
					&StringArg{Name: "arg1", Required: true},
				},
				Commands: []*Command{
					{Name: "child", Run: func(ctx context.Context, cmd *Command) error { return nil }},
				},
			},
			args:       []string{"parent", "unknown"},
			wantErr:    true,
			errContains: "unknown command",
		},
		{
			name: "parent with arguments, valid subcommand call",
			cmd: &Command{
				Name: "parent",
				Commands: []*Command{
					{
						Name: "child",
						Arguments: []Argument{
							&StringArg{Name: "childarg", Required: true},
						},
						Run: func(ctx context.Context, cmd *Command) error { return nil },
					},
				},
			},
			args:    []string{"parent", "child", "value"},
			wantErr: false,
		},
		{
			name: "parent with arguments, subcommand missing required arg",
			cmd: &Command{
				Name: "parent",
				Commands: []*Command{
					{
						Name: "child",
						Arguments: []Argument{
							&StringArg{Name: "childarg", Required: true},
						},
						Run: func(ctx context.Context, cmd *Command) error { return nil },
					},
				},
			},
			args:       []string{"parent", "child"},
			wantErr:    true,
			errContains: "required",
		},
		{
			name: "parent with MinArgs and arguments, unknown subcommand",
			cmd: &Command{
				Name:    "parent",
				MinArgs: 1,
				Arguments: []Argument{
					&StringArg{Name: "arg1", Required: true},
				},
				Commands: []*Command{
					{Name: "child", Run: func(ctx context.Context, cmd *Command) error { return nil }},
				},
			},
			args:       []string{"parent", "unknown"},
			wantErr:    true,
			errContains: "unknown command",
		},
		{
			name: "parent with required arguments and no args (has subcommands)",
			cmd: &Command{
				Name: "parent",
				Arguments: []Argument{
					&StringArg{Name: "arg1", Required: true},
				},
				Commands: []*Command{
					{Name: "child", Run: func(ctx context.Context, cmd *Command) error { return nil }},
				},
			},
			args:       []string{"parent"},
			wantErr:    true,
			errContains: "missing required argument",
		},
		{
			name: "parent with optional arguments and valid subcommand",
			cmd: &Command{
				Name: "parent",
				Arguments: []Argument{
					&StringArg{Name: "arg1", Required: false},
				},
				Commands: []*Command{
					{
						Name: "child",
						Run:  func(ctx context.Context, cmd *Command) error { return nil },
					},
				},
			},
			args:    []string{"parent", "child"},
			wantErr: false,
		},
		{
			name: "nested subcommands with arguments at different levels",
			cmd: &Command{
				Name: "parent",
				Commands: []*Command{
					{
						Name: "child",
						Commands: []*Command{
							{
								Name: "grandchild",
								Arguments: []Argument{
									&StringArg{Name: "garg", Required: true},
								},
								Run: func(ctx context.Context, cmd *Command) error { return nil },
							},
						},
					},
				},
			},
			args:    []string{"parent", "child", "grandchild", "value"},
			wantErr: false,
		},
		{
			name: "parent with MaxArgs and arguments, unknown subcommand",
			cmd: &Command{
				Name:    "parent",
				MaxArgs: 0,
				Arguments: []Argument{
					&StringArg{Name: "arg1", Required: false},
				},
				Commands: []*Command{
					{Name: "child", Run: func(ctx context.Context, cmd *Command) error { return nil }},
				},
			},
			args:       []string{"parent", "unknown"},
			wantErr:    true,
			errContains: "unknown command",
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

