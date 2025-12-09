package cli

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// mockConfigSource implements ConfigFileSource for testing
type mockConfigSource struct {
	data map[string]any
}

func (m *mockConfigSource) GetValue(path string) (any, bool) {
	keys := strings.Split(path, ".")
	current := m.data

	for i, key := range keys {
		if i == len(keys)-1 {
			value, exists := current[key]
			return value, exists
		}
		if nextMap, ok := current[key].(map[string]any); ok {
			current = nextMap
		} else if nextMap, ok := current[key].(map[string]interface{}); ok {
			// Convert map[string]interface{} to map[string]any
			converted := make(map[string]any)
			for k, v := range nextMap {
				converted[k] = v
			}
			current = converted
		} else {
			return nil, false
		}
	}

	return nil, false
}

func (m *mockConfigSource) GetKeys(path string) []string {
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
		} else if nextMap, ok := current[key].(map[string]interface{}); ok {
			converted := make(map[string]any)
			for k, v := range nextMap {
				converted[k] = v
			}
			current = converted
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
	} else if objMap, ok := current[lastKey].(map[string]interface{}); ok {
		result := make([]string, 0, len(objMap))
		for k := range objMap {
			result = append(result, k)
		}
		return result
	}

	return nil
}

func (m *mockConfigSource) SetValue(path string, value any) error {
	keys := strings.Split(path, ".")
	current := m.data

	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
			return nil
		}
		if nextMap, ok := current[key].(map[string]any); ok {
			current = nextMap
		} else if nextMap, ok := current[key].(map[string]interface{}); ok {
			converted := make(map[string]any)
			for k, v := range nextMap {
				converted[k] = v
			}
			current[key] = converted
			current = converted
		} else {
			// Create new nested map
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		}
	}

	return nil
}

func (m *mockConfigSource) DeleteKey(path string) error {
	keys := strings.Split(path, ".")
	current := m.data

	for i, key := range keys {
		if i == len(keys)-1 {
			delete(current, key)
			return nil
		}
		if nextMap, ok := current[key].(map[string]any); ok {
			current = nextMap
		} else if nextMap, ok := current[key].(map[string]interface{}); ok {
			converted := make(map[string]any)
			for k, v := range nextMap {
				converted[k] = v
			}
			current[key] = converted
			current = converted
		} else {
			return nil
		}
	}

	return nil
}

func (m *mockConfigSource) Save() error {
	return nil
}

func (m *mockConfigSource) OnChange(h ConfigFileChangeHandler) error {
	return nil
}

func (m *mockConfigSource) FileUsed() string {
	return "[mock]"
}

func (m *mockConfigSource) LoadData() error {
	return nil
}

// Test data
func createTestConfig() *ConfigFileTypedWrapper {
	mock := &mockConfigSource{
		data: map[string]any{
			"string_val":   "hello world",
			"int_val":      int64(42),
			"float_val":    float64(3.14),
			"bool_val":     true,
			"string_slice": []any{"a", "b", "c"},
			"int_slice":    []any{1, 2, 3},
			"nested": map[string]any{
				"nested_string": "nested value",
				"nested_int":    int64(100),
				"deep": map[string]any{
					"value": "deeply nested",
				},
			},
			"object_slice": []map[string]interface{}{
				{
					"name":  "obj1",
					"value": int64(10),
					"tags":  []any{"tag1", "tag2"},
				},
				{
					"name":  "obj2",
					"value": int64(20),
					"tags":  []any{"tag3", "tag4"},
				},
			},
		},
	}
	return NewTypedConfigFile(mock)
}

