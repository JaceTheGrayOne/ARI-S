# Remove Git from PATH to avoid link.exe conflict
$env:PATH = ($env:PATH -split ';' | Where-Object { $_ -notlike '*Git*' }) -join ';'

# Add VS Installer to PATH
$env:PATH += ';C:\Program Files (x86)\Microsoft Visual Studio\Installer'

# Set up VS 2019 Build Tools environment
$vcvarsPath = 'C:\Program Files (x86)\Microsoft Visual Studio\2019\BuildTools\VC\Auxiliary\Build\vcvars64.bat'
cmd /c "`"$vcvarsPath`" && set" | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}

Write-Host "Using Visual Studio 2019 Build Tools" -ForegroundColor Cyan

# Navigate to project directory
Set-Location 'D:\Development\ARIS\ARI-S'

# Build the NativeAOT DLL
dotnet publish UAssetBridge\UAssetBridge\UAssetBridge.csproj `
    -r win-x64 `
    -c Release `
    -o bin\nativeaot\win-x64 `
    /p:NativeLib=Shared `
    /p:SelfContained=true

# Report status
if ($LASTEXITCODE -eq 0) {
    Write-Host "`n=== BUILD SUCCESSFUL ===" -ForegroundColor Green
    Write-Host "DLL Location: bin\nativeaot\win-x64\UAssetBridge.dll" -ForegroundColor Green
    Get-ChildItem bin\nativeaot\win-x64\UAssetBridge.dll | Select-Object Name, Length
} else {
    Write-Host "`n=== BUILD FAILED ===" -ForegroundColor Red
    exit $LASTEXITCODE
}
