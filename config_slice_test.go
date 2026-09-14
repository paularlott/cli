package cli_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	cli "github.com/paularlott/cli"
	cli_toml "github.com/paularlott/cli/toml"
)

func runSlice(t *testing.T, name string, toml string, expect []string) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "t.toml")
	if toml != "" {
		os.WriteFile(cfgPath, []byte(toml), 0644)
	}
	var configFile = cfgPath
	var got []string
	cmd := &cli.Command{
		Name:       name,
		ConfigFile: cli_toml.NewConfigFile(&configFile, nil),
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:         "backends",
				ConfigPath:   []string{"server.enabled_backends"},
				DefaultValue: []string{"ALL"},
			},
		},
		Run: func(ctx context.Context, c *cli.Command) error {
			got = c.GetStringSlice("backends")
			return nil
		},
	}
	saved := os.Args
	os.Args = []string{name}
	defer func() { os.Args = saved }()
	if err := cmd.Execute(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(expect) {
		t.Fatalf("%s: got %v want %v", name, got, expect)
	}
	for i := range expect {
		if got[i] != expect[i] {
			t.Fatalf("%s: got %v want %v", name, got, expect)
		}
	}
}

func TestStringSliceConfigSet(t *testing.T) {
	runSlice(t, "p1", "[server]\nenabled_backends = [\"docker\", \"podman\"]\n", []string{"docker", "podman"})
}
func TestStringSliceConfigEmpty(t *testing.T) {
	runSlice(t, "p2", "[server]\nenabled_backends = []\n", []string{"ALL"})
}
func TestStringSliceConfigAbsent(t *testing.T) {
	runSlice(t, "p3", "[server]\nother = 1\n", []string{"ALL"})
}
func TestStringSliceConfigBlankEntry(t *testing.T) {
	runSlice(t, "p4", "[server]\nenabled_backends = [\"docker\", \"\", \"podman\"]\n", []string{"docker", "podman"})
}
func TestStringSliceConfigStringValue(t *testing.T) {
	runSlice(t, "p5", "[server]\nenabled_backends = \"docker, podman\"\n", []string{"ALL"})
}
