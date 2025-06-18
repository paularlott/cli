package cli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Flag interface {
	getName() string
	getAliases() []string
	isGlobal() bool
	register(longFlags, shortFlags map[string]Flag)
	parseString(value string, hasValue bool, parsedFlags map[string]interface{}) error
	setFromEnvVar(parsedFlags map[string]interface{})
	setFromDefault(parsedFlags map[string]interface{})
	configPaths() []string
	isSlice() bool
	isRequired() bool
	isHidden() bool
	flagDefinition() string   // Returns flag name and aliases with type (e.g., --port int, -p int)
	getUsage() string         // Returns description of the flag (e.g., "Port to run the server on")
	defaultValueText() string // Returns formatted default value (e.g., "8080")
	typeText() string         // Returns type information (e.g., "int", "string", etc.)
}

type FlagTyped[T any] struct {
	Name         string   // Name of the flag, e.g. "server"
	Usage        string   // Short description of the flag, e.g. "The server to connect to"
	Aliases      []string // Aliases for the flag, e.g. "s" for "server"
	ConfigPath   []string // Configuration paths for the flag, e.g. "cli.server"
	DefaultValue T        // Default value for the flag, e.g. "localhost" for server
	DefaultText  string   // Text to show in usage as the default value, e.g. "localhost"
	AssignTo     *T       // Optional pointer to the variable where the value should be stored
	EnvVars      []string // Environment variables that can be used to set this flag, first found will be used
	Required     bool     // Whether this flag is required
	Global       bool     // Whether this flag is global, i.e. available in all commands
	HideDefault  bool     // Whether to hide the default value in usage output
	HideType     bool     // Whether to hide the type in usage output
	Hidden       bool     // Whether this flag is hidden from help and command completions
}

type StringFlag = FlagTyped[string]
type IntFlag = FlagTyped[int]
type Int8Flag = FlagTyped[int8]
type Int16Flag = FlagTyped[int16]
type Int32Flag = FlagTyped[int32]
type Int64Flag = FlagTyped[int64]
type UintFlag = FlagTyped[uint]
type Uint8Flag = FlagTyped[uint8]
type Uint16Flag = FlagTyped[uint16]
type Uint32Flag = FlagTyped[uint32]
type Uint64Flag = FlagTyped[uint64]
type Float32Flag = FlagTyped[float32]
type Float64Flag = FlagTyped[float64]
type BoolFlag = FlagTyped[bool]

type StringSliceFlag = FlagTyped[[]string]
type IntSliceFlag = FlagTyped[[]int]
type Int8SliceFlag = FlagTyped[[]int8]
type Int16SliceFlag = FlagTyped[[]int16]
type Int32SliceFlag = FlagTyped[[]int32]
type Int64SliceFlag = FlagTyped[[]int64]
type UintSliceFlag = FlagTyped[[]uint]
type Uint8SliceFlag = FlagTyped[[]uint8]
type Uint16SliceFlag = FlagTyped[[]uint16]
type Uint32SliceFlag = FlagTyped[[]uint32]
type Uint64SliceFlag = FlagTyped[[]uint64]
type Float32SliceFlag = FlagTyped[[]float32]
type Float64SliceFlag = FlagTyped[[]float64]

func (f *FlagTyped[T]) getName() string {
	return f.Name
}

func (f *FlagTyped[T]) getAliases() []string {
	return f.Aliases
}

func (f *FlagTyped[T]) isGlobal() bool {
	return f.Global
}

func (f *FlagTyped[T]) configPaths() []string {
	return f.ConfigPath
}

func (f *FlagTyped[T]) isSlice() bool {
	return reflect.TypeOf(f.DefaultValue).Kind() == reflect.Slice
}

func (f *FlagTyped[T]) isRequired() bool {
	return f.Required
}

func (f *FlagTyped[T]) isHidden() bool {
	return f.Hidden
}

func (f *FlagTyped[T]) register(longFlags, shortFlags map[string]Flag) {
	longFlags[f.Name] = f
	for _, alias := range f.Aliases {
		if len(alias) == 1 {
			shortFlags[alias] = f
		} else {
			longFlags[alias] = f
		}
	}
}

