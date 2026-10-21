package cli

// FlagSource says where a flag's value came from.
type FlagSource int

const (
	SourceNone        FlagSource = iota // The flag has no value: not given anywhere and no default
	SourceDefault                       // The flag's DefaultValue
	SourceConfigFile                    // A ConfigPath in the configuration file
	SourceEnv                           // One of the flag's EnvVars
	SourceCommandLine                   // The command line
)

func (s FlagSource) String() string {
	switch s {
	case SourceDefault:
		return "default"
	case SourceConfigFile:
		return "config file"
	case SourceEnv:
		return "environment"
	case SourceCommandLine:
		return "command line"
	}
	return "none"
}

// FlagSource returns where the named flag's value came from. Unlike HasFlag, which is true for a value from any
// source but the default, it tells a flag given on the command line from one set by an environment variable or the
// configuration file. Aliases are resolved to the flag's name.
func (c *Command) FlagSource(name string) FlagSource {
	if s, ok := c.flagSources[c.flagName(name)]; ok {
		return s
	}
	return SourceNone
}

// flagName resolves an alias of one of the command's flags (or the global flags) to the flag's name.
func (c *Command) flagName(name string) string {
	for _, list := range [][]Flag{c.Flags, c.globalFlags} {
		for _, f := range list {
			if f.getName() == name {
				return name
			}
			if matchesAlias(name, f.getAliases()) {
				return f.getName()
			}
		}
	}
	return name
}
