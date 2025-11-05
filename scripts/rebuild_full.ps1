Set-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Full ARI-S Rebuild (Frontend + Backend) ===" -ForegroundColor Cyan

# Step 1: Regenerate Wails bindings
Write-Host "`nStep 1: Regenerating Wails bindings..." -ForegroundColor Yellow
wails3 generate bindings
if ($LASTEXITCODE -ne 0) {
    Write-Host "WARNING: Binding generation had warnings, continuing..." -ForegroundColor Yellow
}

# Step 2: Build frontend
Write-Host "`nStep 2: Building frontend..." -ForegroundColor Yellow
Set-Location "frontend"
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Frontend build failed!" -ForegroundColor Red
    exit 1
}
Set-Location ".."

# Step 3: Build backend with CGO
Write-Host "`nStep 3: Building backend with CGO..." -ForegroundColor Yellow
$env:PATH = "C:\msys64\ucrt64\bin;" + $env:PATH
$env:CGO_ENABLED = "1"

go build -tags "production cgo" -trimpath -ldflags="-w -s -H windowsgui" -o bin\ARI-S.exe 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nBuild successful!" -ForegroundColor Green

    # Get file sizes
    $exeInfo = Get-Item "bin\ARI-S.exe"
    $exeSizeMB = [math]::Round($exeInfo.Length / 1MB, 2)
    Write-Host "ARI-S.exe: $exeSizeMB MB" -ForegroundColor Cyan

    # Copy UAssetBridge.dll
    Write-Host "`nCopying UAssetBridge.dll..." -ForegroundColor Yellow
    Copy-Item "internal\uasset\UAssetBridge.dll" "bin\UAssetBridge.dll" -Force
    $dllInfo = Get-Item "bin\UAssetBridge.dll"
    $dllSizeMB = [math]::Round($dllInfo.Length / 1MB, 2)
    Write-Host "UAssetBridge.dll: $dllSizeMB MB" -ForegroundColor Cyan

    Write-Host "`n=== Build Complete ===" -ForegroundColor Green
    Write-Host "Ready to run: bin\ARI-S.exe" -ForegroundColor Green
} else {
    Write-Host "`nBackend build failed!" -ForegroundColor Red
    exit 1
}
