# Commands

You can define commands and subcommands by using the `Commands` field in the `Command` struct. Each command is an instance of the `Command` struct, which can include properties such as `Name`, `Usage`, `Description`, `Flags`, and nested subcommands. This structure allows you to organize your CLI application hierarchically, making it easier to manage complex command trees.

There is only one top-level command which is the root command, it can have functionality if so desired.

```go
func main() {
  var myCommand = &cli.Command{
    Name:    "mycommand",
    Usage:   "This is my command",
    Commands: []*cli.Command{
      {
        Name:    "subcommand",
        Usage:   "This is a subcommand",
        Run: func(ctx context.Context, cmd *cli.Command) error {
          fmt.Println("Running subcommand")
          return nil
        },
      },
    },
  }

  err := myCommand.Execute(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
```

## Builtin Commands

The CLI package includes a set of built-in commands that are always available. These commands provide basic functionality and can be disabled if required.

### Automatic Help

By default usage guidance is automatically generated for each command and subcommand. This includes information about the command's flags and arguments. The help text is displayed when the user invokes the command with the `-h` or `--help` flag, alternativly if there's no `Run` function defined for a command the help guide will be displayed.

The help can be disabled by setting `DisableHelp: true` field on the root command.

### Version Display

As part of the default functionality, the version information is displayed when the user invokes the command with the `-v` or `--version` flag.

The version display can be disabled by setting the `DisableVersion: true` field on the root command or by not providing a version string.

## Command Actions

Command actions are the core functionality of each command. They are defined by the `Run` field within the `Command` struct. This function is executed when the command is invoked, and it receives the command context and the command instance as parameters.

### PreRun Actions

If present `PreRun` actions are executed before the `Run` function. The `PreRun` is inherited by subcommands, if subcommands define their own `PreRun` action then the one closest to the command being executed is used.

The command object passed to the `PreRun` function is the same as the one passed to the `Run` function.

### PostRun Actions

If present `PostRun` actions are executed after the `Run` function. The `PostRun` is inherited by subcommands, if subcommands define their own `PostRun` action then the one closest to the command being executed is used.

The command object passed to the `PostRun` function is the same as the one passed to the `Run` function.

## Command Suggestions

Command suggestions are disabled by default but can be enabled by setting `Suggestions: true` on the root command. Once enabled a typo in a command name will generate suggestions for similar commands.

```go
func main() {
  var myCommand = &cli.Command{
    Name:        "mycommand",
    Usage:       "This is my command",
    Suggestions: true,
    Commands: []*cli.Command{
      {
        Name:    "greet",
        Usage:   "This is a subcommand",
        Run: func(ctx context.Context, cmd *cli.Command) error {
          fmt.Println("Running subcommand")
          return nil
        },
      },
    },
  }

  err := myCommand.Execute(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
```

In this example attempting to run `mycommand gree` will generate a suggestion for the `greet` subcommand.
