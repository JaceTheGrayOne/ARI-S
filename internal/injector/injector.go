package injector

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
	"golang.org/x/sys/windows"
)

// InjectorService handles DLL injection into running processes using the
// CreateRemoteThread technique. It provides process enumeration, privilege
// checking, and UAC elevation handling.
//
// InjectorService is safe for concurrent use by multiple goroutines.
type InjectorService struct {
	app *app.App
}

// ProcessInfo describes a single running Windows process.
// It is returned by [InjectorService.GetRunningProcesses] for display
// in the UI process selector.
type ProcessInfo struct {
	PID  uint32 `json:"pid"`
	Name string `json:"name"`
}

// InjectionResult contains the outcome of a DLL injection operation.
// If Success is false, the Error field contains details. If the error
// code is "NEEDS_ELEVATION", the application must be restarted with
// administrator privileges.
type InjectionResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Output   string `json:"output"`
	Error    string `json:"error"`
	Duration string `json:"duration"`
}

// NewInjectorService creates a new InjectorService bound to the given App.
func NewInjectorService(a *app.App) *InjectorService {
	return &InjectorService{
		app: a,
	}
}

// GetRunningProcesses returns a snapshot of all running Windows processes.
// Each entry includes the process ID and executable name. The list is not
// sorted. This method uses ToolHelp32 to enumerate processes.
func (i *InjectorService) GetRunningProcesses(ctx context.Context) ([]ProcessInfo, error) {
	// Create a snapshot of all running processes
	handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	defer windows.CloseHandle(handle)

	// The ProcessEntry32 struct must have its dwSize field initialized
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	// Get the first process in the snapshot
	if err := windows.Process32First(handle, &entry); err != nil {
		return nil, fmt.Errorf("Process32First failed: %w", err)
	}

	var processes []ProcessInfo

	// Loop through all processes
	for {
		// The ExeFile field is a null-terminated array of uint16 (UTF-16)
		// We convert it to a Go string for display
		processName := windows.UTF16ToString(entry.ExeFile[:])

		processes = append(processes, ProcessInfo{
			PID:  entry.ProcessID,
			Name: processName,
		})

		// Move to the next process
		if err := windows.Process32Next(handle, &entry); err != nil {
			// ERROR_NO_MORE_FILES is expected at the end of the list
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			return nil, fmt.Errorf("Process32Next failed: %w", err)
		}
	}

	return processes, nil
}

// InjectDLL injects the DLL at dllPath into the process with the given PID
// using the CreateRemoteThread technique. This operation requires administrator
// privileges. If not running as admin, the method returns an InjectionResult
// with Error set to "NEEDS_ELEVATION".
//
// The injection process:
//  1. Opens the target process with required access rights
//  2. Allocates memory in the target process
//  3. Writes the DLL path to that memory
//  4. Creates a remote thread starting at LoadLibraryW
//  5. Waits for the thread to complete
//
// If the DLL fails to load in the target process (exit code 0), possible
// causes include architecture mismatch (32-bit vs 64-bit), missing
// dependencies, or an invalid DLL.
func (i *InjectorService) InjectDLL(ctx context.Context, targetPID uint32, dllPath string) InjectionResult {
	startTime := time.Now()

	// Emit initial status
	i.emitStatus(ctx, "Starting injection process...")

	// Validate DLL path exists
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		return InjectionResult{
			Success:  false,
			Error:    fmt.Sprintf("DLL file not found: %s", dllPath),
			Message:  "Injection failed: DLL not found",
			Duration: time.Since(startTime).String(),
		}
	}

	// Check for admin privileges
	i.emitStatus(ctx, "Checking administrator privileges...")
	isAdmin, err := i.checkAdminPrivileges()
	if err != nil {
		return InjectionResult{
			Success:  false,
			Error:    fmt.Sprintf("Failed to check privileges: %v", err),
			Message:  "Injection failed: privilege check error",
			Duration: time.Since(startTime).String(),
		}
	}

	if !isAdmin {
		// Need to elevate - but warn user first that app will restart
		// The frontend should show a confirmation dialog before calling this
		// For now, return a special status to trigger frontend confirmation
		return InjectionResult{
			Success:  false,
			Error:    "NEEDS_ELEVATION", // Special error code for frontend to detect
			Message:  "Administrator privileges required",
			Output:   "DLL injection requires administrator privileges.\n\nThe application will restart with elevated privileges.\nYour DLL selection will be saved.\n\nClick 'Inject DLL' again after restart to proceed.",
			Duration: time.Since(startTime).String(),
		}
	}

	// Perform the injection
	i.emitStatus(ctx, fmt.Sprintf("Injecting DLL into process %d...", targetPID))
	if err := i.injectDLL(ctx, targetPID, dllPath); err != nil {
		return InjectionResult{
			Success:  false,
			Error:    formatSystemError(err),
			Message:  "Injection failed",
			Output:   fmt.Sprintf("Failed to inject %s into process %d", dllPath, targetPID),
			Duration: time.Since(startTime).String(),
		}
	}

	i.emitStatus(ctx, "Injection completed successfully!")

	return InjectionResult{
		Success:  true,
		Message:  "DLL injected successfully",
		Output:   fmt.Sprintf("Successfully injected %s into process %d", dllPath, targetPID),
		Duration: time.Since(startTime).String(),
	}
}

