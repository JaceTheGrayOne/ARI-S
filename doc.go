// Package main implements ARI-S (Asset Reconfiguration and Integration System),
// a desktop application for Unreal Engine modding workflows.
//
// ARI-S provides a unified graphical interface for three primary operations:
//
// 1. IoStore Package Management - Converting between Unreal Engine PAK formats
// using the embedded retoc tool
//
// 2. UAsset Serialization - Converting binary .uasset/.uexp files to/from JSON
// using either an IPC bridge or native CGO integration with UAssetAPI
//
// 3. DLL Injection - Runtime code injection into game processes using
// CreateRemoteThread technique
//
// The application is built with Go 1.24+ and Wails v3, providing a web-based
// UI rendered in WebView2. All dependencies (retoc.exe, UAssetBridge) are
// embedded at compile-time and extracted on first run.
//
// # Architecture
//
// The backend follows a service-oriented pattern:
//   - [App]: Core application lifecycle and configuration
//   - [RetocService]: Wraps retoc.exe for pak operations
//   - [UAssetService]: IPC-based UAsset serialization (default)
//   - [UAssetNativeService]: CGO-based UAsset serialization (experimental)
//   - [InjectorService]: Windows DLL injection via CreateRemoteThread
//
// # Configuration
//
// Application settings are persisted to a JSON file in the user's local
// AppData directory. See [Config] for details.
//
// # Platform Support
//
// ARI-S is Windows-only, requiring Windows 10/11 (64-bit) due to WebView2
// and Windows-specific API dependencies.
package main
