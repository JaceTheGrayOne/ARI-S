Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Running CGO Test (Verbose) ===" -ForegroundColor Cyan

# Set CGO environment variables with correct PowerShell syntax
$env:CGO_ENABLED = "1"
$env:CGO_LDFLAGS = "-L. -lUAssetBridge"
$env:CGO_CFLAGS = "-I."

Write-Host "`nRunning verbose build..." -ForegroundColor Cyan
go test -v -x -tags cgo -run TestNativeUAssetAPI_GetVersion ./internal/uasset 2>&1 | Tee-Object -FilePath "cgo_build.log"

Write-Host "`nTest completed with exit code: $LASTEXITCODE" -ForegroundColor $(if ($LASTEXITCODE -eq 0) { "Green" } else { "Red" })

if ($LASTEXITCODE -ne 0) {
    Write-Host "`nShowing last 50 lines of build log:" -ForegroundColor Yellow
    Get-Content "cgo_build.log" -Tail 50
}
