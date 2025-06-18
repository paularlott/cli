package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
)

const (
	NoArgs        = 0  // No unnamed arguments allowed
	UnlimitedArgs = -1 // Unlimited unnamed arguments allowed
)

type Command struct {
	Name           string                                                           // Name of the command, e.g. "server", "config", etc.
	Version        string                                                           // Version of the command, e.g. "1.0.0"
	Usage          string                                                           // Short description of the command, e.g. "Start the server", "Show config", etc.
	Description    string                                                           // Longer description of the command, e.g. "This command starts the server with the given configuration", "This command shows the current configuration", etc.
	Flags          []Flag                                                           // Flags that are available for this command only
	Arguments      []Argument                                                       // Arguments that can be passed to this command, e.g. "server start <config-file>", "config show <section>", etc.
	MaxArgs        int                                                              // Maximum number of unnamed arguments that are allowed for this command e.g. 0 for no arguments, -1 for unlimited, or a specific number like 2 for "server start <config-file> <port>
	ConfigFile     ConfigFileSource                                                 // Configuration file reader.
	Commands       []*Command                                                       // Subcommands that can be executed under this command, e.g. "server start", "server stop", etc.
	Run            func(ctx context.Context, cmd *Command) error                    // Function to run when this command is executed, e.g. to start the server, show the config, etc.
	PreRun         func(ctx context.Context, cmd *Command) (context.Context, error) // Function to run before command is executed, e.g. to set up logging, read config files, etc.
	PostRun        func(ctx context.Context, cmd *Command) error                    // Function to run after command is executed, e.g. to clean up resources, log the result, etc.
	GlobalPreRun   func(ctx context.Context, cmd *Command) (context.Context, error) // Function to run before any command or subcommand is executed, only the version closest to the command being executed is used.
	GlobalPostRun  func(ctx context.Context, cmd *Command) error                    // Function to run after any command or subcommand is executed, the version from the command supplying GlobalPreRun is used.
	DisableHelp    bool                                                             // Disable the automatic help command for this command
	DisableVersion bool                                                             // Disable the automatic version command for this command
	Suggestions    bool                                                             // Enable suggestions for unknown commands, if true then the command will try to suggest similar commands if the command is not found
	parsedFlags    map[string]interface{}                                           // Parsed flags for this command
	parsedArgs     map[string]interface{}                                           // Parsed arguments for this command
	givenFlags     map[string]bool                                                  // Flags that were given and not defaulted
	remainingArgs  []string                                                         // Remaining arguments after parsing flags and subcommands
	globalFlags    []Flag                                                           // Global flags that are available for this command and all subcommands
	commandChain   []*Command                                                       // Tack the command chain to the active command
}