// injectDLL performs the actual DLL injection using CreateRemoteThread
func (i *InjectorService) injectDLL(ctx context.Context, targetPID uint32, dllPath string) error {
	// Step 1: Open the target process with required permissions
	i.emitStatus(ctx, "Opening target process...")

	// Request minimal required access rights as per the methodology document:
	// - PROCESS_VM_OPERATION: Required by VirtualAllocEx
	// - PROCESS_VM_WRITE: Required by WriteProcessMemory
	// - PROCESS_CREATE_THREAD: Required by CreateRemoteThread
	const requiredAccess = windows.PROCESS_VM_OPERATION | windows.PROCESS_VM_WRITE | windows.PROCESS_CREATE_THREAD

	processHandle, err := windows.OpenProcess(requiredAccess, false, targetPID)
	if err != nil {
		return fmt.Errorf("OpenProcess failed: %w (ensure you have administrator privileges)", err)
	}
	defer windows.CloseHandle(processHandle)

	// Step 2: Convert the DLL path to UTF-16 (Windows Unicode)
	i.emitStatus(ctx, "Preparing DLL path...")

	// Convert Go string to null-terminated UTF-16 string for Windows API
	utf16DllPath, err := windows.UTF16PtrFromString(dllPath)
	if err != nil {
		return fmt.Errorf("UTF16PtrFromString failed: %w", err)
	}

	// Calculate size in bytes: UTF-16 uses 2 bytes per character + 2 bytes for null terminator
	dllPathSize := uintptr((len(dllPath) + 1) * 2)

	// Step 3: Allocate memory in the remote process
	i.emitStatus(ctx, "Allocating memory in target process...")

	// VirtualAllocEx is not directly wrapped, so we use syscall
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	virtualAllocEx := kernel32.NewProc("VirtualAllocEx")

	remoteAddr, _, err := virtualAllocEx.Call(
		uintptr(processHandle),
		0, // Let the OS choose the address
		dllPathSize,
		windows.MEM_COMMIT|windows.MEM_RESERVE,
		windows.PAGE_READWRITE,
	)
	if remoteAddr == 0 {
		return fmt.Errorf("VirtualAllocEx failed: %w", err)
	}

	// Step 4: Write the DLL path into the allocated memory
	i.emitStatus(ctx, "Writing DLL path to target process memory...")

	var bytesWritten uintptr
	err = windows.WriteProcessMemory(
		processHandle,
		remoteAddr,
		(*byte)(unsafe.Pointer(utf16DllPath)),
		dllPathSize,
		&bytesWritten,
	)
	if err != nil {
		return fmt.Errorf("WriteProcessMemory failed: %w", err)
	}
	if bytesWritten != dllPathSize {
		return fmt.Errorf("incomplete write: wrote %d bytes, expected %d", bytesWritten, dllPathSize)
	}

	// Step 5: Get the address of LoadLibraryW from kernel32.dll
	i.emitStatus(ctx, "Resolving LoadLibraryW address...")

	// kernel32.dll is loaded into virtually every process, so we can safely get its address
	loadLibraryW := kernel32.NewProc("LoadLibraryW")
	if err := loadLibraryW.Find(); err != nil {
		return fmt.Errorf("failed to find LoadLibraryW: %w", err)
	}
	loadLibraryWAddr := loadLibraryW.Addr()

	// Step 6: Create a remote thread to execute LoadLibraryW
	i.emitStatus(ctx, "Creating remote thread...")

	// CreateRemoteThread is not directly wrapped in x/sys/windows, so we use dynamic resolution
	createRemoteThread := kernel32.NewProc("CreateRemoteThread")
	if err := createRemoteThread.Find(); err != nil {
		return fmt.Errorf("failed to find CreateRemoteThread: %w", err)
	}

	threadHandle, _, err := createRemoteThread.Call(
		uintptr(processHandle), // hProcess: Handle to target process
		0,                      // lpThreadAttributes: Default security attributes
		0,                      // dwStackSize: Default stack size
		loadLibraryWAddr,       // lpStartAddress: Address of LoadLibraryW
		remoteAddr,             // lpParameter: Address of DLL path in remote process
		0,                      // dwCreationFlags: Run immediately
		0,                      // lpThreadId: We don't need the thread ID
	)

	if threadHandle == 0 {
		return fmt.Errorf("CreateRemoteThread failed: %w", err)
	}
	defer windows.CloseHandle(windows.Handle(threadHandle))

	// Step 7: Wait for the remote thread to finish (optional but recommended)
	i.emitStatus(ctx, "Waiting for DLL to load...")

	event, err := windows.WaitForSingleObject(windows.Handle(threadHandle), windows.INFINITE)
	if err != nil {
		return fmt.Errorf("WaitForSingleObject failed: %w", err)
	}
	if event != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("unexpected wait event: %d", event)
	}

	// Get the exit code of the thread (this is the return value of LoadLibraryW)
	getExitCodeThread := kernel32.NewProc("GetExitCodeThread")
	var exitCode uint32
	ret, _, err := getExitCodeThread.Call(threadHandle, uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		return fmt.Errorf("GetExitCodeThread failed: %w", err)
	}

	// A non-zero exit code means LoadLibraryW succeeded and returned a module handle
	if exitCode == 0 {
		return fmt.Errorf("LoadLibraryW failed in target process (possible causes: DLL architecture mismatch, missing dependencies, or invalid DLL)")
	}

	return nil
}

