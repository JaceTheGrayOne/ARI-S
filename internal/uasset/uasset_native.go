//go:build cgo
// +build cgo

package uasset

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lUAssetBridge

#include <stdlib.h>

// Forward declarations for exported C functions from UAssetBridge.dll
extern char* GetVersion();
extern void FreeString(void* ptr);
extern char* GetLastErrorString();

extern void* LoadMappings(const char* path);
extern void FreeMappings(void* handle);

extern void* LoadAsset(const char* path, int engineVersion, void* mappingsHandle);
extern void FreeAsset(void* handle);

extern int GetAssetExportCount(void* handle);
extern char* SerializeAssetToJson(void* handle);
extern void* DeserializeAssetFromJson(const char* json);
extern int WriteAssetToFile(void* handle, const char* path, void* mappingsHandle);
extern char* GetAssetFilePath(void* handle);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// NativeUAssetAPI provides Go wrappers around the NativeAOT UAssetBridge library.
// All methods handle memory management and CGO marshalling.
type NativeUAssetAPI struct {
	// No state needed - all operations are stateless or use handles
}

// NewNativeUAssetAPI creates a new NativeUAssetAPI instance.
func NewNativeUAssetAPI() *NativeUAssetAPI {
	return &NativeUAssetAPI{}
}

// GetVersion returns the UAssetAPI version string.
// This is a simple PoC function to validate the CGO toolchain.
func (n *NativeUAssetAPI) GetVersion() (string, error) {
	// Call C function
	cVersion := C.GetVersion()
	if cVersion == nil {
		return "", n.getLastError("GetVersion returned null")
	}
	defer C.FreeString(unsafe.Pointer(cVersion))

	// Convert C string to Go string
	version := C.GoString(cVersion)
	return version, nil
}

// MappingsHandle is an opaque handle to a C# Usmap object.
// Memory is owned by C#. Must call FreeMappings when done.
type MappingsHandle uintptr

// LoadMappings loads a .usmap file and returns an opaque handle.
// The handle must be freed with FreeMappings when no longer needed.
func (n *NativeUAssetAPI) LoadMappings(path string) (MappingsHandle, error) {
	// Convert Go string to C string
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	// Call C function
	handle := C.LoadMappings(cPath)
	if handle == nil {
		return 0, n.getLastError("LoadMappings failed")
	}

	return MappingsHandle(uintptr(handle)), nil
}

// FreeMappings releases a mappings handle previously returned by LoadMappings.
func (n *NativeUAssetAPI) FreeMappings(handle MappingsHandle) {
	if handle == 0 {
		return
	}
	C.FreeMappings(unsafe.Pointer(uintptr(handle)))
}

// AssetHandle is an opaque handle to a C# UAsset object.
// Memory is owned by C#. Must call FreeAsset when done.
type AssetHandle uintptr

// Unreal Engine version constants for use with LoadAsset. These values must
// match the C# EngineVersion enum in UAssetAPI. The version determines which
// serialization format is used when reading and writing assets.
const (
	EngineVersionUE4_0  = 0    // Unreal Engine 4.0
	EngineVersionUE4_27 = 510  // Unreal Engine 4.27
	EngineVersionUE5_0  = 1004 // Unreal Engine 5.0
	EngineVersionUE5_1  = 1007 // Unreal Engine 5.1
	EngineVersionUE5_2  = 1008 // Unreal Engine 5.2
	EngineVersionUE5_3  = 1009 // Unreal Engine 5.3
	EngineVersionUE5_4  = 1010 // Unreal Engine 5.4 (used by Grounded 2)
)

