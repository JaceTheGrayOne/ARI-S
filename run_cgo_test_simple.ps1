Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Running CGO Test ===" -ForegroundColor Cyan
Write-Host "UAssetBridge.dll in root: $(Test-Path 'UAssetBridge.dll')" -ForegroundColor Yellow
Write-Host "UAssetBridge.dll in uasset: $(Test-Path 'internal\uasset\UAssetBridge.dll')" -ForegroundColor Yellow

# CGO directives are in the Go files, so no need to set env vars
Write-Host "`nRunning test..." -ForegroundColor Cyan
go test -v -tags cgo -run TestNativeUAssetAPI_GetVersion ./internal/uasset 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nTEST PASSED!" -ForegroundColor Green
} else {
    Write-Host "`nTEST FAILED with exit code: $LASTEXITCODE" -ForegroundColor Red
}
