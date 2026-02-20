package cli

import (
	"fmt"
	"reflect"
	"strings"
)

// mapConfigSource implements ConfigFileSource for a map[string]any
// This is used as the base for objects returned by GetObject and GetObjectSlice
type mapConfigSource struct {
	data     map[string]any
	readOnly bool
}

func (m *mapConfigSource) GetValue(path string) (any, bool) {
	keys := strings.Split(path, ".")
	current := m.data

	var value any
	var exists bool

	for i, key := range keys {
		if i == len(keys)-1 {
			value, exists = current[key]
			break
		}
		if nextMap, ok := current[key].(map[string]any); ok {
			current = nextMap
		} else {
			exists = false
			break
		}
	}

	return value, exists
}

func (m *mapConfigSource) GetKeys(path string) []string {
	if path == "" {
		result := make([]string, 0, len(m.data))
		for k := range m.data {
			result = append(result, k)
		}
		return result
	}

	keys := strings.Split(path, ".")
	current := m.data

	for _, key := range keys[:len(keys)-1] {
		if nextMap, ok := current[key].(map[string]any); ok {
			current = nextMap
		} else {
			return nil
		}
	}

	lastKey := keys[len(keys)-1]
	if objMap, ok := current[lastKey].(map[string]any); ok {
		result := make([]string, 0, len(objMap))
		for k := range objMap {
			result = append(result, k)
		}
		return result
	}

	return nil
}

func (m *mapConfigSource) SetValue(path string, value any) error {
	if m.readOnly {
		return fmt.Errorf("mapConfigSource is read-only")
	}

	keys := strings.Split(path, ".")
	current := m.data

	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
			return nil
		}
		if nextMap, ok := current[key].(map[string]any); ok {
			current = nextMap
		} else {
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		}
	}

	return nil
}

func (m *mapConfigSource) DeleteKey(path string) error {
	if m.readOnly {
		return fmt.Errorf("mapConfigSource is read-only")
	}

	keys := strings.Split(path, ".")
	current := m.data

	for i, key := range keys {
		if i == len(keys)-1 {
			delete(current, key)
			return nil
		}
		if nextMap, ok := current[key].(map[string]any); ok {
			current = nextMap
		} else {
			return nil
		}
	}

	return nil
}

func (m *mapConfigSource) Save() error {
	if m.readOnly {
		return fmt.Errorf("mapConfigSource is read-only")
	}
	return nil
}

func (m *mapConfigSource) OnChange(h ConfigFileChangeHandler) error {
	if m.readOnly {
		return fmt.Errorf("mapConfigSource is read-only")
	}
	return nil
}

func (m *mapConfigSource) FileUsed() string {
	return "[map source]"
}

func (m *mapConfigSource) LoadData() error {
	return nil // Data is already loaded
}

