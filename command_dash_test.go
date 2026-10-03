package cli

import (
	"context"
	"os"
	"testing"
)

// dashCmd builds root -> get, where get takes two string arguments, a bool
// flag -r, and string flags --out/-o.
func dashCmd(t *testing.T, got *[]string, out *string, recursive *bool) *Command {
	t.Helper()
	get := &Command{
		Name: "get",
		Arguments: []Argument{
			&StringArg{Name: "source", Required: true},
			&StringArg{Name: "dest"},
		},
		Flags: []Flag{
			&BoolFlag{Name: "recursive", Aliases: []string{"r"}},
			&StringFlag{Name: "out", Aliases: []string{"o"}},
			&StringFlag{Name: "server", Global: true},
		},
		MaxArgs: NoArgs,
		Run: func(ctx context.Context, cmd *Command) error {
			*got = []string{cmd.GetStringArg("source"), cmd.GetStringArg("dest")}
			*out = cmd.GetString("out")
			*recursive = cmd.GetBool("recursive")
			return nil
		},
	}
	return &Command{Name: "tool", Commands: []*Command{get}}
}

// A lone "-" means stdin or stdout by convention, so it is an argument where
// it is given, not a flag, and keeps its place among the arguments.
func TestLoneDashIsPositional(t *testing.T) {
	cases := []struct {
		args      []string
		source    string
		dest      string
		out       string
		recursive bool
	}{
		{[]string{"get", "bucket/key", "-"}, "bucket/key", "-", "", false},
		{[]string{"get", "-", "bucket/key"}, "-", "bucket/key", "", false},
		{[]string{"get", "-r", "bucket/prefix", "-"}, "bucket/prefix", "-", "", true},
		{[]string{"get", "bucket/key", "-", "--server=x"}, "bucket/key", "-", "", false},
		{[]string{"get", "bucket/key", "-", "-r"}, "bucket/key", "-", "", true},
		{[]string{"get", "--out", "-", "bucket/key"}, "bucket/key", "", "-", false},
		{[]string{"get", "-o", "-", "bucket/key"}, "bucket/key", "", "-", false},
		{[]string{"get", "-ro", "-", "bucket/key"}, "bucket/key", "", "-", true},
		{[]string{"get", "--", "-r", "-"}, "-r", "-", "", false},
	}
	for _, tc := range cases {
		var got []string
		var out string
		var recursive bool
		root := dashCmd(t, &got, &out, &recursive)
		os.Args = append([]string{"tool"}, tc.args...)
		if err := root.Execute(context.Background()); err != nil {
			t.Errorf("%v: %v", tc.args, err)
			continue
		}
		if got[0] != tc.source || got[1] != tc.dest || out != tc.out || recursive != tc.recursive {
			t.Errorf("%v: source %q dest %q out %q recursive %v, want %q %q %q %v",
				tc.args, got[0], got[1], out, recursive, tc.source, tc.dest, tc.out, tc.recursive)
		}
	}
}

// A flag needing a value still refuses a following flag as its value.
func TestFlagValueIsNotAFlag(t *testing.T) {
	var got []string
	var out string
	var recursive bool
	root := dashCmd(t, &got, &out, &recursive)
	os.Args = []string{"tool", "get", "--out", "-r", "bucket/key"}
	if err := root.Execute(context.Background()); err == nil && out == "-r" {
		t.Error("a flag was taken as another flag's value")
	}
}
