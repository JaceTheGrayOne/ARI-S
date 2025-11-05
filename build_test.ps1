Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Building ARI-S Test Binary ===" -ForegroundColor Cyan

# Put MSYS2 gcc at the front of PATH for CGO
$env:PATH = "C:\msys64\ucrt64\bin;" + $env:PATH
$env:CGO_ENABLED = "1"

Write-Host "Building without CGO tags (production build)..." -ForegroundColor Yellow
go build -tags production -o ari-s-test.exe 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nBuild successful!" -ForegroundColor Green
    Write-Host "Output: ari-s-test.exe" -ForegroundColor Green

    # Get file size
    $fileInfo = Get-Item "ari-s-test.exe"
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "Size: $sizeMB MB" -ForegroundColor Cyan
} else {
    Write-Host "`nBuild failed with exit code: $LASTEXITCODE" -ForegroundColor Red
}