type ConfigFileTyped interface {
	ConfigFileSource

	GetString(string) string                 // Get the string value from the configuration file at the specified path.
	GetInt(string) int                       // Get the int value from the configuration file at the specified path.
	GetInt64(string) int64                   // Get the int64 value from the configuration file at the specified path.
	GetInt32(string) int32                   // Get the int32 value from the configuration file at the specified path.
	GetInt16(string) int16                   // Get the int16 value from the configuration file at the specified path.
	GetInt8(string) int8                     // Get the int8 value from the configuration file at the specified path.
	GetUint(string) uint                     // Get the uint value from the configuration file at the specified path.
	GetUint64(string) uint64                 // Get the uint64 value from the configuration file at the specified path.
	GetUint32(string) uint32                 // Get the uint32 value from the configuration file at the specified path.
	GetUint16(string) uint16                 // Get the uint16 value from the configuration file at the specified path.
	GetUint8(string) uint8                   // Get the uint8 value from the configuration file at the specified path.
	GetFloat32(string) float32               // Get the float32 value from the configuration file at the specified path.
	GetFloat64(string) float64               // Get the float64 value from the configuration file at the specified path.
	GetBool(string) bool                     // Get the bool value from the configuration file at the specified path.
	GetStringSlice(string) []string          // Get the string slice value from the configuration file at the specified path.
	GetIntSlice(string) []int                // Get the int slice value from the configuration file at the specified path.
	GetInt64Slice(string) []int64            // Get the int64 slice value from the configuration file at the specified path.
	GetInt32Slice(string) []int32            // Get the int32 slice value from the configuration file at the specified path.
	GetInt16Slice(string) []int16            // Get the int16 slice value from the configuration file at the specified path.
	GetInt8Slice(string) []int8              // Get the int8 slice value from the configuration file at the specified path.
	GetUintSlice(string) []uint              // Get the uint slice value from the configuration file at the specified path.
	GetUint64Slice(string) []uint64          // Get the uint64 slice value from the configuration file at the specified path.
	GetUint32Slice(string) []uint32          // Get the uint32 slice value from the configuration file at the specified path.
	GetUint16Slice(string) []uint16          // Get the uint16 slice value from the configuration file at the specified path.
	GetUint8Slice(string) []uint8            // Get the uint8 slice value from the configuration file at the specified path.
	GetFloat32Slice(string) []float32        // Get the float32 slice value from the configuration file at the specified path.
	GetFloat64Slice(string) []float64        // Get the float64 slice value from the configuration file at the specified path.
	SetString(string, string) error          // Set the string value in the configuration file at the specified path.
	SetInt(string, int) error                // Set the int value in the configuration file at the specified path.
	SetInt64(string, int64) error            // Set the int64 value in the configuration file at the specified path.
	SetInt32(string, int32) error            // Set the int32 value in the configuration file at the specified path.
	SetInt16(string, int16) error            // Set the int16 value in the configuration file at the specified path.
	SetInt8(string, int8) error              // Set the int8 value in the configuration file at the specified path.
	SetUint(string, uint) error              // Set the uint value in the configuration file at the specified path.
	SetUint64(string, uint64) error          // Set the uint64 value in the configuration file at the specified path.
	SetUint32(string, uint32) error          // Set the uint32 value in the configuration file at the specified path.
	SetUint16(string, uint16) error          // Set the uint16 value in the configuration file at the specified path.
	SetUint8(string, uint8) error            // Set the uint8 value in the configuration file at the specified path.
	SetFloat32(string, float32) error        // Set the float32 value in the configuration file at the specified path.
	SetFloat64(string, float64) error        // Set the float64 value in the configuration file at the specified path.
	SetBool(string, bool) error              // Set the bool value in the configuration file at the specified path.
	SetStringSlice(string, []string) error   // Set the string slice value in the configuration file at the specified path.
	SetIntSlice(string, []int) error         // Set the int slice value in the configuration file at the specified path.
	SetInt64Slice(string, []int64) error     // Set the int64 slice value in the configuration file at the specified path.
	SetInt32Slice(string, []int32) error     // Set the int32 slice value in the configuration file at the specified path.
	SetInt16Slice(string, []int16) error     // Set the int16 slice value in the configuration file at the specified path.
	SetInt8Slice(string, []int8) error       // Set the int8 slice value in the configuration file at the specified path.
	SetUintSlice(string, []uint) error       // Set the uint slice value in the configuration file at the specified path.
	SetUint64Slice(string, []uint64) error   // Set the uint64 slice value in the configuration file at the specified path.
	SetUint32Slice(string, []uint32) error   // Set the uint32 slice value in the configuration file at the specified path.
	SetUint16Slice(string, []uint16) error   // Set the uint16 slice value in the configuration file at the specified path.
	SetUint8Slice(string, []uint8) error     // Set the uint8 slice value in the configuration file at the specified path.
	SetFloat32Slice(string, []float32) error // Set the float32 slice value in the configuration file at the specified path.
	SetFloat64Slice(string, []float64) error // Set the float64 slice value in the configuration file at the specified path.

	// Object getters
	GetObjectSlice(string) []ConfigFileTyped   // Get a slice of objects from the configuration file at the specified path.
	GetObject(string) ConfigFileTyped         // Get a single object from the configuration file at the specified path.

	// Object setters
	SetObjectSlice(string, []ConfigFileTyped) error   // Set a slice of objects in the configuration file at the specified path.
	SetObject(string, ConfigFileTyped) error         // Set an object in the configuration file at the specified path.
}

type ConfigFileTypedWrapper struct {
	inner ConfigFileSource
}

var _ ConfigFileTyped = (*ConfigFileTypedWrapper)(nil)

func NewTypedConfigFile(inner ConfigFileSource) *ConfigFileTypedWrapper {
	return &ConfigFileTypedWrapper{
		inner: inner,
	}
}

