Write-Host "=== CGO Diagnostics ===" -ForegroundColor Cyan

Write-Host "`nGo Environment:" -ForegroundColor Yellow
go env CGO_ENABLED
go env CC
go env CXX
go env CGO_CFLAGS
go env CGO_LDFLAGS

Write-Host "`nGCC Check:" -ForegroundColor Yellow
Write-Host "GCC path: $(Get-Command gcc -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source)"
Write-Host "GCC version:"
gcc --version | Select-Object -First 1

Write-Host "`nPATH (first 5 entries):" -ForegroundColor Yellow
$env:PATH -split ';' | Select-Object -First 5

Write-Host "`nAttempting simple CGO build with verbose output:" -ForegroundColor Yellow
$env:CGO_ENABLED = "1"
Set-Location "D:\Development\ARIS\ARI-S"
go build -x -tags cgo -work test_cgo_minimal.go 2>&1 | Select-Object -First 100
