@echo off
echo Testing ARI.S Functionality
echo =====================================

echo.
echo 1. Testing UAssetBridge...
echo -------------------------
.\build\uasset_bridge.exe --help
if %errorlevel% neq 0 (
    echo ERROR: UAssetBridge test failed!
    pause
    exit /b 1
)
echo UAssetBridge test passed!

echo.
echo 2. Testing Retoc...
echo ------------------
REM Prefer embedded dependency path; fall back to build output
set "RETOC_BIN=.\dependencies\retoc\retoc.exe"
if not exist "%RETOC_BIN%" set "RETOC_BIN=.\build\retoc.exe"
"%RETOC_BIN%" --help
if %errorlevel% neq 0 (
    echo ERROR: Retoc test failed! Tried: %RETOC_BIN%
    pause
    exit /b 1
)
echo Retoc test passed! Using: %RETOC_BIN%

echo.
echo 3. Testing Application Build...
echo ------------------------------
if exist "bin\ARI-S.exe" (
    echo Application executable found!
    echo File size:
    dir "bin\ARI-S.exe" | findstr ARI-S.exe
) else (
    echo ERROR: Application executable not found!
    pause
    exit /b 1
)

echo.
echo 4. Testing Required Dependencies...
echo ----------------------------------
if exist "dependencies\retoc\retoc.exe" (
    echo retoc.exe found in dependencies\retoc\
) else if exist "build\retoc.exe" (
    echo retoc.exe found in build\
) else (
    echo ERROR: retoc.exe not found in dependencies\retoc\ or build\
    pause
    exit /b 1
)

if exist "dependencies\retoc\oo2core_9_win64.dll" (
    echo oo2core_9_win64.dll found in dependencies\retoc\
) else (
    echo ERROR: oo2core_9_win64.dll not found in dependencies\retoc\
    pause
    exit /b 1
)

if exist "build\uasset_bridge.exe" (
    echo uasset_bridge.exe found!
) else (
    echo ERROR: uasset_bridge.exe not found!
    pause
    exit /b 1
)

echo.
echo =====================================
echo ALL TESTS PASSED!
echo ARI.S is ready for use.
echo =====================================
echo.
echo To run the application, execute:
echo .\bin\ARI-S.exe
echo.
pause