func TestConfigFileTyped_Getters(t *testing.T) {
	config := createTestConfig()

	tests := []struct {
		name     string
		path     string
		expected any
		actual   any
	}{
		{"String", "string_val", "hello world", config.GetString("string_val")},
		{"Int", "int_val", 42, config.GetInt("int_val")},
		{"Int64", "int_val", int64(42), config.GetInt64("int_val")},
		{"Float64", "float_val", 3.14, config.GetFloat64("float_val")},
		{"Float32", "float_val", float32(3.14), config.GetFloat32("float_val")},
		{"Bool", "bool_val", true, config.GetBool("bool_val")},
		{"StringSlice", "string_slice", []string{"a", "b", "c"}, config.GetStringSlice("string_slice")},
		{"IntSlice", "int_slice", []int{1, 2, 3}, config.GetIntSlice("int_slice")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.expected, tt.actual) {
				t.Errorf("Expected %v, got %v", tt.expected, tt.actual)
			}
		})
	}
}

func TestConfigFileTyped_GetObject(t *testing.T) {
	config := createTestConfig()

	// Test getting nested object
	nested := config.GetObject("nested")
	if nested == nil {
		t.Fatal("Expected nested object, got nil")
	}

	// Test accessing nested object properties
	if got := nested.GetString("nested_string"); got != "nested value" {
		t.Errorf("Expected 'nested value', got %q", got)
	}

	if got := nested.GetInt("nested_int"); got != 100 {
		t.Errorf("Expected 100, got %d", got)
	}

	// Test deeply nested access
	deep := nested.GetObject("deep")
	if deep == nil {
		t.Fatal("Expected deep object, got nil")
	}

	if got := deep.GetString("value"); got != "deeply nested" {
		t.Errorf("Expected 'deeply nested', got %q", got)
	}

	// Test non-existent object
	if obj := config.GetObject("non_existent"); obj != nil {
		t.Errorf("Expected nil for non-existent object, got %v", obj)
	}
}

func TestConfigFileTyped_GetObjectSlice(t *testing.T) {
	config := createTestConfig()

	objects := config.GetObjectSlice("object_slice")
	if objects == nil {
		t.Fatal("Expected object slice, got nil")
	}

	if len(objects) != 2 {
		t.Fatalf("Expected 2 objects, got %d", len(objects))
	}

	// Test first object
	obj1 := objects[0]
	if got := obj1.GetString("name"); got != "obj1" {
		t.Errorf("Expected 'obj1', got %q", got)
	}

	if got := obj1.GetInt("value"); got != 10 {
		t.Errorf("Expected 10, got %d", got)
	}

	tags := obj1.GetStringSlice("tags")
	if !reflect.DeepEqual(tags, []string{"tag1", "tag2"}) {
		t.Errorf("Expected ['tag1', 'tag2'], got %v", tags)
	}

	// Test second object
	obj2 := objects[1]
	if got := obj2.GetString("name"); got != "obj2" {
		t.Errorf("Expected 'obj2', got %q", got)
	}

	if got := obj2.GetInt("value"); got != 20 {
		t.Errorf("Expected 20, got %d", got)
	}

	// Test non-existent slice
	if slice := config.GetObjectSlice("non_existent"); slice != nil {
		t.Errorf("Expected nil for non-existent slice, got %v", slice)
	}
}

func TestConfigFileTyped_NewTypedConfigObject(t *testing.T) {
	// Test creating empty object
	obj := NewTypedConfigObject()
	if obj == nil {
		t.Fatal("Expected new config object, got nil")
	}

	// Test it's initially empty
	if got := obj.GetString("test"); got != "" {
		t.Errorf("Expected empty string for non-existent key, got %q", got)
	}

	// Test setting and getting values
	if err := obj.SetString("test", "value"); err != nil {
		t.Errorf("Error setting string: %v", err)
	}

	if got := obj.GetString("test"); got != "value" {
		t.Errorf("Expected 'value', got %q", got)
	}
}

