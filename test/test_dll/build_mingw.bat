@echo off
REM Build TestMessageBox.dll using MinGW-w64
REM
REM Requirements:
REM - MinGW-w64 installed and in PATH
REM - Can be installed via MSYS2: pacman -S mingw-w64-x86_64-gcc
REM
REM Usage: Just run this batch file

echo ================================================
echo Building TestMessageBox.dll (64-bit) with MinGW
echo ================================================
echo.

REM Clean previous builds
if exist TestMessageBox.dll del TestMessageBox.dll
if exist TestMessageBox.o del TestMessageBox.o

REM Compile and link
REM -shared - Create DLL
REM -o - Output file name
REM -luser32 - Link against user32.dll for MessageBox
REM -m64 - Build 64-bit (should be default on x86_64 MinGW)
gcc -shared -m64 -o TestMessageBox.dll TestMessageBox.c -luser32

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
    echo Make sure MinGW-w64 is installed and in your PATH.
    echo.
)

pause
