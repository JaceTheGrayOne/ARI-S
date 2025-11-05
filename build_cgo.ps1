Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Building ARI-S with CGO/NativeAOT Support ===" -ForegroundColor Cyan

# Put MSYS2 gcc at the front of PATH for CGO
$env:PATH = "C:\msys64\ucrt64\bin;" + $env:PATH
$env:CGO_ENABLED = "1"

Write-Host "Environment:" -ForegroundColor Yellow
Write-Host "  CGO_ENABLED: $env:CGO_ENABLED" -ForegroundColor Yellow
Write-Host "  GCC: $(Get-Command gcc | Select-Object -ExpandProperty Source)" -ForegroundColor Yellow

Write-Host "`nBuilding with CGO and production tags..." -ForegroundColor Yellow
go build -tags "production cgo" -trimpath -ldflags="-w -s -H windowsgui" -o bin\ARI-S.exe 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nBuild successful!" -ForegroundColor Green
    Write-Host "Output: bin\ARI-S.exe" -ForegroundColor Green

    # Get file size
    $fileInfo = Get-Item "bin\ARI-S.exe"
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "Size: $sizeMB MB" -ForegroundColor Cyan

    # Copy UAssetBridge.dll next to the executable
    Write-Host "`nCopying UAssetBridge.dll to bin directory..." -ForegroundColor Yellow
    $dllSource = "internal\uasset\UAssetBridge.dll"
    $dllDest = "bin\UAssetBridge.dll"

    if (Test-Path $dllSource) {
        Copy-Item $dllSource $dllDest -Force
        $dllInfo = Get-Item $dllDest
        $dllSizeMB = [math]::Round($dllInfo.Length / 1MB, 2)
        Write-Host "UAssetBridge.dll copied ($dllSizeMB MB)" -ForegroundColor Green
    } else {
        Write-Host "WARNING: UAssetBridge.dll not found at $dllSource" -ForegroundColor Red
        Write-Host "The CGO build requires UAssetBridge.dll to be next to the .exe" -ForegroundColor Red
    }
} else {
    Write-Host "`nBuild failed with exit code: $LASTEXITCODE" -ForegroundColor Red
}
