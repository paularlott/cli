# Arguments

The CLI library supports both named and positional arguments.

## Named Arguments

Named arguments are specified using the `Argument` field on the `cli.Command` definition.

```go
var ageValue int

cmd := &cli.Command{
  Arguments: []cli.Argument{
    &cli.StringArg{
      Name:     "name",
      Usage:    "Your name",
      Required: true,
    },
    &cli.IntArg{
      Name:  "age",
      Usage: "Your age",
      AssignTo: &ageValue,
    },
  },
}
```

Named arguments can optionally be assigned to a variable using the `AssignTo` field.

Marking an argument as required by setting the `Required` field to `true` will ensure that the user provides a value for that argument, otherwise an error will be returned.

The value of the arguments can be accessed with the `Get*Arg` methods on the `cli.Command` instance.

```go
name := cmd.GetString("name")
age := cmd.GetInt("age")
```

In the case of age it's value will also be available in the variable `ageValue`.

### Argument Types

The CLI library supports the following argument types:

| Arg Type    | Go Type    | Getters                   |
|-------------|------------|---------------------------|
| StringArg   | `string`   | `GetStringArg(name)`      |
| IntArg      | `int`      | `GetIntArg(name)`         |
| Int8Arg     | `int8`     | `GetInt8Arg(name)`        |
| Int16Arg    | `int16`    | `GetInt16Arg(name)`       |
| Int32Arg    | `int32`    | `GetInt32Arg(name)`       |
| Int64Arg    | `int64`    | `GetInt64Arg(name)`       |
| UintArg     | `uint`     | `GetUintArg(name)`        |
| Uint8Arg    | `uint8`    | `GetUint8Arg(name)`       |
| Uint16Arg   | `uint16`   | `GetUint16Arg(name)`      |
| Uint32Arg   | `uint32`   | `GetUint32Arg(name)`      |
| Uint64Arg   | `uint64`   | `GetUint64Arg(name)`      |
| Float32Arg  | `float32`  | `GetFloat32Arg(name)`     |
| Float64Arg  | `float64`  | `GetFloat64Arg(name)`     |
| BoolArg     | `bool`     | `GetBoolArg(name)`        |

## Positional Arguments

Positional arguments are the arguments left over after named arguments have been processed.

The commands `MinArgs` and `MaxArgs` can be used to specify the minimum and maximum number of positional arguments that a command can accept.

If `MaxArgs` is set to `cli.NoArgs` then the command will not accept any additional arguments and will return an error if any are provided.

If `MinArgs` is set to `cli.UnlimitedArgs` then the command will accept any number of additional arguments.

The positional arguments can be fetched as a string slice using `GetArgs()` on the `cli.Command` instance.

```go
args := cmd.GetArgs()
```

## Flag Validation

Flags can be validated using the `ValidateArg` method. This method is called on each argument once all named arguments have been processed.

```go
cmd := &cli.Command{
  Arguments: []cli.Argument{
    &cli.IntArg{
      Name:  "age",
      Usage: "Your age",
      ValidateArg: func(c *Command) error {
        if c.GetInt("age") < 0 {
          return fmt.Errorf("age must be positive")
        }
        return nil
      },
    },
  },
}
```

From the validator it's possible to query the values of flags and other named arguments so that complex validations can be performed.
