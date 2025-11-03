package app

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/JaceTheGrayOne/ARI-S/internal/config"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// App is the main application service managing lifecycle, configuration,
// and file dialogs. It is registered as a Wails service and bound to the
// frontend for UI interactions.
//
// App handles:
//   - Configuration loading and persistence
//   - Windows folder/file selection dialogs
//   - Path validation and directory utilities
//
// The zero value is not usable; instances must be created with [NewApp].
type App struct {
	config  *config.Config
	ctx     context.Context
	depsDir string // Path to extracted dependencies
}

// NewApp creates a new App instance ready for use with Wails.
// The returned App has no configuration loaded; call LoadConfiguration
// after creation to restore saved settings.
func NewApp() *App {
	return &App{}
}

// LoadConfiguration loads the application configuration from disk and stores
// it in the App instance. If the config file does not exist or cannot be read,
// a default configuration is created instead. This method should be called
// once during application initialization before any other App methods.
func (a *App) LoadConfiguration() error {
	configPath := filepath.Join(getAppDataDir(), "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		cfg = config.NewDefaultConfig()
	}
	a.config = cfg

	// Log successfully loaded paths for debugging
	log.Printf("Configuration loaded from: %s", configPath)
	if cfg != nil {
		pathCount := 0
		for key, path := range cfg.LastUsedPaths {
			if path != "" {
				pathCount++
				log.Printf("  Loaded path [%s]: %s", key, path)
			}
		}
		if pathCount > 0 {
			log.Printf("Successfully loaded %d saved paths", pathCount)
		} else {
			log.Printf("No saved paths found, starting with empty configuration")
		}
	}
	return nil
}

// WailsShutdown is called by Wails when the application is closing.
// It persists the current configuration to disk. Any errors during save
// are logged but do not prevent shutdown.
func (a *App) WailsShutdown(ctx context.Context) {
	// Save configuration
	if a.config != nil {
		configPath := filepath.Join(getAppDataDir(), "config.json")
		if err := config.SaveConfig(configPath, a.config); err != nil {
			log.Printf("Failed to save config: %v", err)
		}
	}
}

// getAppDataDir returns the application data directory using cross-platform standard
func getAppDataDir() string {
	// Use os.UserConfigDir() for cross-platform compatibility
	// Windows: C:\Users\<Username>\AppData\Roaming
	// macOS: /Users/<Username>/Library/Application Support
	// Linux: /home/<username>/.config
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Failed to get user config directory, falling back to current directory: %v", err)
		return "."
	}

	// Create ARI-S subdirectory within the config location
	appConfigDir := filepath.Join(configDir, "ARI-S")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		log.Printf("Failed to create config directory, falling back to current directory: %v", err)
		return "."
	}

	log.Printf("Config directory: %s", appConfigDir)
	return appConfigDir
}

// GetLastUsedPath retrieves the last used path for the given key from the
// configuration. Returns an empty string if the key does not exist or if
// the configuration is not loaded.
func (a *App) GetLastUsedPath(key string) string {
	if a.config == nil {
		return ""
	}
	return a.config.GetLastUsedPath(key)
}

// SetLastUsedPath stores the given path for the given key in the configuration
// and immediately saves it to disk. If the configuration is not loaded, this
// method logs an error and returns without saving.
func (a *App) SetLastUsedPath(key, path string) {
	if a.config == nil {
		log.Printf("ERROR: Config is nil in SetLastUsedPath - this should never happen!")
		return
	}
	a.config.SetLastUsedPath(key, path)

	// Save config immediately after setting path
	configPath := filepath.Join(getAppDataDir(), "config.json")
	if err := config.SaveConfig(configPath, a.config); err != nil {
		log.Printf("Failed to save config after setting path: %v", err)
	} else {
		log.Printf("Saved path for key '%s': %s", key, path)
	}
}

// GetPreference retrieves the preference value for the given key from the
// configuration. Returns an empty string if the key does not exist or if
// the configuration is not loaded.
func (a *App) GetPreference(key string) string {
	if a.config == nil {
		return ""
	}
	return a.config.GetPreference(key)
}

// SetPreference stores the given preference value for the given key in the
// configuration and immediately saves it to disk. If the configuration is not
// loaded, this method logs an error and returns without saving.
func (a *App) SetPreference(key, value string) {
	if a.config == nil {
		log.Printf("ERROR: Config is nil in SetPreference - this should never happen!")
		return
	}
	a.config.SetPreference(key, value)

	// Save config immediately after setting preference
	configPath := filepath.Join(getAppDataDir(), "config.json")
	if err := config.SaveConfig(configPath, a.config); err != nil {
		log.Printf("Failed to save config after setting preference: %v", err)
	} else {
		log.Printf("Saved preference '%s': %s", key, value)
	}
}

