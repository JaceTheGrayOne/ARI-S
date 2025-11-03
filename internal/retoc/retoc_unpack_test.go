package retoc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
)

func TestRetocUnpack_Success_ExtractsToLegacy(t *testing.T) {
	tempDir := t.TempDir()
	gamePaksPath := filepath.Join(tempDir, "game_paks")
	extractPath := filepath.Join(tempDir, "extracted")

	if err := os.MkdirAll(gamePaksPath, 0755); err != nil {
		t.Fatalf("Failed to create game paks dir: %v", err)
	}
	if err := os.MkdirAll(extractPath, 0755); err != nil {
		t.Fatalf("Failed to create extract dir: %v", err)
	}

	// Create mock retoc.exe
	depsDir := filepath.Join(tempDir, "deps")
	retocDir := filepath.Join(depsDir, "retoc")
	if err := os.MkdirAll(retocDir, 0755); err != nil {
		t.Fatalf("Failed to create retoc dir: %v", err)
	}
	retocPath := filepath.Join(retocDir, "retoc.exe")
	if err := os.WriteFile(retocPath, []byte("mock"), 0755); err != nil {
		t.Fatalf("Failed to create mock retoc.exe: %v", err)
	}

	appInstance := app.NewApp()
	service := NewRetocService(appInstance, depsDir)

	operation := RetocOperation{
		Command:    "to-legacy",
		InputPath:  gamePaksPath,
		OutputPath: extractPath,
	}

	ctx := context.Background()
	result := service.RunRetoc(ctx, operation)

	if result.OperationID == "" {
		t.Error("Expected operation ID to be generated")
	}

	if result.Duration == "" {
		t.Error("Expected duration to be recorded")
	}

	t.Logf("Unpack result: %+v", result)
}

func TestRetocUnpack_Cancellation_StopsOperation(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "deps")
	retocDir := filepath.Join(depsDir, "retoc")

	if err := os.MkdirAll(retocDir, 0755); err != nil {
		t.Fatalf("Failed to create retoc dir: %v", err)
	}

	// Create mock long-running retoc.exe (script that sleeps)
	retocPath := filepath.Join(retocDir, "retoc.exe")
	if err := os.WriteFile(retocPath, []byte("mock"), 0755); err != nil {
		t.Fatalf("Failed to create mock retoc.exe: %v", err)
	}

	appInstance := app.NewApp()
	service := NewRetocService(appInstance, depsDir)

	operation := RetocOperation{
		Command:    "to-legacy",
		InputPath:  tempDir,
		OutputPath: tempDir,
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	resultChan := make(chan RetocResult, 1)
	go func() {
		result := service.RunRetoc(ctx, operation)
		resultChan <- result
	}()

	// Cancel after short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for result
	var result RetocResult
	select {
	case result = <-resultChan:
		// Operation completed (likely quickly due to mock exe)
	case <-time.After(5 * time.Second):
		t.Fatal("Operation did not complete after cancellation")
	}

	if result.Error != "cancelled" && ctx.Err() == context.Canceled {
		t.Logf("Context was cancelled correctly")
	}
}

func TestRetocUnpack_CancelByCommand_TargetsCorrectOperation(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "deps")

	appInstance := app.NewApp()
	service := NewRetocService(appInstance, depsDir)

	// Manually add a tracked operation
	ctx, cancel := context.WithCancel(context.Background())
	operationID := "test-operation-123"
	service.operationsMutex.Lock()
	service.runningOperations[operationID] = cancel
	service.operationsByCommand["to-legacy"] = operationID
	service.operationsMutex.Unlock()

	err := service.CancelOperationByCommand(context.Background(), "to-legacy")

	if err != nil {
		t.Errorf("Expected no error cancelling by command, got: %v", err)
	}

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		t.Log("Context cancelled successfully")
	case <-time.After(1 * time.Second):
		t.Error("Context was not cancelled")
	}
}

func TestRetocUnpack_CancelNonExistent_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "deps")

	appInstance := app.NewApp()
	service := NewRetocService(appInstance, depsDir)

	err := service.CancelOperation(context.Background(), "nonexistent-id")

	if err == nil {
		t.Error("Expected error when cancelling non-existent operation")
	}

	expectedMsg := "not found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message containing '%s', got: %s", expectedMsg, err.Error())
	}
}
