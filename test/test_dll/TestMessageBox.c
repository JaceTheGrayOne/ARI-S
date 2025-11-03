/*
 * TestMessageBox.c
 *
 * Simple test DLL for verifying DLL injection functionality.
 * Displays a message box when injected into a process.
 *
 * Compile with:
 * x64 Native Tools Command Prompt for VS 2022:
 *   cl /LD /MD TestMessageBox.c user32.lib /Fe:TestMessageBox.dll
 *
 * Or with MinGW-w64:
 *   gcc -shared -o TestMessageBox.dll TestMessageBox.c -luser32
 */

#include <windows.h>

// DllMain is called when the DLL is loaded/unloaded
BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved)
{
    switch (fdwReason)
    {
        case DLL_PROCESS_ATTACH:
            // DLL is being loaded into a process
            // Display a message box to confirm successful injection
            MessageBoxW(
                NULL,
                L"DLL Injection Successful!\n\n"
                L"This message confirms that the DLL was successfully injected "
                L"into the target process using ARI-S.\n\n"
                L"Process ID: (check Task Manager)\n"
                L"DLL: TestMessageBox.dll",
                L"ARI-S Injection Test",
                MB_OK | MB_ICONINFORMATION
            );
            break;

        case DLL_PROCESS_DETACH:
            // DLL is being unloaded from a process
            // We don't show a message here as the process might be terminating
            break;

        case DLL_THREAD_ATTACH:
        case DLL_THREAD_DETACH:
            // We don't need to do anything for thread attach/detach
            break;
    }

    return TRUE; // Indicate success
}

/*
 * Optional: Export a test function that can be called after injection
 * This is useful for more complex testing scenarios
 */
__declspec(dllexport) void TestFunction(void)
{
    MessageBoxW(
        NULL,
        L"TestFunction() was called successfully!",
        L"ARI-S Test DLL",
        MB_OK | MB_ICONINFORMATION
    );
}
