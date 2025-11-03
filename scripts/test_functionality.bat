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
.\retoc\retoc.exe --help
if %errorlevel% neq 0 (
    echo ERROR: Retoc test failed!
    pause
    exit /b 1
)
echo Retoc test passed!

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
if exist "retoc\retoc.exe" (
    echo retoc.exe found!
) else (
    echo ERROR: retoc.exe not found!
    pause
    exit /b 1
)

if exist "retoc\oo2core_9_win64.dll" (
    echo oo2core_9_win64.dll found!
) else (
    echo ERROR: oo2core_9_win64.dll not found!
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
