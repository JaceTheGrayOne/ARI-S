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
cd ..\..\..\Resources\retoc
copy retoc.exe ..\..\ARI-S\build\
copy oo2core_9_win64.dll ..\..\ARI-S\build\

echo.
echo Setting up UAssetAPI directory...
cd ..\..\ARI-S
if not exist bin\UAssetAPI mkdir bin\UAssetAPI
for %%f in (build\*.*) do copy /Y "%%f" bin\UAssetAPI\ >nul 2>&1

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
echo Executable: build\bin\ARI-S.exe
pause
