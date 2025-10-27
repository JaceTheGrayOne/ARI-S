package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// App struct represents the main application
type App struct {
	config *Config
	ctx    context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// WailsInit is called when the app starts
func (a *App) WailsInit(ctx context.Context) {
	a.ctx = ctx

	// Load configuration
	configPath := filepath.Join(getAppDataDir(), "config.json")
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		config = NewDefaultConfig()
	}
	a.config = config
}

// WailsShutdown is called when the app is shutting down
func (a *App) WailsShutdown(ctx context.Context) {
	// Save configuration
	if a.config != nil {
		configPath := filepath.Join(getAppDataDir(), "config.json")
		if err := SaveConfig(configPath, a.config); err != nil {
			log.Printf("Failed to save config: %v", err)
		}
	}
}

// getAppDataDir returns the application data directory
func getAppDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, "AppData", "Local", "ARI-S")
}

// GetLastUsedPath returns the last used path for a given key
func (a *App) GetLastUsedPath(key string) string {
	if a.config == nil {
		return ""
	}
	return a.config.GetLastUsedPath(key)
}

// SetLastUsedPath sets the last used path for a given key
func (a *App) SetLastUsedPath(key, path string) {
	if a.config == nil {
		a.config = NewDefaultConfig()
	}
	a.config.SetLastUsedPath(key, path)
}

// GetPreference returns a preference value
func (a *App) GetPreference(key string) string {
	if a.config == nil {
		return ""
	}
	return a.config.GetPreference(key)
}

// SetPreference sets a preference value
func (a *App) SetPreference(key, value string) {
	if a.config == nil {
		a.config = NewDefaultConfig()
	}
	a.config.SetPreference(key, value)
}

// BrowseFolder opens a folder selection dialog using Windows API
func (a *App) BrowseFolder(title string) string {
	log.Printf("BrowseFolder called with title: %s", title)

	// Try to get the last used path for this field type
	var lastPath string
	if a.config != nil {
		// Try to get a sensible default path
		lastPath = a.config.GetLastUsedPath("input_mod_folder")
		if lastPath == "" {
			lastPath = a.config.GetLastUsedPath("pak_output_dir")
		}
		if lastPath == "" {
			lastPath = a.config.GetLastUsedPath("extract_output_dir")
		}
	}

	// If no last path, use a default
	if lastPath == "" {
		lastPath = "C:\\Users\\Public\\Documents"
	}

	// Try to open Windows folder dialog
	result := a.openWindowsFolderDialog(title, lastPath)
	if result != "" {
		return result
	}

	// Fallback to last used path if dialog fails
	return lastPath
}

// openWindowsFolderDialog opens a Windows folder selection dialog
func (a *App) openWindowsFolderDialog(title, initialPath string) string {
	// Windows API constants
	const (
		BIF_RETURNONLYFSDIRS = 0x00000001
		BIF_NEWDIALOGSTYLE   = 0x00000040
		BIF_EDITBOX          = 0x00000010
	)

	// Convert strings to UTF-16
	titleUTF16, _ := syscall.UTF16PtrFromString(title)
	_, _ = syscall.UTF16PtrFromString(initialPath) // For future use

	// Create buffer for selected path
	buffer := make([]uint16, 260)

	// Load shell32.dll
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shBrowseForFolder := shell32.NewProc("SHBrowseForFolderW")
	shGetPathFromIDList := shell32.NewProc("SHGetPathFromIDListW")

	// Set up BROWSEINFO structure
	bi := struct {
		hwndOwner      uintptr
		pidlRoot       uintptr
		pszDisplayName uintptr
		lpszTitle      uintptr
		ulFlags        uint32
		lpfn           uintptr
		lParam         uintptr
		iImage         int32
	}{
		hwndOwner:      0,
		pidlRoot:       0,
		pszDisplayName: uintptr(unsafe.Pointer(&buffer[0])),
		lpszTitle:      uintptr(unsafe.Pointer(titleUTF16)),
		ulFlags:        BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE | BIF_EDITBOX,
		lpfn:           0,
		lParam:         0,
		iImage:         0,
	}

	// Call SHBrowseForFolder
	ret, _, _ := shBrowseForFolder.Call(uintptr(unsafe.Pointer(&bi)))
	if ret == 0 {
		return "" // User cancelled
	}

	// Get the selected path
	pathBuffer := make([]uint16, 260)
	ret, _, _ = shGetPathFromIDList.Call(ret, uintptr(unsafe.Pointer(&pathBuffer[0])))
	if ret == 0 {
		return ""
	}

	// Convert back to Go string
	return syscall.UTF16ToString(pathBuffer)
}