func (c *Command) Execute(ctx context.Context) error {
	args := os.Args
	if len(args) > 0 {
		args = args[1:]
	}

	// Match subcommands
	remainingArgs, matchedCommand, commandSequence, suggestions := c.matchSubcommands(args)
	matchedCommand.commandChain = commandSequence

	// Inject help and version flags
	if !matchedCommand.DisableHelp {
		matchedCommand.Flags = append(matchedCommand.Flags, &BoolFlag{
			Name:         "help",
			Aliases:      []string{"h"},
			Usage:        "Show help for the command",
			DefaultValue: true,
			Global:       true,
			HideDefault:  true,
			HideType:     true,
		})
	}

	if c == matchedCommand && !c.DisableVersion {
		matchedCommand.Flags = append(matchedCommand.Flags, &BoolFlag{
			Name:         "version",
			Aliases:      []string{"v"},
			Usage:        "Show version information",
			DefaultValue: true,
			HideDefault:  true,
			HideType:     true,
		})
	}

	// Parse the command line flags first
	remainingArgs, err := matchedCommand.parseFlags(remainingArgs)
	if err != nil {
		return err
	}

	// Merge the global and command flags
	combinedFlags := make([]Flag, 0, len(matchedCommand.globalFlags)+len(matchedCommand.Flags))
	combinedFlags = append(combinedFlags, matchedCommand.globalFlags...)
	combinedFlags = append(combinedFlags, matchedCommand.Flags...)

	// For flags that are not set on the command line see if they can be set from an environment variable
	for _, flag := range combinedFlags {
		if _, ok := matchedCommand.parsedFlags[flag.getName()]; !ok {
			flag.setFromEnvVar(matchedCommand.parsedFlags)
		}
	}

	// For flags that are still not set, check if they can be set from a config
	if c.ConfigFile != nil {
		for _, flag := range combinedFlags {
			if _, ok := matchedCommand.parsedFlags[flag.getName()]; !ok {
				cfgPaths := flag.configPaths()
				if len(cfgPaths) > 0 {
					for _, path := range cfgPaths {
						if v, ok := c.ConfigFile.Value(path); ok {
							isSlice := reflect.TypeOf(v).Kind() == reflect.Slice
							if isSlice == flag.isSlice() {
								if isSlice {
									switch vals := v.(type) {
									case []interface{}:
										for _, val := range vals {
											flag.parseString(fmt.Sprintf("%v", val), true, matchedCommand.parsedFlags)
										}
									case []string:
										for _, val := range vals {
											flag.parseString(fmt.Sprintf("%v", val), true, matchedCommand.parsedFlags)
										}
									default:
									}
								} else {
									flag.parseString(fmt.Sprintf("%v", v), true, matchedCommand.parsedFlags)
								}
							}
						}
					}
				}
			}
		}
	}

	// For flags that are not set, set the default values
	matchedCommand.givenFlags = make(map[string]bool)
	for _, flag := range combinedFlags {
		if _, ok := matchedCommand.parsedFlags[flag.getName()]; !ok {
			flag.setFromDefault(matchedCommand.parsedFlags)
		} else {
			matchedCommand.givenFlags[flag.getName()] = true
		}
	}

	// Are we showing version information
	if !matchedCommand.DisableVersion && matchedCommand.HasFlag("version") {
		fmt.Printf("%s version %s\n", matchedCommand.Name, matchedCommand.Version)
		return nil
	}

	// Are we showing help
	if !matchedCommand.DisableHelp && matchedCommand.HasFlag("help") {
		matchedCommand.ShowHelp()
		return nil
	}

	// Check if we have suggestions for a failed command match
	if c.Suggestions && len(suggestions) > 0 && matchedCommand == c && len(remainingArgs) > 0 {
		c.displaySuggestions(suggestions, remainingArgs)
		return fmt.Errorf("unknown command")
	}

	// Check required flags are present
	for _, flag := range combinedFlags {
		if _, ok := matchedCommand.parsedFlags[flag.getName()]; !ok {
			if flag.isRequired() {
				return fmt.Errorf("required flag '%s' not set", flag.getName())
			}
		}
	}

	// Parse named arguments
	matchedCommand.remainingArgs, err = matchedCommand.parseArgs(remainingArgs)
	if err != nil {
		return err
	}

	// Check the limits on the number of unnamed arguments
	if matchedCommand.MaxArgs != UnlimitedArgs && len(matchedCommand.remainingArgs) > matchedCommand.MaxArgs {
		return fmt.Errorf("too many arguments")
	}

	// From the command look back towards the root for the first command that implements GlobalPreRun
	var globalPreRunCmd *Command = nil
	for i := len(commandSequence) - 1; i >= 0; i-- {
		if commandSequence[i].GlobalPreRun != nil {
			globalPreRunCmd = commandSequence[i]
			break
		}
	}

	// Execute, GlobalPreRun, PreRun, Run, PostRun and GlobalPostRun
	var globalPreErr error = nil
	var preErr error = nil
	var runErr error = nil
	var postErr error = nil
	var globalPostErr error = nil
	if globalPreRunCmd != nil {
		ctx, globalPreErr = globalPreRunCmd.GlobalPreRun(ctx, globalPreRunCmd)
	}
	if globalPreErr == nil && matchedCommand.PreRun != nil {
		ctx, preErr = matchedCommand.PreRun(ctx, matchedCommand)
	}
	if preErr == nil && globalPreErr == nil {
		if matchedCommand.Run != nil {
			runErr = matchedCommand.Run(ctx, matchedCommand)
		} else {
			if matchedCommand.DisableHelp {
				// Handle the case when no Run is defined and help is disabled
				if c.Suggestions && len(remainingArgs) > 0 {
					suggestions := c.findSimilarCommands(remainingArgs[0], matchedCommand.Commands, 2)
					if len(suggestions) > 0 {
						c.displaySuggestions(suggestions, remainingArgs)
					} else {
						fmt.Printf("Unknown command: %s\n", remainingArgs[0])
					}
				}
				runErr = fmt.Errorf("unknown command")
			} else {
				matchedCommand.ShowHelp()
			}
		}
	}
	if preErr == nil && matchedCommand.PostRun != nil {
		postErr = matchedCommand.PostRun(ctx, matchedCommand)
	}
	if globalPreRunCmd != nil && globalPreErr == nil {
		globalPostErr = globalPreRunCmd.GlobalPostRun(ctx, globalPreRunCmd)
	}

	return errors.Join(globalPreErr, preErr, runErr, postErr, globalPostErr)
}

// matchSubcommands walks through args to find the deepest matching subcommand
func (c *Command) matchSubcommands(args []string) ([]string, *Command, []*Command, []string) {
	current := c
	remaining := args
	globalFlags := []Flag{}
	commandSequence := []*Command{c}
	var suggestions []string

	for len(remaining) > 0 && len(current.Commands) > 0 {
		found := false
		for _, subcmd := range current.Commands {
			if remaining[0] == subcmd.Name {
				// Save the global flags or the parent command
				for _, flag := range current.Flags {
					if flag.isGlobal() {
						globalFlags = append(globalFlags, flag)
					}
				}

				current = subcmd
				remaining = remaining[1:]
				found = true
				commandSequence = append(commandSequence, subcmd)
				break
			}
		}
		if !found {
			if len(remaining) > 0 {
				suggestions = c.findSimilarCommands(remaining[0], current.Commands, 2)
			}
			break
		}
	}

	current.globalFlags = globalFlags
	return remaining, current, commandSequence, suggestions
}

// Args returns the list of arguments that were passed to the command and have not been consumed by subcommands, flags and arguments
func (c *Command) GetArgs() []string {
	return c.remainingArgs
}

// HasFlag checks if a flag with the given name was set for this command
func (c *Command) HasFlag(name string) bool {
	_, ok := c.givenFlags[name]
	return ok
}

func (c *Command) HasArg(name string) bool {
	_, ok := c.parsedArgs[name]
	return ok
}

func (c *Command) GetRootCmd() *Command {
	if len(c.commandChain) > 0 {
		return c.commandChain[0]
	}
	return c
}
