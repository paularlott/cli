package cli

import (
	"fmt"
	"strconv"
)

func (c *Command) parseArgs(args []string) ([]string, error) {
	c.parsedArgs = make(map[string]interface{})

	// Parse the arguments
	for _, arg := range c.Arguments {
		if len(args) == 0 {
			if arg.isRequired() {
				return args, fmt.Errorf("missing required argument: %s", arg.name())
			}
			break // No more args to parse
		}

		// Get the next argument
		value := args[0]
		args = args[1:]

		switch arg := arg.(type) {
		case *StringArg:
			c.parsedArgs[arg.name()] = value
			if arg.AssignTo != nil {
				*arg.AssignTo = value
			}
		case *IntArg:
			intVal, err := strconv.Atoi(value)
			if err != nil {
				return args, fmt.Errorf("invalid integer value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = intVal
			if arg.AssignTo != nil {
				*arg.AssignTo = intVal
			}
		case *Int8Arg:
			int8Val, err := strconv.ParseInt(value, 10, 8)
			if err != nil {
				return args, fmt.Errorf("invalid int8 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = int8(int8Val)
			if arg.AssignTo != nil {
				*arg.AssignTo = int8(int8Val)
			}
		case *Int16Arg:
			int16Val, err := strconv.ParseInt(value, 10, 16)
			if err != nil {
				return args, fmt.Errorf("invalid int16 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = int16(int16Val)
			if arg.AssignTo != nil {
				*arg.AssignTo = int16(int16Val)
			}
		case *Int32Arg:
			int32Val, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return args, fmt.Errorf("invalid int32 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = int32(int32Val)
			if arg.AssignTo != nil {
				*arg.AssignTo = int32(int32Val)
			}
		case *Int64Arg:
			int64Val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return args, fmt.Errorf("invalid int64 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = int64Val
			if arg.AssignTo != nil {
				*arg.AssignTo = int64Val
			}
		case *UintArg:
			uintVal, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return args, fmt.Errorf("invalid uint value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = uint(uintVal)
			if arg.AssignTo != nil {
				*arg.AssignTo = uint(uintVal)
			}
		case *Uint8Arg:
			uint8Val, err := strconv.ParseUint(value, 10, 8)
			if err != nil {
				return args, fmt.Errorf("invalid uint8 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = uint8(uint8Val)
			if arg.AssignTo != nil {
				*arg.AssignTo = uint8(uint8Val)
			}
		case *Uint16Arg:
			uint16Val, err := strconv.ParseUint(value, 10, 16)
			if err != nil {
				return args, fmt.Errorf("invalid uint16 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = uint16(uint16Val)
			if arg.AssignTo != nil {
				*arg.AssignTo = uint16(uint16Val)
			}
		case *Uint32Arg:
			uint32Val, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return args, fmt.Errorf("invalid uint32 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = uint32(uint32Val)
			if arg.AssignTo != nil {
				*arg.AssignTo = uint32(uint32Val)
			}
		case *Uint64Arg:
			uint64Val, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return args, fmt.Errorf("invalid uint64 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = uint64Val
			if arg.AssignTo != nil {
				*arg.AssignTo = uint64Val
			}
		case *Float32Arg:
			float32Val, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return args, fmt.Errorf("invalid float32 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = float32(float32Val)
			if arg.AssignTo != nil {
				*arg.AssignTo = float32(float32Val)
			}
		case *Float64Arg:
			float64Val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return args, fmt.Errorf("invalid float64 value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = float64Val
			if arg.AssignTo != nil {
				*arg.AssignTo = float64Val
			}
		case *BoolArg:
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return args, fmt.Errorf("invalid boolean value for argument %s: %s", arg.name(), value)
			}
			c.parsedArgs[arg.name()] = boolVal
			if arg.AssignTo != nil {
				*arg.AssignTo = boolVal
			}
		}
	}

	return args, nil
}
