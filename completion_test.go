package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

// testCompletionCmd builds a reusable command tree for completion tests:
//
//	root (--config global, --verbose bool global, --name string, --count int, --flag-only bool)
//	  └── greet (--greeting string, -g alias)
//	        └── formal (--title string, -t alias)
func testCompletionCmd() *Command {
	return &Command{
		Name: "app",
		Flags: []Flag{
			&StringFlag{Name: "config", Global: true},
			&BoolFlag{Name: "verbose", Global: true},
			&StringFlag{Name: "name", Aliases: []string{"n", "name-alias"}},
			&IntFlag{Name: "count", Aliases: []string{"c"}},
			&BoolFlag{Name: "flag-only"},
		},
		Commands: []*Command{
			GenerateCompletionCommand(),
			{
				Name: "greet",
				Flags: []Flag{
					&StringFlag{Name: "greeting", Aliases: []string{"g"}},
				},
				Commands: []*Command{
					{
						Name: "formal",
						Flags: []Flag{
							&StringFlag{Name: "title", Aliases: []string{"t"}},
						},
					},
				},
			},
		},
	}
}

func captureCompletionOutput(t *testing.T, args []string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = append([]string{"app"}, args...)
	_ = testCompletionCmd().Execute(nil) //nolint

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func assertContains(t *testing.T, output, want string) {
	t.Helper()
	if !strings.Contains(output, want) {
		t.Errorf("expected output to contain %q\ngot:\n%s", want, output)
	}
}

func assertNotContains(t *testing.T, output, want string) {
	t.Helper()
	if strings.Contains(output, want) {
		t.Errorf("expected output NOT to contain %q\ngot:\n%s", want, output)
	}
}

func assertLines(t *testing.T, output string, want []string) {
	t.Helper()
	lines := map[string]bool{}
	for _, l := range strings.Split(strings.TrimSpace(output), "\n") {
		if l != "" {
			lines[l] = true
		}
	}
	for _, w := range want {
		if !lines[w] {
			t.Errorf("expected line %q in output\ngot:\n%s", w, output)
		}
	}
}

// ---------------------------------------------------------------------------
// --command completions
// ---------------------------------------------------------------------------

func TestCompletion_Command_RootSubcommands(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell, "--command=app"})
			assertContains(t, out, "greet")
			assertContains(t, out, "completion")
		})
	}
}

func TestCompletion_Command_NestedSubcommands(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell, "--command=app greet"})
			assertContains(t, out, "formal")
			assertNotContains(t, out, "greet") // sibling, not child
		})
	}
}

func TestCompletion_Command_UnknownPath(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell, "--command=app unknown"})
			if strings.TrimSpace(out) != "" {
				t.Errorf("expected empty output for unknown path, got: %s", out)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// --flag completions
// ---------------------------------------------------------------------------

func TestCompletion_Flag_RootFlags(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell, "--flag=app"})
			assertContains(t, out, "--name")
			assertContains(t, out, "--count")
			assertContains(t, out, "--flag-only")
		})
	}
}

func TestCompletion_Flag_SubcommandFlags(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell, "--flag=app greet"})
			assertContains(t, out, "--greeting")
			// global flags propagate
			assertContains(t, out, "--config")
			assertContains(t, out, "--verbose")
			// root-only flags should not appear
			assertNotContains(t, out, "--name")
		})
	}
}

func TestCompletion_Flag_NestedSubcommandFlags(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell, "--flag=app greet formal"})
			assertContains(t, out, "--title")
			// global flags from greet propagate
			assertContains(t, out, "--config")
		})
	}
}

func TestCompletion_Flag_HiddenFlagsExcluded(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell, "--flag=app"})
			// the internal completion flags are hidden and must not appear as exact entries
			for _, hidden := range []string{"--command", "--value-flags"} {
				if strings.Contains(out, hidden) {
					t.Errorf("hidden flag %q must not appear in output\ngot:\n%s", hidden, out)
				}
			}
			// "--flag" is a hidden flag but "--flag-only" is not; check exact line match
			for _, line := range strings.Split(out, "\n") {
				if line == "--flag" {
					t.Errorf("hidden flag --flag must not appear as its own line\ngot:\n%s", out)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// --value-flags (zsh only — used to skip flag arguments in cmdpath building)
// ---------------------------------------------------------------------------

func TestCompletion_ValueFlags_Root(t *testing.T) {
	out := captureCompletionOutput(t, []string{"completion", "zsh", "--value-flags=app"})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	lineSet := map[string]bool{}
	for _, l := range lines {
		lineSet[l] = true
	}
	// value-taking flags and their aliases
	for _, want := range []string{"config", "name", "n", "count", "c"} {
		if !lineSet[want] {
			t.Errorf("expected %q in value-flags output\ngot: %s", want, out)
		}
	}
	// bool flags must NOT appear
	for _, notWant := range []string{"verbose", "flag-only"} {
		if lineSet[notWant] {
			t.Errorf("bool flag %q must not appear in value-flags output\ngot: %s", notWant, out)
		}
	}
}

func TestCompletion_ValueFlags_Subcommand(t *testing.T) {
	out := captureCompletionOutput(t, []string{"completion", "zsh", "--value-flags=app greet"})
	assertLines(t, out, []string{"greeting", "g", "config"})
	assertNotContains(t, out, "verbose")
}

// ---------------------------------------------------------------------------
// Alias completions (fish & powershell include descriptions; bash/zsh plain)
// ---------------------------------------------------------------------------

func TestCompletion_Flag_AliasesLongOnly(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell, "--flag=app"})
			// multi-char alias must appear
			assertContains(t, out, "--name-alias")
			// single-char aliases must appear with single dash prefix
			assertContains(t, out, "-n")
			assertContains(t, out, "-c")
		})
	}
}