func TestConfigFileTyped_SetObject(t *testing.T) {
	config := createTestConfig()

	// Create a new object
	newObj := NewTypedConfigObject()
	newObj.SetString("name", "test object")
	newObj.SetInt("value", 999)
	newObj.SetBool("enabled", true)

	// Set the object
	if err := config.SetObject("new_object", newObj); err != nil {
		t.Fatalf("Error setting object: %v", err)
	}

	// Retrieve and verify
	retrieved := config.GetObject("new_object")
	if retrieved == nil {
		t.Fatal("Expected to retrieve new object, got nil")
	}

	if got := retrieved.GetString("name"); got != "test object" {
		t.Errorf("Expected 'test object', got %q", got)
	}

	if got := retrieved.GetInt("value"); got != 999 {
		t.Errorf("Expected 999, got %d", got)
	}

	if got := retrieved.GetBool("enabled"); got != true {
		t.Errorf("Expected true, got %t", got)
	}
}

func TestConfigFileTyped_SetObjectSlice(t *testing.T) {
	config := createTestConfig()

	// Create multiple objects
	obj1 := NewTypedConfigObject()
	obj1.SetString("name", "object 1")
	obj1.SetInt("id", 1)

	obj2 := NewTypedConfigObject()
	obj2.SetString("name", "object 2")
	obj2.SetInt("id", 2)

	obj3 := NewTypedConfigObject()
	obj3.SetString("name", "object 3")
	obj3.SetInt("id", 3)

	// Set the object slice
	objects := []ConfigFileTyped{obj1, obj2, obj3}
	if err := config.SetObjectSlice("new_objects", objects); err != nil {
		t.Fatalf("Error setting object slice: %v", err)
	}

	// Retrieve and verify
	retrieved := config.GetObjectSlice("new_objects")
	if retrieved == nil {
		t.Fatal("Expected to retrieve object slice, got nil")
	}

	if len(retrieved) != 3 {
		t.Fatalf("Expected 3 objects, got %d", len(retrieved))
	}

	for i, obj := range retrieved {
		expectedName := fmt.Sprintf("object %d", i+1)
		if got := obj.GetString("name"); got != expectedName {
			t.Errorf("Object %d: Expected name %q, got %q", i, expectedName, got)
		}

		expectedID := i + 1
		if got := obj.GetInt("id"); got != expectedID {
			t.Errorf("Object %d: Expected id %d, got %d", i, expectedID, got)
		}
	}
}

func TestConfigFileTyped_NestedObjectSetting(t *testing.T) {
	config := createTestConfig()

	// Create a nested object structure
	parent := NewTypedConfigObject()
	parent.SetString("name", "parent")

	child := NewTypedConfigObject()
	child.SetString("name", "child")
	child.SetInt("value", 42)

	grandchild := NewTypedConfigObject()
	grandchild.SetString("message", "hello from grandchild")

	// Nest the objects
	if err := child.SetObject("nested", grandchild); err != nil {
		t.Fatalf("Error setting grandchild: %v", err)
	}

	if err := parent.SetObject("child", child); err != nil {
		t.Fatalf("Error setting child: %v", err)
	}

	// Set the parent
	if err := config.SetObject("nested_structure", parent); err != nil {
		t.Fatalf("Error setting parent: %v", err)
	}

	// Verify the nested structure
	retrievedParent := config.GetObject("nested_structure")
	if retrievedParent == nil {
		t.Fatal("Expected to retrieve parent, got nil")
	}

	if got := retrievedParent.GetString("name"); got != "parent" {
		t.Errorf("Expected parent name 'parent', got %q", got)
	}

	retrievedChild := retrievedParent.GetObject("child")
	if retrievedChild == nil {
		t.Fatal("Expected to retrieve child, got nil")
	}

	if got := retrievedChild.GetString("name"); got != "child" {
		t.Errorf("Expected child name 'child', got %q", got)
	}

	if got := retrievedChild.GetInt("value"); got != 42 {
		t.Errorf("Expected child value 42, got %d", got)
	}

	retrievedGrandchild := retrievedChild.GetObject("nested")
	if retrievedGrandchild == nil {
		t.Fatal("Expected to retrieve grandchild, got nil")
	}

	if got := retrievedGrandchild.GetString("message"); got != "hello from grandchild" {
		t.Errorf("Expected grandchild message 'hello from grandchild', got %q", got)
	}
}

