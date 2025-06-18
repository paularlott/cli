package cli

import (
	"fmt"
	"strings"
)

// POSIX flag parser
func (c *Command) parseFlags(args []string) ([]string, error) {
	var parsed = make(map[string]interface{})
	var remainingArgs []string

	// Create lookup maps for flags
	shortFlags := make(map[string]Flag)
	longFlags := make(map[string]Flag)

	for _, flag := range c.Flags {
		flag.register(longFlags, shortFlags)
	}

	for _, flag := range c.globalFlags {
		flag.register(longFlags, shortFlags)
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		// End of flags marker
		if arg == "--" {
			remainingArgs = append(remainingArgs, args[i+1:]...)
			break
		}

		// Long flag (--flag or --flag=value)
		if strings.HasPrefix(arg, "--") {
			flagName := arg[2:]
			var value string
			hasValue := false

			// Check for --flag=value format
			if eq := strings.Index(flagName, "="); eq != -1 {
				value = flagName[eq+1:]
				flagName = flagName[:eq]
				hasValue = true
			}

			flag, exists := longFlags[flagName]
			if !exists {
				return remainingArgs, fmt.Errorf("unknown flag: --%s", flagName)
			}

			if err := c.parseFlag(flag, value, hasValue, args, &i, parsed); err != nil {
				return remainingArgs, err
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Short flag(s) (-f or -abc for bundled flags)
			flagChars := arg[1:]

			// Handle bundled short flags
			for j, char := range flagChars {
				flagName := string(char)
				flag, exists := shortFlags[flagName]
				if !exists {
					return remainingArgs, fmt.Errorf("unknown flag: -%s", flagName)
				}

				// For bundled flags, only the last one can take a value
				isLast := j == len(flagChars)-1
				var value string
				hasValue := false

				// If this is the last flag in a bundle and it's not a bool, check for value
				if isLast {
					switch flag.(type) {
					case *BoolFlag:
						// Bool flags don't need values
					default:
						// Check if next arg is a value
						if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
							value = args[i+1]
							hasValue = true
							i++ // consume the value
						}
					}
				}

				if err := c.parseFlag(flag, value, hasValue, args, &i, parsed); err != nil {
					return remainingArgs, err
				}
			}
		} else {
			// Regular argument
			remainingArgs = append(remainingArgs, arg)
		}

		i++
	}

	c.parsedFlags = parsed

	return remainingArgs, nil
}

func (c *Command) parseFlag(flag Flag, value string, hasValue bool, args []string, i *int, parsed map[string]interface{}) error {
	if _, ok := flag.(*BoolFlag); !ok {
		if !hasValue {
			if *i+1 >= len(args) || strings.HasPrefix(args[*i+1], "-") {
				return fmt.Errorf("flag --%s requires a value", flag.getName())
			}
			value = args[*i+1]
			*i++
		}
	}

	return flag.parseString(value, hasValue, parsed)
}
