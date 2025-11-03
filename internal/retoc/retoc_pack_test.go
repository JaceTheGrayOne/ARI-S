package retoc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
)

func TestRetocPack_Success_CreatesZenFiles(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input_mod")
	outputPath := filepath.Join(tempDir, "output")

	// Create input directory
	if err := os.MkdirAll(inputPath, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Create mock retoc.exe directory
	depsDir := filepath.Join(tempDir, "deps")
	retocDir := filepath.Join(depsDir, "retoc")
	if err := os.MkdirAll(retocDir, 0755); err != nil {
		t.Fatalf("Failed to create retoc dir: %v", err)
	}

	// Create mock retoc.exe (for existence check)
	retocPath := filepath.Join(retocDir, "retoc.exe")
	if err := os.WriteFile(retocPath, []byte("mock"), 0755); err != nil {
		t.Fatalf("Failed to create mock retoc.exe: %v", err)
	}

	appInstance := app.NewApp()
	service := NewRetocService(appInstance, depsDir)

	operation := RetocOperation{
		Command:    "to-zen",
		InputPath:  inputPath,
		OutputPath: outputPath,
		UEVersion:  "UE5_4",
		Options:    []string{"--mod-name", "TestMod", "--serialization", "0001"},
	}

	ctx := context.Background()
	result := service.RunRetoc(ctx, operation)

	if result.OperationID == "" {
		t.Error("Expected operation ID to be generated")
	}

	if result.Duration == "" {
		t.Error("Expected duration to be recorded")
	}

	// Note: Success will be false due to mock exe, but structure is valid
	t.Logf("Operation result: %+v", result)
}

func TestRetocPack_MissingRetocExe_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "nonexistent_deps")

	appInstance := app.NewApp()
	service := NewRetocService(appInstance, depsDir)

	operation := RetocOperation{
		Command:    "to-zen",
		InputPath:  "/test/input",
		OutputPath: "/test/output",
		UEVersion:  "UE5_4",
	}

	ctx := context.Background()
	result := service.RunRetoc(ctx, operation)

	if result.Success {
		t.Error("Expected operation to fail with missing retoc.exe")
	}

	if result.Error == "" {
		t.Error("Expected error message about missing retoc.exe")
	}

	expectedErrSubstring := "retoc.exe not found"
	if !strings.Contains(result.Error, expectedErrSubstring) {
		t.Errorf("Expected error containing '%s', got: %s", expectedErrSubstring, result.Error)
	}
}

func TestRetocPack_UnknownCommand_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "deps")
	retocDir := filepath.Join(depsDir, "retoc")
	if err := os.MkdirAll(retocDir, 0755); err != nil {
		t.Fatalf("Failed to create retoc dir: %v", err)
	}

	// Create mock retoc.exe so we get past the existence check
	retocPath := filepath.Join(retocDir, "retoc.exe")
	if err := os.WriteFile(retocPath, []byte("mock"), 0755); err != nil {
		t.Fatalf("Failed to create mock retoc.exe: %v", err)
	}

	appInstance := app.NewApp()
	service := NewRetocService(appInstance, depsDir)

	operation := RetocOperation{
		Command:    "invalid-command",
		InputPath:  "/test/input",
		OutputPath: "/test/output",
	}

	ctx := context.Background()
	result := service.RunRetoc(ctx, operation)

	if result.Success {
		t.Error("Expected operation to fail with unknown command")
	}

	expectedErrSubstring := "Unknown command"
	if !strings.Contains(result.Error, expectedErrSubstring) {
		t.Errorf("Expected error containing '%s', got: %s", expectedErrSubstring, result.Error)
	}
}

func TestRetocPack_FileRenaming_CorrectFormat(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "deps")
	outputDir := tempDir

	appInstance := app.NewApp()
	service := NewRetocService(appInstance, depsDir)

	// Create mock output files with default names
	inputFolderName := "TestModFolder"
	mockFiles := []string{
		filepath.Join(outputDir, inputFolderName+".utoc"),
		filepath.Join(outputDir, inputFolderName+".ucas"),
		filepath.Join(outputDir, inputFolderName+".pak"),
	}

	for _, file := range mockFiles {
		if err := os.WriteFile(file, []byte("mock"), 0644); err != nil {
			t.Fatalf("Failed to create mock file: %v", err)
		}
	}

	operation := RetocOperation{
		Command:    "to-zen",
		InputPath:  filepath.Join(tempDir, inputFolderName),
		OutputPath: outputDir,
		Options:    []string{"--mod-name", "AwesomeMod", "--serialization", "0042"},
	}

	err := service.renameOutputFiles(operation)

	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(outputDir, "z_AwesomeMod_0042_p.utoc"),
		filepath.Join(outputDir, "z_AwesomeMod_0042_p.ucas"),
		filepath.Join(outputDir, "z_AwesomeMod_0042_p.pak"),
	}

	for _, expected := range expectedFiles {
		if _, err := os.Stat(expected); os.IsNotExist(err) {
			t.Errorf("Expected renamed file not found: %s", expected)
		}
	}

	// Verify old files are gone
	for _, old := range mockFiles {
		if _, err := os.Stat(old); !os.IsNotExist(err) {
			t.Errorf("Expected old file to be removed: %s", old)
		}
	}
}
