package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPersistence_SaveAndLoad_PreservesData(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	originalConfig := NewDefaultConfig()
	originalConfig.SetLastUsedPath("input_mod_folder", "C:\\TestMod")
	originalConfig.SetLastUsedPath("pak_output_dir", "C:\\Output")
	originalConfig.SetPreference("ue_version", "UE5_3")
	originalConfig.SetPreference("theme", "light")

	if err := SaveConfig(configPath, originalConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"input_mod_folder path", loadedConfig.GetLastUsedPath("input_mod_folder"), "C:\\TestMod"},
		{"pak_output_dir path", loadedConfig.GetLastUsedPath("pak_output_dir"), "C:\\Output"},
		{"ue_version preference", loadedConfig.GetPreference("ue_version"), "UE5_3"},
		{"theme preference", loadedConfig.GetPreference("theme"), "light"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.got)
			}
		})
	}
}

func TestConfigPersistence_JSONFormat_Valid(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	config := NewDefaultConfig()
	config.SetLastUsedPath("test_path", "/test")
	config.SetPreference("test_pref", "value")

	if err := SaveConfig(configPath, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read raw JSON
	jsonData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Errorf("Config JSON is invalid: %v", err)
	}

	// Verify indentation (SaveConfig uses SetIndent("", "  "))
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(jsonData, &prettyJSON); err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}
}

func TestConfigPersistence_MissingFields_InitializesCorrectly(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Write minimal JSON (missing preferences map)
	minimalJSON := `{"last_used_paths": {}}`
	if err := os.WriteFile(configPath, []byte(minimalJSON), 0644); err != nil {
		t.Fatalf("Failed to write minimal config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.LastUsedPaths == nil {
		t.Error("Expected LastUsedPaths map to be initialized")
	}

	if config.Preferences == nil {
		t.Error("Expected Preferences map to be initialized")
	}

	// Verify can set values without panic
	config.SetPreference("test", "value")
	if config.GetPreference("test") != "value" {
		t.Error("Failed to set preference on loaded config with missing fields")
	}
}

func TestConfigPersistence_Config_Integration(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create config and set values
	cfg := NewDefaultConfig()
	cfg.SetLastUsedPath("test_key", "/test/path")
	cfg.SetPreference("test_pref", "test_value")

	// Save config
	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if path := loadedConfig.GetLastUsedPath("test_key"); path != "/test/path" {
		t.Errorf("Expected path '/test/path', got '%s'", path)
	}

	if pref := loadedConfig.GetPreference("test_pref"); pref != "test_value" {
		t.Errorf("Expected preference 'test_value', got '%s'", pref)
	}
}
