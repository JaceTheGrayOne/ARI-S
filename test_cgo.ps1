# Set CGO environment variables
$env:CGO_ENABLED = "1"
$env:CGO_LDFLAGS = "-L. -lUAssetBridge"
$env:CGO_CFLAGS = "-I."

# Navigate to project directory
Set-Location "D:\Development\ARIS\ARI-S"

# Run the proof-of-concept test
Write-Host "Running CGO proof-of-concept test..." -ForegroundColor Cyan
go test -v -tags cgo -run TestNativeUAssetAPI_GetVersion ./internal/uasset
