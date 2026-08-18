package glowconfig

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetString(t *testing.T) {
	// Create a dummy config.yaml
	configContent := map[string]any{
		"foo": "bar",
		"port": 8080,
		"debug": true,
		"nested": map[string]any{
			"key": "value",
		},
	}
	data, _ := yaml.Marshal(configContent)
	if err := os.WriteFile("config.yaml", data, 0644); err != nil {
		t.Fatalf("Failed to create config.yaml: %v", err)
	}
	defer os.Remove("config.yaml")

	// Reset global state for clean test
	Reload()

	// Test package-level functions
	if GetString("foo") != "bar" {
		t.Errorf("Expected foo=bar, got %v", GetString("foo"))
	}
	if GetInt("port") != 8080 {
		t.Errorf("Expected port=8080, got %v", GetInt("port"))
	}
	if !GetBool("debug") {
		t.Errorf("Expected debug=true, got %v", GetBool("debug"))
	}
	if GetString("nested.key") != "value" {
		t.Errorf("Expected nested.key=value, got %v", GetString("nested.key"))
	}
}

func TestLazyLoad(t *testing.T) {
	// Create config
	configContent := map[string]any{
		"test": "value",
	}
	data, _ := yaml.Marshal(configContent)
	if err := os.WriteFile("config.yaml", data, 0644); err != nil {
		t.Fatalf("Failed to create config.yaml: %v", err)
	}
	defer os.Remove("config.yaml")

	// Reset global state
	globalConfig = nil
	initialized = false

	// First call should load the file
	val1 := GetString("test")
	if val1 != "value" {
		t.Errorf("Expected test=value, got %v", val1)
	}

	// Second call should use cached value
	val2 := GetString("test")
	if val2 != "value" {
		t.Errorf("Expected test=value, got %v", val2)
	}

	// Verify caching works
	if !initialized {
		t.Error("Expected config to be initialized after second GetString()")
	}
}

func TestReload(t *testing.T) {
	// Create initial config
	configContent := map[string]any{
		"key": "value1",
	}
	data, _ := yaml.Marshal(configContent)
	if err := os.WriteFile("config.yaml", data, 0644); err != nil {
		t.Fatalf("Failed to create config.yaml: %v", err)
	}
	defer os.Remove("config.yaml")

	// Reset and load
	globalConfig = nil
	initialized = false

	if GetString("key") != "value1" {
		t.Errorf("Expected key=value1, got %v", GetString("key"))
	}

	// Update config file
	configContent["key"] = "value2"
	data, _ = json.Marshal(configContent)
	if err := os.WriteFile("config.yaml", data, 0644); err != nil {
		t.Fatalf("Failed to update config.yaml: %v", err)
	}

	// Reload
	if err := Reload(); err != nil {
		t.Fatalf("Reload() failed: %v", err)
	}

	if GetString("key") != "value2" {
		t.Errorf("Expected key=value2 after reload, got %v", GetString("key"))
	}
}
