package cli

import (
	"reflect"
)

// GetTypeText returns a string representation of a type for help text display
func GetTypeText(value interface{}) string {
	t := reflect.TypeOf(value)

	// Handle nil value
	if t == nil {
		return "value"
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "bool"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			return "strings"
		} else if t.Elem().Kind() == reflect.Int || t.Elem().Kind() == reflect.Int32 || t.Elem().Kind() == reflect.Int64 {
			return "ints"
		} else if t.Elem().Kind() == reflect.Float32 || t.Elem().Kind() == reflect.Float64 {
			return "floats"
		}
		return "values"
	default:
		return "value"
	}
}

// StrToPtr converts a string to a pointer to a string.
func StrToPtr(v string) *string {
	return &v
}