// NewTypedConfigObject creates a new empty ConfigFileTyped instance that can be used for setting objects
// This returns a ConfigFileTyped wrapped around an empty map that you can populate and then set
func NewTypedConfigObject() ConfigFileTyped {
	mapSource := &mapConfigSource{data: make(map[string]any), readOnly: false}
	return &ConfigFileTypedWrapper{inner: mapSource}
}

// NewTypedConfigObjectWithData creates a new ConfigFileTyped instance from a map[string]any
func NewTypedConfigObjectWithData(data map[string]any) ConfigFileTyped {
	mapSource := &mapConfigSource{data: data, readOnly: false}
	return &ConfigFileTypedWrapper{inner: mapSource}
}

func (w *ConfigFileTypedWrapper) GetValue(path string) (any, bool)  { return w.inner.GetValue(path) }
func (w *ConfigFileTypedWrapper) GetKeys(path string) []string      { return w.inner.GetKeys(path) }
func (w *ConfigFileTypedWrapper) SetValue(path string, v any) error { return w.inner.SetValue(path, v) }
func (w *ConfigFileTypedWrapper) DeleteKey(path string) error       { return w.inner.DeleteKey(path) }
func (w *ConfigFileTypedWrapper) Save() error                       { return w.inner.Save() }
func (w *ConfigFileTypedWrapper) OnChange(h ConfigFileChangeHandler) error {
	return w.inner.OnChange(h)
}
func (w *ConfigFileTypedWrapper) FileUsed() string { return w.inner.FileUsed() }

