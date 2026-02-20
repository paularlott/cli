package cli

import (
	"context"
	"os"
	"testing"
)

func TestGetters_AllFlagTypes(t *testing.T) {
	var (
		strVal     string
		intVal     int
		int8Val    int8
		int16Val   int16
		int32Val   int32
		int64Val   int64
		uintVal    uint
		uint8Val   uint8
		uint16Val  uint16
		uint32Val  uint32
		uint64Val  uint64
		float32Val float32
		float64Val float64
		boolVal    bool
	)

	cmd := &Command{
		Name: "test",
		Flags: []Flag{
			&StringFlag{Name: "str", AssignTo: &strVal},
			&IntFlag{Name: "int", AssignTo: &intVal},
			&Int8Flag{Name: "int8", AssignTo: &int8Val},
			&Int16Flag{Name: "int16", AssignTo: &int16Val},
			&Int32Flag{Name: "int32", AssignTo: &int32Val},
			&Int64Flag{Name: "int64", AssignTo: &int64Val},
			&UintFlag{Name: "uint", AssignTo: &uintVal},
			&Uint8Flag{Name: "uint8", AssignTo: &uint8Val},
			&Uint16Flag{Name: "uint16", AssignTo: &uint16Val},
			&Uint32Flag{Name: "uint32", AssignTo: &uint32Val},
			&Uint64Flag{Name: "uint64", AssignTo: &uint64Val},
			&Float32Flag{Name: "float32", AssignTo: &float32Val},
			&Float64Flag{Name: "float64", AssignTo: &float64Val},
			&BoolFlag{Name: "bool", AssignTo: &boolVal},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if cmd.GetString("str") != "hello" {
				t.Errorf("GetString: got %q", cmd.GetString("str"))
			}
			if cmd.GetInt("int") != 1 {
				t.Errorf("GetInt: got %d", cmd.GetInt("int"))
			}
			if cmd.GetInt8("int8") != 2 {
				t.Errorf("GetInt8: got %d", cmd.GetInt8("int8"))
			}
			if cmd.GetInt16("int16") != 3 {
				t.Errorf("GetInt16: got %d", cmd.GetInt16("int16"))
			}
			if cmd.GetInt32("int32") != 4 {
				t.Errorf("GetInt32: got %d", cmd.GetInt32("int32"))
			}
			if cmd.GetInt64("int64") != 5 {
				t.Errorf("GetInt64: got %d", cmd.GetInt64("int64"))
			}
			if cmd.GetUint("uint") != 6 {
				t.Errorf("GetUint: got %d", cmd.GetUint("uint"))
			}
			if cmd.GetUint8("uint8") != 7 {
				t.Errorf("GetUint8: got %d", cmd.GetUint8("uint8"))
			}
			if cmd.GetUint16("uint16") != 8 {
				t.Errorf("GetUint16: got %d", cmd.GetUint16("uint16"))
			}
			if cmd.GetUint32("uint32") != 9 {
				t.Errorf("GetUint32: got %d", cmd.GetUint32("uint32"))
			}
			if cmd.GetUint64("uint64") != 10 {
				t.Errorf("GetUint64: got %d", cmd.GetUint64("uint64"))
			}
			if cmd.GetFloat32("float32") != 1.5 {
				t.Errorf("GetFloat32: got %f", cmd.GetFloat32("float32"))
			}
			if cmd.GetFloat64("float64") != 2.5 {
				t.Errorf("GetFloat64: got %f", cmd.GetFloat64("float64"))
			}
			if !cmd.GetBool("bool") {
				t.Errorf("GetBool: got false")
			}
			return nil
		},
	}

	os.Args = []string{"test",
		"--str", "hello",
		"--int", "1",
		"--int8", "2",
		"--int16", "3",
		"--int32", "4",
		"--int64", "5",
		"--uint", "6",
		"--uint8", "7",
		"--uint16", "8",
		"--uint32", "9",
		"--uint64", "10",
		"--float32", "1.5",
		"--float64", "2.5",
		"--bool",
	}
	if err := cmd.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetters_MissingFlagReturnsZero(t *testing.T) {
	cmd := &Command{
		Name: "test",
		Run:  func(ctx context.Context, cmd *Command) error { return nil },
	}
	os.Args = []string{"test"}
	if err := cmd.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmd.GetString("missing") != "" {
		t.Error("GetString missing should return empty string")
	}
	if cmd.GetInt("missing") != 0 {
		t.Error("GetInt missing should return 0")
	}
	if cmd.GetInt8("missing") != 0 {
		t.Error("GetInt8 missing should return 0")
	}
	if cmd.GetInt16("missing") != 0 {
		t.Error("GetInt16 missing should return 0")
	}
	if cmd.GetInt32("missing") != 0 {
		t.Error("GetInt32 missing should return 0")
	}
	if cmd.GetInt64("missing") != 0 {
		t.Error("GetInt64 missing should return 0")
	}
	if cmd.GetUint("missing") != 0 {
		t.Error("GetUint missing should return 0")
	}
	if cmd.GetUint8("missing") != 0 {
		t.Error("GetUint8 missing should return 0")
	}
	if cmd.GetUint16("missing") != 0 {
		t.Error("GetUint16 missing should return 0")
	}
	if cmd.GetUint32("missing") != 0 {
		t.Error("GetUint32 missing should return 0")
	}
	if cmd.GetUint64("missing") != 0 {
		t.Error("GetUint64 missing should return 0")
	}
	if cmd.GetFloat32("missing") != 0 {
		t.Error("GetFloat32 missing should return 0")
	}
	if cmd.GetFloat64("missing") != 0 {
		t.Error("GetFloat64 missing should return 0")
	}
	if cmd.GetBool("missing") {
		t.Error("GetBool missing should return false")
	}
	if cmd.GetStringSlice("missing") != nil {
		t.Error("GetStringSlice missing should return nil")
	}
	if cmd.GetIntSlice("missing") != nil {
		t.Error("GetIntSlice missing should return nil")
	}
	if cmd.GetInt8Slice("missing") != nil {
		t.Error("GetInt8Slice missing should return nil")
	}
	if cmd.GetInt16Slice("missing") != nil {
		t.Error("GetInt16Slice missing should return nil")
	}
	if cmd.GetInt32Slice("missing") != nil {
		t.Error("GetInt32Slice missing should return nil")
	}
	if cmd.GetInt64Slice("missing") != nil {
		t.Error("GetInt64Slice missing should return nil")
	}
	if cmd.GetUintSlice("missing") != nil {
		t.Error("GetUintSlice missing should return nil")
	}
	if cmd.GetUint8Slice("missing") != nil {
		t.Error("GetUint8Slice missing should return nil")
	}
	if cmd.GetUint16Slice("missing") != nil {
		t.Error("GetUint16Slice missing should return nil")
	}
	if cmd.GetUint32Slice("missing") != nil {
		t.Error("GetUint32Slice missing should return nil")
	}
	if cmd.GetUint64Slice("missing") != nil {
		t.Error("GetUint64Slice missing should return nil")
	}
	if cmd.GetFloat32Slice("missing") != nil {
		t.Error("GetFloat32Slice missing should return nil")
	}
	if cmd.GetFloat64Slice("missing") != nil {
		t.Error("GetFloat64Slice missing should return nil")
	}
}

func TestGetters_SliceFlags(t *testing.T) {
	cmd := &Command{
		Name: "test",
		Flags: []Flag{
			&Int8SliceFlag{Name: "i8"},
			&Int16SliceFlag{Name: "i16"},
			&Int32SliceFlag{Name: "i32"},
			&Int64SliceFlag{Name: "i64"},
			&UintSliceFlag{Name: "u"},
			&Uint8SliceFlag{Name: "u8"},
			&Uint16SliceFlag{Name: "u16"},
			&Uint32SliceFlag{Name: "u32"},
			&Uint64SliceFlag{Name: "u64"},
			&Float32SliceFlag{Name: "f32"},
			&Float64SliceFlag{Name: "f64"},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if s := cmd.GetInt8Slice("i8"); len(s) != 2 || s[0] != 1 || s[1] != 2 {
				t.Errorf("GetInt8Slice: got %v", s)
			}
			if s := cmd.GetInt16Slice("i16"); len(s) != 2 || s[0] != 3 || s[1] != 4 {
				t.Errorf("GetInt16Slice: got %v", s)
			}
			if s := cmd.GetInt32Slice("i32"); len(s) != 2 || s[0] != 5 || s[1] != 6 {
				t.Errorf("GetInt32Slice: got %v", s)
			}
			if s := cmd.GetInt64Slice("i64"); len(s) != 2 || s[0] != 7 || s[1] != 8 {
				t.Errorf("GetInt64Slice: got %v", s)
			}
			if s := cmd.GetUintSlice("u"); len(s) != 2 || s[0] != 9 || s[1] != 10 {
				t.Errorf("GetUintSlice: got %v", s)
			}
			if s := cmd.GetUint8Slice("u8"); len(s) != 2 || s[0] != 11 || s[1] != 12 {
				t.Errorf("GetUint8Slice: got %v", s)
			}
			if s := cmd.GetUint16Slice("u16"); len(s) != 2 || s[0] != 13 || s[1] != 14 {
				t.Errorf("GetUint16Slice: got %v", s)
			}
			if s := cmd.GetUint32Slice("u32"); len(s) != 2 || s[0] != 15 || s[1] != 16 {
				t.Errorf("GetUint32Slice: got %v", s)
			}
			if s := cmd.GetUint64Slice("u64"); len(s) != 2 || s[0] != 17 || s[1] != 18 {
				t.Errorf("GetUint64Slice: got %v", s)
			}
			if s := cmd.GetFloat32Slice("f32"); len(s) != 2 || s[0] != 1.1 || s[1] != 2.2 {
				t.Errorf("GetFloat32Slice: got %v", s)
			}
			if s := cmd.GetFloat64Slice("f64"); len(s) != 2 || s[0] != 3.3 || s[1] != 4.4 {
				t.Errorf("GetFloat64Slice: got %v", s)
			}
			return nil
		},
	}

	os.Args = []string{"test",
		"--i8", "1", "--i8", "2",
		"--i16", "3", "--i16", "4",
		"--i32", "5", "--i32", "6",
		"--i64", "7", "--i64", "8",
		"--u", "9", "--u", "10",
		"--u8", "11", "--u8", "12",
		"--u16", "13", "--u16", "14",
		"--u32", "15", "--u32", "16",
		"--u64", "17", "--u64", "18",
		"--f32", "1.1", "--f32", "2.2",
		"--f64", "3.3", "--f64", "4.4",
	}
	if err := cmd.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetters_ArgTypes(t *testing.T) {
	var (
		strArg     string
		intArg     int
		int8Arg    int8
		int16Arg   int16
		int32Arg   int32
		int64Arg   int64
		uintArg    uint
		uint8Arg   uint8
		uint16Arg  uint16
		uint32Arg  uint32
		uint64Arg  uint64
		float32Arg float32
		float64Arg float64
		boolArg    bool
	)

	cmd := &Command{
		Name: "test",
		Arguments: []Argument{
			&StringArg{Name: "str", Required: true, AssignTo: &strArg},
			&IntArg{Name: "int", Required: true, AssignTo: &intArg},
			&Int8Arg{Name: "int8", Required: true, AssignTo: &int8Arg},
			&Int16Arg{Name: "int16", Required: true, AssignTo: &int16Arg},
			&Int32Arg{Name: "int32", Required: true, AssignTo: &int32Arg},
			&Int64Arg{Name: "int64", Required: true, AssignTo: &int64Arg},
			&UintArg{Name: "uint", Required: true, AssignTo: &uintArg},
			&Uint8Arg{Name: "uint8", Required: true, AssignTo: &uint8Arg},
			&Uint16Arg{Name: "uint16", Required: true, AssignTo: &uint16Arg},
			&Uint32Arg{Name: "uint32", Required: true, AssignTo: &uint32Arg},
			&Uint64Arg{Name: "uint64", Required: true, AssignTo: &uint64Arg},
			&Float32Arg{Name: "float32", Required: true, AssignTo: &float32Arg},
			&Float64Arg{Name: "float64", Required: true, AssignTo: &float64Arg},
			&BoolArg{Name: "bool", Required: true, AssignTo: &boolArg},
		},
		Run: func(ctx context.Context, cmd *Command) error {
			if cmd.GetStringArg("str") != "hello" {
				t.Errorf("GetStringArg: got %q", cmd.GetStringArg("str"))
			}
			if cmd.GetIntArg("int") != 1 {
				t.Errorf("GetIntArg: got %d", cmd.GetIntArg("int"))
			}
			if cmd.GetInt8Arg("int8") != 2 {
				t.Errorf("GetInt8Arg: got %d", cmd.GetInt8Arg("int8"))
			}
			if cmd.GetInt16Arg("int16") != 3 {
				t.Errorf("GetInt16Arg: got %d", cmd.GetInt16Arg("int16"))
			}
			if cmd.GetInt32Arg("int32") != 4 {
				t.Errorf("GetInt32Arg: got %d", cmd.GetInt32Arg("int32"))
			}
			if cmd.GetInt64Arg("int64") != 5 {
				t.Errorf("GetInt64Arg: got %d", cmd.GetInt64Arg("int64"))
			}
			if cmd.GetUintArg("uint") != 6 {
				t.Errorf("GetUintArg: got %d", cmd.GetUintArg("uint"))
			}
			if cmd.GetUint8Arg("uint8") != 7 {
				t.Errorf("GetUint8Arg: got %d", cmd.GetUint8Arg("uint8"))
			}
			if cmd.GetUint16Arg("uint16") != 8 {
				t.Errorf("GetUint16Arg: got %d", cmd.GetUint16Arg("uint16"))
			}
			if cmd.GetUint32Arg("uint32") != 9 {
				t.Errorf("GetUint32Arg: got %d", cmd.GetUint32Arg("uint32"))
			}
			if cmd.GetUint64Arg("uint64") != 10 {
				t.Errorf("GetUint64Arg: got %d", cmd.GetUint64Arg("uint64"))
			}
			if cmd.GetFloat32Arg("float32") != 1.5 {
				t.Errorf("GetFloat32Arg: got %f", cmd.GetFloat32Arg("float32"))
			}
			if cmd.GetFloat64Arg("float64") != 2.5 {
				t.Errorf("GetFloat64Arg: got %f", cmd.GetFloat64Arg("float64"))
			}
			if !cmd.GetBoolArg("bool") {
				t.Errorf("GetBoolArg: got false")
			}
			return nil
		},
	}

	os.Args = []string{"test", "hello", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "1.5", "2.5", "true"}
	if err := cmd.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetters_ArgMissingReturnsZero(t *testing.T) {
	cmd := &Command{
		Name: "test",
		Run:  func(ctx context.Context, cmd *Command) error { return nil },
	}
	os.Args = []string{"test"}
	if err := cmd.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmd.GetStringArg("missing") != "" {
		t.Error("GetStringArg missing should return empty string")
	}
	if cmd.GetBoolArg("missing") {
		t.Error("GetBoolArg missing should return false")
	}
	if cmd.GetInt64Arg("missing") != 0 {
		t.Error("GetInt64Arg missing should return 0")
	}
	if cmd.GetIntArg("missing") != 0 {
		t.Error("GetIntArg missing should return 0")
	}
	if cmd.GetInt32Arg("missing") != 0 {
		t.Error("GetInt32Arg missing should return 0")
	}
	if cmd.GetInt16Arg("missing") != 0 {
		t.Error("GetInt16Arg missing should return 0")
	}
	if cmd.GetInt8Arg("missing") != 0 {
		t.Error("GetInt8Arg missing should return 0")
	}
	if cmd.GetUint64Arg("missing") != 0 {
		t.Error("GetUint64Arg missing should return 0")
	}
	if cmd.GetUintArg("missing") != 0 {
		t.Error("GetUintArg missing should return 0")
	}
	if cmd.GetUint32Arg("missing") != 0 {
		t.Error("GetUint32Arg missing should return 0")
	}
	if cmd.GetUint16Arg("missing") != 0 {
		t.Error("GetUint16Arg missing should return 0")
	}
	if cmd.GetUint8Arg("missing") != 0 {
		t.Error("GetUint8Arg missing should return 0")
	}
	if cmd.GetFloat64Arg("missing") != 0 {
		t.Error("GetFloat64Arg missing should return 0")
	}
	if cmd.GetFloat32Arg("missing") != 0 {
		t.Error("GetFloat32Arg missing should return 0")
	}
}

func TestReloadFlags(t *testing.T) {
	var value string
	cmd := &Command{
		Name: "test",
		Flags: []Flag{
			&StringFlag{Name: "flag", AssignTo: &value},
		},
		Run: func(ctx context.Context, cmd *Command) error { return nil },
	}

	os.Args = []string{"test", "--flag", "initial"}
	if err := cmd.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "initial" {
		t.Fatalf("expected 'initial', got %q", value)
	}

	os.Args = []string{"test", "--flag", "reloaded"}
	if err := cmd.ReloadFlags(); err != nil {
		t.Fatalf("ReloadFlags error: %v", err)
	}
	if value != "reloaded" {
		t.Fatalf("expected 'reloaded', got %q", value)
	}
}

func TestGetRootCmd(t *testing.T) {
	var capturedCmd *Command

	root := &Command{
		Name: "root",
		Commands: []*Command{
			{
				Name: "child",
				Run: func(ctx context.Context, cmd *Command) error {
					capturedCmd = cmd
					return nil
				},
			},
		},
	}

	os.Args = []string{"root", "child"}
	if err := root.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedCmd == nil {
		t.Fatal("command was not executed")
	}
	if got := capturedCmd.GetRootCmd(); got != root {
		t.Errorf("GetRootCmd: expected root command, got %v", got)
	}
}

func TestGetRootCmd_NoChain(t *testing.T) {
	cmd := &Command{Name: "solo"}
	if got := cmd.GetRootCmd(); got != cmd {
		t.Error("GetRootCmd on command with no chain should return itself")
	}
}
