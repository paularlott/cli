package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GenerateCompletionCommand creates a new command that outputs shell completion scripts
func GenerateCompletionCommand() *Command {
	return &Command{
		Name:        "completion",
		Usage:       "Generate shell completion scripts",
		Description: "Output shell completion scripts for bash, zsh, fish or powershell",
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
			}

			// Generate completion script for the requested shell
			rootCmd := cmd.GetRootCmd()

			switch strings.ToLower(shell) {
			case "bash":
				err := generateDynamicBashCompletion(os.Stdout, rootCmd)
				if err == nil {
					fmt.Fprintln(os.Stderr, "\nBash completion has been generated. To use it, run:")
					fmt.Fprintln(os.Stderr, "    source <("+rootCmd.Name+" completion bash)")
					fmt.Fprintln(os.Stderr, "\nTo load completions for each session, execute once:")
					fmt.Fprintln(os.Stderr, "    "+rootCmd.Name+" completion bash > ~/.bash_completion")
				}
				return err
			case "zsh":
				err := generateDynamicZshCompletion(os.Stdout, rootCmd)
				if err == nil {
					fmt.Fprintln(os.Stderr, "\nZsh completion has been generated. To use it, run:")
					fmt.Fprintln(os.Stderr, "    source <("+rootCmd.Name+" completion zsh)")
					fmt.Fprintln(os.Stderr, "\nTo load completions for each session, add to your ~/.zshrc:")
					fmt.Fprintln(os.Stderr, "    "+rootCmd.Name+" completion zsh > \"${fpath[1]}/_"+rootCmd.Name+"\"")
				}
				return err
			case "fish":
				err := generateDynamicFishCompletion(os.Stdout, rootCmd)
				if err == nil {
					fmt.Fprintln(os.Stderr, "\nFish completion has been generated. To use it, run:")
					fmt.Fprintln(os.Stderr, "    "+rootCmd.Name+" completion fish | source")
					fmt.Fprintln(os.Stderr, "\nTo load completions for each session, run once:")
					fmt.Fprintln(os.Stderr, "    "+rootCmd.Name+" completion fish > ~/.config/fish/completions/"+rootCmd.Name+".fish")
				}
				return err
			case "powershell":
				err := generateDynamicPowershellCompletion(os.Stdout, rootCmd)
				if err == nil {
					fmt.Fprintln(os.Stderr, "\nPowerShell completion has been generated. To use it, run:")
					fmt.Fprintln(os.Stderr, "    "+rootCmd.Name+" completion powershell | Out-String | Invoke-Expression")
					fmt.Fprintln(os.Stderr, "\nTo load completions for each session, add to your PowerShell profile:")
					fmt.Fprintln(os.Stderr, "    "+rootCmd.Name+" completion powershell | Out-String | Invoke-Expression")
				}
				return err
			default:
				return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
			}
		},
	}
}

// handleCommandCompletion prints available commands for the given path
func handleCommandCompletion(cmd *Command, shell string) {
	cmdPath := filepath.Base(cmd.GetString("command"))
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
	cmdPath := filepath.Base(cmd.GetString("flag"))
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
	for _, flag := range current.Flags {
		if flag.isHidden() {
			continue
		}

		switch shell {
		case "fish":
			if flag.getUsage() == "" {
				fmt.Printf("--%s\n", flag.getName())
			} else {
				fmt.Printf("--%s\t%s\n", flag.getName(), flag.getUsage())
			}

		case "powershell":
			// Powershell uses value:description format
			if flag.getUsage() != "" {
				fmt.Printf("--%s:%s\n", flag.getName(), flag.getUsage())
			} else {
				fmt.Printf("--%s\n", flag.getName())
			}

		default:
			fmt.Printf("--%s\n", flag.getName())
		}
	}

	for _, flag := range globalFlags {
		switch shell {
		case "fish":
			if flag.getUsage() == "" {
				fmt.Printf("--%s\n", flag.getName())
			} else {
				fmt.Printf("--%s\t%s\n", flag.getName(), flag.getUsage())
			}

		default:
			fmt.Printf("--%s\n", flag.getName())
		}
	}
}

