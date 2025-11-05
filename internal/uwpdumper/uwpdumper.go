package uwpdumper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
)

// UWPDumperService wraps the UWPInjector.exe tool for dumping UWP application
// packages. It provides methods to launch the interactive dumper tool which
// allows users to extract UWP application files that are normally encrypted
// by Windows Store protection.
//
// UWPDumperService is safe for concurrent use by multiple goroutines.
type UWPDumperService struct {
	app     *app.App
	depsDir string // Path to extracted dependencies
}

// LaunchResult contains the outcome of launching UWPInjector.exe.
// Since UWPInjector runs interactively, this result indicates whether
// the tool was successfully started, not whether the dump completed.
type LaunchResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Output   string `json:"output"`
	Error    string `json:"error"`
	Duration string `json:"duration"`
	ToolPath string `json:"tool_path"` // Path to UWPInjector.exe for reference
}

// NewUWPDumperService creates a new UWPDumperService using the UWPInjector.exe
// binary located in the given depsDir. The depsDir should contain a "uwpdumper"
// subdirectory with UWPInjector.exe and UWPDumper.dll.
func NewUWPDumperService(a *app.App, depsDir string) *UWPDumperService {
	return &UWPDumperService{
		app:     a,
		depsDir: depsDir,
	}
}

// LaunchUWPDumper opens UWPInjector.exe in a new console window, allowing the
// user to interact with it directly. The tool will prompt for a UWP process ID
// and perform the dump operation. This is an interactive process - the user
// must:
//   1. Ensure the target UWP application is running
//   2. Find the process ID (using Task Manager or the injector's process list)
//   3. Enter the PID when prompted by UWPInjector.exe
//
// The dumped files will be located in:
// %LOCALAPPDATA%\Packages\<PackageFamilyName>\TempState\DUMP
//
// Note: Administrator privileges may be required for some UWP applications.
func (u *UWPDumperService) LaunchUWPDumper(ctx context.Context) LaunchResult {
	startTime := time.Now()

	// Use the extracted dependencies directory
	dumperPath := filepath.Join(u.depsDir, "uwpdumper", "UWPInjector.exe")

	// Check if UWPInjector.exe exists
	if _, err := os.Stat(dumperPath); os.IsNotExist(err) {
		return LaunchResult{
			Success:  false,
			Error:    fmt.Sprintf("UWPInjector.exe not found at: %s", dumperPath),
			Message:  "UWPDumper tool not found",
			ToolPath: dumperPath,
			Duration: time.Since(startTime).String(),
		}
	}

	// Check if UWPDumper.dll exists (required dependency)
	dllPath := filepath.Join(u.depsDir, "uwpdumper", "UWPDumper.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		return LaunchResult{
			Success:  false,
			Error:    fmt.Sprintf("UWPDumper.dll not found at: %s", dllPath),
			Message:  "UWPDumper DLL not found",
			ToolPath: dumperPath,
			Duration: time.Since(startTime).String(),
		}
	}

	// Launch UWPInjector.exe in a new console window for interactive use
	// We use cmd.exe /c start to open a new window that persists after launch
	// Using empty string "" as title to avoid Windows interpreting the path as the title
	cmd := exec.Command("cmd.exe", "/c", "start", "", dumperPath)
	cmd.Dir = filepath.Dir(dumperPath) // Set working directory to tool location

	// Start the command (non-blocking)
	if err := cmd.Start(); err != nil {
		return LaunchResult{
			Success:  false,
			Error:    fmt.Sprintf("Failed to launch UWPInjector.exe: %v", err),
			Message:  "Failed to launch UWPDumper",
			ToolPath: dumperPath,
			Duration: time.Since(startTime).String(),
		}
	}

	duration := time.Since(startTime)

	return LaunchResult{
		Success:  true,
		Message:  "UWPDumper launched successfully",
		Output:   fmt.Sprintf("UWPInjector.exe opened in new window.\n\nInstructions:\n1. Ensure your target UWP app is running\n2. Enter the process ID when prompted\n3. Files will be dumped to: %%LOCALAPPDATA%%\\Packages\\<PackageFamilyName>\\TempState\\DUMP"),
		ToolPath: dumperPath,
		Duration: duration.String(),
	}
}

// GetDumperPath returns the absolute path to the UWPInjector.exe tool.
// This method can be used by the frontend to display the tool location
// or to verify the tool is available before attempting to launch it.
func (u *UWPDumperService) GetDumperPath(ctx context.Context) string {
	return filepath.Join(u.depsDir, "uwpdumper", "UWPInjector.exe")
}

// GetDumperInfo returns information about the UWPDumper tool installation,
// including whether the tool and its dependencies are available.
func (u *UWPDumperService) GetDumperInfo(ctx context.Context) map[string]interface{} {
	dumperPath := filepath.Join(u.depsDir, "uwpdumper", "UWPInjector.exe")
	dllPath := filepath.Join(u.depsDir, "uwpdumper", "UWPDumper.dll")

	dumperExists := false
	dllExists := false

	if _, err := os.Stat(dumperPath); err == nil {
		dumperExists = true
	}

	if _, err := os.Stat(dllPath); err == nil {
		dllExists = true
	}

	return map[string]interface{}{
		"dumper_path":    dumperPath,
		"dll_path":       dllPath,
		"dumper_exists":  dumperExists,
		"dll_exists":     dllExists,
		"ready":          dumperExists && dllExists,
		"output_location": "%LOCALAPPDATA%\\Packages\\<PackageFamilyName>\\TempState\\DUMP",
	}
}