// checkAdminPrivileges determines if the current process is running with administrator rights
func (i *InjectorService) checkAdminPrivileges() (bool, error) {
	var sid *windows.SID

	// The SID for the Administrators group is S-1-5-32-544
	// This is defined by SECURITY_BUILTIN_DOMAIN_RID (32) and DOMAIN_ALIAS_RID_ADMINS (544)
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2, // Two sub-authorities
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false, fmt.Errorf("AllocateAndInitializeSid failed: %w", err)
	}
	defer windows.FreeSid(sid)

	// Check if current process token is a member of the Administrators group
	token := windows.GetCurrentProcessToken()
	member, err := token.IsMember(sid)
	if err != nil {
		return false, fmt.Errorf("Token.IsMember failed: %w", err)
	}

	return member, nil
}

// elevatePrivileges re-launches the current application with a UAC prompt for admin rights
func (i *InjectorService) elevatePrivileges() error {
	verb := "runas" // This triggers the UAC prompt
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Preserve command-line arguments
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := windows.UTF16PtrFromString(verb)
	exePtr, _ := windows.UTF16PtrFromString(exe)
	var argPtr *uint16
	if args != "" {
		argPtr, _ = windows.UTF16PtrFromString(args)
	}

	// SW_SHOWNORMAL = 1 (show window in normal state)
	err = windows.ShellExecute(0, verbPtr, exePtr, argPtr, nil, 1)
	if err != nil {
		return fmt.Errorf("ShellExecute failed: %w", err)
	}

	// If ShellExecute succeeds, exit the current non-elevated process
	// The new elevated process will start
	os.Exit(0)
	return nil
}

// emitStatus sends a status update to the frontend
func (i *InjectorService) emitStatus(ctx context.Context, message string) {
	// In Wails v3, we emit events through the application context
	// The frontend will listen for these events
	// Note: This requires the application instance to be available
	// For now, we'll use a simple log. The Wails runtime will be integrated when called from frontend
	// runtime.EventsEmit(ctx, "injection-status", message)

	// TODO: Integrate with Wails runtime.EventsEmit when available
	// For now, just log to console
	fmt.Println("[Injection Status]:", message)
}

// formatSystemError converts a Windows syscall.Errno into a human-readable string
func formatSystemError(err error) string {
	errno, ok := err.(syscall.Errno)
	if !ok {
		return err.Error()
	}

	// Use FormatMessage to get a human-readable error description
	var flags uint32 = windows.FORMAT_MESSAGE_FROM_SYSTEM | windows.FORMAT_MESSAGE_IGNORE_INSERTS
	b := make([]uint16, 300)
	n, err := windows.FormatMessage(flags, 0, uint32(errno), 0, b, nil)
	if err != nil {
		return fmt.Sprintf("system error code %d (FormatMessage failed: %v)", errno, err)
	}

	// Trim terminating \r\n and convert to string
	return strings.TrimSpace(windows.UTF16ToString(b[:n]))
}