func (f *FlagTyped[T]) setFromEnvVar(parsedFlags map[string]interface{}) {
	if len(f.EnvVars) > 0 {
		for _, envVar := range f.EnvVars {
			if value, ok := os.LookupEnv(envVar); ok {
				// If slice then split by comma
				if f.isSlice() {
					values := strings.Split(value, ",")
					for _, v := range values {
						v = strings.TrimSpace(v)
						f.parseString(v, true, parsedFlags)
					}
				} else {
					f.parseString(value, true, parsedFlags)
				}
				return // Use the first found environment variable
			}
		}
	}
}

func (f *FlagTyped[T]) setFromDefault(parsedFlags map[string]interface{}) {
	zero := reflect.Zero(reflect.TypeOf(f.DefaultValue)).Interface()
	if !reflect.DeepEqual(f.DefaultValue, zero) {
		parsedFlags[f.Name] = f.DefaultValue
		if f.AssignTo != nil {
			*f.AssignTo = f.DefaultValue
		}
	}
}

func (f *FlagTyped[T]) parseString(value string, hasValue bool, parsedFlags map[string]interface{}) error {
	switch f := any(f).(type) {
	case *StringFlag:
		parsedFlags[f.Name] = value
		if f.AssignTo != nil {
			*f.AssignTo = value
		}

	case *IntFlag:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = intVal
		if f.AssignTo != nil {
			*f.AssignTo = intVal
		}

	case *Int8Flag:
		int8Val, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return fmt.Errorf("invalid int8 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = int8(int8Val)
		if f.AssignTo != nil {
			*f.AssignTo = int8(int8Val)
		}

	case *Int16Flag:
		int16Val, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid int16 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = int16(int16Val)
		if f.AssignTo != nil {
			*f.AssignTo = int16(int16Val)
		}

	case *Int32Flag:
		int32Val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid int32 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = int32(int32Val)
		if f.AssignTo != nil {
			*f.AssignTo = int32(int32Val)
		}

	case *Int64Flag:
		int64Val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int64 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = int64Val
		if f.AssignTo != nil {
			*f.AssignTo = int64Val
		}

	case *UintFlag:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = uintVal
		if f.AssignTo != nil {
			*f.AssignTo = uint(uintVal)
		}

	case *Uint8Flag:
		uint8Val, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return fmt.Errorf("invalid uint8 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = uint8(uint8Val)
		if f.AssignTo != nil {
			*f.AssignTo = uint8(uint8Val)
		}

	case *Uint16Flag:
		uint16Val, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid uint16 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = uint16(uint16Val)
		if f.AssignTo != nil {
			*f.AssignTo = uint16(uint16Val)
		}

	case *Uint32Flag:
		uint32Val, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid uint32 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = uint32(uint32Val)
		if f.AssignTo != nil {
			*f.AssignTo = uint32(uint32Val)
		}

	case *Uint64Flag:
		uint64Val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint64 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = uint64Val
		if f.AssignTo != nil {
			*f.AssignTo = uint64Val
		}

	case *Float32Flag:
		float32Val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return fmt.Errorf("invalid float32 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = float32(float32Val)
		if f.AssignTo != nil {
			*f.AssignTo = float32(float32Val)
		}

	case *Float64Flag:
		float64Val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float64 value for flag --%s: %s", f.Name, value)
		}
		parsedFlags[f.Name] = float64Val
		if f.AssignTo != nil {
			*f.AssignTo = float64Val
		}

	case *BoolFlag:
		if hasValue {
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value for flag --%s: %s", f.Name, value)
			}
			parsedFlags[f.Name] = boolVal
			if f.AssignTo != nil {
				*f.AssignTo = boolVal
			}
		} else {
			parsedFlags[f.Name] = true
			if f.AssignTo != nil {
				*f.AssignTo = true
			}
		}

	case *StringSliceFlag:
		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]string), value)
		} else {
			parsedFlags[f.Name] = []string{value}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]string)
		}

	case *IntSliceFlag:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]int), intVal)
		} else {
			parsedFlags[f.Name] = []int{intVal}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]int)
		}

	case *Int8SliceFlag:
		int8Val, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return fmt.Errorf("invalid int8 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]int8), int8(int8Val))
		} else {
			parsedFlags[f.Name] = []int8{int8(int8Val)}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]int8)
		}

	case *Int16SliceFlag:
		int16Val, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid int16 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]int16), int16(int16Val))
		} else {
			parsedFlags[f.Name] = []int16{int16(int16Val)}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]int16)
		}

	case *Int32SliceFlag:
		int32Val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid int32 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]int32), int32(int32Val))
		} else {
			parsedFlags[f.Name] = []int32{int32(int32Val)}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]int32)
		}

	case *Int64SliceFlag:
		int64Val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int64 value for flag --%s: %s", f.Name, value)
		}
		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]int64), int64Val)
		} else {
			parsedFlags[f.Name] = []int64{int64Val}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]int64)
		}

	case *UintSliceFlag:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]uint), uint(uintVal))
		} else {
			parsedFlags[f.Name] = []uint{uint(uintVal)}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]uint)
		}

	case *Uint8SliceFlag:
		uint8Val, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return fmt.Errorf("invalid uint8 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]uint8), uint8(uint8Val))
		} else {
			parsedFlags[f.Name] = []uint8{uint8(uint8Val)}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]uint8)
		}

	case *Uint16SliceFlag:
		uint16Val, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid uint16 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]uint16), uint16(uint16Val))
		} else {
			parsedFlags[f.Name] = []uint16{uint16(uint16Val)}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]uint16)
		}

	case *Uint32SliceFlag:
		uint32Val, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid uint32 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]uint32), uint32(uint32Val))
		} else {
			parsedFlags[f.Name] = []uint32{uint32(uint32Val)}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]uint32)
		}

	case *Uint64SliceFlag:
		uint64Val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint64 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]uint64), uint64Val)
		} else {
			parsedFlags[f.Name] = []uint64{uint64Val}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]uint64)
		}

	case *Float32SliceFlag:
		float32Val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return fmt.Errorf("invalid float32 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]float32), float32(float32Val))
		} else {
			parsedFlags[f.Name] = []float32{float32(float32Val)}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]float32)
		}

	case *Float64SliceFlag:
		float64Val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float64 value for flag --%s: %s", f.Name, value)
		}

		if existing, ok := parsedFlags[f.Name]; ok {
			parsedFlags[f.Name] = append(existing.([]float64), float64Val)
		} else {
			parsedFlags[f.Name] = []float64{float64Val}
		}

		if f.AssignTo != nil {
			*f.AssignTo = parsedFlags[f.Name].([]float64)
		}
	}

	return nil
}

