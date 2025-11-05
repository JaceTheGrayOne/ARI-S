Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Attempting CGO build with full error capture ===" -ForegroundColor Cyan

# Ensure MSYS2 gcc is in PATH first
$env:PATH = "C:\msys64\ucrt64\bin;" + $env:PATH
$env:CGO_ENABLED = "1"

Write-Host "`nBuilding..." -ForegroundColor Yellow

# Capture both stdout and stderr to file
$output = go build -v -tags cgo test_cgo_minimal.go 2>&1 | Out-String
Write-Host $output

Write-Host "`nExit code: $LASTEXITCODE" -ForegroundColor $(if ($LASTEXITCODE -eq 0) { "Green" } else { "Red" })

if (Test-Path "test_cgo_minimal.exe") {
    Write-Host "`nExecutable created!" -ForegroundColor Green
    & .\test_cgo_minimal.exe
} else {
    Write-Host "`nExecutable NOT created" -ForegroundColor Red

    # Try to get more details from cgo directly
    Write-Host "`nTrying to run cgo manually..." -ForegroundColor Yellow
    $cgoTestFile = "C:\Program Files\Go\src\runtime\cgo\cgo.go"
    $workDir = "D:\Development\ARIS\ARI-S\test_cgo_work"
    New-Item -ItemType Directory -Force -Path $workDir | Out-Null

    Write-Host "Running cgo.exe directly..." -ForegroundColor Yellow
    & "C:\Program Files\Go\pkg\tool\windows_amd64\cgo.exe" -objdir $workDir -importpath runtime/cgo -import_runtime_cgo=false -import_syscall=false -- -I $workDir -O2 -g -Wall -Werror -fno-stack-protector -Wdeclaration-after-statement $cgoTestFile 2>&1
}
