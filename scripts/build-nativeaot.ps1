#Requires -Version 5.0

<#
.SYNOPSIS
    Builds UAssetBridge as a NativeAOT shared library.

.DESCRIPTION
    This script compiles the UAssetBridge C# project using .NET NativeAOT,
    producing a single native DLL with no .NET runtime dependencies.

.PARAMETER Configuration
    Build configuration (Debug or Release). Default: Release

.PARAMETER Runtime
    Target runtime identifier. Default: win-x64
    Options: win-x64, win-arm64, linux-x64, linux-arm64, osx-x64, osx-arm64

.PARAMETER OutputDir
    Output directory for the native library. Default: .\bin\nativeaot\{runtime}

.EXAMPLE
    .\build-nativeaot.ps1
    Builds Release configuration for win-x64

.EXAMPLE
    .\build-nativeaot.ps1 -Configuration Debug -Runtime linux-x64
    Builds Debug configuration for linux-x64
#>

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("Debug", "Release")]
    [string]$Configuration = "Release",

    [Parameter(Mandatory=$false)]
    [ValidateSet("win-x64", "win-arm64", "linux-x64", "linux-arm64", "osx-x64", "osx-arm64")]
    [string]$Runtime = "win-x64",

    [Parameter(Mandatory=$false)]
    [string]$OutputDir = ""
)

$ErrorActionPreference = "Stop"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "UAssetBridge NativeAOT Build Script" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Resolve paths
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$csprojPath = Join-Path $projectRoot "UAssetBridge\UAssetBridge\UAssetBridge.csproj"

if (-not (Test-Path $csprojPath)) {
    Write-Host "ERROR: Project file not found: $csprojPath" -ForegroundColor Red
    exit 1
}

# Set output directory
if ([string]::IsNullOrEmpty($OutputDir)) {
    $OutputDir = Join-Path $projectRoot "bin\nativeaot\$Runtime"
}

Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Project:       $csprojPath" -ForegroundColor Gray
Write-Host "  Configuration: $Configuration" -ForegroundColor Gray
Write-Host "  Runtime:       $Runtime" -ForegroundColor Gray
Write-Host "  Output:        $OutputDir" -ForegroundColor Gray
Write-Host ""

# Check for .NET SDK
Write-Host "Checking .NET SDK..." -ForegroundColor Yellow
try {
    $dotnetVersion = dotnet --version
    Write-Host "  Found .NET SDK: $dotnetVersion" -ForegroundColor Green
} catch {
    Write-Host "ERROR: .NET SDK not found. Please install .NET 8 or later." -ForegroundColor Red
    Write-Host "Download from: https://dotnet.microsoft.com/download" -ForegroundColor Gray
    exit 1
}

# Clean previous build
Write-Host ""
Write-Host "Cleaning previous build..." -ForegroundColor Yellow
$binPath = Join-Path (Split-Path -Parent $csprojPath) "bin"
$objPath = Join-Path (Split-Path -Parent $csprojPath) "obj"

if (Test-Path $binPath) {
    Remove-Item -Recurse -Force $binPath
    Write-Host "  Removed: $binPath" -ForegroundColor Gray
}
if (Test-Path $objPath) {
    Remove-Item -Recurse -Force $objPath
    Write-Host "  Removed: $objPath" -ForegroundColor Gray
}

# Build with NativeAOT
Write-Host ""
Write-Host "Building NativeAOT library..." -ForegroundColor Yellow
Write-Host "  (This may take several minutes on first build)" -ForegroundColor Gray
Write-Host ""

$publishArgs = @(
    "publish",
    $csprojPath,
    "-r", $Runtime,
    "-c", $Configuration,
    "-o", $OutputDir,
    "/p:NativeLib=Shared",
    "/p:SelfContained=true",
    "--nologo"
)

Write-Host "Command: dotnet $($publishArgs -join ' ')" -ForegroundColor DarkGray
Write-Host ""

$stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

try {
    & dotnet @publishArgs

    if ($LASTEXITCODE -ne 0) {
        throw "dotnet publish failed with exit code $LASTEXITCODE"
    }
} catch {
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "BUILD FAILED" -ForegroundColor Red
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}

$stopwatch.Stop()

# Verify output
Write-Host ""
Write-Host "Verifying build output..." -ForegroundColor Yellow

$expectedFiles = @()
switch -Regex ($Runtime) {
    "^win-" {
        $expectedFiles += "UAssetBridge.dll"
        $expectedFiles += "UAssetBridge.lib"  # Import library for linking
    }
    "^linux-" {
        $expectedFiles += "libUAssetBridge.so"
    }
    "^osx-" {
        $expectedFiles += "libUAssetBridge.dylib"
    }
}

$allFound = $true
foreach ($file in $expectedFiles) {
    $filePath = Join-Path $OutputDir $file
    if (Test-Path $filePath) {
        $fileSize = (Get-Item $filePath).Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        Write-Host "  $file - ${fileSizeMB} MB" -ForegroundColor Green
    } else {
        Write-Host "  $file - NOT FOUND" -ForegroundColor Red
        $allFound = $false
    }
}

Write-Host ""

if ($allFound) {
    Write-Host "================================================" -ForegroundColor Green
    Write-Host "BUILD SUCCESSFUL" -ForegroundColor Green
    Write-Host "================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Build time: $($stopwatch.Elapsed.ToString('mm\:ss'))" -ForegroundColor Gray
    Write-Host "Output directory: $OutputDir" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  1. Copy the DLL to your Go project directory" -ForegroundColor Gray
    Write-Host "  2. Run: .\scripts\build-go.ps1" -ForegroundColor Gray
    Write-Host ""

    # Provide copy command for convenience
    $goProjectDir = Join-Path $projectRoot "."
    Write-Host "Quick copy command:" -ForegroundColor Cyan
    $mainDll = $expectedFiles[0]
    Write-Host "  Copy-Item '$OutputDir\$mainDll' -Destination '$goProjectDir'" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "BUILD INCOMPLETE" -ForegroundColor Red
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "Some expected files were not found." -ForegroundColor Red
    exit 1
}
