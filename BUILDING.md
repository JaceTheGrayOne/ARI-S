# Building ARI-S from Source

This guide covers building ARI-S from source code, including setting up dependencies and understanding the build process.

## Prerequisites

### Required Software

1. **Go 1.21 or later**
   - Download: https://go.dev/dl/
   - Verify: `go version`

2. **Node.js 18+ and npm**
   - Download: https://nodejs.org/
   - Verify: `node --version` and `npm --version`

3. **Wails v3 CLI**
   ```bash
   go install github.com/wailsapp/wails/v3/cmd/wails3@latest
   ```
   - Verify: `wails3 version`

4. **Git**
   - Download: https://git-scm.com/
   - Verify: `git --version`

### Platform-Specific Requirements

**Windows:**
- Windows 10/11 (64-bit)
- WebView2 Runtime (usually pre-installed on Windows 11)
- No additional C compiler needed for standard builds

**Linux (future):**
- gcc
- gtk3
- webkit2gtk

**macOS (future):**
- Xcode Command Line Tools

## Quick Build

```bash
# Clone the repository
git clone https://github.com/YourOrg/ARI-S.git
cd ARI-S

# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..

# Build the application
wails3 build

# Output will be in ./bin/ARI-S.exe
```

## Development Build

For development with hot-reload:

```bash
# Terminal 1: Start frontend dev server
cd frontend
npm run dev

# Terminal 2: Run Wails in dev mode
wails3 dev
```

This provides:
- Hot reload for frontend changes
- Live Go code recompilation
- Debug console in the app

## Understanding the Build Process

### Dependency Embedding

ARI-S uses Go's `embed` package to bundle all binary dependencies directly into the executable. This is why the output binary is ~106 MB instead of just a few MB.

**What gets embedded:**

1. **Frontend Assets** (~10-15 MB)
   - Built JavaScript/CSS from `frontend/dist/`
   - Embedded via `//go:embed all:frontend/dist` in `main.go`

2. **Binary Dependencies** (~88 MB)
   - Retoc: `dependencies/retoc/` (2 files)
   - UAssetAPI: `dependencies/UAssetAPI/` (194 files)
   - Embedded via `//go:embed all:dependencies` in `main.go`

### Build Stages

When you run `wails3 build`, the following happens:

1. **Frontend Build** (`frontend/` → `frontend/dist/`)
   ```bash
   npm run build
   ```
   - Compiles JavaScript/Vue components
   - Optimizes and bundles assets
   - Output: `frontend/dist/`

2. **Go Embedding** (`deps.go`, `main.go`)
   - Go compiler reads `//go:embed` directives
   - Embeds `frontend/dist/` into `assets` variable
   - Embeds `dependencies/` into `depsFS` variable
   - These become part of the compiled binary

3. **Binary Compilation** (`*.go` → `bin/ARI-S.exe`)
   ```bash
   go build -o bin/ARI-S.exe
   ```
   - Compiles all Go source files
   - Links embedded assets
   - Produces single executable

4. **Runtime Extraction** (first launch)
   - On first run, `deps.go::ensureDependencies()` executes
   - Extracts embedded dependencies to `dependencies/` folder
   - Creates `version.txt` for version tracking
   - Subsequent runs skip extraction if version matches

## Updating Dependencies

### Updating Retoc

1. Download/build new `retoc.exe` and `oo2core_9_win64.dll`
2. Replace files in `dependencies/retoc/`
3. Increment version in `dependencies/version.txt`
4. Rebuild: `wails3 build`

### Updating UAssetAPI

1. Rebuild UAssetBridge as self-contained .NET application:
   ```bash
   cd UAssetBridge/UAssetBridge
   dotnet publish -c Release -r win-x64 --self-contained true -p:PublishSingleFile=false
   ```

2. Copy output from `bin/Release/net9.0/win-x64/` to `dependencies/UAssetAPI/`
3. Increment version in `dependencies/version.txt`
4. Rebuild: `wails3 build`

### Version Management

The `dependencies/version.txt` file controls when dependencies are re-extracted:

- **Same version**: Skip extraction (fast startup)
- **Different version**: Delete old files, extract new ones
- **Missing file**: Full extraction

**To force re-extraction for all users:**
1. Update dependency files
2. Bump version: `1.0.0` → `1.0.1`
3. Rebuild and release

## Build Configuration

### Wails Configuration

The build is configured via `Taskfile.yml` and Wails metadata:

Key settings:
- **Target**: Windows AMD64
- **Binary name**: `ARI-S.exe`
- **Output directory**: `bin/`
- **Build flags**: `-buildvcs=false -gcflags=all="-l"`

### Custom Build Flags

You can customize the build:

```bash
# Development build (faster, larger, includes debug symbols)
wails3 task build

# Production build (slower, optimized, no debug symbols)
wails3 task build -ldflags "-s -w"

# Cross-compile (from the Taskfile)
wails3 task windows:build
```

