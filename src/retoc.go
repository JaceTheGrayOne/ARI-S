package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RetocService handles Retoc operations
type RetocService struct {
	app               *App
	runningOperations map[string]context.CancelFunc
	// Track operations by command type so we can cancel specific operation types
	operationsByCommand map[string]string // command -> operationID
	operationsMutex     sync.Mutex
}

// RetocOperation represents a Retoc operation
type RetocOperation struct {
	Command    string   `json:"command"`
	InputPath  string   `json:"input_path"`
	OutputPath string   `json:"output_path"`
	UEVersion  string   `json:"ue_version"`
	Options    []string `json:"options"`
}

// RetocResult represents the result of a Retoc operation
type RetocResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Output      string `json:"output"`
	Error       string `json:"error"`
	Duration    string `json:"duration"`
	OperationID string `json:"operation_id"` // ID for tracking/cancelling
}

// NewRetocService creates a new RetocService
func NewRetocService(app *App) *RetocService {
	return &RetocService{
		app:                 app,
		runningOperations:   make(map[string]context.CancelFunc),
		operationsByCommand: make(map[string]string),
	}
}


// RunRetoc executes a Retoc operation
func (r *RetocService) RunRetoc(ctx context.Context, operation RetocOperation) RetocResult {
	startTime := time.Now()

	// Generate unique operation ID
	operationID := uuid.New().String()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)

	// Store cancel function and track by command type
	r.operationsMutex.Lock()
	r.runningOperations[operationID] = cancel
	r.operationsByCommand[operation.Command] = operationID
	r.operationsMutex.Unlock()

	// Ensure cleanup on exit
	defer func() {
		r.operationsMutex.Lock()
		delete(r.runningOperations, operationID)
		// Only delete from operationsByCommand if this is still the current operation for this command
		if r.operationsByCommand[operation.Command] == operationID {
			delete(r.operationsByCommand, operation.Command)
		}
		r.operationsMutex.Unlock()
	}()

	// Get the directory where the application executable is located
	execPath, err := os.Executable()
	if err != nil {
		return RetocResult{
			Success:     false,
			Error:       fmt.Sprintf("Failed to get executable path: %v", err),
			OperationID: operationID,
		}
	}
	execDir := filepath.Dir(execPath)

	// Look for retoc.exe in a "retoc" folder next to the executable
	retocPath := filepath.Join(execDir, "retoc", "retoc.exe")

	// Check if retoc.exe exists
	if _, err := os.Stat(retocPath); os.IsNotExist(err) {
		return RetocResult{
			Success:     false,
			Error:       fmt.Sprintf("retoc.exe not found at: %s", retocPath),
			OperationID: operationID,
		}
	}

	// Build command arguments
	args := []string{}

	switch operation.Command {
	case "to-legacy":
		args = append(args, "to-legacy", operation.InputPath, operation.OutputPath)
	case "to-zen":
		// Build: retoc.exe to-zen --version UE5_4 "inputdirectory" "outputdirectory\basename.utoc"
		// retoc requires a .utoc filename as the output parameter, not just a directory
		args = append(args, "to-zen")
		if operation.UEVersion != "" {
			args = append(args, "--version", operation.UEVersion)
		}

		// Get the base name from the input folder
		inputBaseName := filepath.Base(operation.InputPath)
		// Construct output path with .utoc extension (retoc will create .utoc, .ucas, and .pak)
		outputFilePath := filepath.Join(operation.OutputPath, inputBaseName+".utoc")

		args = append(args, operation.InputPath, outputFilePath)
	case "unpack":
		args = append(args, "unpack", operation.InputPath, operation.OutputPath)
	case "info":
		args = append(args, "info", operation.InputPath)
	case "list":
		args = append(args, "list", operation.InputPath)
	default:
		return RetocResult{
			Success:     false,
			Error:       fmt.Sprintf("Unknown command: %s", operation.Command),
			OperationID: operationID,
		}
	}

	// Note: We don't add operation.Options anymore since retoc doesn't use --priority

	// Create command
	cmd := exec.CommandContext(ctx, retocPath, args...)

	// Set working directory to the directory containing retoc.exe
	cmd.Dir = filepath.Dir(retocPath)

	// Set up pipes for streaming output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return RetocResult{
			Success:     false,
			Error:       fmt.Sprintf("Failed to create stdout pipe: %v", err),
			OperationID: operationID,
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return RetocResult{
			Success:     false,
			Error:       fmt.Sprintf("Failed to create stderr pipe: %v", err),
			OperationID: operationID,
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return RetocResult{
			Success:     false,
			Error:       fmt.Sprintf("Failed to start retoc: %v", err),
			OperationID: operationID,
		}
	}

	// Stream output in real-time to the frontend
	var wg sync.WaitGroup
	wg.Add(2)

	go r.streamOutput(stdout, operationID, &wg)
	go r.streamOutput(stderr, operationID, &wg)

	// Wait for streams to finish
	wg.Wait()

	// Wait for command to complete
	err = cmd.Wait()

	duration := time.Since(startTime)

	result := RetocResult{
		Duration:    duration.String(),
		OperationID: operationID,
	}

	// Check if cancelled
	if ctx.Err() == context.Canceled {
		result.Success = false
		result.Message = "Operation cancelled by user"
		result.Error = "cancelled"
		return result
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "Retoc operation failed"
	} else {
		result.Success = true
		result.Message = "Retoc operation completed successfully"

		// If this was a to-zen operation, always rename output files with proper UE mod naming convention
		if operation.Command == "to-zen" {
			renameErr := r.renameOutputFiles(operation)
			if renameErr != nil {
				result.Message += fmt.Sprintf(" (Warning: File renaming failed: %v)", renameErr)
			}
		}
	}

	return result
}

