# Flags

Flags can be passed on the command line to control the behaviour of the application, their values can also be set from environment variables or a configuration file.

On the command line flags are specified with the syntax `--flag=value` and can optionally support a shorthand version such as `-f value`.

## Flag Source Precedence

When resolving the value of a flag, the following sources are checked in order:

1. Command line arguments
2. Environment variables
3. Configuration files
4. Default values

The first non-empty value found in this order is used.

## Defining Flags

Flags can be defined in the command struct using the `Flags` field. Each flag is represented by a `Flag` struct, which includes the flag name, default value, and description.

```go
var myCommand = &cli.Command{
  Name:    "mycommand",
  Usage:   "This is my command",
  Flags: []cli.Flag{
    cli.StringFlag{
      Name:  "config",
      Value: "config.yaml",
      Usage: "Path to the config file",
    },
    cli.BoolFlag{
      Name:  "verbose",
      Usage: "Enable verbose output",
    },
  },
}
```

### Global Flags

By default flags only apply to the command that they are defined against, subcommands don't inherit the flags. However setting `Global: true` on a flag will make it available to all subcommands.

### Hidden Flags

In some cases it may be desirable to hide a flag from the help text or command line usage. This can be achieved by setting the `Hidden: true` field on the flag.

## Parsing Flags

Flags are automatically parsed from the command line arguments when the command is executed. The parsed flag values can be accessed using the `Get*` methods on the `cli.Command` instance.

```go
func (c *Command) Run(ctx context.Context, cmd *Command) error {
    config := cmd.GetString("config")
    verbose := cmd.GetBool("verbose")

    // Use the flag values
    return nil
}
```

### Flag Types

The CLI library supports the following flag types:

| Flag Type        | Go Type       | Getter Method             |
|------------------|---------------|---------------------------|
| StringFlag       | `string`      | `GetString(name)`         |
| IntFlag          | `int`         | `GetInt(name)`            |
| Int8Flag         | `int8`        | `GetInt8(name)`           |
| Int16Flag        | `int16`       | `GetInt16(name)`          |
| Int32Flag        | `int32`       | `GetInt32(name)`          |
| Int64Flag        | `int64`       | `GetInt64(name)`          |
| UintFlag         | `uint`        | `GetUint(name)`           |
| Uint8Flag        | `uint8`       | `GetUint8(name)`          |
| Uint16Flag       | `uint16`      | `GetUint16(name)`         |
| Uint32Flag       | `uint32`      | `GetUint32(name)`         |
| Uint64Flag       | `uint64`      | `GetUint64(name)`         |
| Float32Flag      | `float32`     | `GetFloat32(name)`        |
| Float64Flag      | `float64`     | `GetFloat64(name)`        |
| BoolFlag         | `bool`        | `GetBool(name)`           |
| StringSliceFlag  | `[]string`    | `GetStringSlice(name)`    |
| IntSliceFlag     | `[]int`       | `GetIntSlice(name)`       |
| Int8SliceFlag    | `[]int8`      | `GetInt8Slice(name)`      |
| Int16SliceFlag   | `[]int16`     | `GetInt16Slice(name)`     |
| Int32SliceFlag   | `[]int32`     | `GetInt32Slice(name)`     |
| Int64SliceFlag   | `[]int64`     | `GetInt64Slice(name)`     |
| UintSliceFlag    | `[]uint`      | `GetUintSlice(name)`      |
| Uint8SliceFlag   | `[]uint8`     | `GetUint8Slice(name)`     |
| Uint16SliceFlag  | `[]uint16`    | `GetUint16Slice(name)`    |
| Uint32SliceFlag  | `[]uint32`    | `GetUint32Slice(name)`    |
| Uint64SliceFlag  | `[]uint64`    | `GetUint64Slice(name)`    |
| Float32SliceFlag | `[]float32`   | `GetFloat32Slice(name)`   |
| Float64SliceFlag | `[]float64`   | `GetFloat64Slice(name)`   |

## Flag Validation

Flags can be validated using the `ValidateFlag` method. This method is called on each flag once all flags have been processed.

```go
var myCommand = &cli.Command{
  Name:    "mycommand",
  Usage:   "This is my command",
  Flags: []cli.Flag{
    cli.StringFlag{
      Name:  "config",
      Value: "config.yaml",
      Usage: "Path to the config file",
    },
    cli.IntFlag{
      Name:  "age",
      Usage: "User's age",
      ValidateFlag: (c *cli.Command) error {
        if c.GetInt("age") < 0 {
          return fmt.Errorf("age must be positive")
        }
        return nil
      },
    },
  },
}
```

From the validator it's possible to query the values of other flags so that complex validations can be performed. However the values of named arguments are not available.

