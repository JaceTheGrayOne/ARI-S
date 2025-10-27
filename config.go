package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	RetocPath        string            `json:"retoc_path"`
	UAssetBridgePath string            `json:"uasset_bridge_path"`
	LastUsedPaths    map[string]string `json:"last_used_paths"`
	Preferences      map[string]string `json:"preferences"`
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig() *Config {
	return &Config{
		RetocPath:        "retoc/retoc.exe",
		UAssetBridgePath: "UAssetAPI/UAssetBridge.exe", // Relative to ARI-S executable directory
		LastUsedPaths:    make(map[string]string),
		Preferences: map[string]string{
			"ue_version": "UE5_4",
			"theme":      "dark",
		},
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filePath string) (*Config, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewDefaultConfig(), nil
		}
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	// Ensure maps are initialized
	if config.LastUsedPaths == nil {
		config.LastUsedPaths = make(map[string]string)
	}
	if config.Preferences == nil {
		config.Preferences = make(map[string]string)
	}

	return config, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(filePath string, config *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

// GetLastUsedPath returns the last used path for a given key
func (c *Config) GetLastUsedPath(key string) string {
	if path, exists := c.LastUsedPaths[key]; exists {
		return path
	}
	return ""
}

// SetLastUsedPath sets the last used path for a given key
func (c *Config) SetLastUsedPath(key, path string) {
	if c.LastUsedPaths == nil {
		c.LastUsedPaths = make(map[string]string)
	}
	c.LastUsedPaths[key] = path
}

// GetPreference returns a preference value
func (c *Config) GetPreference(key string) string {
	if value, exists := c.Preferences[key]; exists {
		return value
	}
	return ""
}

// SetPreference sets a preference value
func (c *Config) SetPreference(key, value string) {
	if c.Preferences == nil {
		c.Preferences = make(map[string]string)
	}
	c.Preferences[key] = value
}
