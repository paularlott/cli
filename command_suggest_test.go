package cli

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func noopRun(ctx context.Context, cmd *Command) error { return nil }

// Every candidate is scored, not just the first that matches.
func TestFindSimilarCommandsScoresEveryCommand(t *testing.T) {
	commands := []*Command{{Name: "port"}, {Name: "pool"}, {Name: "template"}, {Name: "space"}}
	got := (&Command{}).findSimilarCommands("templat", commands, 2)
	if len(got) == 0 || got[0] != "template" {
		t.Fatalf("suggestions for templat = %v, want template first", got)
	}
	if len(got) != 1 {
		t.Errorf("suggestions for templat = %v, want only template", got)
	}
	// A short typo isn't matched to a long, unrelated name.
	got = (&Command{}).findSimilarCommands("lst", []*Command{{Name: "list"}, {Name: "delete"}}, 2)
	if len(got) != 1 || got[0] != "list" {
		t.Errorf("suggestions for lst = %v, want only list", got)
	}
	got = (&Command{}).findSimilarCommands("spce", commands, 2)
	if len(got) == 0 || got[0] != "space" {
		t.Fatalf("suggestions for spce = %v, want space first", got)
	}
}

// Aliases are matched too, and a command is suggested once, with its aliases.
func TestFindSimilarCommandsIncludesAliases(t *testing.T) {
	commands := []*Command{{Name: "port"}, {Name: "ls", Aliases: []string{"list"}}}
	for _, typo := range []string{"lst", "list3"} {
		got := (&Command{}).findSimilarCommands(typo, commands, 2)
		if len(got) != 1 || got[0] != "ls, list" {
			t.Errorf("suggestions for %s = %q, want [\"ls, list\"]", typo, got)
		}
	}
}

// The error carries the suggestions, so the caller prints them once and the
// library prints nothing itself.
func TestUnknownCommandErrorPrintsOnce(t *testing.T) {
	root := &Command{
		Name:        "app",
		Suggestions: true,
		Commands: []*Command{{
			Name:     "file",
			Commands: []*Command{{Name: "ls", Aliases: []string{"list"}, Run: noopRun}},
		}},
	}

	saved, savedOut := os.Args, os.Stdout
	t.Cleanup(func() { os.Args, os.Stdout = saved, savedOut })
	os.Args = []string{"app", "file", "list3", "bucket:path"}
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := root.Execute(context.Background())
	w.Close()
	os.Stdout = savedOut
	printed, _ := io.ReadAll(r)

	if len(printed) > 0 {
		t.Errorf("the library printed %q", printed)
	}
	want := "unknown command \"list3\"\n\nDid you mean this?\n   - ls, list\n\nRun 'app file --help' for usage"
	if err == nil || err.Error() != want {
		t.Fatalf("error:\n%v\nwant:\n%s", err, want)
	}
}

// A mistyped subcommand of a group command is reported before any PreRun,
// so setup that needs configuration never runs for a command that doesn't
// exist, and the first unknown word is the one matched.
func TestUnknownSubcommandBeforePreRun(t *testing.T) {
	preRan, listRan := false, false
	root := &Command{
		Name:        "app",
		Suggestions: true,
		PreRun: func(ctx context.Context, cmd *Command) (context.Context, error) {
			preRan = true
			return ctx, nil
		},
		Commands: []*Command{{
			Name: "file",
			Commands: []*Command{
				{Name: "bucket", Commands: []*Command{{Name: "list", Run: noopRun}}},
				{Name: "copy", Run: noopRun},
				// A later word naming a sibling must not run it.
				{Name: "ls", Aliases: []string{"list"}, MaxArgs: UnlimitedArgs, Run: func(ctx context.Context, cmd *Command) error {
					listRan = true
					return nil
				}},
			},
		}},
	}

	saved := os.Args
	t.Cleanup(func() { os.Args = saved })
	os.Args = []string{"app", "file", "buket", "list"}

	err := root.Execute(context.Background())
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("got %v, want an unknown command error", err)
	}
	if preRan || listRan {
		t.Errorf("for an unknown command PreRun ran: %v, file ls ran: %v", preRan, listRan)
	}
}

// A command with both a Run and subcommands may take arguments, so they are
// not treated as a mistyped subcommand.
func TestRunnableGroupCommandKeepsArguments(t *testing.T) {
	var got []string
	root := &Command{
		Name:        "app",
		Suggestions: true,
		Commands: []*Command{{
			Name:     "space",
			MaxArgs:  UnlimitedArgs,
			Run:      func(ctx context.Context, cmd *Command) error { got = cmd.GetArgs(); return nil },
			Commands: []*Command{{Name: "list", Run: noopRun}},
		}},
	}

	saved := os.Args
	t.Cleanup(func() { os.Args = saved })
	os.Args = []string{"app", "space", "myspace"}

	if err := root.Execute(context.Background()); err != nil {
		t.Fatalf("got %v, want the command to run", err)
	}
	if len(got) != 1 || got[0] != "myspace" {
		t.Fatalf("args = %v, want [myspace]", got)
	}
}
