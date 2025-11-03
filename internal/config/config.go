package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config stores application settings including last-used paths for file/folder
// selection dialogs and user preferences such as the selected Unreal Engine
// version. The configuration is persisted to disk as JSON in the user's
// local AppData directory.
//
// Config is safe for concurrent use by multiple goroutines after initialization.
type Config struct {
	LastUsedPaths map[string]string `json:"last_used_paths"`
	Preferences   map[string]string `json:"preferences"`
}

// NewDefaultConfig creates a Config with default values. The returned Config
// has empty path maps and default preferences (UE version: UE5_4, theme: dark).
func NewDefaultConfig() *Config {
	return &Config{
		LastUsedPaths: make(map[string]string),
		Preferences: map[string]string{
			"ue_version": "UE5_4",
			"theme":      "dark",
		},
	}
}

// LoadConfig reads and decodes a Config from the JSON file at the given path.
// If the file does not exist, it returns a default Config with no error. If
// the file exists but cannot be read or parsed, it returns an error. The
// directory containing the file is created if it does not exist.
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

// SaveConfig writes the given Config to disk as JSON at the specified path.
// The file is written with indentation for human readability. The directory
// containing the file is created if it does not exist.
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

// GetLastUsedPath retrieves the last-used path for the given key.
// Returns an empty string if the key does not exist.
func (c *Config) GetLastUsedPath(key string) string {
	if path, exists := c.LastUsedPaths[key]; exists {
		return path
	}
	return ""
}

// SetLastUsedPath stores the given path for the given key. If the LastUsedPaths
// map is nil, it is initialized. This method does not persist the change to disk;
// call SaveConfig to write changes.
func (c *Config) SetLastUsedPath(key, path string) {
	if c.LastUsedPaths == nil {
		c.LastUsedPaths = make(map[string]string)
	}
	c.LastUsedPaths[key] = path
}

// GetPreference retrieves the preference value for the given key.
// Returns an empty string if the key does not exist.
func (c *Config) GetPreference(key string) string {
	if value, exists := c.Preferences[key]; exists {
		return value
	}
	return ""
}

// SetPreference stores the given value for the given preference key. If the
// Preferences map is nil, it is initialized. This method does not persist the
// change to disk; call SaveConfig to write changes.
func (c *Config) SetPreference(key, value string) {
	if c.Preferences == nil {
		c.Preferences = make(map[string]string)
	}
	c.Preferences[key] = value
}
