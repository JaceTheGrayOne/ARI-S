//go:build windows
// +build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32      = syscall.NewLazyDLL("user32.dll")
	messageBoxW = user32.NewProc("MessageBoxW")
)

// MessageBox displays a Windows message box
func MessageBox(hwnd uintptr, caption, title string, flags uint) int {
	captionPtr, _ := syscall.UTF16PtrFromString(caption)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	ret, _, _ := messageBoxW.Call(
		hwnd,
		uintptr(unsafe.Pointer(captionPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)

	return int(ret)
}

// DllMain is called when the DLL is loaded
//
//export DllMain
func DllMain(hinstDLL uintptr, fdwReason uint32, lpvReserved uintptr) bool {
	switch fdwReason {
	case 1: // DLL_PROCESS_ATTACH
		// Show message box when DLL is injected
		MessageBox(
			0,
			"DLL Injection Successful!\n\n"+
				"This message confirms that the DLL was successfully injected "+
				"into the target process using ARI-S.\n\n"+
				"DLL: TestMessageBox.dll (built with Go)",
			"ARI-S Injection Test",
			0x00000040, // MB_ICONINFORMATION
		)
	}
	return true
}

func main() {
	// Required for Go DLL, but not called
}
