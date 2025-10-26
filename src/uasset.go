package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// UAssetService handles UAsset operations
type UAssetService struct {
	app *App
}

// UAssetOperation represents a UAsset operation
type UAssetOperation struct {
	Command    string `json:"command"`
	FolderPath string `json:"folder_path"`
}

// UAssetResult represents the result of a UAsset operation
type UAssetResult struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	Output         string `json:"output"`
	Error          string `json:"error"`
	Duration       string `json:"duration"`
	FilesProcessed int    `json:"files_processed"`
}

// NewUAssetService creates a new UAssetService
func NewUAssetService(app *App) *UAssetService {
	return &UAssetService{app: app}
}

// ExportUAssets exports .uasset/.uexp files to JSON
func (u *UAssetService) ExportUAssets(ctx context.Context, folderPath string) UAssetResult {
	return u.runUAssetOperation(ctx, "export", folderPath)
}

// ImportUAssets imports JSON files back to .uasset/.uexp
func (u *UAssetService) ImportUAssets(ctx context.Context, folderPath string) UAssetResult {
	return u.runUAssetOperation(ctx, "import", folderPath)
}

// runUAssetOperation executes a UAsset operation
func (u *UAssetService) runUAssetOperation(ctx context.Context, command, folderPath string) UAssetResult {
	startTime := time.Now()

	// Get the directory where the application executable is located
	execPath, err := os.Executable()
	if err != nil {
		return UAssetResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to get executable path: %v", err),
		}
	}
	execDir := filepath.Dir(execPath)

	// Look for UAssetBridge.exe in a "UAssetAPI" folder next to the executable
	bridgePath := filepath.Join(execDir, "UAssetAPI", "UAssetBridge.exe")

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

// CountUAssetFiles counts .uasset and .uexp files in a directory
func (u *UAssetService) CountUAssetFiles(ctx context.Context, folderPath string) (int, int, error) {
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

// CountJSONFiles counts .json files in a directory
func (u *UAssetService) CountJSONFiles(ctx context.Context, folderPath string) (int, error) {
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