func (f *FlagTyped[T]) flagDefinition() string {
	var typeInfo string

	// Determine type text based on type T
	if !f.HideType {
		typeInfo = " " + f.typeText()
	}

	// Start building the flag definition
	var result string

	// Check if the first alias is a single character
	if len(f.Aliases) > 0 && len(f.Aliases[0]) == 1 {
		// First alias is a single character, use it at the beginning
		result = fmt.Sprintf("-%s, --%s%s", f.Aliases[0], f.Name, typeInfo)

		// Add any remaining aliases (starting from the second one)
		for _, alias := range f.Aliases[1:] {
			if len(alias) == 1 {
				result += fmt.Sprintf(", -%s", alias)
			} else {
				result += fmt.Sprintf(", --%s", alias)
			}
		}
	} else {
		// No single-character alias as the first alias, use padding
		result = fmt.Sprintf("    --%s%s", f.Name, typeInfo)

		// Add all aliases
		for _, alias := range f.Aliases {
			if len(alias) == 1 {
				result += fmt.Sprintf(", -%s", alias)
			} else {
				result += fmt.Sprintf(", --%s", alias)
			}
		}
	}

	return result
}

func (f *FlagTyped[T]) getUsage() string {
	return f.Usage
}

func (f *FlagTyped[T]) defaultValueText() string {
	if f.HideDefault {
		return ""
	}

	if f.DefaultText != "" {
		return f.DefaultText
	}

	// Use reflection to check if default value is set
	zero := reflect.Zero(reflect.TypeOf(f.DefaultValue)).Interface()
	if reflect.DeepEqual(f.DefaultValue, zero) {
		return "" // Return empty string for zero values
	}

	// Format the default value based on type
	return fmt.Sprintf("%v", f.DefaultValue)
}

func (f *FlagTyped[T]) typeText() string {
	return getTypeText(f.DefaultValue)
}