// renameOutputFiles renames the output files from retoc to the correct UE mod naming convention
// Expected format: z_modname_serialization_p.utoc, z_modname_serialization_p.ucas, z_modname_serialization_p.pak
func (r *RetocService) renameOutputFiles(operation RetocOperation) error {
	// Get the input folder name (retoc uses this as the default output name)
	inputFolderName := filepath.Base(operation.InputPath)

	// Get the mod name and serialization from the options, with defaults
	modName := inputFolderName
	serialization := "0001"

	// Parse options to get mod name and serialization if provided
	for i, opt := range operation.Options {
		if opt == "--mod-name" && i+1 < len(operation.Options) {
			if operation.Options[i+1] != "" {
				modName = operation.Options[i+1]
			}
		}
		if opt == "--serialization" && i+1 < len(operation.Options) {
			if operation.Options[i+1] != "" {
				// Ensure serialization is exactly 4 digits (pad with zeros or truncate)
				rawSerial := operation.Options[i+1]
				// Parse as integer and format with zero-padding to exactly 4 digits
				var val int
				if _, err := fmt.Sscanf(rawSerial, "%d", &val); err == nil {
					serialization = fmt.Sprintf("%04d", val%10000) // Limit to 4 digits max
				} else {
					// If not a valid number, truncate or pad the string
					if len(rawSerial) >= 4 {
						serialization = rawSerial[:4]
					} else {
						// Pad with leading zeros
						serialization = fmt.Sprintf("%04s", rawSerial)
						// fmt pads with spaces, so manually pad with zeros
						for len(serialization) > 0 && serialization[0] == ' ' {
							serialization = "0" + serialization[1:]
						}
					}
				}
			}
		}
	}

	// Get the output directory (OutputPath should be just a directory for to-zen)
	outputDir := operation.OutputPath

	// Define the old and new file names
	oldFiles := []string{
		filepath.Join(outputDir, inputFolderName+".utoc"),
		filepath.Join(outputDir, inputFolderName+".ucas"),
		filepath.Join(outputDir, inputFolderName+".pak"),
	}

	newFiles := []string{
		filepath.Join(outputDir, fmt.Sprintf("z_%s_%s_p.utoc", modName, serialization)),
		filepath.Join(outputDir, fmt.Sprintf("z_%s_%s_p.ucas", modName, serialization)),
		filepath.Join(outputDir, fmt.Sprintf("z_%s_%s_p.pak", modName, serialization)),
	}

	// Rename each file
	for i, oldFile := range oldFiles {
		if _, err := os.Stat(oldFile); err == nil {
			err := os.Rename(oldFile, newFiles[i])
			if err != nil {
				return fmt.Errorf("failed to rename %s to %s: %v", oldFile, newFiles[i], err)
			}
		}
	}

	return nil
}

// streamOutput reads from an output pipe and emits it to the frontend in real-time
func (r *RetocService) streamOutput(reader io.Reader, operationID string, wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// TODO: Emit event to frontend with the output line
		// For now, just print to stdout
		fmt.Printf("[retoc:%s] %s\n", operationID, line)
	}
}

// CancelOperation cancels a running retoc operation by its ID
func (r *RetocService) CancelOperation(ctx context.Context, operationID string) error {
	r.operationsMutex.Lock()
	cancel, exists := r.runningOperations[operationID]
	r.operationsMutex.Unlock()

	if !exists {
		return fmt.Errorf("operation %s not found or already completed", operationID)
	}

	// Trigger cancellation
	cancel()
	return nil
}

// CancelOperationByCommand cancels the currently running operation of a specific command type
func (r *RetocService) CancelOperationByCommand(ctx context.Context, command string) error {
	r.operationsMutex.Lock()
	operationID := r.operationsByCommand[command]
	r.operationsMutex.Unlock()

	if operationID == "" {
		return fmt.Errorf("no %s operation currently running", command)
	}

	return r.CancelOperation(ctx, operationID)
}
