package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/JaceTheGrayOne/ARI-S/internal/config"
)

func TestApplicationLifecycle_Success(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	app := NewApp()
	ctx := context.Background()

	// Simulate WailsInit
	app.ctx = ctx
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	app.config = cfg

	if app.config == nil {
		t.Error("Expected config to be initialized")
	}

	if app.config.Preferences == nil {
		t.Error("Expected preferences map to be initialized")
	}

	// Verify default preferences
	if ueVersion := app.config.GetPreference("ue_version"); ueVersion != "UE5_4" {
		t.Errorf("Expected default UE version 'UE5_4', got '%s'", ueVersion)
	}

	if theme := app.config.GetPreference("theme"); theme != "dark" {
		t.Errorf("Expected default theme 'dark', got '%s'", theme)
	}
}

func TestApplicationLifecycle_Shutdown_PersistsConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	app := NewApp()
	app.ctx = context.Background()
	app.config = config.NewDefaultConfig()
	app.config.SetLastUsedPath("test_key", "/test/path")
	app.config.SetPreference("test_pref", "test_value")

	if err := config.SaveConfig(configPath, app.config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Reload config to verify persistence
	reloadedConfig, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if path := reloadedConfig.GetLastUsedPath("test_key"); path != "/test/path" {
		t.Errorf("Expected path '/test/path', got '%s'", path)
	}

	if pref := reloadedConfig.GetPreference("test_pref"); pref != "test_value" {
		t.Errorf("Expected preference 'test_value', got '%s'", pref)
	}
}

func TestApplicationLifecycle_ConfigLoadError_UsesDefaults(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent", "config.json")

	cfg, err := config.LoadConfig(nonExistentPath)

	if err != nil {
		t.Errorf("Expected no error with non-existent config, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected default config to be created")
	}

	if cfg.GetPreference("ue_version") != "UE5_4" {
		t.Error("Expected default preferences in new config")
	}
}
