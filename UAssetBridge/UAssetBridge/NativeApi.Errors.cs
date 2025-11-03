using System;
using System.Runtime.InteropServices;
using System.Text;

namespace UAssetBridge
{
    /// <summary>
    /// Thread-local error handling for native interop.
    /// Stores the last error message for diagnostic purposes.
    /// </summary>
    public static class NativeErrors
    {
        [ThreadStatic]
        private static string lastError;

        /// <summary>
        /// Sets the last error message for the current thread.
        /// </summary>
        /// <param name="error">Error message to store</param>
        internal static void SetLastError(string error)
        {
            lastError = error ?? string.Empty;
        }

        /// <summary>
        /// Clears the last error for the current thread.
        /// </summary>
        internal static void ClearLastError()
        {
            lastError = null;
        }

        /// <summary>
        /// Gets the last error message for the current thread.
        /// Returns empty string if no error is set.
        /// </summary>
        /// <returns>Last error message</returns>
        internal static string GetLastError()
        {
            return lastError ?? string.Empty;
        }

        /// <summary>
        /// Exports the last error string to unmanaged code.
        /// Memory ownership: C# allocates, caller must call FreeString to release.
        /// </summary>
        /// <returns>Pointer to UTF-8 encoded error string (never null)</returns>
        [UnmanagedCallersOnly(EntryPoint = "GetLastErrorString")]
        public static IntPtr GetLastErrorString()
        {
            try
            {
                string error = GetLastError();
                if (string.IsNullOrEmpty(error))
                {
                    error = "No error";
                }

                // Allocate unmanaged memory for the UTF-8 string
                return Marshal.StringToCoTaskMemUTF8(error);
            }
            catch (Exception ex)
            {
                // Fallback: allocate memory for a generic error message
                string fallback = $"Error retrieving last error: {ex.Message}";
                return Marshal.StringToCoTaskMemUTF8(fallback);
            }
        }

        /// <summary>
        /// Frees a string previously returned by any native API function.
        /// Must be called by the caller to prevent memory leaks.
        /// </summary>
        /// <param name="ptr">Pointer to string allocated by C# (can be IntPtr.Zero)</param>
        [UnmanagedCallersOnly(EntryPoint = "FreeString")]
        public static void FreeString(IntPtr ptr)
        {
            if (ptr != IntPtr.Zero)
            {
                Marshal.FreeCoTaskMem(ptr);
            }
        }
    }
}