// convertValue handles type conversion from any value to target type T
func convertValue[T any](value any) T {
	var zero T

	// Try direct type assertion first
	if cast, ok := value.(T); ok {
		return cast
	}

	// Handle type conversions based on target type
	switch any(zero).(type) {
	case int:
		if f, ok := value.(float64); ok {
			return any(int(f)).(T)
		} else if i, ok := value.(int64); ok {
			return any(int(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(int(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(int(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(int(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(int(u)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(int(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(int(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(int(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(int(u)).(T)
		}
	case int64:
		if f, ok := value.(float64); ok {
			return any(int64(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(int64(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(int64(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(int64(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(int64(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(int64(u)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(int64(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(int64(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(int64(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(int64(u)).(T)
		}
	case int32:
		if f, ok := value.(float64); ok {
			return any(int32(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(int32(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(int32(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(int32(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(int32(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(int32(u)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(int32(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(int32(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(int32(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(int32(u)).(T)
		}
	case int16:
		if f, ok := value.(float64); ok {
			return any(int16(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(int16(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(int16(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(int16(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(int16(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(int16(u)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(int16(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(int16(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(int16(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(int16(u)).(T)
		}
	case int8:
		if f, ok := value.(float64); ok {
			return any(int8(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(int8(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(int8(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(int8(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(int8(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(int8(u)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(int8(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(int8(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(int8(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(int8(u)).(T)
		}
	case uint:
		if f, ok := value.(float64); ok {
			return any(uint(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(uint(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(uint(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(uint(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(uint(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(uint(i)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(uint(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(uint(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(uint(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(uint(u)).(T)
		}
	case uint64:
		if f, ok := value.(float64); ok {
			return any(uint64(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(uint64(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(uint64(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(uint64(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(uint64(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(uint64(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(uint64(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(uint64(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(uint64(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(uint64(u)).(T)
		}
	case uint32:
		if f, ok := value.(float64); ok {
			return any(uint32(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(uint32(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(uint32(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(uint32(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(uint32(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(uint32(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(uint32(u)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(uint32(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(uint32(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(uint32(u)).(T)
		}
	case uint16:
		if f, ok := value.(float64); ok {
			return any(uint16(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(uint16(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(uint16(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(uint16(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(uint16(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(uint16(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(uint16(u)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(uint16(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(uint16(u)).(T)
		} else if u, ok := value.(uint8); ok {
			return any(uint16(u)).(T)
		}
	case uint8:
		if f, ok := value.(float64); ok {
			return any(uint8(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(uint8(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(uint8(i)).(T)
		} else if i, ok := value.(int32); ok {
			return any(uint8(i)).(T)
		} else if i, ok := value.(int16); ok {
			return any(uint8(i)).(T)
		} else if i, ok := value.(int8); ok {
			return any(uint8(i)).(T)
		} else if u, ok := value.(uint); ok {
			return any(uint8(u)).(T)
		} else if u, ok := value.(uint64); ok {
			return any(uint8(u)).(T)
		} else if u, ok := value.(uint32); ok {
			return any(uint8(u)).(T)
		} else if u, ok := value.(uint16); ok {
			return any(uint8(u)).(T)
		}
	case float32:
		if f, ok := value.(float64); ok {
			return any(float32(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(float32(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(float32(i)).(T)
		}
	case float64:
		if f, ok := value.(float32); ok {
			return any(float64(f)).(T)
		} else if i, ok := value.(int); ok {
			return any(float64(i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(float64(i)).(T)
		}
	case string:
		if s, ok := value.(string); ok {
			return any(s).(T)
		} else if i, ok := value.(int); ok {
			return any(fmt.Sprintf("%d", i)).(T)
		} else if i, ok := value.(int64); ok {
			return any(fmt.Sprintf("%d", i)).(T)
		} else if f, ok := value.(float64); ok {
			return any(fmt.Sprintf("%g", f)).(T)
		} else if b, ok := value.(bool); ok {
			return any(fmt.Sprintf("%t", b)).(T)
		}
	case bool:
		if b, ok := value.(bool); ok {
			return any(b).(T)
		} else if s, ok := value.(string); ok {
			// Handle string to bool conversion
			switch strings.ToLower(s) {
			case "true", "yes", "y", "1":
				return any(true).(T)
			case "false", "no", "n", "0":
				return any(false).(T)
			}
		} else if i, ok := value.(int); ok {
			return any(i != 0).(T)
		} else if f, ok := value.(float64); ok {
			return any(f != 0).(T)
		}
	}

	return zero
}

// isZeroValue checks if a value is the zero value for its type
func isZeroValue[T any](value T) bool {
	return reflect.ValueOf(value).IsZero()
}

// getAs gets a value from the config and converts it to type T
func getAs[T any](c ConfigFileSource, path string) T {
	var zero T
	if value, exists := c.GetValue(path); exists {
		converted := convertValue[T](value)
		return converted
	}
	return zero
}

// getAsSlice now uses convertValue for each element
func getAsSlice[T any](c ConfigFileSource, path string) []T {
	if value, exists := c.GetValue(path); exists {
		// If already the correct type, return it
		if cast, ok := value.([]T); ok {
			return cast
		}

		// If it's a []interface{}, convert each element using convertValue
		if arr, ok := value.([]interface{}); ok {
			result := make([]T, 0, len(arr))
			for _, v := range arr {
				converted := convertValue[T](v)
				result = append(result, converted)
			}
			return result
		}

		// Handle single value to slice conversion
		converted := convertValue[T](value)
		if !isZeroValue(converted) {
			return []T{converted}
		}
	}
	return nil
}

func (c *ConfigFileTypedWrapper) GetString(path string) string {
	return getAs[string](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt(path string) int {
	return getAs[int](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt64(path string) int64 {
	return getAs[int64](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt32(path string) int32 {
	return getAs[int32](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt16(path string) int16 {
	return getAs[int16](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt8(path string) int8 {
	return getAs[int8](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint(path string) uint {
	return getAs[uint](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetBool(path string) bool {
	return getAs[bool](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint64(path string) uint64 {
	return getAs[uint64](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint32(path string) uint32 {
	return getAs[uint32](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint16(path string) uint16 {
	return getAs[uint16](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint8(path string) uint8 {
	return getAs[uint8](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetFloat32(path string) float32 {
	return getAs[float32](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetFloat64(path string) float64 {
	return getAs[float64](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetStringSlice(path string) []string {
	return getAsSlice[string](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetIntSlice(path string) []int {
	return getAsSlice[int](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt64Slice(path string) []int64 {
	return getAsSlice[int64](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt32Slice(path string) []int32 {
	return getAsSlice[int32](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt16Slice(path string) []int16 {
	return getAsSlice[int16](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt8Slice(path string) []int8 {
	return getAsSlice[int8](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUintSlice(path string) []uint {
	return getAsSlice[uint](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint64Slice(path string) []uint64 {
	return getAsSlice[uint64](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint32Slice(path string) []uint32 {
	return getAsSlice[uint32](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint16Slice(path string) []uint16 {
	return getAsSlice[uint16](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetUint8Slice(path string) []uint8 {
	return getAsSlice[uint8](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetFloat32Slice(path string) []float32 {
	return getAsSlice[float32](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetFloat64Slice(path string) []float64 {
	return getAsSlice[float64](c.inner, path)
}

func (c *ConfigFileTypedWrapper) SetString(path string, value string) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt(path string, value int) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt64(path string, value int64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt32(path string, value int32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt16(path string, value int16) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt8(path string, value int8) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint(path string, value uint) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint64(path string, value uint64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint32(path string, value uint32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint16(path string, value uint16) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint8(path string, value uint8) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetFloat32(path string, value float32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetFloat64(path string, value float64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetBool(path string, value bool) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetStringSlice(path string, value []string) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetIntSlice(path string, value []int) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt64Slice(path string, value []int64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt32Slice(path string, value []int32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt16Slice(path string, value []int16) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt8Slice(path string, value []int8) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUintSlice(path string, value []uint) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint64Slice(path string, value []uint64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint32Slice(path string, value []uint32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint16Slice(path string, value []uint16) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint8Slice(path string, value []uint8) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetFloat32Slice(path string, value []float32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetFloat64Slice(path string, value []float64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) LoadData() error {
	return c.inner.LoadData()
}

func (c *ConfigFileTypedWrapper) GetObject(path string) ConfigFileTyped {
	if v, ok := c.inner.GetValue(path); ok {
		if objMap, ok := v.(map[string]any); ok {
			mapSource := &mapConfigSource{data: objMap, readOnly: true}
			return &ConfigFileTypedWrapper{inner: mapSource}
		}
	}
	return nil
}

func (c *ConfigFileTypedWrapper) GetObjectSlice(path string) []ConfigFileTyped {
	if v, ok := c.inner.GetValue(path); ok {
		// Handle []any (JSON style)
		if slice, ok := v.([]any); ok {
			objects := make([]ConfigFileTyped, 0, len(slice))
			for _, item := range slice {
				if objMap, ok := item.(map[string]any); ok {
					mapSource := &mapConfigSource{data: objMap, readOnly: true}
					objects = append(objects, &ConfigFileTypedWrapper{inner: mapSource})
				}
			}
			return objects
		}
		// Handle []map[string]interface{} (TOML style - BurntSushi/toml produces this type)
		if slice, ok := v.([]map[string]interface{}); ok {
			objects := make([]ConfigFileTyped, 0, len(slice))
			for _, item := range slice {
				mapSource := &mapConfigSource{data: item, readOnly: true}
				objects = append(objects, &ConfigFileTypedWrapper{inner: mapSource})
			}
			return objects
		}
	}
	return nil
}

// SetObject sets an object at the specified path.
// The object parameter should be a ConfigFileTyped instance that contains the data to set.
func (c *ConfigFileTypedWrapper) SetObject(path string, obj ConfigFileTyped) error {
	if obj == nil {
		// If nil is passed, delete the key
		return c.DeleteKey(path)
	}

	// Extract the data from the object
	// If it's a ConfigFileTypedWrapper, try to get the underlying data
	if wrapper, ok := obj.(*ConfigFileTypedWrapper); ok {
		// Check if the inner source is a mapConfigSource
		if mapSrc, ok := wrapper.inner.(*mapConfigSource); ok {
			return c.SetValue(path, mapSrc.data)
		}

		// For other ConfigFileSource implementations, serialize to map
		// This is a simplified approach - in practice you might want to handle this differently
		return fmt.Errorf("SetObject only supports ConfigFileTypedWrapper with mapConfigSource backing")
	}

	return fmt.Errorf("invalid object type for SetObject")
}

// SetObjectSlice sets a slice of objects at the specified path.
func (c *ConfigFileTypedWrapper) SetObjectSlice(path string, objects []ConfigFileTyped) error {
	if objects == nil {
		// If nil is passed, delete the key
		return c.DeleteKey(path)
	}

	// Convert ConfigFileTyped objects to []map[string]any
	result := make([]map[string]any, 0, len(objects))
	for _, obj := range objects {
		if obj == nil {
			continue
		}

		if wrapper, ok := obj.(*ConfigFileTypedWrapper); ok {
			// Check if the inner source is a mapConfigSource
			if mapSrc, ok := wrapper.inner.(*mapConfigSource); ok {
				result = append(result, mapSrc.data)
			} else {
				// For other sources, we would need to extract all values
				// This is a simplified approach
				return fmt.Errorf("SetObjectSlice only supports ConfigFileTypedWrapper with mapConfigSource backing")
			}
		} else {
			return fmt.Errorf("invalid object type in slice for SetObjectSlice")
		}
	}

	return c.SetValue(path, result)
}
