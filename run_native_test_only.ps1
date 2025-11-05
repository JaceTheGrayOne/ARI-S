Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Running UAssetBridge Native CGO Test ===" -ForegroundColor Cyan

# Put MSYS2 gcc at the front of PATH
$env:PATH = "C:\msys64\ucrt64\bin;" + $env:PATH
$env:CGO_ENABLED = "1"

Write-Host "Running specific test file..." -ForegroundColor Cyan
go test -v -tags cgo internal/uasset/uasset_native.go internal/uasset/uasset_native_test.go 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n=== TEST PASSED! ===" -ForegroundColor Green
} else {
    Write-Host "`n=== TEST FAILED with exit code: $LASTEXITCODE ===" -ForegroundColor Red
}
