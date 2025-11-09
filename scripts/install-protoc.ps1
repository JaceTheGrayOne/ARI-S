#!/usr/bin/env pwsh

# Script to download and install protoc (Protocol Buffer Compiler)

param(
    [string]$Version = "29.3",  # Latest stable as of Nov 2025
    [string]$InstallDir = "tools/protoc"
)

$ErrorActionPreference = "Stop"

Write-Host "================================" -ForegroundColor Cyan
Write-Host " Installing protoc $Version" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Determine platform
$Platform = "win64"
$FileName = "protoc-$Version-$Platform.zip"
$Url = "https://github.com/protocolbuffers/protobuf/releases/download/v$Version/$FileName"

Write-Host "Download URL: $Url" -ForegroundColor Gray
Write-Host "Install directory: $InstallDir" -ForegroundColor Gray
Write-Host ""

# Create install directory
if (-not (Test-Path $InstallDir)) {
    Write-Host "Creating install directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Download
$ZipPath = Join-Path $InstallDir $FileName
Write-Host "Downloading protoc..." -ForegroundColor Green
try {
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath -UseBasicParsing
    Write-Host "  [OK] Download complete" -ForegroundColor Green
} catch {
    Write-Host "  [ERROR] Download failed: $_" -ForegroundColor Red
    exit 1
}

# Extract
Write-Host ""
Write-Host "Extracting..." -ForegroundColor Green
try {
    Expand-Archive -Path $ZipPath -DestinationPath $InstallDir -Force
    Write-Host "  [OK] Extraction complete" -ForegroundColor Green
} catch {
    Write-Host "  [ERROR] Extraction failed: $_" -ForegroundColor Red
    exit 1
}

# Clean up zip
Remove-Item $ZipPath

# Verify
$ProtocExe = Join-Path $InstallDir "bin/protoc.exe"
if (Test-Path $ProtocExe) {
    Write-Host ""
    Write-Host "[SUCCESS] protoc installed successfully!" -ForegroundColor Green
    Write-Host ""

    # Test version
    $VersionOutput = & $ProtocExe --version
    Write-Host "  Version: $VersionOutput" -ForegroundColor Gray
    Write-Host "  Path: $ProtocExe" -ForegroundColor Gray

    Write-Host ""
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host " Installation complete!" -ForegroundColor Green
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Add to PATH (temporary):" -ForegroundColor Yellow
    Write-Host "  `$env:PATH = `"$(Resolve-Path $InstallDir)/bin;`$env:PATH`"" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Or use full path in generate-proto.ps1" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "[ERROR] Installation verification failed" -ForegroundColor Red
    exit 1
}
