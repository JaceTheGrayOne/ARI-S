# Automated Test DLL Builder
# Downloads portable MinGW and compiles TestMessageBox.dll

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "ARI-S Test DLL - Automated Builder" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$tempDir = "$env:TEMP\aris_compiler"
$compilerUrl = "https://github.com/brechtsanders/winlibs_mingw/releases/download/13.2.0-16.0.6-11.0.0-msvcrt-r1/winlibs-x86_64-posix-seh-gcc-13.2.0-mingw-w64msvcrt-11.0.0-r1.zip"
$compilerZip = "$tempDir\mingw.zip"
$compilerDir = "$tempDir\mingw64"

# Create temp directory
if (!(Test-Path $tempDir)) {
    New-Item -ItemType Directory -Path $tempDir | Out-Null
}

# Check if DLL already exists
if (Test-Path "TestMessageBox.dll") {
    Write-Host "TestMessageBox.dll already exists!" -ForegroundColor Green
    Write-Host ""
    $response = Read-Host "Rebuild it? (y/n)"
    if ($response -ne 'y') {
        Write-Host "Using existing DLL." -ForegroundColor Yellow
        exit 0
    }
    Remove-Item "TestMessageBox.dll" -Force
}

# Check if compiler is already downloaded
if (!(Test-Path "$compilerDir\bin\gcc.exe")) {
    Write-Host "Downloading portable MinGW compiler..." -ForegroundColor Yellow
    Write-Host "This is a one-time download (~60MB)..." -ForegroundColor Gray
    Write-Host ""

    try {
        # Download with progress
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $compilerUrl -OutFile $compilerZip -UseBasicParsing
        $ProgressPreference = 'Continue'

        Write-Host "Download complete! Extracting..." -ForegroundColor Green

        # Extract
        Expand-Archive -Path $compilerZip -DestinationPath $tempDir -Force

        Write-Host "Compiler ready!" -ForegroundColor Green
    }
    catch {
        Write-Host "ERROR: Failed to download compiler" -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red
        Write-Host ""
        Write-Host "Please try manual installation from SETUP_COMPILER.md" -ForegroundColor Yellow
        pause
        exit 1
    }
} else {
    Write-Host "Using cached compiler from previous download..." -ForegroundColor Green
}

Write-Host ""
Write-Host "Compiling TestMessageBox.dll..." -ForegroundColor Yellow
Write-Host ""

# Set up environment
$env:PATH = "$compilerDir\bin;$env:PATH"

# Compile
$gcc = "$compilerDir\bin\gcc.exe"
$output = & $gcc -shared -m64 -o TestMessageBox.dll TestMessageBox.c -luser32 2>&1

if ($LASTEXITCODE -eq 0 -and (Test-Path "TestMessageBox.dll")) {
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Green
    Write-Host "SUCCESS! DLL Built Successfully" -ForegroundColor Green
    Write-Host "================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Output: TestMessageBox.dll" -ForegroundColor White
    Write-Host "Location: $(Get-Location)\TestMessageBox.dll" -ForegroundColor White
    Write-Host ""
    Write-Host "File size: $((Get-Item TestMessageBox.dll).Length) bytes" -ForegroundColor Gray
    Write-Host ""
    Write-Host "You can now use this DLL with ARI-S injector!" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Quick test steps:" -ForegroundColor Yellow
    Write-Host "  1. Launch notepad.exe" -ForegroundColor Gray
    Write-Host "  2. Run ARI-S as Administrator" -ForegroundColor Gray
    Write-Host "  3. Navigate to DLL Injector pane" -ForegroundColor Gray
    Write-Host "  4. Browse and select this DLL" -ForegroundColor Gray
    Write-Host "  5. Select notepad.exe from process list" -ForegroundColor Gray
    Write-Host "  6. Click 'Inject DLL'" -ForegroundColor Gray
    Write-Host "  7. You should see a message box appear!" -ForegroundColor Gray
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "ERROR: Compilation failed" -ForegroundColor Red
    Write-Host "================================================" -ForegroundColor Red
    Write-Host ""
    Write-Host "Compiler output:" -ForegroundColor Yellow
    Write-Host $output -ForegroundColor Gray
    Write-Host ""
    pause
    exit 1
}

Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