// Generate a dynamic bash completion script
func generateDynamicBashCompletion(w io.Writer, root *Command) error {
	cmdName := root.Name

	fmt.Fprintf(w, `# bash completion script for the command %[1]s

_%[1]s() {
    local exec_path
    local suggestions=()

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

    # Build the command path from all non-flag arguments
    if [[ ${#COMP_WORDS[@]} -gt 1 ]]; then
        for ((i=1; i<COMP_CWORD; i++)); do
            # Only add non-flag tokens to the command path
            if [[ "${COMP_WORDS[i]}" != -* ]]; then
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

    # Split the output into an array of suggestions
    IFS=$'\n' read -r -d '' -a suggestions <<< "$completions"

    # Set completion replies
    COMPREPLY=("${suggestions[@]}")
}

# Register the completion function
complete -o bashdefault -o default -o nospace -F _%[1]s %[1]s`, cmdName)

	return nil
}

// generateDynamicZshCompletion writes a zsh completion script that calls back to the program
func generateDynamicZshCompletion(w io.Writer, root *Command) error {
	cmdName := root.Name

	// Write the function header
	fmt.Fprintf(w, `#compdef %[1]s
# zsh completion script for the command %[1]s

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

		# Skip command name and build from arguments
		if [[ ${#words[@]} -gt 1 ]]; then
			for ((i=2; i<CURRENT; i++)); do
				# Only add subcommands (not flags) to cmdpath
				if [[ "${words[i]}" != -* ]]; then
					# Check if previous word is NOT an option expecting an argument
					if [[ "${words[i-1]}" != -* || "${words[i-1]}" == "--"* ]]; then
						cmdpath+=" ${words[i]}"
					fi
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
compdef _%[1]s %[1]s`, cmdName)

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
        # unless it's a complete word (has a space after it)
        if string match -q -- "*$current_token" $cmd_line[-1] && test -n "$current_token"
            set cmd_parts $cmd_parts[1..-2]
        end

        # Add non-flag tokens to command path
        for part in $cmd_parts
            if not string match -q -- '-*' $part
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
complete -c %[1]s -f -a '(__%[1]s_completion)'`, cmdName)

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
    if (Get-Command %[1]s -ErrorAction SilentlyContinue) {
        $execPath = "%[1]s"
    } elseif (Test-Path -Path "./%[1]s.exe") {
        $execPath = "./%[1]s.exe"
    } elseif (Test-Path -Path "./%[1]s") {
        $execPath = "./%[1]s"
    } else {
        # No executable found
        return @()
    }

    # Build the command path from all tokens before cursor
    $cmdPath = "%[1]s"
    $tokens = $commandAst.CommandElements

    # Start at index 1 to skip the command itself
    for ($i = 1; $i -lt $tokens.Count; $i++) {
        $token = $tokens[$i].ToString()

        # Skip if this is the current word being completed
        if ($i -eq $tokens.Count - 1 -and $token -eq $currentWord) {
            continue
        }

        # Only add non-flag tokens to the command path
        if (-not $token.StartsWith("-")) {
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
`, cmdName)
	fmt.Fprintln(w, "$completions -split \"`n\" | ForEach-Object {")
	fmt.Fprintf(w, `			$line = $_.Trim()
			if ($line) {
				# Parse completion lines - format should be "value:description"
				if ($line -match "^(.*?)(?::(.*))?$") {
					$value = $matches[1]
					$description = if ($matches.Count -gt 2) { $matches[2] } else { $value }

					# Create a CompletionResult
					[System.Management.Automation.CompletionResult]::new(
						$value,    # CompletionText
						$value,    # ListItemText
						'ParameterValue',  # ResultType
						$description  # ToolTip
					)
				}
			}
		}
	}
}

# Note: This script should be dot-sourced or added to your PowerShell profile
# Example: . ./%[1]s_completions.ps1`, cmdName)

	return nil
}
