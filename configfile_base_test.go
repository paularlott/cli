package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func newJSONConfigBase(t *testing.T, content string) (*ConfigFileBase, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	cfg := &ConfigFileBase{}
	cfg.InitConfigFile()
	cfg.FileName = &path
	cfg.Unmarshal = func(data []byte, v any) error { return json.Unmarshal(data, v) }
	cfg.Marshal = func(v any) ([]byte, error) { return json.Marshal(v) }
	return cfg, path
}

func TestConfigFileBase_LoadData(t *testing.T) {
	cfg, _ := newJSONConfigBase(t, `{"key":"value"}`)

	if err := cfg.LoadData(); err != nil {
		t.Fatalf("LoadData error: %v", err)
	}
	// Second call should be a no-op (already loaded)
	if err := cfg.LoadData(); err != nil {
		t.Fatalf("second LoadData error: %v", err)
	}
}

func TestConfigFileBase_LoadData_NotFound(t *testing.T) {
	path := "/nonexistent/path/config.json"
	cfg := &ConfigFileBase{}
	cfg.InitConfigFile()
	cfg.FileName = &path
	cfg.Unmarshal = func(data []byte, v any) error { return json.Unmarshal(data, v) }
	cfg.Marshal = func(v any) ([]byte, error) { return json.Marshal(v) }

	err := cfg.LoadData()
	if err != ConfigFileNotFoundError {
		t.Errorf("expected ConfigFileNotFoundError, got %v", err)
	}
}

func TestConfigFileBase_GetValue(t *testing.T) {
	cfg, _ := newJSONConfigBase(t, `{"name":"alice","nested":{"age":30}}`)

	v, ok := cfg.GetValue("name")
	if !ok || v != "alice" {
		t.Errorf("GetValue(name): got %v, %v", v, ok)
	}

	v, ok = cfg.GetValue("nested.age")
	if !ok {
		t.Error("GetValue(nested.age): not found")
	}
	// JSON numbers decode as float64
	if v.(float64) != 30 {
		t.Errorf("GetValue(nested.age): got %v", v)
	}

	_, ok = cfg.GetValue("missing")
	if ok {
		t.Error("GetValue(missing): should not be found")
	}
}

func TestConfigFileBase_GetKeys(t *testing.T) {
	cfg, _ := newJSONConfigBase(t, `{"a":1,"b":2,"c":3}`)

	keys := cfg.GetKeys("")
	if len(keys) != 3 {
		t.Errorf("GetKeys: expected 3 keys, got %d: %v", len(keys), keys)
	}
}

func TestConfigFileBase_SetValue_DeleteKey(t *testing.T) {
	cfg, _ := newJSONConfigBase(t, `{"key":"original"}`)

	// Load first so the in-memory data is populated
	if err := cfg.LoadData(); err != nil {
		t.Fatalf("LoadData error: %v", err)
	}

	if err := cfg.SetValue("key", "updated"); err != nil {
		t.Fatalf("SetValue error: %v", err)
	}
	v, ok := cfg.GetValue("key")
	if !ok || v != "updated" {
		t.Errorf("after SetValue: got %v, %v", v, ok)
	}

	if err := cfg.DeleteKey("key"); err != nil {
		t.Fatalf("DeleteKey error: %v", err)
	}
	_, ok = cfg.GetValue("key")
	if ok {
		t.Error("after DeleteKey: key should not exist")
	}
}

func TestConfigFileBase_SetValue_NestedCreate(t *testing.T) {
	cfg, _ := newJSONConfigBase(t, `{}`)

	if err := cfg.SetValue("a.b.c", "deep"); err != nil {
		t.Fatalf("SetValue nested error: %v", err)
	}
	v, ok := cfg.GetValue("a.b.c")
	if !ok || v != "deep" {
		t.Errorf("GetValue(a.b.c): got %v, %v", v, ok)
	}
}

func TestConfigFileBase_Save(t *testing.T) {
	cfg, path := newJSONConfigBase(t, `{"key":"value"}`)

	if err := cfg.SetValue("key", "saved"); err != nil {
		t.Fatalf("SetValue error: %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Reload and verify
	cfg2, _ := newJSONConfigBase(t, "")
	cfg2.FileName = &path
	v, ok := cfg2.GetValue("key")
	if !ok || v != "saved" {
		t.Errorf("after Save+reload: got %v, %v", v, ok)
	}
}

func TestConfigFileBase_FileUsed(t *testing.T) {
	cfg, path := newJSONConfigBase(t, `{"key":"value"}`)

	used := cfg.FileUsed()
	if used != path {
		t.Errorf("FileUsed: got %q, want %q", used, path)
	}
}

func TestConfigFileBase_FileUsed_NotFound(t *testing.T) {
	path := "/nonexistent/config.json"
	cfg := &ConfigFileBase{}
	cfg.InitConfigFile()
	cfg.FileName = &path
	cfg.Unmarshal = func(data []byte, v any) error { return json.Unmarshal(data, v) }
	cfg.Marshal = func(v any) ([]byte, error) { return json.Marshal(v) }

	used := cfg.FileUsed()
	if used != "" {
		t.Errorf("FileUsed on missing file: got %q, want empty", used)
	}
}

func TestConfigFileBase_SearchForDotFile(t *testing.T) {
	dir := t.TempDir()
	dotPath := filepath.Join(dir, ".myconfig")
	if err := os.WriteFile(dotPath, []byte(`{"x":1}`), 0644); err != nil {
		t.Fatalf("failed to write dot config: %v", err)
	}

	name := ".myconfig"
	cfg := &ConfigFileBase{}
	cfg.InitConfigFile()
	cfg.FileName = &name
	cfg.SearchPath = func() []string { return []string{dir} }
	cfg.Unmarshal = func(data []byte, v any) error { return json.Unmarshal(data, v) }
	cfg.Marshal = func(v any) ([]byte, error) { return json.Marshal(v) }

	v, ok := cfg.GetValue("x")
	if !ok {
		t.Fatal("GetValue(x): not found")
	}
	if v.(float64) != 1 {
		t.Errorf("GetValue(x): got %v", v)
	}
}

func TestConfigFileBase_Save_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.json")

	cfg := &ConfigFileBase{}
	cfg.InitConfigFile()
	cfg.FileName = &path
	cfg.Unmarshal = func(data []byte, v any) error { return json.Unmarshal(data, v) }
	cfg.Marshal = func(v any) ([]byte, error) { return json.Marshal(v) }

	cfg.SetValue("hello", "world")
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save to new file error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if len(data) == 0 {
		t.Error("saved file is empty")
	}
}
