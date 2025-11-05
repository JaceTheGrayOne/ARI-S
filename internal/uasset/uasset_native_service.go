//go:build cgo
// +build cgo

package uasset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
)

// UAssetService handles UAsset operations using the NativeAOT library via CGO.
// This replaces the IPC-based service with direct in-process calls via CGO.
// When built with -tags cgo, this is the native implementation.
type UAssetService struct {
	app *app.App
	api *NativeUAssetAPI
}

// NewUAssetService creates a new UAssetService using the NativeAOT CGO implementation.
// This service uses CGO to call the NativeAOT UAssetBridge library directly.
// The depsDir parameter is ignored in the CGO build.
func NewUAssetService(a *app.App, depsDir string) *UAssetService {
	return &UAssetService{
		app: a,
		api: NewNativeUAssetAPI(),
	}
}

// ExportUAssets exports .uasset/.uexp files to JSON using the native library.
// This method signature matches the IPC-based service for drop-in replacement.
func (u *UAssetService) ExportUAssets(ctx context.Context, folderPath, mappingsPath string) UAssetResult {
	startTime := time.Now()

	// Validate folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return UAssetResult{
			Success: false,
			Error:   fmt.Sprintf("Folder does not exist: %s", folderPath),
		}
	}

	// Load mappings if provided
	var mappingsHandle MappingsHandle
	var mappingsLoaded bool

	if mappingsPath != "" {
		if _, err := os.Stat(mappingsPath); os.IsNotExist(err) {
			return UAssetResult{
				Success: false,
				Error:   fmt.Sprintf("Mappings file does not exist: %s", mappingsPath),
			}
		}

		handle, err := u.api.LoadMappings(mappingsPath)
		if err != nil {
			return UAssetResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to load mappings: %v", err),
			}
		}

		mappingsHandle = handle
		mappingsLoaded = true
		defer u.api.FreeMappings(mappingsHandle)
	}

	// Find all .uasset files
	var uassetFiles []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".uasset" {
			uassetFiles = append(uassetFiles, path)
		}

		return nil
	})

	if err != nil {
		return UAssetResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to scan directory: %v", err),
		}
	}

	// Process each file
	processed := 0
	failed := 0
	var outputBuilder strings.Builder

	outputBuilder.WriteString(fmt.Sprintf("=== UAsset Export (NativeAOT) ===\n"))
	outputBuilder.WriteString(fmt.Sprintf("Folder: %s\n", folderPath))
	if mappingsLoaded {
		outputBuilder.WriteString(fmt.Sprintf("Mappings: %s\n", mappingsPath))
	} else {
		outputBuilder.WriteString("Mappings: None (export may be incomplete)\n")
	}
	outputBuilder.WriteString(fmt.Sprintf("Found %d .uasset file(s)\n\n", len(uassetFiles)))

	for i, file := range uassetFiles {
		// Check for cancellation
		select {
		case <-ctx.Done():
			outputBuilder.WriteString("\n*** Operation cancelled by user ***\n")
			return UAssetResult{
				Success:        false,
				Error:          "Operation cancelled",
				Output:         outputBuilder.String(),
				Duration:       time.Since(startTime).String(),
				FilesProcessed: processed,
			}
		default:
		}

		fileName := filepath.Base(file)
		outputBuilder.WriteString(fmt.Sprintf("[%d/%d] Processing: %s\n", i+1, len(uassetFiles), fileName))

		// Load asset
		assetHandle, err := u.api.LoadAsset(file, EngineVersionUE5_4, mappingsHandle)
		if err != nil {
			outputBuilder.WriteString(fmt.Sprintf("  ✗ Failed to load: %v\n", err))
			failed++
			continue
		}

		// Serialize to JSON
		json, err := u.api.SerializeAssetToJson(assetHandle)
		if err != nil {
			outputBuilder.WriteString(fmt.Sprintf("  ✗ Failed to serialize: %v\n", err))
			u.api.FreeAsset(assetHandle)
			failed++
			continue
		}

		// Write JSON to file
		jsonPath := strings.TrimSuffix(file, filepath.Ext(file)) + ".json"
		err = os.WriteFile(jsonPath, []byte(json), 0644)
		if err != nil {
			outputBuilder.WriteString(fmt.Sprintf("  ✗ Failed to write JSON: %v\n", err))
			u.api.FreeAsset(assetHandle)
			failed++
			continue
		}

		u.api.FreeAsset(assetHandle)
		outputBuilder.WriteString(fmt.Sprintf("  ✓ Exported to: %s\n", filepath.Base(jsonPath)))
		processed++

		// Log progress every 10 files
		if (processed+failed)%10 == 0 {
			outputBuilder.WriteString(fmt.Sprintf("Progress: %d/%d files processed\n", processed+failed, len(uassetFiles)))
		}
	}

	duration := time.Since(startTime)

	outputBuilder.WriteString(fmt.Sprintf("\n=== Export Complete ===\n"))
	outputBuilder.WriteString(fmt.Sprintf("Successfully processed: %d files\n", processed))
	outputBuilder.WriteString(fmt.Sprintf("Failed: %d files\n", failed))
	outputBuilder.WriteString(fmt.Sprintf("Duration: %s\n", duration))

	success := failed == 0

	return UAssetResult{
		Success:        success,
		Message:        fmt.Sprintf("Export completed - %d successful, %d failed", processed, failed),
		Output:         outputBuilder.String(),
		Duration:       duration.String(),
		FilesProcessed: processed,
	}
}