func TestCompletion_Flag_AliasesInFishOutput(t *testing.T) {
	out := captureCompletionOutput(t, []string{"completion", "fish", "--flag=app"})
	assertContains(t, out, "--name")
	assertContains(t, out, "--count")
	// single-char aliases must appear with single dash prefix
	assertContains(t, out, "-n")
	assertContains(t, out, "-c")
}

func TestCompletion_Command_FishDescriptions(t *testing.T) {
	cmd := &Command{
		Name: "app",
		Commands: []*Command{
			GenerateCompletionCommand(),
			{Name: "greet", Usage: "Greet someone"},
		},
	}
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Args = []string{"app", "completion", "fish", "--command=app"}
	_ = cmd.Execute(nil)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()
	assertContains(t, out, "greet\tGreet someone")
}

func TestCompletion_Command_PowershellDescriptions(t *testing.T) {
	cmd := &Command{
		Name: "app",
		Commands: []*Command{
			GenerateCompletionCommand(),
			{Name: "greet", Usage: "Greet someone"},
		},
	}
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Args = []string{"app", "completion", "powershell", "--command=app"}
	_ = cmd.Execute(nil)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()
	assertContains(t, out, "greet:Greet someone")
}

// ---------------------------------------------------------------------------
// Dynamic argument completions (--arg)
// ---------------------------------------------------------------------------

func testDynamicCompletionCmd() *Command {
	return &Command{
		Name: "app",
		Commands: []*Command{
			GenerateCompletionCommand(),
			{
				Name:  "space",
				Usage: "Manage spaces",
				Commands: []*Command{
					{
						Name:  "restart",
						Usage: "Restart a space",
						Arguments: []Argument{
							&StringArg{
								Name:  "name",
								Usage: "Space name",
								CompletionFunc: func(_ context.Context, _ *Command) []CompletionItem {
									return []CompletionItem{
										{Value: "apple", Description: "Apple space"},
										{Value: "banana", Description: "Banana space"},
										{Value: "cherry", Description: "Cherry space"},
									}
								},
							},
						},
					},
				},
			},
		},
	}
}

func captureArgCompletionOutput(t *testing.T, shell, argFlag string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Args = []string{"app", "completion", shell, "--arg=" + argFlag}
	_ = testDynamicCompletionCmd().Execute(nil) //nolint
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestCompletion_Arg_AllItems(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureArgCompletionOutput(t, shell, "app space restart:0:")
			assertContains(t, out, "apple")
			assertContains(t, out, "banana")
			assertContains(t, out, "cherry")
		})
	}
}

func TestCompletion_Arg_Filtered(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureArgCompletionOutput(t, shell, "app space restart:0:b")
			assertContains(t, out, "banana")
			assertNotContains(t, out, "apple")
			assertNotContains(t, out, "cherry")
		})
	}
}

func TestCompletion_Arg_FishDescriptions(t *testing.T) {
	out := captureArgCompletionOutput(t, "fish", "app space restart:0:")
	assertContains(t, out, "apple\tApple space")
	assertContains(t, out, "banana\tBanana space")
}

func TestCompletion_Arg_PowershellDescriptions(t *testing.T) {
	out := captureArgCompletionOutput(t, "powershell", "app space restart:0:")
	assertContains(t, out, "apple:Apple space")
	assertContains(t, out, "banana:Banana space")
}

func TestCompletion_Arg_OutOfRange(t *testing.T) {
	out := captureArgCompletionOutput(t, "bash", "app space restart:5:")
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty output for out-of-range arg index, got: %s", out)
	}
}

func TestCompletion_Arg_NoCompletionFunc(t *testing.T) {
	// A command with an argument but no CompletionFunc should return nothing
	cmd := &Command{
		Name: "app",
		Commands: []*Command{
			GenerateCompletionCommand(),
			{
				Name: "run",
				Arguments: []Argument{
					&StringArg{Name: "target"},
				},
			},
		},
	}
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Args = []string{"app", "completion", "bash", "--arg=app run:0:"}
	_ = cmd.Execute(nil) //nolint
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty output when no CompletionFunc, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// Script generation smoke tests
// ---------------------------------------------------------------------------

func TestCompletion_ScriptGeneration(t *testing.T) {
	shells := map[string][]string{
		"bash":       {"_app()", "complete -o bashdefault"},
		"zsh":        {"compdef _app app", "autoload -U compinit"},
		"fish":       {"complete -c app", "__app_completion"},
		"powershell": {"Register-ArgumentCompleter", "param($wordToComplete"},
	}
	for shell, markers := range shells {
		t.Run(shell, func(t *testing.T) {
			out := captureCompletionOutput(t, []string{"completion", shell})
			for _, m := range markers {
				assertContains(t, out, m)
			}
		})
	}
}
