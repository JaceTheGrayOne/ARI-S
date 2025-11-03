# Test Message Box DLL

This is a simple test DLL for verifying DLL injection functionality in ARI-S.

## What It Does

When injected into a process, this DLL displays a message box confirming successful injection. This is a safe, harmless test that allows you to verify your injector is working correctly before using it with more complex DLLs like UnrealMappingsDumper or Dumper7.

## Building the DLL

### Option 1: Using Visual Studio (MSVC)

1. Open **"x64 Native Tools Command Prompt for VS 2022"** (or 2019)
2. Navigate to this directory
3. Run: `build_msvc.bat`

### Option 2: Using MinGW-w64

1. Install MinGW-w64 (via MSYS2 or standalone)
2. Open a command prompt
3. Navigate to this directory
4. Run: `build_mingw.bat`

### Manual Compilation

**MSVC:**
```cmd
cl /LD /MD TestMessageBox.c user32.lib /Fe:TestMessageBox.dll
```

**MinGW:**
```cmd
gcc -shared -o TestMessageBox.dll TestMessageBox.c -luser32
```

## Testing the DLL

### Method 1: Using ARI-S (Your Injector)

1. Launch a test process (e.g., `notepad.exe`)
2. Open ARI-S
3. Navigate to the DLL Injector pane
4. Click "Browse" and select `TestMessageBox.dll`
5. Click "Refresh" to load processes
6. Select your test process (e.g., "notepad.exe")
7. Click "Inject DLL"
8. If successful, you should see a message box appear!

### Method 2: Using Third-Party Injector

To verify the DLL itself works before testing ARI-S:

1. Use a known-good injector (Extreme Injector, Process Hacker, etc.)
2. Inject `TestMessageBox.dll` into any running process
3. Message box should appear immediately

## What You Should See

**Success:**
- A message box with title "ARI-S Injection Test"
- Message: "DLL Injection Successful!"
- Detailed information about the injection

**Failure Scenarios:**

| Error | Cause | Solution |
|-------|-------|----------|
| "Access denied" | Not running as admin | Run ARI-S as Administrator |
| "LoadLibraryW failed" | DLL architecture mismatch | Ensure DLL is 64-bit for 64-bit processes |
| "Module not found" | DLL path incorrect | Check the full path is correct |
| No message box | DLL injected but DllMain didn't run | Check target process is still running |

## Architecture Notes

**Important:** This DLL must match the architecture of the target process:

- **64-bit DLL** (built with x64 compiler) → **64-bit process**
- **32-bit DLL** (built with x86 compiler) → **32-bit process**

The build scripts default to **64-bit** since Grounded 2 and most modern games are 64-bit.

To build 32-bit (for testing with 32-bit processes):
```cmd
# MSVC: Use "x86 Native Tools Command Prompt"
cl /LD /MD TestMessageBox.c user32.lib /Fe:TestMessageBox.dll

# MinGW: Add -m32 flag
gcc -shared -m32 -o TestMessageBox.dll TestMessageBox.c -luser32
```

## Safe Testing Targets

Good processes to test injection on:

1. **notepad.exe** - Simple, single-threaded, easy to restart
2. **calc.exe** - Calculator app, safe to inject into
3. **explorer.exe** - File explorer (be careful, can affect desktop)
4. **Your own test program** - Launch a simple program you create

**Avoid injecting into:**
- System-critical processes (csrss.exe, winlogon.exe, etc.)
- Antivirus processes
- Processes with anti-cheat protection

## Cleanup

After testing, you can safely:
- Close the injected process (the DLL unloads automatically)
- Delete the compiled DLL file
- The injection is temporary and leaves no permanent changes

## Next Steps

Once this test DLL works successfully:

1. ✅ Your injector is working correctly
2. ✅ You can proceed to test with UnrealMappingsDumper.dll
3. ✅ You can confidently use ARI-S for Grounded 2 modding

## Troubleshooting

**Message box doesn't appear:**
- Check if target process is 64-bit (Task Manager → Details → right-click columns → Platform)
- Verify DLL was compiled as 64-bit
- Check ARI-S console output for error messages
- Try injecting into notepad.exe first as a baseline test

**"Access denied" errors:**
- Right-click ARI-S.exe → "Run as administrator"
- Some processes require elevated privileges to inject into

**DLL compiles but doesn't show message:**
- Verify MessageBoxW is being called (add MessageBoxA as alternative)
- Check if target process has a GUI thread
- Try injecting into a known GUI process like notepad.exe

## Additional Notes

This DLL is **completely safe** to inject into test processes. It:
- Only displays a message box
- Doesn't modify process memory (except for loading itself)
- Doesn't hook any functions
- Unloads cleanly when the process exits
- Contains no malicious code whatsoever

It's specifically designed for testing and education purposes in modding research.
