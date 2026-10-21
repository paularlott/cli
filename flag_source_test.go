package cli

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestFlagSource(t *testing.T) {
	t.Setenv("T_ENV", "e")
	t.Setenv("T_ENV_FILE", "e")
	t.Setenv("T_CLI", "e")
	cfg := &mockConfigSource{data: map[string]any{"from_file": "f", "env_and_file": "f"}}
	var got *Command
	cmd := &Command{
		Name:       "app",
		ConfigFile: cfg,
		Flags: []Flag{
			&StringFlag{Name: "cli", Aliases: []string{"c"}, EnvVars: []string{"T_CLI"}},
			&StringFlag{Name: "from-env", EnvVars: []string{"T_ENV"}},
			&StringFlag{Name: "from-file", ConfigPath: []string{"from_file"}},
			&StringFlag{Name: "env-and-file", EnvVars: []string{"T_ENV_FILE"}, ConfigPath: []string{"env_and_file"}},
			&StringFlag{Name: "defaulted", DefaultValue: "d"},
			&StringFlag{Name: "unset"},
		},
		Run: func(_ context.Context, c *Command) error { got = c; return nil },
	}
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"app", "-c", "x"}
	if err := cmd.Execute(context.Background()); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]FlagSource{"cli": SourceCommandLine, "c": SourceCommandLine,
		"from-env": SourceEnv, "from-file": SourceConfigFile, "env-and-file": SourceEnv, "defaulted": SourceDefault,
		"unset": SourceNone, "nope": SourceNone} {
		if s := got.FlagSource(name); s != want {
			t.Errorf("%s: %v, want %v", name, s, want)
		}
	}
}

func TestBadEnvAndConfigValuesFail(t *testing.T) {
	run := func(env string, cfg map[string]any) error {
		t.Setenv("T_PORT", env)
		if env == "" {
			os.Unsetenv("T_PORT")
		}
		cmd := &Command{Name: "app", ConfigFile: &mockConfigSource{data: cfg},
			Flags: []Flag{
				&IntFlag{Name: "port", EnvVars: []string{"T_PORT"}, ConfigPath: []string{"port"}},
				&StringFlag{Name: "name", ConfigPath: []string{"name"}},
			},
			Run: func(context.Context, *Command) error { return nil }}
		old := os.Args
		defer func() { os.Args = old }()
		os.Args = []string{"app"}
		return cmd.Execute(context.Background())
	}
	for name, tc := range map[string]struct {
		env  string
		cfg  map[string]any
		want string
	}{
		"bad env":         {"abc", nil, "environment variable T_PORT"},
		"bad config":      {"", map[string]any{"port": "abc"}, "config port"},
		"list for scalar": {"", map[string]any{"name": []any{"a", "b"}}, "config name: a list"},
	} {
		err := run(tc.env, tc.cfg)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if err := run("8080", map[string]any{"port": int64(9000), "name": "x"}); err != nil {
		t.Fatal(err)
	}
}
