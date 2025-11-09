package main

import (
    "embed"
    "log"
    "os"
    "path/filepath"

    "github.com/JaceTheGrayOne/ARI-S/internal/app"
    "github.com/JaceTheGrayOne/ARI-S/internal/injector"
    "github.com/JaceTheGrayOne/ARI-S/internal/retoc"
    "github.com/JaceTheGrayOne/ARI-S/internal/uasset"
    "github.com/JaceTheGrayOne/ARI-S/internal/uwpdumper"
    "github.com/wailsapp/wails/v3/pkg/application"
)

// assets holds the embedded frontend build output from frontend/dist.
// This directory contains the compiled HTML, CSS, and JavaScript files
// that are served by the Wails application to render the UI.
//
//go:embed frontend/dist
var assets embed.FS

// depsFS holds the embedded external dependencies extracted on first run.
// This includes retoc.exe, oo2core_9_win64.dll, UAssetBridge.exe,
// the .NET runtime DLLs required for UAsset operations, Dumper7.dll,
// and UnrealMappingsDumper.dll for SDK and mappings dumping.
//
//go:embed dependencies
var depsFS embed.FS

// Embed the managed entry assembly produced by the bridge publish output.
// This guards against environments where the embedded dependencies folder
// might be missing UAssetBridge.dll (the managed assembly), even though the
// apphost UAssetBridge.exe is present.
//go:embed build/UAssetBridge.dll
var uassetBridgeManagedDLL []byte

func main() {
	// Extract native libraries first (if CGO is enabled)
	if err := initNativeLibraries(); err != nil {
		log.Fatalf("Fatal error during native library setup: %v", err)
	}

    extractedDepsDir, err := app.EnsureDependencies(depsFS)
    if err != nil {
        log.Fatalf("Fatal error during dependency setup: %v", err)
    }
    log.Printf("Dependencies directory: %s", extractedDepsDir)

    // Ensure UAssetBridge.dll exists next to UAssetBridge.exe. Some builds may
    // only include the apphost EXE; .NET still expects the managed DLL listed
    // in the deps.json. If it's missing, provide the embedded copy from build/.
    {
        dllPath := filepath.Join(extractedDepsDir, "UAssetAPI", "UAssetBridge.dll")
        if _, statErr := os.Stat(dllPath); os.IsNotExist(statErr) {
            if len(uassetBridgeManagedDLL) == 0 {
                log.Printf("Warning: Embedded UAssetBridge.dll not available; cannot repair missing managed assembly")
            } else {
                if writeErr := os.WriteFile(dllPath, uassetBridgeManagedDLL, 0644); writeErr != nil {
                    log.Printf("Warning: Failed to write UAssetBridge.dll: %v", writeErr)
                } else {
                    log.Printf("Restored missing UAssetBridge.dll to: %s", dllPath)
                }
            }
        }
    }

	// Create a new App instance
	appInstance := app.NewApp()

	// Set the dependencies directory
	appInstance.SetDepsDir(extractedDepsDir)

	// Load configuration before starting Wails
	log.Println("Loading application configuration...")
	if err := appInstance.LoadConfiguration(); err != nil {
		log.Printf("Warning: Failed to load configuration: %v", err)
	}

	// Create a new Wails application by providing the necessary options.
	wailsApp := application.New(application.Options{
		Name:        "ARI.S",
		Description: "Asset Reconfiguration and Integration System",
		Services: []application.Service{
			application.NewService(appInstance),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Windows: application.WindowsOptions{
			// Store WebView2 data in ARI-S folder instead of ARI-S.exe folder
			WebviewUserDataPath: filepath.Join(extractedDepsDir, "..", "webview"),
		},
	})

	// Register services that need access to the Wails app instance
	retocService := retoc.NewRetocService(appInstance, extractedDepsDir)
	uassetService := uasset.NewUAssetService(appInstance, extractedDepsDir)
	injectorService := injector.NewInjectorService(appInstance)
	uwpDumperService := uwpdumper.NewUWPDumperService(appInstance, extractedDepsDir)
	wailsApp.RegisterService(application.NewService(retocService))
	wailsApp.RegisterService(application.NewService(uassetService))
	wailsApp.RegisterService(application.NewService(injectorService))
	wailsApp.RegisterService(application.NewService(uwpDumperService))

	// Create a new window with the necessary options.
	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "ARI.S",
		Width:     1080,
		Height:    900,
		MinWidth:  1050,
		MinHeight: 720,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Run the application. This blocks until the application has been exited.
	err = wailsApp.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
