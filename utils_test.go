package cli

import "testing"

func TestGetTypeText(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected string
	}{
		{"hello", "string"},
		{42, "int"},
		{int8(1), "int"},
		{int16(1), "int"},
		{int32(1), "int"},
		{int64(1), "int"},
		{uint(1), "uint"},
		{uint8(1), "uint"},
		{uint16(1), "uint"},
		{uint32(1), "uint"},
		{uint64(1), "uint"},
		{float32(1.0), "float"},
		{float64(1.0), "float"},
		{true, "bool"},
		{[]string{"a"}, "strings"},
		{[]int{1}, "ints"},
		{[]int32{1}, "ints"},
		{[]int64{1}, "ints"},
		{[]float32{1.0}, "floats"},
		{[]float64{1.0}, "floats"},
		{[]bool{true}, "values"},
		{nil, "value"},
	}

	for _, tt := range tests {
		got := GetTypeText(tt.value)
		if got != tt.expected {
			t.Errorf("GetTypeText(%T) = %q, want %q", tt.value, got, tt.expected)
		}
	}
}

func TestStrToPtr(t *testing.T) {
	s := "hello"
	p := StrToPtr(s)
	if p == nil {
		t.Fatal("StrToPtr returned nil")
	}
	if *p != s {
		t.Errorf("StrToPtr: got %q, want %q", *p, s)
	}
	// Ensure it's a new pointer, not the address of s
	if p == &s {
		t.Error("StrToPtr should return a new pointer, not the address of the argument")
	}
}
