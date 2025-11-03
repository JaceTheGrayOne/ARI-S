using System;
using System.IO;
using System.Runtime.InteropServices;
using UAssetAPI;
using UAssetAPI.UnrealTypes;
using UAssetAPI.Unversioned;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace UAssetBridge
{
    /// <summary>
    /// Native C-style API facade for UAssetAPI library.
    /// All functions use [UnmanagedCallersOnly] for direct CGO interop.
    ///
    /// Memory Ownership Rules:
    /// - Handles (IntPtr): C# owns via GCHandle, caller must call Free* functions
    /// - Strings returned from C#: C# allocates, caller must call FreeString
    /// - Strings passed to C#: Caller owns, C# reads only
    /// </summary>
    public static class NativeApi
    {
        /// <summary>
        /// Returns the UAssetAPI version string.
        /// Memory ownership: C# allocates, caller must call FreeString.
        /// </summary>
        /// <returns>Pointer to UTF-8 encoded version string</returns>
        [UnmanagedCallersOnly(EntryPoint = "GetVersion")]
        public static IntPtr GetVersion()
        {
            try
            {
                NativeErrors.ClearLastError();

                // Get UAssetAPI assembly version
                var assembly = typeof(UAsset).Assembly;
                string version = assembly.GetName().Version?.ToString() ?? "Unknown";

                string versionInfo = $"UAssetAPI v{version} (NativeAOT Bridge)";

                // Allocate unmanaged memory and return to caller
                return Marshal.StringToCoTaskMemUTF8(versionInfo);
            }
            catch (Exception ex)
            {
                NativeErrors.SetLastError($"GetVersion failed: {ex.Message}");
                return IntPtr.Zero;
            }
        }

        /// <summary>
        /// Loads a .usmap mappings file and returns an opaque handle.
        /// Memory ownership: C# owns the Usmap object, caller must call FreeMappings.
        /// </summary>
        /// <param name="path">Pointer to UTF-8 null-terminated path string</param>
        /// <returns>Handle to Usmap object, or IntPtr.Zero on failure</returns>
        [UnmanagedCallersOnly(EntryPoint = "LoadMappings")]
        public static IntPtr LoadMappings(IntPtr path)
        {
            try
            {
                NativeErrors.ClearLastError();

                if (path == IntPtr.Zero)
                {
                    NativeErrors.SetLastError("LoadMappings: path is null");
                    return IntPtr.Zero;
                }

                // Marshal C string to managed string
                string managedPath = Marshal.PtrToStringUTF8(path);
                if (string.IsNullOrEmpty(managedPath))
                {
                    NativeErrors.SetLastError("LoadMappings: path is empty");
                    return IntPtr.Zero;
                }

                if (!File.Exists(managedPath))
                {
                    NativeErrors.SetLastError($"LoadMappings: file not found: {managedPath}");
                    return IntPtr.Zero;
                }

                // Create Usmap object
                Usmap mappings = new Usmap(managedPath);

                // Allocate GCHandle to pin the object and prevent GC
                GCHandle handle = GCHandle.Alloc(mappings);

                // Return handle as IntPtr
                return GCHandle.ToIntPtr(handle);
            }
            catch (Exception ex)
            {
                NativeErrors.SetLastError($"LoadMappings exception: {ex.Message}");
                return IntPtr.Zero;
            }
        }

        /// <summary>
        /// Frees a Usmap object previously returned by LoadMappings.
        /// </summary>
        /// <param name="mappingsHandle">Handle to Usmap object</param>
        [UnmanagedCallersOnly(EntryPoint = "FreeMappings")]
        public static void FreeMappings(IntPtr mappingsHandle)
        {
            try
            {
                if (mappingsHandle == IntPtr.Zero)
                    return;

                GCHandle handle = GCHandle.FromIntPtr(mappingsHandle);
                handle.Free();
            }
            catch (Exception ex)
            {
                // Cannot set error here as this is a cleanup function
                // Silently fail to avoid cascading errors
                Console.Error.WriteLine($"FreeMappings warning: {ex.Message}");
            }
        }

        /// <summary>
        /// Loads a .uasset file and returns an opaque handle.
        /// Memory ownership: C# owns the UAsset object, caller must call FreeAsset.
        /// </summary>
        /// <param name="path">Pointer to UTF-8 null-terminated path string</param>
        /// <param name="engineVersion">Engine version as integer (e.g., EngineVersion.VER_UE5_4)</param>
        /// <param name="mappingsHandle">Optional handle to Usmap object (can be IntPtr.Zero)</param>
        /// <returns>Handle to UAsset object, or IntPtr.Zero on failure</returns>
        [UnmanagedCallersOnly(EntryPoint = "LoadAsset")]
        public static IntPtr LoadAsset(IntPtr path, int engineVersion, IntPtr mappingsHandle)
        {
            try
            {
                NativeErrors.ClearLastError();

                if (path == IntPtr.Zero)
                {
                    NativeErrors.SetLastError("LoadAsset: path is null");
                    return IntPtr.Zero;
                }

                // Marshal path
                string managedPath = Marshal.PtrToStringUTF8(path);
                if (string.IsNullOrEmpty(managedPath))
                {
                    NativeErrors.SetLastError("LoadAsset: path is empty");
                    return IntPtr.Zero;
                }

                if (!File.Exists(managedPath))
                {
                    NativeErrors.SetLastError($"LoadAsset: file not found: {managedPath}");
                    return IntPtr.Zero;
                }

                // Get mappings object if handle provided
                Usmap mappings = null;
                if (mappingsHandle != IntPtr.Zero)
                {
                    try
                    {
                        GCHandle mappingsGCHandle = GCHandle.FromIntPtr(mappingsHandle);
                        mappings = mappingsGCHandle.Target as Usmap;
                    }
                    catch (Exception ex)
                    {
                        NativeErrors.SetLastError($"LoadAsset: invalid mappings handle: {ex.Message}");
                        return IntPtr.Zero;
                    }
                }

                // Cast to EngineVersion enum
                EngineVersion version = (EngineVersion)engineVersion;

                // Load the asset
                UAsset asset = new UAsset(managedPath, version, mappings);

                // Allocate GCHandle
                GCHandle handle = GCHandle.Alloc(asset);

                return GCHandle.ToIntPtr(handle);
            }
            catch (Exception ex)
            {
                NativeErrors.SetLastError($"LoadAsset exception: {ex.Message}");
                return IntPtr.Zero;
            }
        }

        /// <summary>
        /// Frees a UAsset object previously returned by LoadAsset or DeserializeAssetFromJson.
        /// </summary>
        /// <param name="assetHandle">Handle to UAsset object</param>
        [UnmanagedCallersOnly(EntryPoint = "FreeAsset")]
        public static void FreeAsset(IntPtr assetHandle)
        {
            try
            {
                if (assetHandle == IntPtr.Zero)
                    return;

                GCHandle handle = GCHandle.FromIntPtr(assetHandle);
                handle.Free();
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine($"FreeAsset warning: {ex.Message}");
            }
        }

        /// <summary>
        /// Gets the number of exports in a loaded UAsset.
        /// </summary>
        /// <param name="assetHandle">Handle to UAsset object</param>
        /// <returns>Number of exports, or -1 on error</returns>
        [UnmanagedCallersOnly(EntryPoint = "GetAssetExportCount")]
        public static int GetAssetExportCount(IntPtr assetHandle)
        {
            try
            {
                NativeErrors.ClearLastError();

                if (assetHandle == IntPtr.Zero)
                {
                    NativeErrors.SetLastError("GetAssetExportCount: handle is null");
                    return -1;
                }

                GCHandle handle = GCHandle.FromIntPtr(assetHandle);
                UAsset asset = handle.Target as UAsset;

                if (asset == null)
                {
                    NativeErrors.SetLastError("GetAssetExportCount: invalid handle");
                    return -1;
                }

                return asset.Exports?.Count ?? 0;
            }
            catch (Exception ex)
            {
                NativeErrors.SetLastError($"GetAssetExportCount exception: {ex.Message}");
                return -1;
            }
        }

        /// <summary>
        /// Serializes a UAsset to formatted JSON string.
        /// Memory ownership: C# allocates, caller must call FreeString.
        /// </summary>
        /// <param name="assetHandle">Handle to UAsset object</param>
        /// <returns>Pointer to UTF-8 JSON string, or IntPtr.Zero on failure</returns>
        [UnmanagedCallersOnly(EntryPoint = "SerializeAssetToJson")]
        public static IntPtr SerializeAssetToJson(IntPtr assetHandle)
        {
            try
            {
                NativeErrors.ClearLastError();

                if (assetHandle == IntPtr.Zero)
                {
                    NativeErrors.SetLastError("SerializeAssetToJson: handle is null");
                    return IntPtr.Zero;
                }

                GCHandle handle = GCHandle.FromIntPtr(assetHandle);
                UAsset asset = handle.Target as UAsset;

                if (asset == null)
                {
                    NativeErrors.SetLastError("SerializeAssetToJson: invalid handle");
                    return IntPtr.Zero;
                }

                // Serialize to JSON
                string json = asset.SerializeJson();

                // Format for readability
                JObject jsonObject = JObject.Parse(json);
                string formattedJson = jsonObject.ToString(Formatting.Indented);

                // Allocate and return
                return Marshal.StringToCoTaskMemUTF8(formattedJson);
            }
            catch (Exception ex)
            {
                NativeErrors.SetLastError($"SerializeAssetToJson exception: {ex.Message}");
                return IntPtr.Zero;
            }
        }

        /// <summary>
        /// Deserializes a JSON string to a UAsset object.
        /// Memory ownership: C# owns the UAsset, caller must call FreeAsset.
        /// </summary>
        /// <param name="json">Pointer to UTF-8 null-terminated JSON string</param>
        /// <returns>Handle to UAsset object, or IntPtr.Zero on failure</returns>
        [UnmanagedCallersOnly(EntryPoint = "DeserializeAssetFromJson")]
        public static IntPtr DeserializeAssetFromJson(IntPtr json)
        {
            try
            {
                NativeErrors.ClearLastError();

                if (json == IntPtr.Zero)
                {
                    NativeErrors.SetLastError("DeserializeAssetFromJson: json is null");
                    return IntPtr.Zero;
                }

                string managedJson = Marshal.PtrToStringUTF8(json);
                if (string.IsNullOrEmpty(managedJson))
                {
                    NativeErrors.SetLastError("DeserializeAssetFromJson: json is empty");
                    return IntPtr.Zero;
                }

                // Deserialize from JSON
                UAsset asset = UAsset.DeserializeJson(managedJson);

                if (asset == null)
                {
                    NativeErrors.SetLastError("DeserializeAssetFromJson: deserialization returned null");
                    return IntPtr.Zero;
                }

                // Allocate GCHandle
                GCHandle handle = GCHandle.Alloc(asset);

                return GCHandle.ToIntPtr(handle);
            }
            catch (Exception ex)
            {
                NativeErrors.SetLastError($"DeserializeAssetFromJson exception: {ex.Message}");
                return IntPtr.Zero;
            }
        }

        /// <summary>
        /// Writes a UAsset object to .uasset/.uexp files.
        /// </summary>
        /// <param name="assetHandle">Handle to UAsset object</param>
        /// <param name="path">Pointer to UTF-8 null-terminated output path</param>
        /// <param name="mappingsHandle">Optional handle to Usmap object (can be IntPtr.Zero)</param>
        /// <returns>1 on success, 0 on failure</returns>
        [UnmanagedCallersOnly(EntryPoint = "WriteAssetToFile")]
        public static int WriteAssetToFile(IntPtr assetHandle, IntPtr path, IntPtr mappingsHandle)
        {
            try
            {
                NativeErrors.ClearLastError();

                if (assetHandle == IntPtr.Zero)
                {
                    NativeErrors.SetLastError("WriteAssetToFile: assetHandle is null");
                    return 0;
                }

                if (path == IntPtr.Zero)
                {
                    NativeErrors.SetLastError("WriteAssetToFile: path is null");
                    return 0;
                }

                // Get asset
                GCHandle handle = GCHandle.FromIntPtr(assetHandle);
                UAsset asset = handle.Target as UAsset;

                if (asset == null)
                {
                    NativeErrors.SetLastError("WriteAssetToFile: invalid asset handle");
                    return 0;
                }

                // Marshal path
                string managedPath = Marshal.PtrToStringUTF8(path);
                if (string.IsNullOrEmpty(managedPath))
                {
                    NativeErrors.SetLastError("WriteAssetToFile: path is empty");
                    return 0;
                }

                // Apply mappings if provided
                if (mappingsHandle != IntPtr.Zero)
                {
                    try
                    {
                        GCHandle mappingsGCHandle = GCHandle.FromIntPtr(mappingsHandle);
                        Usmap mappings = mappingsGCHandle.Target as Usmap;
                        if (mappings != null)
                        {
                            asset.Mappings = mappings;
                        }
                    }
                    catch (Exception ex)
                    {
                        NativeErrors.SetLastError($"WriteAssetToFile: invalid mappings handle: {ex.Message}");
                        return 0;
                    }
                }

                // Write to file
                asset.Write(managedPath);

                return 1; // Success
            }
            catch (Exception ex)
            {
                NativeErrors.SetLastError($"WriteAssetToFile exception: {ex.Message}");
                return 0;
            }
        }

        /// <summary>
        /// Gets the file path of a loaded UAsset.
        /// Memory ownership: C# allocates, caller must call FreeString.
        /// </summary>
        /// <param name="assetHandle">Handle to UAsset object</param>
        /// <returns>Pointer to UTF-8 path string, or IntPtr.Zero on failure</returns>
        [UnmanagedCallersOnly(EntryPoint = "GetAssetFilePath")]
        public static IntPtr GetAssetFilePath(IntPtr assetHandle)
        {
            try
            {
                NativeErrors.ClearLastError();

                if (assetHandle == IntPtr.Zero)
                {
                    NativeErrors.SetLastError("GetAssetFilePath: handle is null");
                    return IntPtr.Zero;
                }

                GCHandle handle = GCHandle.FromIntPtr(assetHandle);
                UAsset asset = handle.Target as UAsset;

                if (asset == null)
                {
                    NativeErrors.SetLastError("GetAssetFilePath: invalid handle");
                    return IntPtr.Zero;
                }

                string filePath = asset.FilePath ?? string.Empty;

                return Marshal.StringToCoTaskMemUTF8(filePath);
            }
            catch (Exception ex)
            {
                NativeErrors.SetLastError($"GetAssetFilePath exception: {ex.Message}");
                return IntPtr.Zero;
            }
        }
    }
}