## Project Structure for Building

```
ARI-S/
├── main.go                    # Entry point, embed directives
├── deps.go                    # Extraction logic
├── *.go                       # Application services
├── go.mod                     # Go dependencies
├── Taskfile.yml              # Build configuration
├── dependencies/              # TO BE EMBEDDED
│   ├── retoc/                # Retoc binary + DLL
│   ├── UAssetAPI/            # UAssetBridge + runtime
│   └── version.txt           # Dependency version
├── frontend/                  # Frontend source
│   ├── src/                  # Vue/JS source
│   ├── dist/                 # Built assets (TO BE EMBEDDED)
│   ├── package.json          # Frontend dependencies
│   └── vite.config.js        # Build config
└── bin/                      # Build output (gitignored)
    └── ARI-S.exe             # Final executable (~106 MB)
```

## Common Build Issues

### Issue: `embed` directive not found

**Symptom:**
```
pattern dependencies: no matching files found
```

**Solution:**
Ensure the `dependencies/` folder exists and contains files:
```bash
ls dependencies/retoc/
ls dependencies/UAssetAPI/
```

### Issue: Frontend build fails

**Symptom:**
```
npm ERR! Missing script: "build"
```

**Solution:**
```bash
cd frontend
npm install
npm run build
cd ..
```

### Issue: Go module errors

**Symptom:**
```
missing go.sum entry
```

**Solution:**
```bash
go mod tidy
go mod download
```

### Issue: Wails command not found

**Symptom:**
```
wails3: command not found
```

**Solution:**
```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
# Ensure $GOPATH/bin is in your PATH
```

### Issue: Build succeeds but app crashes on startup

**Symptom:**
App launches but immediately crashes

**Possible causes:**
1. Corrupted embedded dependencies
2. Missing files in `dependencies/` folder
3. Incorrect file permissions (Linux/macOS)

**Solution:**
```bash
# Clean build
rm -rf bin/ frontend/dist/
wails3 build

# Verify dependencies are complete
wails3 task build > build.log 2>&1
grep -i "embed" build.log
```

## Testing the Build

### Basic Test

```bash
# Build
wails3 build

# Run
cd bin
./ARI-S.exe

# Check console output
# Should see: "Dependencies directory: ..."
```

### Dependency Extraction Test

```bash
# Clean any existing dependencies
rm -rf bin/dependencies

# Run app
cd bin
./ARI-S.exe

# Verify extraction
ls dependencies/
ls dependencies/retoc/
ls dependencies/UAssetAPI/
cat dependencies/version.txt
```

Should output:
```
dependencies/retoc/retoc.exe
dependencies/retoc/oo2core_9_win64.dll
dependencies/UAssetAPI/UAssetBridge.exe
dependencies/UAssetAPI/... (194 total files)
dependencies/version.txt (contains: 1.0.0)
```

### Version Upgrade Test

```bash
# Build with version 1.0.0
wails3 build
cd bin
./ARI-S.exe  # Extracts dependencies

# Update version in source
echo "1.0.1" > ../dependencies/version.txt

# Rebuild
cd ..
wails3 build
cd bin
./ARI-S.exe  # Should re-extract with new version
cat dependencies/version.txt  # Should show 1.0.1
```

## Advanced Build Topics

### Reducing Binary Size

The embedded dependencies make the binary large (~106 MB). This is intentional for zero-dependency distribution.

If size is critical:
1. Use external dependency folder instead of embedding
2. Remove debug symbols: `-ldflags "-s -w"`
3. Use UPX compression (may trigger antivirus)

### Code Signing (Windows)

For distribution without SmartScreen warnings:

```bash
# After building
signtool sign /f certificate.pfx /p password /tr http://timestamp.digicert.com /td sha256 /fd sha256 bin/ARI-S.exe
```

### CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Build ARI-S

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v3/cmd/wails3@latest

      - name: Build
        run: wails3 build

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: ARI-S-Windows
          path: bin/ARI-S.exe
```

## Building for Other Platforms

### Linux (Future)

```bash
# Prerequisites
sudo apt-get install gcc libgtk-3-dev libwebkit2gtk-4.0-dev

# Build
GOOS=linux GOARCH=amd64 wails3 build

# Output: bin/ARI-S (ELF binary)
```

### macOS (Future)

```bash
# Prerequisites
xcode-select --install

# Build
GOOS=darwin GOARCH=amd64 wails3 build

# Output: bin/ARI-S.app (macOS application bundle)
```

## Getting Help

- **Wails Documentation**: https://wails.io/docs/
- **Go Embed Documentation**: https://pkg.go.dev/embed
- **Project Issues**: https://github.com/YourOrg/ARI-S/issues

## Contributing

When contributing code:

1. Test builds on clean environment
2. Verify dependency extraction works
3. Update `dependencies/version.txt` if changing dependencies
4. Run `go fmt` on all Go files
5. Test both dev and production builds

---

**Happy Building!** 🚀
