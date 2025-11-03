@echo off
REM Build TestMessageBox.dll using MSVC (Visual Studio)
REM
REM Requirements:
REM - Visual Studio 2019 or 2022 installed
REM - Run from "x64 Native Tools Command Prompt for VS 2022"
REM
REM Usage: Just run this batch file from the x64 Native Tools prompt

echo ================================================
echo Building TestMessageBox.dll (64-bit) with MSVC
echo ================================================
echo.

REM Clean previous builds
if exist TestMessageBox.dll del TestMessageBox.dll
if exist TestMessageBox.obj del TestMessageBox.obj
if exist TestMessageBox.exp del TestMessageBox.exp
if exist TestMessageBox.lib del TestMessageBox.lib

REM Compile and link
REM /LD - Create DLL
REM /MD - Use multithreaded DLL runtime
REM /O2 - Optimize for speed
REM /W3 - Warning level 3
REM user32.lib - Required for MessageBox
cl /LD /MD /O2 /W3 TestMessageBox.c user32.lib /Fe:TestMessageBox.dll

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ================================================
    echo Build successful!
    echo Output: TestMessageBox.dll
    echo ================================================
    echo.
    echo You can now test this DLL with ARI-S injector.
    echo Inject it into any running process (e.g., notepad.exe)
    echo.
) else (
    echo.
    echo ================================================
    echo Build failed! Error code: %ERRORLEVEL%
    echo ================================================
    echo.
    echo Make sure you're running from:
    echo "x64 Native Tools Command Prompt for VS 2022"
    echo.
)

pause
