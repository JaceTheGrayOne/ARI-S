Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Running CGO Test ===" -ForegroundColor Cyan

# Set CGO environment variables with correct PowerShell syntax
$env:CGO_ENABLED = "1"
$env:CGO_LDFLAGS = "-L. -lUAssetBridge"
$env:CGO_CFLAGS = "-I."

Write-Host "Environment variables set:" -ForegroundColor Yellow
Write-Host "  CGO_ENABLED: $env:CGO_ENABLED" -ForegroundColor Yellow
Write-Host "  CGO_LDFLAGS: $env:CGO_LDFLAGS" -ForegroundColor Yellow
Write-Host "  CGO_CFLAGS: $env:CGO_CFLAGS" -ForegroundColor Yellow
Write-Host "  Working directory: $(Get-Location)" -ForegroundColor Yellow
Write-Host "  UAssetBridge.dll exists: $(Test-Path 'UAssetBridge.dll')" -ForegroundColor Yellow

Write-Host "`nRunning test..." -ForegroundColor Cyan
go test -v -tags cgo -run TestNativeUAssetAPI_GetVersion ./internal/uasset

Write-Host "`nTest completed with exit code: $LASTEXITCODE" -ForegroundColor $(if ($LASTEXITCODE -eq 0) { "Green" } else { "Red" })
