package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// GenerateCompletionCommand creates a new command that outputs shell completion scripts
func GenerateCompletionCommand() *Command {
	return &Command{
		Name:        "completion",
		Usage:       "Generate shell completion scripts",
		Description: "Output shell completion scripts for bash, zsh, fish or powershell",
		PreRun:      func(ctx context.Context, cmd *Command) (context.Context, error) { return ctx, nil },
		PostRun:     func(ctx context.Context, cmd *Command) error { return nil },
		Arguments: []Argument{
			&StringArg{
				Name:     "shell",
				Usage:    "Shell type (bash, zsh, fish, powershell)",
				Required: true,
			},
		},
		Flags: []Flag{
			&StringFlag{
				Name:   "command",
				Usage:  "Return command completions for the given command path",
				Hidden: true,
			},
			&StringFlag{
				Name:   "flag",
				Usage:  "Return flag completions for the given command path",
				Hidden: true,
			},
			&StringFlag{
				Name:   "value-flags",
				Usage:  "Return flags that take a value for the given command path",
				Hidden: true,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			shell := cmd.GetStringArg("shell")

			// Handle the dynamic completion mode when called with the completion flags
			if cmd.HasFlag("command") {
				handleCommandCompletion(cmd, shell)
				return nil
			} else if cmd.HasFlag("flag") {
				handleFlagCompletion(cmd, shell)
				return nil
			} else if cmd.HasFlag("value-flags") {
				handleValueFlagsCompletion(cmd)
				return nil
			}

			// Generate completion script for the requested shell
			rootCmd := cmd.GetRootCmd()

			switch strings.ToLower(shell) {
			case "bash":
				return generateDynamicBashCompletion(os.Stdout, rootCmd)
			case "zsh":
				return generateDynamicZshCompletion(os.Stdout, rootCmd)
			case "fish":
				return generateDynamicFishCompletion(os.Stdout, rootCmd)
			case "powershell":
				return generateDynamicPowershellCompletion(os.Stdout, rootCmd)
			default:
				return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
			}
		},
	}
}

// handleCommandCompletion prints available commands for the given path
func handleCommandCompletion(cmd *Command, shell string) {
	cmdPath := cmd.GetString("command")
	rootCmd := cmd.GetRootCmd()

	// Parse the command path to find the target command
	pathParts := strings.Split(cmdPath, " ")
	current := rootCmd

	// Navigate to the specified command
	for _, part := range pathParts {
		if part == "" || part == rootCmd.Name {
			continue
		}

		found := false
		for _, subCmd := range current.Commands {
			if subCmd.Name == part {
				current = subCmd
				found = true
				break
			}
		}

		if !found {
			return
		}
	}

	// Output available subcommands
	for _, subCmd := range current.Commands {
		switch shell {
		case "fish":
			// Fish uses tab-separated description format
			if subCmd.Usage != "" {
				fmt.Printf("%s\t%s\n", subCmd.Name, subCmd.Usage)
			} else {
				fmt.Println(subCmd.Name)
			}

		case "powershell":
			// Powershell uses value:description format
			if subCmd.Usage != "" {
				fmt.Printf("%s:%s\n", subCmd.Name, subCmd.Usage)
			} else {
				fmt.Println(subCmd.Name)
			}

		default:
			// Just need command names
			fmt.Println(subCmd.Name)
		}
	}
}

// handleFlagCompletion prints available flags for the given command path
func handleFlagCompletion(cmd *Command, shell string) {
	cmdPath := cmd.GetString("flag")
	rootCmd := cmd.GetRootCmd()

	// Parse the command path to find the target command
	pathParts := strings.Split(cmdPath, " ")
	current := rootCmd

	var globalFlags []Flag

	// Navigate to the specified command
	for _, part := range pathParts {
		if part == "" || part == rootCmd.Name {
			continue
		}

		for _, flag := range current.Flags {
			if flag.isGlobal() && !flag.isHidden() {
				globalFlags = append(globalFlags, flag)
			}
		}

		found := false
		for _, subCmd := range current.Commands {
			if subCmd.Name == part {
				current = subCmd
				found = true
				break
			}
		}

		if !found {
			return
		}
	}

	// Output available flags & global flags
	printFlag := func(flag Flag) {
		if flag.isHidden() {
			return
		}
		name := flag.getName()
		switch shell {
		case "fish":
			if flag.getUsage() == "" {
				fmt.Printf("--%s\n", name)
			} else {
				fmt.Printf("--%s\t%s\n", name, flag.getUsage())
			}
			for _, alias := range flag.getAliases() {
				if len(alias) == 1 {
					fmt.Printf("-%s\n", alias)
				} else {
					fmt.Printf("--%s\n", alias)
				}
			}
		case "powershell":
			if flag.getUsage() != "" {
				fmt.Printf("--%s:%s\n", name, flag.getUsage())
			} else {
				fmt.Printf("--%s\n", name)
			}
			for _, alias := range flag.getAliases() {
				if len(alias) == 1 {
					fmt.Printf("-%s\n", alias)
				} else {
					fmt.Printf("--%s\n", alias)
				}
			}
		default:
			fmt.Printf("--%s\n", name)
			for _, alias := range flag.getAliases() {
				if len(alias) == 1 {
					fmt.Printf("-%s\n", alias)
				} else {
					fmt.Printf("--%s\n", alias)
				}
			}
		}
	}

	for _, flag := range current.Flags {
		printFlag(flag)
	}
	for _, flag := range globalFlags {
		printFlag(flag)
	}
}

// handleValueFlagsCompletion prints flag names (and aliases) that take a value for the given command path
func handleValueFlagsCompletion(cmd *Command) {
	cmdPath := cmd.GetString("value-flags")
	rootCmd := cmd.GetRootCmd()

	pathParts := strings.Split(cmdPath, " ")
	current := rootCmd

	var globalFlags []Flag

	for _, part := range pathParts {
		if part == "" || part == rootCmd.Name {
			continue
		}
		for _, flag := range current.Flags {
			if flag.isGlobal() {
				globalFlags = append(globalFlags, flag)
			}
		}
		found := false
		for _, subCmd := range current.Commands {
			if subCmd.Name == part {
				current = subCmd
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	printValueFlag := func(flag Flag) {
		if _, isBool := flag.(*BoolFlag); isBool {
			return
		}
		fmt.Println(flag.getName())
		for _, alias := range flag.getAliases() {
			fmt.Println(alias)
		}
	}

	for _, flag := range current.Flags {
		printValueFlag(flag)
	}
	for _, flag := range globalFlags {
		printValueFlag(flag)
	}
}

// Generate a dynamic bash completion script
func generateDynamicBashCompletion(w io.Writer, root *Command) error {
	cmdName := root.Name

	fmt.Fprintf(w, `# bash completion script for the command %[1]s

_%[1]s() {
    local exec_path

    # Check if exec is in the PATH, otherwise assume a local executable
    if command -v %[1]s >/dev/null 2>&1; then
        exec_path="%[1]s"
    else
        exec_path="./%[1]s"
    fi

    # Exit if the command is not executable
    if ! [[ -x "$exec_path" ]]; then
        return 1
    fi

    # Capture the current command line words
    local cmdpath="%[1]s"
    local current_word="${COMP_WORDS[COMP_CWORD]}"
    local completions

    # Get flags that take a value so we can skip their arguments
    local value_flags
    value_flags=$($exec_path completion bash --value-flags="$cmdpath")

    # Build the command path from all non-flag arguments
    if [[ ${#COMP_WORDS[@]} -gt 1 ]]; then
        local skip_next=0
        for ((i=1; i<COMP_CWORD; i++)); do
            if [[ $skip_next -eq 1 ]]; then
                skip_next=0
                continue
            fi
            if [[ "${COMP_WORDS[i]}" == -* ]]; then
                local fname="${COMP_WORDS[i]#--}"
                fname="${fname#-}"
                if echo "$value_flags" | grep -qx "$fname"; then
                    skip_next=1
                fi
            else
                cmdpath+=" ${COMP_WORDS[i]}"
            fi
        done
    fi

    # Request completions from the binary
    if [[ "$current_word" == -* ]]; then
        # Flag completion
        completions=$($exec_path completion bash --flag="$cmdpath")
    else
        # Command/subcommand/argument completion
        completions=$($exec_path completion bash --command="$cmdpath")
    fi

    # Filter completions against the current word (use mapfile for safe handling)
    mapfile -t COMPREPLY < <(compgen -W "${completions}" -- "$current_word")
}

# Register the completion function
complete -o bashdefault -o default -o nospace -F _%[1]s %[1]s
`, cmdName)

	return nil
}

// generateDynamicZshCompletion writes a zsh completion script that calls back to the program
func generateDynamicZshCompletion(w io.Writer, root *Command) error {
	cmdName := root.Name

	// Write the function header
	fmt.Fprintf(w, `# zsh completion script for the command %[1]s
autoload -U compinit && compinit

_%[1]s() {
    local exec_path
    local -a suggestions

    # Check if exec is in the PATH, otherwise assume a local executable
    if command -v %[1]s >/dev/null 2>&1; then
        exec_path="%[1]s"
    else
        exec_path="./%[1]s"
    fi

    # Exit if the command is not executable
    if ! [[ -x "$exec_path" ]]; then
        return 1
    fi

    # Capture the current command line words
    local cmdpath="%[1]s"
    local current_word="${words[$CURRENT]}"
    local completions

    # Get flags that take a value so we can skip their arguments
    local value_flags
    value_flags=$($exec_path completion zsh --value-flags="$cmdpath")

    # Skip command name and build from arguments
    if [[ ${#words[@]} -gt 1 ]]; then
        local skip_next=0
        for ((i=2; i<CURRENT; i++)); do
            if [[ $skip_next -eq 1 ]]; then
                skip_next=0
                continue
            fi
            if [[ "${words[i]}" == -* ]]; then
                # Check if this flag takes a value
                local fname="${words[i]#--}"
                fname="${fname#-}"
                if echo "$value_flags" | grep -qx "$fname"; then
                    skip_next=1
                fi
            else
                cmdpath+=" ${words[i]}"
            fi
        done
    fi

    # Determine whether we are completing a flag or a command/argument
    if [[ "$current_word" == -* ]]; then
        # Request flag completions
        completions=$($exec_path completion zsh --flag="$cmdpath")
    else
        # Request command or argument completions
        completions=$($exec_path completion zsh --command="$cmdpath")
    fi

    # Split the output from the command into an array of suggestions
    suggestions=("${(@f)completions}")

    # Add the suggestions to the completion list
    compadd -- "${suggestions[@]}"
}

# Register the completion function
compdef _%[1]s %[1]s
`, cmdName)

	return nil
}

// generateDynamicFishCompletion writes a fish completion script that calls back to the program
func generateDynamicFishCompletion(w io.Writer, root *Command) error {
	cmdName := root.Name

	fmt.Fprintf(w, `# fish completion script for the command %[1]s

function __%[1]s_completion
    set -l exec_path
    set -l cmd_line (commandline -opc)
    set -l current_token (commandline -ct)
    set -l cmd_path "%[1]s"

    # Check if exec is in the PATH, otherwise assume a local executable
    if command -sq %[1]s
        set exec_path "%[1]s"
    else
        set exec_path "./%[1]s"
    end

    # Exit if the command is not executable
    if not test -x "$exec_path"
        return 1
    end

    # Build the command path including all executed subcommands (not the current token)
    # First token is always the exec name which we've already included
    if test (count $cmd_line) -gt 1
        # Skip first token (the command itself)
        set -l cmd_parts $cmd_line[2..-1]

        # Remove the last token if it's incomplete (current token being completed)
        if string match -q -- "*$current_token" $cmd_line[-1] && test -n "$current_token"
            set cmd_parts $cmd_parts[1..-2]
        end

        # Get flags that take a value so we can skip their arguments
        set -l value_flags (eval $exec_path completion fish --value-flags=\"$cmd_path\")

        set -l skip_next 0
        for part in $cmd_parts
            if test $skip_next -eq 1
                set skip_next 0
                continue
            end
            if string match -q -- '-*' $part
                set -l fname (string replace --regex -- '^--?' '' $part)
                if contains -- $fname $value_flags
                    set skip_next 1
                end
            else
                set cmd_path "$cmd_path $part"
            end
        end
    end

    # Request completions from the binary
    if string match -q -- '-*' $current_token
        # Flag completion
        eval $exec_path completion fish --flag=\"$cmd_path\"
    else
        # Command/subcommand/argument completion
        eval $exec_path completion fish --command=\"$cmd_path\"
    end
end

# Register the completion function
complete -c %[1]s -f -a '(__%[1]s_completion)'
`, cmdName)

	return nil
}

// generateDynamicPowershellCompletion generates a PowerShell completion script
func generateDynamicPowershellCompletion(w io.Writer, root *Command) error {
	cmdName := root.Name

	fmt.Fprintf(w, `# PowerShell completion script for the command %[1]s

Register-ArgumentCompleter -Native -CommandName %[1]s -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # Get the command line and the current word
    $cmdLine = $commandAst.ToString()
    $currentWord = $wordToComplete

    # Set the executable path
    $execPath = $null
    if (Test-Path -Path "./%[1]s.exe") {
        $execPath = "./%[1]s.exe"
    } elseif (Test-Path -Path "./%[1]s") {
        $execPath = "./%[1]s"
    } elseif (Get-Command %[1]s -ErrorAction SilentlyContinue) {
        $execPath = "%[1]s"
    } else {
        # No executable found
        return @()
    }

    # Build the command path from all tokens before cursor
    $cmdPath = "%[1]s"
    $tokens = $commandAst.CommandElements

    # Get flags that take a value so we can skip their arguments
`, cmdName)
	fmt.Fprintln(w, "    $valueFlags = (& $execPath completion powershell --value-flags=\"$cmdPath\" 2>$null) -split \"`n\" | Where-Object { $_ -ne \"\" }")
	fmt.Fprint(w, `
    # Start at index 1 to skip the command itself
    $skipNext = $false
    for ($i = 1; $i -lt $tokens.Count; $i++) {
        $token = $tokens[$i].ToString()

        # Skip if this is the current word being completed
        if ($i -eq $tokens.Count - 1 -and $token -eq $currentWord) {
            continue
        }

        if ($skipNext) {
            $skipNext = $false
            continue
        }

        if ($token.StartsWith("-")) {
            $fname = $token -replace '^--?', ''
            if ($valueFlags -contains $fname) {
                $skipNext = $true
            }
        } else {
            $cmdPath += " $token"
        }
    }

    # Determine if we're completing a flag or a command/argument
    $completions = $null
    if ($currentWord -match "^-") {
        # Flag completion
        $completions = & $execPath completion powershell --flag="$cmdPath" 2>$null
    } else {
        # Command/subcommand/argument completion
        $completions = & $execPath completion powershell --command="$cmdPath" 2>$null
    }

	# Process completions and return them as CompletionResults
	if ($completions) {
`)
	fmt.Fprintln(w, "    $completions -split \"`n\" | Where-Object { $_ -ne '' -and $_.Split(':')[0] -like \"$currentWord*\" } | ForEach-Object {")
	fmt.Fprint(w, `        $parts = $_ -split ':', 2
        $value = $parts[0]
        $description = if ($parts.Count -gt 1 -and $parts[1] -ne '') { $parts[1] } else { $value }
        $resultType = if ($value.StartsWith('-')) { 'ParameterName' } else { 'Text' }
        [System.Management.Automation.CompletionResult]::new(
            $value,
            $value,
            $resultType,
            $description
        )
    }
}
}

`)
	fmt.Fprintf(w, "# Note: This script should be dot-sourced or added to your PowerShell profile\n# Example: . ./%s_completions.ps1\n", cmdName)

	return nil
}
