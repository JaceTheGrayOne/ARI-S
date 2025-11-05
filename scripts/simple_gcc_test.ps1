Push-Location "D:\Development\ARIS\ARI-S"

Write-Host "=== Simple GCC Test ===" -ForegroundColor Cyan
Write-Host "Working directory: $(Get-Location)" -ForegroundColor Yellow
Write-Host "Test file exists: $(Test-Path test_gcc.c)" -ForegroundColor Yellow

try {
    Write-Host "`nAttempting compilation..." -ForegroundColor Cyan
    $result = Start-Process -FilePath "gcc" -ArgumentList "test_gcc.c", "-o", "test_gcc.exe" -Wait -NoNewWindow -PassThru -RedirectStandardError "gcc_error.txt" -RedirectStandardOutput "gcc_output.txt"

    Write-Host "Exit code: $($result.ExitCode)" -ForegroundColor Yellow

    if (Test-Path "gcc_output.txt") {
        Write-Host "`nStdout:" -ForegroundColor Cyan
        Get-Content "gcc_output.txt"
    }

    if (Test-Path "gcc_error.txt") {
        Write-Host "`nStderr:" -ForegroundColor Cyan
        Get-Content "gcc_error.txt"
    }

    if (Test-Path "test_gcc.exe") {
        Write-Host "`nExecutable created! Running..." -ForegroundColor Green
        & .\test_gcc.exe
    } else {
        Write-Host "`nExecutable NOT created" -ForegroundColor Red
    }
} finally {
    Pop-Location
}
