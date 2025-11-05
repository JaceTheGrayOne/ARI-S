# UWPDumper Tool

This directory should contain the UWPDumper tool binaries.

## Required Files

Place the following files in this directory:

- `UWPInjector.exe` - The main UWP dumper executable
- `UWPDumper.dll` - Required DLL for injection

## Download Location

Download the latest release from the official UWPDumper repository:
https://github.com/Wunkolo/UWPDumper/releases

1. Download `UWPDumper.zip` from the latest release
2. Extract the contents
3. Copy `UWPInjector.exe` and `UWPDumper.dll` to this directory

## Expected Directory Structure

```
dependencies/uwpdumper/
├── README.md (this file)
├── UWPInjector.exe
└── UWPDumper.dll
```

## Usage

Once the files are in place:
1. Build/run ARIS
2. Navigate to the UWP Dumper section
3. Click "Launch UWPDumper"
4. The tool will open in a new console window
5. Follow the interactive prompts to dump your UWP application

## Output Location

Dumped files will be saved to:
```
%LOCALAPPDATA%\Packages\<PackageFamilyName>\TempState\DUMP
```

Where `<PackageFamilyName>` is specific to your UWP application (e.g., `Microsoft.WindowsStore_8wekyb3d8bbwe`).