// ValidateDirectory reports whether the given path exists and is a directory.
// It returns false for empty paths, non-existent paths, or paths that are files.
func (a *App) ValidateDirectory(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// GetInitialConfig returns a combined map of all last-used paths and
// preferences from the configuration. This method is called by the frontend
// on startup to populate form fields. If the configuration is not loaded,
// it returns an empty map.
func (a *App) GetInitialConfig() map[string]string {
	if a.config == nil {
		return make(map[string]string)
	}

	// Return a copy of all stored paths and preferences combined
	result := make(map[string]string)

	// Add all last used paths
	for key, value := range a.config.LastUsedPaths {
		result[key] = value
	}

	// Add all preferences
	for key, value := range a.config.Preferences {
		result[key] = value
	}

	return result
}

// BrowseFolder opens a native Windows folder selection dialog and returns the
// selected path. The title parameter sets the dialog's title bar text. The key
// parameter is used to remember the last-used path for this operation and will
// default to that location in the dialog. Returns an empty string if the user
// cancels the dialog.
func (a *App) BrowseFolder(title string, key string) string {
	log.Printf("BrowseFolder called with title: %s, key: %s", title, key)

	// Get the last used path for this specific field
	var lastPath string
	if a.config != nil && key != "" {
		lastPath = a.config.GetLastUsedPath(key)
	}

	// If no last path, try some common defaults
	if lastPath == "" && a.config != nil {
		// Try to get a sensible default path from other fields
		lastPath = a.config.GetLastUsedPath("input_mod_folder")
		if lastPath == "" {
			lastPath = a.config.GetLastUsedPath("pak_output_dir")
		}
		if lastPath == "" {
			lastPath = a.config.GetLastUsedPath("extract_output_dir")
		}
	}

	// If still no last path, use a default
	if lastPath == "" {
		lastPath = "C:\\Users\\Public\\Documents"
	}

	// Use Wails v3 OpenFileDialog configured for directory selection
	// This will use the modern IFileOpenDialog on Windows with FOS_PICKFOLDERS flag
	result, err := application.OpenFileDialog().
		SetTitle(title).
		SetDirectory(lastPath).
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()

	if err != nil {
		// Dialog was cancelled or failed
		log.Printf("Folder selection cancelled or failed: %v", err)
		return ""
	}

	return result
}

// BrowseFile opens a native Windows file selection dialog and returns the
// selected file path. The title parameter sets the dialog's title text. The
// filter parameter specifies the file type filter in Windows format (e.g.,
// "USMAP Files\x00*.usmap\x00\x00"). The key parameter is used to remember
// the last-used path. Returns an empty string if the user cancels.
func (a *App) BrowseFile(title, filter, key string) string {
	log.Printf("BrowseFile called with title: %s, filter: %s, key: %s", title, filter, key)

	// Get last used path if config is available and key is provided
	var lastPath string
	if a.config != nil && key != "" {
		lastPath = a.config.GetLastUsedPath(key)
		log.Printf("Last used path for key '%s': %s", key, lastPath)
	}

	// Try to open Windows file dialog
	result := a.openWindowsFileDialog(title, filter, lastPath)
	return result
}

// openWindowsFileDialog opens a Windows file selection dialog
func (a *App) openWindowsFileDialog(title, filter, initialPath string) string {
	// Load comdlg32.dll
	comdlg32 := syscall.NewLazyDLL("comdlg32.dll")
	getOpenFileName := comdlg32.NewProc("GetOpenFileNameW")

	// Create buffer for selected file path
	fileBuffer := make([]uint16, 260)

	// Convert filter string to Windows format (e.g., "USMAP Files\x00*.usmap\x00All Files\x00*.*\x00\x00")
	var filterUTF16 *uint16
	if filter == "" {
		// Default filter for .usmap files
		filterStr := "USMAP Files\x00*.usmap\x00All Files\x00*.*\x00\x00"
		filterUTF16, _ = syscall.UTF16PtrFromString(filterStr)
	} else {
		filterUTF16, _ = syscall.UTF16PtrFromString(filter)
	}

	titleUTF16, _ := syscall.UTF16PtrFromString(title)

	// Set initial directory if provided
	var initialDirUTF16 *uint16
	if initialPath != "" {
		// Extract directory from file path
		initialDir := filepath.Dir(initialPath)
		initialDirUTF16, _ = syscall.UTF16PtrFromString(initialDir)
		log.Printf("Setting initial directory: %s", initialDir)
	}

	// Set up OPENFILENAME structure
	ofn := struct {
		lStructSize       uint32
		hwndOwner         uintptr
		hInstance         uintptr
		lpstrFilter       *uint16
		lpstrCustomFilter *uint16
		nMaxCustFilter    uint32
		nFilterIndex      uint32
		lpstrFile         *uint16
		nMaxFile          uint32
		lpstrFileTitle    *uint16
		nMaxFileTitle     uint32
		lpstrInitialDir   *uint16
		lpstrTitle        *uint16
		flags             uint32
		nFileOffset       uint16
		nFileExtension    uint16
		lpstrDefExt       *uint16
		lCustData         uintptr
		lpfnHook          uintptr
		lpTemplateName    *uint16
		pvReserved        uintptr
		dwReserved        uint32
		FlagsEx           uint32
	}{
		lStructSize: uint32(unsafe.Sizeof(struct {
			lStructSize       uint32
			hwndOwner         uintptr
			hInstance         uintptr
			lpstrFilter       *uint16
			lpstrCustomFilter *uint16
			nMaxCustFilter    uint32
			nFilterIndex      uint32
			lpstrFile         *uint16
			nMaxFile          uint32
			lpstrFileTitle    *uint16
			nMaxFileTitle     uint32
			lpstrInitialDir   *uint16
			lpstrTitle        *uint16
			flags             uint32
			nFileOffset       uint16
			nFileExtension    uint16
			lpstrDefExt       *uint16
			lCustData         uintptr
			lpfnHook          uintptr
			lpTemplateName    *uint16
			pvReserved        uintptr
			dwReserved        uint32
			FlagsEx           uint32
		}{})),
		hwndOwner:       0,
		lpstrFilter:     filterUTF16,
		lpstrFile:       &fileBuffer[0],
		nMaxFile:        260,
		lpstrInitialDir: initialDirUTF16,
		lpstrTitle:      titleUTF16,
		flags:           0x00000800 | 0x00001000, // OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST
	}

	// Call GetOpenFileName
	ret, _, _ := getOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "" // User cancelled or error
	}

	// Convert back to Go string
	return syscall.UTF16ToString(fileBuffer)
}
