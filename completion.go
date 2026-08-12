package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
			&StringFlag{
				Name:   "arg",
				Usage:  "Return dynamic argument completions: 'cmdpath:argindex:partial'",
				Hidden: true,
			},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			shell := cmd.GetStringArg("shell")

			// Handle the dynamic completion mode when called with the completion flags
			if cmd.HasFlag("arg") {
				handleArgCompletion(cmd, shell)
				return nil
			} else if cmd.HasFlag("command") {
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
	binName := execName(rootCmd)
	for _, part := range pathParts {
		if part == "" || part == rootCmd.Name || part == binName {
			continue
		}

		found := false
		for _, subCmd := range current.Commands {
			if subCmd.Name == part || matchesAlias(part, subCmd.Aliases) {
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
			// Also output aliases
			for _, alias := range subCmd.Aliases {
				if subCmd.Usage != "" {
					fmt.Printf("%s\t%s\n", alias, subCmd.Usage)
				} else {
					fmt.Println(alias)
				}
			}

		case "powershell":
			// Powershell uses value:description format
			if subCmd.Usage != "" {
				fmt.Printf("%s:%s\n", subCmd.Name, subCmd.Usage)
			} else {
				fmt.Println(subCmd.Name)
			}
			// Also output aliases
			for _, alias := range subCmd.Aliases {
				if subCmd.Usage != "" {
					fmt.Printf("%s:%s\n", alias, subCmd.Usage)
				} else {
					fmt.Println(alias)
				}
			}

		default:
			// Just need command names
			fmt.Println(subCmd.Name)
			for _, alias := range subCmd.Aliases {
				fmt.Println(alias)
			}
		}
	}
}

// handleArgCompletion calls the CompletionFunc for the argument at the given index.
// The flag value format is "cmdpath:argindex:partial".
func handleArgCompletion(cmd *Command, shell string) {
	raw := cmd.GetString("arg")
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 2 {
		return
	}
	cmdPath := parts[0]
	argIdx, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	partial := ""
	if len(parts) == 3 {
		partial = parts[2]
	}

	rootCmd := cmd.GetRootCmd()
	current := rootCmd
	binName := execName(rootCmd)
	for _, part := range strings.Split(cmdPath, " ") {
		if part == "" || part == rootCmd.Name || part == binName {
			continue
		}
		found := false
		for _, subCmd := range current.Commands {
			if subCmd.Name == part || matchesAlias(part, subCmd.Aliases) {
				current = subCmd
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	if argIdx < 0 || argIdx >= len(current.Arguments) {
		return
	}
	fn := current.Arguments[argIdx].completionFunc()
	if fn == nil {
		return
	}

	items := fn(context.Background(), current)
	if partial != "" {
		filtered := items[:0]
		for _, item := range items {
			if strings.HasPrefix(item.Value, partial) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	printCompletionItems(items, shell)
}

// printCompletionItems outputs completion items in the format expected by each shell.
func printCompletionItems(items []CompletionItem, shell string) {
	for _, item := range items {
		switch shell {
		case "fish":
			if item.Description != "" {
				fmt.Printf("%s\t%s\n", item.Value, item.Description)
			} else {
				fmt.Println(item.Value)
			}
		case "powershell":
			if item.Description != "" {
				fmt.Printf("%s:%s\n", item.Value, item.Description)
			} else {
				fmt.Println(item.Value)
			}
		default:
			fmt.Println(item.Value)
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
	binName := execName(rootCmd)
	for _, part := range pathParts {
		if part == "" || part == rootCmd.Name || part == binName {
			continue
		}

		for _, flag := range current.Flags {
			if flag.isGlobal() && !flag.isHidden() {
				globalFlags = append(globalFlags, flag)
			}
		}

		found := false
		for _, subCmd := range current.Commands {
			if subCmd.Name == part || matchesAlias(part, subCmd.Aliases) {
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
					if flag.getUsage() == "" {
						fmt.Printf("-%s\n", alias)
					} else {
						fmt.Printf("-%s\t%s\n", alias, flag.getUsage())
					}
				} else {
					if flag.getUsage() == "" {
						fmt.Printf("--%s\n", alias)
					} else {
						fmt.Printf("--%s\t%s\n", alias, flag.getUsage())
					}
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
					if flag.getUsage() != "" {
						fmt.Printf("-%s:%s\n", alias, flag.getUsage())
					} else {
						fmt.Printf("-%s\n", alias)
					}
				} else {
					if flag.getUsage() != "" {
						fmt.Printf("--%s:%s\n", alias, flag.getUsage())
					} else {
						fmt.Printf("--%s\n", alias)
					}
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

	binName := execName(rootCmd)
	for _, part := range pathParts {
		if part == "" || part == rootCmd.Name || part == binName {
			continue
		}
		for _, flag := range current.Flags {
			if flag.isGlobal() {
				globalFlags = append(globalFlags, flag)
			}
		}
		found := false
		for _, subCmd := range current.Commands {
			if subCmd.Name == part || matchesAlias(part, subCmd.Aliases) {
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

// execName returns the actual binary name from os.Args[0], falling back to root.Name
func execName(root *Command) string {
	if len(os.Args) > 0 {
		if name := filepath.Base(os.Args[0]); name != "" && name != "." {
			return name
		}
	}
	return root.Name
}

// execPath returns the absolute path of the running binary
func execPath() string {
	if p, err := filepath.Abs(os.Args[0]); err == nil {
		return p
	}
	return os.Args[0]
}

// Generate a dynamic bash completion script
func generateDynamicBashCompletion(w io.Writer, root *Command) error {
	cmdName := execName(root)

	fmt.Fprintf(w, `# bash completion script for the command %[1]s

_%[1]s() {
    local exec_path
    exec_path=$(eval echo "${COMP_WORDS[0]}")

    # Capture the current command line words
    local cmdpath="%[1]s"
    local current_word="${COMP_WORDS[COMP_CWORD]}"
    local completions

    # Get flags that take a value so we can skip their arguments
    local value_flags

    # Build the command path: greedily match subcommands, then count remaining positionals
    local arg_index=0
    local in_args=0
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
                value_flags=$($exec_path completion bash --value-flags="$cmdpath")
                if echo "$value_flags" | grep -qx "$fname"; then
                    skip_next=1
                fi
            elif [[ $in_args -eq 0 ]]; then
                local sub_completions
                sub_completions=$($exec_path completion bash --command="$cmdpath")
                if echo "$sub_completions" | grep -qx "${COMP_WORDS[i]}"; then
                    cmdpath+=" ${COMP_WORDS[i]}"
                else
                    in_args=1
                    arg_index=$((arg_index + 1))
                fi
            else
                arg_index=$((arg_index + 1))
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
        if [[ -z "$completions" ]]; then
            completions=$($exec_path completion bash --arg="$cmdpath:$arg_index:$current_word")
        fi
    fi

    # Filter completions against the current word (use mapfile for safe handling)
    mapfile -t COMPREPLY < <(compgen -W "${completions}" -- "$current_word")
}

# Register the completion function
complete -o bashdefault -o default -o nospace -F _%[1]s %[1]s '%[2]s'
`, cmdName, execPath())

	return nil
}

// generateDynamicZshCompletion writes a zsh completion script that calls back to the program
func generateDynamicZshCompletion(w io.Writer, root *Command) error {
	cmdName := execName(root)

	// Write the function header
	fmt.Fprintf(w, `# zsh completion script for the command %[1]s
autoload -U compinit && compinit

_%[1]s() {
    local exec_path
    exec_path=$(eval echo "${words[1]}")
    local -a suggestions

    # Capture the current command line words
    local cmdpath="%[1]s"
    local current_word="${words[$CURRENT]}"
    local completions

    # Get flags that take a value so we can skip their arguments
    local value_flags

    # Skip command name and build from arguments: greedily match subcommands, then count positionals
    local arg_index=0
    local in_args=0
    if [[ ${#words[@]} -gt 1 ]]; then
        local skip_next=0
        for ((i=2; i<CURRENT; i++)); do
            if [[ $skip_next -eq 1 ]]; then
                skip_next=0
                continue
            fi
            if [[ "${words[i]}" == -* ]]; then
                local fname="${words[i]#--}"
                fname="${fname#-}"
                value_flags=$($exec_path completion zsh --value-flags="$cmdpath")
                if echo "$value_flags" | grep -qx "$fname"; then
                    skip_next=1
                fi
            elif [[ $in_args -eq 0 ]]; then
                local sub_completions=$($exec_path completion zsh --command="$cmdpath")
                if echo "$sub_completions" | grep -qx "${words[i]}"; then
                    cmdpath+=" ${words[i]}"
                else
                    in_args=1
                    arg_index=$((arg_index + 1))
                fi
            else
                arg_index=$((arg_index + 1))
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
        if [[ -z "$completions" ]]; then
            completions=$($exec_path completion zsh --arg="$cmdpath:$arg_index:$current_word")
        fi
    fi

    # Split the output from the command into an array of suggestions
    suggestions=("${(@f)completions}")

    # Add the suggestions to the completion list
    if [[ ${#suggestions[@]} -gt 0 && -n "${suggestions[1]}" ]]; then
        compadd -- "${suggestions[@]}"
    else
        _default
    fi
}

# Register the completion function
compdef _%[1]s %[1]s %[2]s
`, cmdName, execPath())

	return nil
}

// generateDynamicFishCompletion writes a fish completion script that calls back to the program
func generateDynamicFishCompletion(w io.Writer, root *Command) error {
	cmdName := execName(root)

	fmt.Fprintf(w, `# fish completion script for the command %[1]s

function __%[1]s_completion
    set -l cmd_line (commandline -opc)
    set -l exec_path $cmd_line[1]
    set -l current_token (commandline -ct)
    set -l cmd_path "%[1]s"
    set -l arg_index 0

    # Build the command path including all executed subcommands (not the current token)
    # First token is always the exec name which we've already included
    if test (count $cmd_line) -gt 1
        # Skip first token (the command itself)
        set -l cmd_parts $cmd_line[2..-1]

        # Remove the last token if it's incomplete (current token being completed)
        if string match -q -- "*$current_token" $cmd_line[-1] && test -n "$current_token"
            set cmd_parts $cmd_parts[1..-2]
        end

        # Two-pass: first greedily match subcommands, then count remaining positionals as arg_index
        set -l in_args 0
        set -l skip_next 0
        for part in $cmd_parts
            if test $skip_next -eq 1
                set skip_next 0
                continue
            end
            if string match -q -- '-*' $part
                set -l value_flags (eval $exec_path completion fish --value-flags=\"$cmd_path\")
                set -l fname (string replace --regex -- '^--?' '' $part)
                if contains -- $fname $value_flags
                    set skip_next 1
                end
            else if test $in_args -eq 0
                set -l sub_names (eval $exec_path completion fish --command=\"$cmd_path\" | string replace --regex '\t.*' '')
                if contains -- $part $sub_names
                    set cmd_path "$cmd_path $part"
                else
                    set in_args 1
                    set arg_index (math $arg_index + 1)
                end
            else
                set arg_index (math $arg_index + 1)
            end
        end
    end
    set -l flag_path $cmd_path

    # Request completions from the binary
    if string match -q -- '-*' $current_token
        # Flag completion
        eval $exec_path completion fish --flag=\"$flag_path\"
    else
        # Command/subcommand/argument completion
        set -l completions (eval $exec_path completion fish --command=\"$flag_path\")
        if test -z "$completions"
            eval $exec_path completion fish --arg=\"$flag_path:$arg_index:$current_token\"
        else
            string join \n $completions
        end
    end
end

# Register the completion function
complete -c %[1]s -f -a '(__%[1]s_completion)'
`, cmdName)

	return nil
}

// generateDynamicPowershellCompletion generates a PowerShell completion script
func generateDynamicPowershellCompletion(w io.Writer, root *Command) error {
	cmdName := execName(root)

	fmt.Fprintf(w, `# PowerShell completion script for the command %[1]s

Register-ArgumentCompleter -Native -CommandName %[1]s -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # Get the current word
    $currentWord = $wordToComplete

    # Build the command path from all tokens before cursor
    $cmdPath = "%[1]s"
    $tokens = $commandAst.CommandElements
    $execPath = $ExecutionContext.InvokeCommand.ExpandString($tokens[0].ToString())

    # Get flags that take a value so we can skip their arguments
`, cmdName)
	fmt.Fprint(w, `
    # Start at index 1 to skip the command itself
    $skipNext = $false
    $argIndex = 0
    $inArgs = $false
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
`)
	fmt.Fprintln(w, "            $vf = (& $execPath completion powershell --value-flags=\"$cmdPath\" 2>$null) -split \"`n\" | Where-Object { $_ -ne \"\" }")
	fmt.Fprint(w, `
            $fname = $token -replace '^--?', ''
            if ($vf -contains $fname) {
                $skipNext = $true
            }
        } elseif (-not $inArgs) {
`)
	fmt.Fprintln(w, "            $subNames = (& $execPath completion powershell --command=\"$cmdPath\" 2>$null) -split \"`n\" | ForEach-Object { ($_ -split ':')[0].Trim() } | Where-Object { $_ -ne '' }")
	fmt.Fprint(w, `
            if ($subNames -contains $token) {
                $cmdPath += " $token"
            } else {
                $inArgs = $true
                $argIndex++
            }
        } else {
            $argIndex++
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
        if (-not $completions) {
            $completions = & $execPath completion powershell --arg="${cmdPath}:${argIndex}:${currentWord}" 2>$null
        }
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
