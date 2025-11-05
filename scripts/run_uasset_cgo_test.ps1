Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Running UAssetBridge CGO Test ===" -ForegroundColor Cyan

# Put MSYS2 gcc at the front of PATH - this is critical
$env:PATH = "C:\msys64\ucrt64\bin;" + $env:PATH
$env:CGO_ENABLED = "1"

Write-Host "Environment:" -ForegroundColor Yellow
Write-Host "  CGO_ENABLED: $env:CGO_ENABLED" -ForegroundColor Yellow
Write-Host "  GCC: $(Get-Command gcc | Select-Object -ExpandProperty Source)" -ForegroundColor Yellow
Write-Host "  UAssetBridge.dll exists (root): $(Test-Path 'UAssetBridge.dll')" -ForegroundColor Yellow
Write-Host "  UAssetBridge.dll exists (uasset): $(Test-Path 'internal\uasset\UAssetBridge.dll')" -ForegroundColor Yellow

Write-Host "`nRunning test..." -ForegroundColor Cyan
go test -v -tags cgo -run TestNativeUAssetAPI_GetVersion ./internal/uasset 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n=== TEST PASSED! ===" -ForegroundColor Green
} else {
    Write-Host "`n=== TEST FAILED with exit code: $LASTEXITCODE ===" -ForegroundColor Red
}
