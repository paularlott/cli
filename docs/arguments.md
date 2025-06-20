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

- `StringArg`
- `IntArg`
- `Int8Arg`
- `Int16Arg`
- `Int32Arg`
- `Int64Arg`
- `UintArg`
- `Uint8Arg`
- `Uint16Arg`
- `Uint32Arg`
- `Uint64Arg`
- `Float32Arg`
- `Float64Arg`
- `BoolArg`

## Positional Arguments

Positional arguments are the arguments left over after named arguments have been processed.

The commands `MinArgs` and `MaxArgs` can be used to specify the minimum and maximum number of positional arguments that a command can accept.

If `MaxArgs` is set to `cli.NoArgs` then the command will not accept any additional arguments and will return an error if any are provided.

If `MinArgs` is set to `cli.UnlimitedArgs` then the command will accept any number of additional arguments.

The positional arguments can be fetched as a string slice using `GetArgs()` on the `cli.Command` instance.

```go
args := cmd.GetArgs()
```
