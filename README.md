<div align="center">

# ARI-S

### Asset Reconfiguration and Integration System

*A multi-tool application for Unreal Engine game modding.*

[![GitHub](https://img.shields.io/badge/GitHub-JaceTheGrayOne/ARI--S-181717?style=for-the-badge&logo=github)](https://github.com/JaceTheGrayOne/ARI-S)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)
[![Release](https://img.shields.io/github/v/release/JaceTheGrayOne/ARI-S?style=for-the-badge)](https://github.com/JaceTheGrayOne/ARI-S/releases)
[![Last Commit](https://img.shields.io/github/last-commit/JaceTheGrayOne/ARI-S?style=for-the-badge)](https://github.com/JaceTheGrayOne/ARI-S/commits/main)

[Features](#-features) • [Installation](#-installation) • [Usage](#-usage) • [Documentation](#-documentation)
</div>

## Overview

**ARI-S** is a multi-tool application I built to make using modding tools like Retoc and UAssetAPI simpler and easier for everyone.<br>
I plan to add more third-party and custom tools to ARI.S over time.

<br>

## Features
### Package Manager
- **Pack (Legacy → Zen)**: Convert edited legacy assets (.uasset/.uexp) into IoStore format (.utoc/.ucas/.pak)
- **Unpack (Zen → Legacy)**: Extract game assets from IoStore packages to legacy layout for inspection and editing
- **Auto-renaming**: Automatically applies UE mod naming convention (`z_modname_0001_p.*`)
- **Mod Naming Guide**: Inline, in-app Markdown guide accessible from the Pack panel
- **Version Support**: Auto-detection and support for Unreal Engine 4.27 through 5.5+

### UAsset Manager
- **Export to JSON**: Convert .uasset/.uexp files to JSON format for editing
- **Import from JSON**: Convert JSON files back to .uasset/.uexp format
- **Batch Processing**: Process entire directories recursively

### DLL Injector
- **Process Enumeration**: Browse running processes
- **Filtering/Sorting**: Text filter with configurable default sort (Name/PID) and option to hide common system processes
- **Native Injection**: CreateRemoteThread-based DLL injection
- **UAC Integration**: Automatic privilege elevation when needed

### UWP Dumper
- **UWP Package Extraction**: Dump encrypted files from Universal Windows Platform applications
- **Game Pass Support**: Extract assets from Xbox Game Pass and Windows Store games
- **Interactive Interface**: Launches UWPInjector tool in separate console window
- **Automatic Output**: Dumps to `%LOCALAPPDATA%\Packages\<PFN>\TempState\DUMP`

<br>

### Settings
- **Accessibility**: Reduce motion/animations
- **Console Log**: Clear on launch, maximum lines to retain
- **Directories**: Default Mods/Exports/Imports directories, default mappings (.usmap) path, remember last used paths
- **Process Injection**: Hide system processes, default sort (name/PID)
- **Misc**: Backup/restore settings, reset to defaults, Save

<br>

## Tech Stack
<table>
<tr>
<td width="33%" valign="top">

### Backend
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/dl/)
[![Wails](https://img.shields.io/badge/Wails_v3-CA4245?style=for-the-badge&logo=go&logoColor=white)](https://wails.io)

</td>
<td width="33%" valign="top">

### Frontend
[![JavaScript](https://img.shields.io/badge/JavaScript-F7DF1E?style=for-the-badge&logo=javascript&logoColor=black)](https://developer.mozilla.org/docs/Web/JavaScript)
[![HTML5](https://img.shields.io/badge/HTML5-E34F26?style=for-the-badge&logo=html5&logoColor=white)](https://developer.mozilla.org/docs/Web/Guide/HTML/HTML5)
[![CSS3](https://img.shields.io/badge/CSS3-1572B6?style=for-the-badge&logo=css3&logoColor=white)](https://developer.mozilla.org/docs/Web/CSS)

</td>
<td width="33%" valign="top">

### External Tools
[![Retoc](https://img.shields.io/badge/Retoc-8A2BE2?style=for-the-badge&logoColor=white)](https://github.com/WorkingRobot/Retoc)
[![UAssetAPI](https://img.shields.io/badge/UAssetAPI-2E8B57?style=for-the-badge&logoColor=white)](https://github.com/atenfyr/UAssetAPI)
[![UWPDumper](https://img.shields.io/badge/UWPDumper-FF6B6B?style=for-the-badge&logoColor=white)](https://github.com/Wunkolo/UWPDumper)

</td>
</tr>
</table>

**UAssetBridge**: Custom .NET 9 executable wrapping UAssetAPI for seamless integration

<br>

## Requirements

| Component | Version | Purpose |
|-----------|---------|---------|
| **Windows** | 10/11 (64-bit) | Operating System |
| **Go** | 1.24+ | Building from source |
| **Node.js** | 18+ | Frontend build |
| **Wails** | v3 (alpha.36+) | Application framework |

> **Note**: Pre-built binaries do not require Go, Node.js, or Wails.

<br>

## Installation

### Option 1: Download Pre-built Binary
<details><summary>Download instructions</summary>

```bash
1. Download the latest release
2. Extract ARI-S.exe to your preferred location
3. Run ARI-S.exe
```
**First Run**: The application will automatically extract dependencies to `%APPDATA%/ARI-S/dependencies` (Windows Roaming profile), not next to the executable.
</details>

<br>

### Option 2: Build from Source
<details>
<summary>Build instructions</summary>

#### 1. Install Prerequisites

```bash
# Install Go 1.24+
Download from: https://golang.org/dl/

# Install Node.js 18+
Download from: https://nodejs.org/

# Install Wails CLI
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

#### 2. Clone the Repository

```bash
# Clone the repo
git clone https://github.com/JaceTheGrayOne/ARI-S.git

# Change to the repo directory
cd ARI-S
```

#### 3. Install Frontend Dependencies

```bash
# Change to the frontend directory
cd frontend

# Use the npm package manager to install the front end dependencies
npm install

# Go back to the previous directory
cd ..
```

#### 4. Build the Application

```bash
# Development build with hot-reload
wails3 dev

# Production build
wails3 build
```

The built executable will be located at: `bin/ARI-S.exe`

</details>

<br>

## Usage
### First Time Setup
1. **Launch ARI-S** - Double-click `ARI-S.exe`
2. **Automatic Extraction** - Dependencies are extracted on first run to `%APPDATA%\ARI-S\dependencies`
3. **Configuration** - Settings are stored in `%APPDATA%\ARI-S\config.json`
4. **Auto-Save** - All paths and preferences are automatically saved

<br>

### Package Operations
<details>
<summary><b>Packing Mods (Legacy → Zen)</b></summary>

1. Select **"Pack / Unpack"** from the sidebar
2. Click **"Browse"** for input folder (your mod assets)
3. Click **"Browse"** for output directory
4. Enter **mod name** and **load order** (serialization number)
5. Select **Unreal Engine version**
6. Click **"Pack to Zen"**

**Output Files**:
```
z_modname_0001_p.utoc
z_modname_0001_p.ucas
z_modname_0001_p.pak
```
</details>

<details>
<summary><b>Unpacking Game Files (Zen → Legacy)</b></summary>

1. Select **"Pack / Unpack"** from the sidebar
2. Click **"Browse"** for game paks folder
3. Click **"Browse"** for extract output directory
4. Select **Unreal Engine version**
5. Click **"Unpack to Legacy"**
</details>

<br>

### UAsset Operations
<details>
<summary><b>Exporting to JSON</b></summary>

1. Select **"UAsset Manager"** from the sidebar
2. Click **"Browse"** to select folder containing `.uasset`/`.uexp` files
3. *(Optional)* Select a `.usmap` mappings file for better accuracy
4. File count updates automatically
5. Click **"Export to JSON"**

</details>

<details>
<summary><b>Importing from JSON</b></summary>

1. Select **"UAsset Manager"** from the sidebar
2. Click **"Browse"** to select folder containing `.json` files
3. *(Optional)* Select a `.usmap` mappings file
4. File count updates automatically
5. Click **"Import from UAsset"**
</details>

<br>

### DLL Injection
<details>
<summary><b>Injection Procedure</b></summary>

1. Select **"Injector"** from the sidebar
2. Click **"Browse Processes"** to see running processes
3. Select the **target process** from the list
4. Click **"Browse DLL"** to select your DLL file
5. Click **"Inject DLL"**
6. Allow **UAC elevation** if prompted

<br>

> **⚠️ Important Notes**:
> - DLL injection requires **administrator privileges**
> - ARI-S will automatically re-launch with elevation if needed
> - The DLL and target process must be the **same architecture** (64-bit | 32-bit)
</details>

<br>

### UWP Dumper
<details>
<summary><b>Setup and Usage</b></summary>

**First Time Setup**:
1. Download UWPDumper from [GitHub Releases](https://github.com/Wunkolo/UWPDumper/releases)
2. Extract `UWPInjector.exe` and `UWPDumper.dll` from the archive
3. Place both files in: `%APPDATA%/ARI-S/dependencies/uwpdumper/`
4. Restart ARI-S (or launch for first time)

**Usage**:
1. **Launch your target UWP app** (e.g., a Game Pass game)
2. Select **"UWPDumper"** from the sidebar
3. Verify tool status shows **"Ready"**
4. Click **"Launch UWPDumper"**
5. A console window will open prompting for the **Process ID (PID)**
6. Find your app's PID in Task Manager (Details tab, enable PID column)
7. Enter the PID in the console window and press Enter
8. Files will be extracted automatically

**Output Location**:
```
%LOCALAPPDATA%\Packages\<PackageFamilyName>\TempState\DUMP\
```

> **📝 Notes**:
> - UWPDumper binaries are **not** included with ARI-S and must be downloaded separately
> - The tool works with all UWP apps (Game Pass, Windows Store, etc.)
> - Administrator privileges may be required for some applications
> - Package Family Name (PFN) varies by application
</details>

<br>

## File Structure
```
ARI-S/
├── ARI-S.exe                    # Main Executable
├── dependencies/                # Auto-Extracted
│   ├── version.txt
│   ├── retoc/
│   │   ├── retoc.exe           # IoStore Package Manager
│   │   └── oo2core_9_win64.dll # Oodle Compression Library
│   ├── UAssetAPI/
│   │   ├── UAssetBridge.exe    # UAssetAPI Language Agnostic Bridge
│   │   └── [.NET runtime DLLs]
│   └── uwpdumper/
│       ├── UWPInjector.exe     # UWP Dumper Tool (User-provided)
│       ├── UWPDumper.dll       # UWP Dumper DLL (User-provided)
│       └── README.md           # Setup Instructions
└── [User configuration]
    └── %LOCALAPPDATA%\ARI-S\
        └── config.json         # Saved Paths and Preferences
```

<br>

### Configuration Location
```
C:\Users\<Username>\AppData\Roaming\ARI-S\config.json
```
**Configuration stores**:
- Last used paths (mods, exports, imports, mappings, DLL, etc.)
- Preferred Unreal Engine version
- Accessibility and console preferences (reduce motion, clear on launch, max lines)
- Directory defaults (mods/exports/imports, mappings path) and path memory preference
- Process list preferences (hide system processes, default sort)

<br>

## Troubleshooting
### Common Issues
<details>
<summary><b>"Dependencies not found" error</b></summary>

**Solution**: Ensure the `dependencies` folder exists at `%APPDATA%\ARI-S\dependencies`. Delete that folder and restart ARI-S to re-extract.
</details>

<details>
<summary><b>"Access denied" errors during DLL injection</b></summary>

**Solution**: Right-click `ARI-S.exe` and select **"Run as administrator"**
</details>

<details>
<summary><b>DLL injection fails with LoadLibraryW error</b></summary>

**Possible causes**:
- Architecture mismatch (32-bit DLL vs 64-bit process or vice versa)
- DLL has missing dependencies
- DLL path contains special characters
</details>

<details>
<summary><b>UAsset export/import fails</b></summary>

**Solution**: For best results, use a `.usmap` mappings file for your game version
</details>

<details>
<summary><b>Retoc operations fail</b></summary>

**Solution**: Ensure you've selected the correct Unreal Engine version matching your game
</details>

<details>
<summary><b>Configuration not saving</b></summary>

**Solution**: Check that `%APPDATA%\ARI-S\` is writable
</details>

### Debug Logs
Application logs are written to:
- **Console output** (when running from terminal)
- **Windows Event Viewer** (for critical errors)

To view detailed logs, run ARI-S from a command prompt:

```cmd
ARI-S.exe
```

<br>

## Development
### Project Structure
```
ARI-S/
├── main.go                 # Application entry point
├── app.go                  # Core app service
├── config.go               # Configuration management
├── deps.go                 # Dependency extraction
├── retoc.go                # Retoc service (pak operations)
├── uasset.go               # UAsset service (export/import)
├── injector.go             # DLL injection service
├── go.mod                  # Go dependencies
├── frontend/               # Web-based UI
│   ├── index.html         # Main HTML
│   ├── src/
│   │   ├── main.js        # Frontend logic
│   │   └── style.css      # Styling
│   ├── package.json       # Node dependencies
│   └── dist/              # Build output
└── dependencies/           # Embedded resources
    ├── retoc/
    └── UAssetAPI/
```

<br>

### Running Tests
```bash
go test ./...
```

<br>

### Development Mode
```bash
cd frontend
npm install
cd ..
wails3 dev
```

This starts the app with **hot-reload** enabled for frontend changes.

<br>

### Architecture Notes
All external dependencies are embedded in the executable at compile-time using Go's `embed` package. On first run, these are extracted to `%APPDATA%/ARI-S/dependencies`. The application checks `dependencies/version.txt` and automatically re-extracts if the version changes, ensuring seamless updates.

<br>

## Contributing
This is an active development project. Contributions are welcome!


### Guidelines
- Follow Go style conventions (`gofmt`, clear naming)
- Test all changes with `go test`
- Update documentation for new features
- Keep commits focused and descriptive


### How to Contribute
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request


## Credits

| Tool | Description | Link |
|------|-------------|------|
| **Retoc** | IoStore package conversion tool | [GitHub](https://github.com/WorkingRobot/Retoc) |
| **UAssetAPI** | .NET library for UAsset serialization | [GitHub](https://github.com/atenfyr/UAssetAPI) |
| **Wails** | Desktop application framework | [Website](https://wails.io) |
| **Oodle** | Compression library | oo2core_9_win64.dll |

<br>

## License
This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.
This project is for **educational and modding purposes**.

<br>

## Support
For issues, questions, or feature requests:
- Check existing documentation in `Resources/Documentation/`
- Review source code comments
- Consult related tool documentation (Retoc, UAssetAPI)
- [Open an issue](https://github.com/JaceTheGrayOne/ARI-S/issues)

<br>

<div align="center">

</div>
