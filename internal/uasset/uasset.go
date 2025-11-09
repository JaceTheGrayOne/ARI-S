//go:build !grpc
// +build !grpc

package uasset

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
)

// UAssetService provides UAsset serialization operations using an external
// UAssetBridge.exe process via IPC. It supports exporting .uasset/.uexp files
// to JSON and importing JSON back to binary format.
//
// UAssetService is safe for concurrent use by multiple goroutines.
type UAssetService struct {
	app     *app.App
	depsDir string // Path to extracted dependencies
}

// NewUAssetService creates a new UAssetService using the UAssetBridge.exe
// binary located in the given depsDir. The depsDir should contain a "UAssetAPI"
// subdirectory with UAssetBridge.exe and all required .NET runtime DLLs.
func NewUAssetService(a *app.App, depsDir string) *UAssetService {
	return &UAssetService{
		app:     a,
		depsDir: depsDir,
	}
}

// ExportUAssets converts all .uasset/.uexp files in folderPath to JSON format.
// If mappingsPath is provided and valid, it is passed to the bridge for
// unversioned property resolution. The operation spawns UAssetBridge.exe as
// a subprocess and waits for it to complete.
func (u *UAssetService) ExportUAssets(ctx context.Context, folderPath, mappingsPath string) UAssetResult {
	return u.runUAssetOperation(ctx, "export", folderPath, mappingsPath)
}

// ImportUAssets converts all .json files in folderPath back to .uasset/.uexp
// format. If mappingsPath is provided and valid, it is passed to the bridge
// for unversioned property serialization. The operation spawns UAssetBridge.exe
// as a subprocess and waits for it to complete.
func (u *UAssetService) ImportUAssets(ctx context.Context, folderPath, mappingsPath string) UAssetResult {
	return u.runUAssetOperation(ctx, "import", folderPath, mappingsPath)
}

func (u *UAssetService) runUAssetOperation(ctx context.Context, command, folderPath, mappingsPath string) UAssetResult {
	startTime := time.Now()

	// Use the extracted dependencies directory
	// Look for UAssetBridge.exe in the UAssetAPI subfolder
	bridgePath := filepath.Join(u.depsDir, "UAssetAPI", "UAssetBridge.exe")

	// Check if UAssetBridge.exe exists
	if _, err := os.Stat(bridgePath); os.IsNotExist(err) {
		return UAssetResult{
			Success: false,
			Error:   fmt.Sprintf("UAssetBridge.exe not found at: %s", bridgePath),
		}
	}

	// Check if folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return UAssetResult{
			Success: false,
			Error:   fmt.Sprintf("Folder does not exist: %s", folderPath),
		}
	}

	// Build command arguments
	args := []string{command, folderPath}

	// Add mappings path if provided (needed for both export and import)
	if mappingsPath != "" {
		// Check if mappings file exists
		if _, err := os.Stat(mappingsPath); os.IsNotExist(err) {
			return UAssetResult{
				Success: false,
				Error:   fmt.Sprintf("Mappings file does not exist: %s", mappingsPath),
			}
		}
		args = append(args, mappingsPath)
	}

	// Create command
	cmd := exec.CommandContext(ctx, bridgePath, args...)

	// Set working directory to the directory containing uasset_bridge.exe
	cmd.Dir = filepath.Dir(bridgePath)

	// Capture output
	output, err := cmd.CombinedOutput()

	duration := time.Since(startTime)

	result := UAssetResult{
		Duration: duration.String(),
		Output:   string(output),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = fmt.Sprintf("UAsset %s operation failed", command)
	} else {
		result.Success = true
		result.Message = fmt.Sprintf("UAsset %s operation completed successfully", command)

		// Try to extract file count from output
		result.FilesProcessed = u.extractFileCount(string(output))
	}

	return result
}

// extractFileCount attempts to extract the number of files processed from output
func (u *UAssetService) extractFileCount(output string) int {
	// Look for patterns like "Processed X files" or "X files processed"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)
		if strings.Contains(line, "processed") && strings.Contains(line, "files") {
			// Try to extract number from the line
			words := strings.Fields(line)
			for i, word := range words {
				if word == "processed" && i > 0 {
					// Previous word might be the number
					var num int
					if _, err := fmt.Sscanf(words[i-1], "%d", &num); err == nil && num > 0 {
						return num
					}
				}
			}
		}
	}
	return 0
}

// CountUAssetFiles recursively counts .uasset and .uexp files in the given
// directory. Returns (uassetCount, uexpCount, error). If the directory does
// not exist, it returns (0, 0, error).
func (u *UAssetService) CountUAssetFiles(ctx context.Context, folderPath string) (int, int, error) {
	// Check if directory exists first
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return 0, 0, fmt.Errorf("directory does not exist: %s", folderPath)
	}

	var uassetCount, uexpCount int

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".uasset":
				uassetCount++
			case ".uexp":
				uexpCount++
			}
		}

		return nil
	})

	return uassetCount, uexpCount, err
}

// CountJSONFiles recursively counts .json files in the given directory.
// If the directory does not exist, it returns (0, error).
func (u *UAssetService) CountJSONFiles(ctx context.Context, folderPath string) (int, error) {
	// Check if directory exists first
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("directory does not exist: %s", folderPath)
	}

	var jsonCount int

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".json" {
				jsonCount++
			}
		}

		return nil
	})

	return jsonCount, err
}
