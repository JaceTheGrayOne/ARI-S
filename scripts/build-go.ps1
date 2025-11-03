#Requires -Version 5.0

<#
.SYNOPSIS
    Builds the ARI-S Go application with CGO enabled for NativeAOT integration.

.DESCRIPTION
    This script sets up the CGO environment and builds the Go application
    with the native UAssetBridge library.

.PARAMETER SkipDLLCheck
    Skip checking for the NativeAOT DLL. Default: false

.EXAMPLE
    .\build-go.ps1
    Builds the Go application with CGO

.EXAMPLE
    .\build-go.ps1 -SkipDLLCheck
    Builds without checking for the DLL (for CI/CD)
#>

param(
    [Parameter(Mandatory=$false)]
    [switch]$SkipDLLCheck = $false
)

$ErrorActionPreference = "Stop"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "ARI-S Go Build Script (with CGO)" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Resolve paths
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir

Write-Host "Project root: $projectRoot" -ForegroundColor Gray
Write-Host ""

# Check for Go
Write-Host "Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "  $goVersion" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Go not found. Please install Go 1.20 or later." -ForegroundColor Red
    Write-Host "Download from: https://go.dev/dl/" -ForegroundColor Gray
    exit 1
}

# Check for GCC (required for CGO on Windows)
Write-Host ""
Write-Host "Checking for GCC (required for CGO)..." -ForegroundColor Yellow
try {
    $gccVersion = gcc --version | Select-Object -First 1
    Write-Host "  $gccVersion" -ForegroundColor Green
} catch {
    Write-Host "WARNING: GCC not found." -ForegroundColor Red
    Write-Host ""
    Write-Host "CGO requires a C compiler. On Windows, install MinGW-w64:" -ForegroundColor Yellow
    Write-Host "  1. Download from: https://www.mingw-w64.org/downloads/" -ForegroundColor Gray
    Write-Host "  2. Or use Chocolatey: choco install mingw" -ForegroundColor Gray
    Write-Host "  3. Ensure gcc.exe is in your PATH" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Alternatively, install TDM-GCC:" -ForegroundColor Yellow
    Write-Host "  https://jmeubank.github.io/tdm-gcc/download/" -ForegroundColor Gray
    Write-Host ""
    $continue = Read-Host "Continue anyway? (y/n)"
    if ($continue -ne 'y') {
        exit 1
    }
}

# Check for NativeAOT DLL
if (-not $SkipDLLCheck) {
    Write-Host ""
    Write-Host "Checking for NativeAOT DLL..." -ForegroundColor Yellow

    $dllPath = Join-Path $projectRoot "UAssetBridge.dll"

    if (Test-Path $dllPath) {
        $fileSize = (Get-Item $dllPath).Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        Write-Host "  Found: UAssetBridge.dll (${fileSizeMB} MB)" -ForegroundColor Green
    } else {
        Write-Host "  NOT FOUND: UAssetBridge.dll" -ForegroundColor Red
        Write-Host ""
        Write-Host "The NativeAOT DLL is required for CGO integration." -ForegroundColor Yellow
        Write-Host "Build it first by running:" -ForegroundColor Yellow
        Write-Host "  .\scripts\build-nativeaot.ps1" -ForegroundColor White
        Write-Host ""
        $continue = Read-Host "Continue anyway? (y/n)"
        if ($continue -ne 'y') {
            exit 1
        }
    }
}

# Set CGO environment variables
Write-Host ""
Write-Host "Setting CGO environment variables..." -ForegroundColor Yellow

$env:CGO_ENABLED = "1"
Write-Host "  CGO_ENABLED=1" -ForegroundColor Gray

# Set library path to project root (where DLL should be)
$env:CGO_LDFLAGS = "-L. -lUAssetBridge"
Write-Host "  CGO_LDFLAGS=-L. -lUAssetBridge" -ForegroundColor Gray

# Set include path
$env:CGO_CFLAGS = "-I."
Write-Host "  CGO_CFLAGS=-I." -ForegroundColor Gray

# Run wails build
Write-Host ""
Write-Host "Building ARI-S with Wails..." -ForegroundColor Yellow
Write-Host "  (This will compile Go code with CGO enabled)" -ForegroundColor Gray
Write-Host ""

$stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

try {
    Push-Location $projectRoot

    Write-Host "Command: wails3 build" -ForegroundColor DarkGray
    Write-Host ""

    & wails3 build

    if ($LASTEXITCODE -ne 0) {
        throw "wails3 build failed with exit code $LASTEXITCODE"
    }
} catch {
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "BUILD FAILED" -ForegroundColor Red
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}

$stopwatch.Stop()

# Verify output
Write-Host ""
Write-Host "Verifying build output..." -ForegroundColor Yellow

$exePath = Join-Path $projectRoot "bin\ARI-S.exe"

if (Test-Path $exePath) {
    $fileSize = (Get-Item $exePath).Length
    $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
    Write-Host "  ARI-S.exe - ${fileSizeMB} MB" -ForegroundColor Green

    Write-Host ""
    Write-Host "================================================" -ForegroundColor Green
    Write-Host "BUILD SUCCESSFUL" -ForegroundColor Green
    Write-Host "================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Build time: $($stopwatch.Elapsed.ToString('mm\:ss'))" -ForegroundColor Gray
    Write-Host "Output: $exePath" -ForegroundColor Gray
    Write-Host ""
    Write-Host "IMPORTANT: Runtime Requirements" -ForegroundColor Yellow
    Write-Host "  The UAssetBridge.dll must be in the same directory as ARI-S.exe" -ForegroundColor Gray
    Write-Host "  Copy it to the bin directory:" -ForegroundColor Gray
    Write-Host "    Copy-Item '.\UAssetBridge.dll' -Destination '.\bin\'" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host "  ARI-S.exe - NOT FOUND" -ForegroundColor Red
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "BUILD INCOMPLETE" -ForegroundColor Red
    Write-Host "================================================" -ForegroundColor Red
    exit 1
}