// LoadAsset loads a .uasset file and returns an opaque handle.
// The mappingsHandle can be 0 if no mappings are needed.
// The handle must be freed with FreeAsset when no longer needed.
func (n *NativeUAssetAPI) LoadAsset(path string, engineVersion int, mappingsHandle MappingsHandle) (AssetHandle, error) {
	// Convert Go string to C string
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	// Call C function
	handle := C.LoadAsset(cPath, C.int(engineVersion), unsafe.Pointer(uintptr(mappingsHandle)))
	if handle == nil {
		return 0, n.getLastError("LoadAsset failed")
	}

	return AssetHandle(uintptr(handle)), nil
}

// FreeAsset releases an asset handle previously returned by LoadAsset or DeserializeAssetFromJson.
func (n *NativeUAssetAPI) FreeAsset(handle AssetHandle) {
	if handle == 0 {
		return
	}
	C.FreeAsset(unsafe.Pointer(uintptr(handle)))
}

// GetAssetExportCount returns the number of exports in the asset.
// Returns -1 on error.
func (n *NativeUAssetAPI) GetAssetExportCount(handle AssetHandle) (int, error) {
	count := C.GetAssetExportCount(unsafe.Pointer(uintptr(handle)))
	if count < 0 {
		return -1, n.getLastError("GetAssetExportCount failed")
	}
	return int(count), nil
}

// SerializeAssetToJson serializes the asset to a JSON string.
// The returned string is a copy - the caller owns the memory.
func (n *NativeUAssetAPI) SerializeAssetToJson(handle AssetHandle) (string, error) {
	// Call C function
	cJson := C.SerializeAssetToJson(unsafe.Pointer(uintptr(handle)))
	if cJson == nil {
		return "", n.getLastError("SerializeAssetToJson failed")
	}
	defer C.FreeString(unsafe.Pointer(cJson))

	// Convert C string to Go string (this makes a copy)
	json := C.GoString(cJson)
	return json, nil
}

// DeserializeAssetFromJson deserializes a JSON string to a UAsset object.
// Returns a handle that must be freed with FreeAsset.
func (n *NativeUAssetAPI) DeserializeAssetFromJson(json string) (AssetHandle, error) {
	// Convert Go string to C string
	cJson := C.CString(json)
	defer C.free(unsafe.Pointer(cJson))

	// Call C function
	handle := C.DeserializeAssetFromJson(cJson)
	if handle == nil {
		return 0, n.getLastError("DeserializeAssetFromJson failed")
	}

	return AssetHandle(uintptr(handle)), nil
}

// WriteAssetToFile writes the asset to a .uasset/.uexp file pair.
// The mappingsHandle can be 0 if no mappings are needed.
func (n *NativeUAssetAPI) WriteAssetToFile(handle AssetHandle, path string, mappingsHandle MappingsHandle) error {
	// Convert Go string to C string
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	// Call C function
	result := C.WriteAssetToFile(
		unsafe.Pointer(uintptr(handle)),
		cPath,
		unsafe.Pointer(uintptr(mappingsHandle)),
	)

	if result == 0 {
		return n.getLastError("WriteAssetToFile failed")
	}

	return nil
}

// GetAssetFilePath returns the file path of a loaded asset.
func (n *NativeUAssetAPI) GetAssetFilePath(handle AssetHandle) (string, error) {
	// Call C function
	cPath := C.GetAssetFilePath(unsafe.Pointer(uintptr(handle)))
	if cPath == nil {
		return "", n.getLastError("GetAssetFilePath failed")
	}
	defer C.FreeString(unsafe.Pointer(cPath))

	// Convert C string to Go string
	path := C.GoString(cPath)
	return path, nil
}

// getLastError retrieves the last error from C# and wraps it in a Go error.
func (n *NativeUAssetAPI) getLastError(context string) error {
	cError := C.GetLastErrorString()
	if cError == nil {
		return fmt.Errorf("%s: unknown error", context)
	}
	defer C.FreeString(unsafe.Pointer(cError))

	errorMsg := C.GoString(cError)
	if errorMsg == "" || errorMsg == "No error" {
		return fmt.Errorf("%s", context)
	}

	return fmt.Errorf("%s: %s", context, errorMsg)
}
