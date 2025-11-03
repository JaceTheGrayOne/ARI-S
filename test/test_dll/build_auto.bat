@echo off
setlocal enabledelayedexpansion

echo ================================================
echo ARI-S Test DLL Auto-Builder
echo ================================================
echo.

REM Try to find Visual Studio compiler
echo Searching for Visual Studio compiler...
echo.

REM Check for VS 2022
set "VSPATH=C:\Program Files\Microsoft Visual Studio\2022"
if exist "%VSPATH%" (
    echo Found Visual Studio 2022
    for %%e in (Community Professional Enterprise) do (
        if exist "%VSPATH%\%%e\VC\Auxiliary\Build\vcvars64.bat" (
            echo Using VS 2022 %%e
            call "%VSPATH%\%%e\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
            goto :compile
        )
    )
)

REM Check for VS 2019
set "VSPATH=C:\Program Files (x86)\Microsoft Visual Studio\2019"
if exist "%VSPATH%" (
    echo Found Visual Studio 2019
    for %%e in (Community Professional Enterprise) do (
        if exist "%VSPATH%\%%e\VC\Auxiliary\Build\vcvars64.bat" (
            echo Using VS 2019 %%e
            call "%VSPATH%\%%e\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
            goto :compile
        )
    )
)

REM Check for Build Tools
set "VSPATH=C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools"
if exist "%VSPATH%\VC\Auxiliary\Build\vcvars64.bat" (
    echo Found VS 2022 Build Tools
    call "%VSPATH%\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
    goto :compile
)

set "VSPATH=C:\Program Files (x86)\Microsoft Visual Studio\2019\BuildTools"
if exist "%VSPATH%\VC\Auxiliary\Build\vcvars64.bat" (
    echo Found VS 2019 Build Tools
    call "%VSPATH%\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
    goto :compile
)

echo.
echo ================================================
echo ERROR: No compiler found!
echo ================================================
echo.
echo Please install one of the following:
echo.
echo 1. Visual Studio 2022 or 2019 (Community Edition is free)
echo    Download from: https://visualstudio.microsoft.com/downloads/
echo.
echo 2. OR Build Tools for Visual Studio
echo    Download from: https://visualstudio.microsoft.com/downloads/
echo    (Scroll down to "Tools for Visual Studio")
echo.
echo After installation, run this script again.
echo.
pause
exit /b 1

:compile
echo.
echo Compiling TestMessageBox.dll...
echo.

REM Clean old files
if exist TestMessageBox.dll del TestMessageBox.dll
if exist TestMessageBox.obj del TestMessageBox.obj
if exist TestMessageBox.exp del TestMessageBox.exp
if exist TestMessageBox.lib del TestMessageBox.lib

REM Compile
cl /LD /MD /O2 /W3 TestMessageBox.c user32.lib /Fe:TestMessageBox.dll

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ================================================
    echo SUCCESS! DLL Built Successfully
    echo ================================================
    echo.
    echo Output: TestMessageBox.dll
    echo Location: %~dp0TestMessageBox.dll
    echo.
    echo You can now use this DLL with ARI-S injector!
    echo.
    echo Quick test:
    echo 1. Launch notepad.exe
    echo 2. Run ARI-S as Administrator
    echo 3. Inject this DLL into notepad
    echo 4. You should see a message box appear!
    echo.
) else (
    echo.
    echo ================================================
    echo ERROR: Compilation failed
    echo ================================================
    echo.
)

pause