// ImportUAssets imports JSON files back to .uasset/.uexp using the native library.
// This method signature matches the IPC-based service for drop-in replacement.
func (u *UAssetService) ImportUAssets(ctx context.Context, folderPath, mappingsPath string) UAssetResult {
	startTime := time.Now()

	// Validate folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return UAssetResult{
			Success: false,
			Error:   fmt.Sprintf("Folder does not exist: %s", folderPath),
		}
	}

	// Load mappings if provided
	var mappingsHandle MappingsHandle
	var mappingsLoaded bool

	if mappingsPath != "" {
		if _, err := os.Stat(mappingsPath); os.IsNotExist(err) {
			return UAssetResult{
				Success: false,
				Error:   fmt.Sprintf("Mappings file does not exist: %s", mappingsPath),
			}
		}

		handle, err := u.api.LoadMappings(mappingsPath)
		if err != nil {
			return UAssetResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to load mappings: %v", err),
			}
		}

		mappingsHandle = handle
		mappingsLoaded = true
		defer u.api.FreeMappings(mappingsHandle)
	}

	// Find all .json files
	var jsonFiles []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".json" {
			jsonFiles = append(jsonFiles, path)
		}

		return nil
	})

	if err != nil {
		return UAssetResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to scan directory: %v", err),
		}
	}

	// Process each file
	processed := 0
	failed := 0
	var outputBuilder strings.Builder

	outputBuilder.WriteString(fmt.Sprintf("=== UAsset Import (NativeAOT) ===\n"))
	outputBuilder.WriteString(fmt.Sprintf("Folder: %s\n", folderPath))
	if mappingsLoaded {
		outputBuilder.WriteString(fmt.Sprintf("Mappings: %s\n", mappingsPath))
	} else {
		outputBuilder.WriteString("Mappings: None (import may fail for unversioned properties)\n")
	}
	outputBuilder.WriteString(fmt.Sprintf("Found %d .json file(s)\n\n", len(jsonFiles)))

	for i, file := range jsonFiles {
		// Check for cancellation
		select {
		case <-ctx.Done():
			outputBuilder.WriteString("\n*** Operation cancelled by user ***\n")
			return UAssetResult{
				Success:        false,
				Error:          "Operation cancelled",
				Output:         outputBuilder.String(),
				Duration:       time.Since(startTime).String(),
				FilesProcessed: processed,
			}
		default:
		}

		fileName := filepath.Base(file)
		outputBuilder.WriteString(fmt.Sprintf("[%d/%d] Processing: %s\n", i+1, len(jsonFiles), fileName))

		// Read JSON file
		jsonData, err := os.ReadFile(file)
		if err != nil {
			outputBuilder.WriteString(fmt.Sprintf("  ✗ Failed to read JSON: %v\n", err))
			failed++
			continue
		}

		// Deserialize from JSON
		assetHandle, err := u.api.DeserializeAssetFromJson(string(jsonData))
		if err != nil {
			outputBuilder.WriteString(fmt.Sprintf("  ✗ Failed to deserialize: %v\n", err))
			failed++
			continue
		}

		// Write to .uasset/.uexp
		uassetPath := strings.TrimSuffix(file, filepath.Ext(file)) + ".uasset"
		err = u.api.WriteAssetToFile(assetHandle, uassetPath, mappingsHandle)
		if err != nil {
			outputBuilder.WriteString(fmt.Sprintf("  ✗ Failed to write asset: %v\n", err))
			u.api.FreeAsset(assetHandle)
			failed++
			continue
		}

		u.api.FreeAsset(assetHandle)
		outputBuilder.WriteString(fmt.Sprintf("  ✓ Imported to: %s\n", filepath.Base(uassetPath)))
		processed++

		// Log progress every 10 files
		if (processed+failed)%10 == 0 {
			outputBuilder.WriteString(fmt.Sprintf("Progress: %d/%d files processed\n", processed+failed, len(jsonFiles)))
		}
	}

	duration := time.Since(startTime)

	outputBuilder.WriteString(fmt.Sprintf("\n=== Import Complete ===\n"))
	outputBuilder.WriteString(fmt.Sprintf("Successfully processed: %d files\n", processed))
	outputBuilder.WriteString(fmt.Sprintf("Failed: %d files\n", failed))
	outputBuilder.WriteString(fmt.Sprintf("Duration: %s\n", duration))

	success := failed == 0

	return UAssetResult{
		Success:        success,
		Message:        fmt.Sprintf("Import completed - %d successful, %d failed", processed, failed),
		Output:         outputBuilder.String(),
		Duration:       duration.String(),
		FilesProcessed: processed,
	}
}

// CountUAssetFiles counts .uasset and .uexp files in a directory.
// This method signature matches the IPC-based service for drop-in replacement.
func (u *UAssetService) CountUAssetFiles(ctx context.Context, folderPath string) (int, int, error) {
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

// CountJSONFiles counts .json files in a directory.
// This method signature matches the IPC-based service for drop-in replacement.
func (u *UAssetService) CountJSONFiles(ctx context.Context, folderPath string) (int, error) {
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
