package cli

import (
	"strings"
)

// preprocessArgs reorders arguments according to the flag positioning rules:
// This function collects all flags (and their values) and moves them to after commands
// but before positional arguments, maintaining order for precedence
func (c *Command) preprocessArgs(args []string) ([]string, error) {
	var commands []string
	var flags []string
	var positionalArgs []string

	i := 0

	for i < len(args) {
		arg := args[i]

		// Check for flag terminator
		if arg == "--" {
			// Everything after -- is positional
			positionalArgs = append(positionalArgs, args[i:]...)
			break
		}

		// Check if it's a flag
		if strings.HasPrefix(arg, "-") {
			// Collect the flag
			if strings.HasPrefix(arg, "--") {
				// Long form
				flagName := arg[2:]

				// Check if it has an inline value (--flag=value)
				if strings.Contains(flagName, "=") {
					flags = append(flags, arg)
					i++
					continue
				}

				// Flag without inline value
				flags = append(flags, arg)

				// Determine if we need to consume next arg as value
				// We need to look up the flag definition
				eqIdx := strings.Index(flagName, "=")
				if eqIdx == -1 {
					flagNameOnly := flagName
					flagObj := c.lookupFlagInTree(flagNameOnly, commands)
					if flagObj != nil {
						// Check if it's a bool flag
						if _, isBool := flagObj.(*BoolFlag); !isBool {
							// Non-bool flag needs a value
							if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && args[i+1] != "--" {
								flags = append(flags, args[i+1])
								i++
							}
						}
					}
				}
			} else {
				// Short form (e.g., -f or -abc)
				flagChars := arg[1:]
				flags = append(flags, arg)

				// For bundled short flags, only the last one can have a value
				if len(flagChars) > 0 {
					lastChar := string(flagChars[len(flagChars)-1])
					flagObj := c.lookupFlagInTree(lastChar, commands)
					if flagObj != nil {
						if _, isBool := flagObj.(*BoolFlag); !isBool {
							// Non-bool flag needs a value
							if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && args[i+1] != "--" {
								flags = append(flags, args[i+1])
								i++
							}
						}
					}
				}
			}
		} else {
			// Not a flag - check if it's a command
			if c.isKnownCommand(arg, commands) {
				commands = append(commands, arg)
			} else {
				// It's a positional argument
				positionalArgs = append(positionalArgs, arg)
			}
		}

		i++
	}

	// Reconstruct: commands + flags + positional args
	result := make([]string, 0, len(commands)+len(flags)+len(positionalArgs))
	result = append(result, commands...)
	result = append(result, flags...)
	result = append(result, positionalArgs...)

	return result, nil
}

// lookupFlagInTree finds a flag definition given the current command path
func (c *Command) lookupFlagInTree(flagName string, commandPath []string) Flag {
	// Navigate to the current command level
	current := c
	allFlags := make([]Flag, 0)

	// Collect flags from root
	allFlags = append(allFlags, current.Flags...)

	// Navigate through command path and collect flags
	for _, cmdName := range commandPath {
		found := false
		for _, subcmd := range current.Commands {
			if subcmd.Name == cmdName {
				current = subcmd
				allFlags = append(allFlags, current.Flags...)
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// Search for the flag by name or alias
	for _, flag := range allFlags {
		if flag.getName() == flagName {
			return flag
		}
		for _, alias := range flag.getAliases() {
			if alias == flagName {
				return flag
			}
		}
	}

	return nil
}

// isKnownCommand checks if an argument is a known subcommand given the current command path
func (c *Command) isKnownCommand(arg string, currentPath []string) bool {
	// Navigate to the current command level
	current := c
	for _, cmdName := range currentPath {
		found := false
		for _, subcmd := range current.Commands {
			if subcmd.Name == cmdName {
				current = subcmd
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check if arg is a subcommand at this level
	for _, subcmd := range current.Commands {
		if subcmd.Name == arg {
			return true
		}
	}

	return false
}
