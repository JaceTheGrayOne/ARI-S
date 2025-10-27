@echo off
echo Building ARI.S...

echo.
echo Building UAssetBridge...
cd UAssetBridge\UAssetBridge
dotnet publish -c Release -r win-x64 --self-contained true -o ..\..\build
if %errorlevel% neq 0 (
    echo Failed to build UAssetBridge
    pause
    exit /b 1
)

echo.
echo Copying retoc files...
cd ..\..
copy retoc\retoc.exe build\
copy retoc\oo2core_9_win64.dll build\

echo.
echo Setting up bin directory structure...
if not exist bin\retoc mkdir bin\retoc
if not exist bin\UAssetAPI mkdir bin\UAssetAPI

echo Copying retoc to bin...
copy retoc\retoc.exe bin\retoc\
copy retoc\oo2core_9_win64.dll bin\retoc\
copy retoc\LICENSE bin\retoc\ 2>nul
copy retoc\README.md bin\retoc\ 2>nul

echo Copying UAssetBridge to bin...
copy build\UAssetBridge.exe bin\UAssetAPI\
copy build\UAssetBridge.dll bin\UAssetAPI\
copy build\UAssetAPI.dll bin\UAssetAPI\
copy build\*.json bin\UAssetAPI\ 2>nul
copy build\ZstdSharp.dll bin\UAssetAPI\ 2>nul
copy build\Newtonsoft.Json.dll bin\UAssetAPI\ 2>nul

echo.
echo Building Wails application...
set PRODUCTION=true
wails3 task build
if %errorlevel% neq 0 (
    echo Failed to build Wails application
    pause
    exit /b 1
)

echo.
echo Build completed successfully!
echo Executable: bin\ARI-S.exe
pause