func TestConfigFileTyped_Setters(t *testing.T) {
	config := NewTypedConfigFile(&mockConfigSource{data: make(map[string]any)})

	tests := []struct {
		name     string
		path     string
		value    any
		setFunc  func(string, any) error
		getFunc  func(string) any
	}{
		{"String", "test_string", "hello world",
			func(p string, v any) error { return config.SetString(p, v.(string)) },
			func(p string) any { return config.GetString(p) } },
		{"Int", "test_int", 42,
			func(p string, v any) error { return config.SetInt(p, v.(int)) },
			func(p string) any { return config.GetInt(p) } },
		{"Int64", "test_int64", int64(1234567890),
			func(p string, v any) error { return config.SetInt64(p, v.(int64)) },
			func(p string) any { return config.GetInt64(p) } },
		{"Float64", "test_float64", 3.14159,
			func(p string, v any) error { return config.SetFloat64(p, v.(float64)) },
			func(p string) any { return config.GetFloat64(p) } },
		{"Bool", "test_bool", true,
			func(p string, v any) error { return config.SetBool(p, v.(bool)) },
			func(p string) any { return config.GetBool(p) } },
		{"StringSlice", "test_string_slice", []string{"a", "b", "c"},
			func(p string, v any) error { return config.SetStringSlice(p, v.([]string)) },
			func(p string) any { return config.GetStringSlice(p) } },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setFunc(tt.path, tt.value); err != nil {
				t.Fatalf("Error setting %s: %v", tt.name, err)
			}

			if got := tt.getFunc(tt.path); !reflect.DeepEqual(tt.value, got) {
				t.Errorf("Expected %v, got %v", tt.value, got)
			}
		})
	}
}

func TestConfigFileTyped_ErrorHandling(t *testing.T) {
	config := createTestConfig()

	// Test getting non-existent values
	if got := config.GetString("non_existent"); got != "" {
		t.Errorf("Expected empty string for non-existent key, got %q", got)
	}

	if got := config.GetInt("non_existent"); got != 0 {
		t.Errorf("Expected 0 for non-existent key, got %d", got)
	}

	if got := config.GetBool("non_existent"); got != false {
		t.Errorf("Expected false for non-existent key, got %t", got)
	}

	// Test type conversion from int to string
	config.SetInt("int_as_string", 1)
	if got := config.GetString("int_as_string"); got != "1" {
		t.Errorf("Expected \"1\" when getting int as string, got %q", got)
	}

	// Test setting invalid object types
	if err := config.SetObject("invalid", nil); err != nil {
		// Should succeed (deletes the key)
		t.Errorf("Expected no error setting nil object, got %v", err)
	}

	// Create an object that's not backed by mapConfigSource
	mock := &mockConfigSource{data: make(map[string]any)}
	wrapped := NewTypedConfigFile(mock)
	if err := config.SetObject("invalid", wrapped); err == nil {
		t.Error("Expected error when setting object not backed by mapConfigSource")
	}
}

func TestConfigFileTyped_NewTypedConfigObjectWithData(t *testing.T) {
	data := map[string]any{
		"string": "test",
		"number": int64(42),
		"nested": map[string]any{
			"value": "nested value",
		},
	}

	obj := NewTypedConfigObjectWithData(data)
	if obj == nil {
		t.Fatal("Expected new config object with data, got nil")
	}

	// Test accessing the data
	if got := obj.GetString("string"); got != "test" {
		t.Errorf("Expected 'test', got %q", got)
	}

	if got := obj.GetInt("number"); got != 42 {
		t.Errorf("Expected 42, got %d", got)
	}

	nested := obj.GetObject("nested")
	if nested == nil {
		t.Fatal("Expected nested object, got nil")
	}

	if got := nested.GetString("value"); got != "nested value" {
		t.Errorf("Expected 'nested value', got %q", got)
	}
}