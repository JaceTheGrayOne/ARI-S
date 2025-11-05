Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "Testing GCC compilation..." -ForegroundColor Cyan
Write-Host "GCC path: $(Get-Command gcc | Select-Object -ExpandProperty Source)" -ForegroundColor Yellow

$output = & gcc test_gcc.c -o test_gcc.exe 2>&1 | Out-String
Write-Host $output

if ($LASTEXITCODE -eq 0) {
    Write-Host "GCC compilation successful!" -ForegroundColor Green
    Write-Host "Running test program..." -ForegroundColor Cyan
    .\test_gcc.exe
} else {
    Write-Host "GCC compilation failed with exit code: $LASTEXITCODE" -ForegroundColor Red
    Write-Host "Output: $output" -ForegroundColor Red
}
