package main

import (
	"embed"
	"log"

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
// This includes retoc.exe, oo2core_9_win64.dll, UAssetBridge.exe, and
// the .NET runtime DLLs required for UAsset operations.
//
//go:embed dependencies
var depsFS embed.FS

func main() {
	extractedDepsDir, err := app.EnsureDependencies(depsFS)
	if err != nil {
		log.Fatalf("Fatal error during dependency setup: %v", err)
	}
	log.Printf("Dependencies directory: %s", extractedDepsDir)

	// Create a new App instance
	appInstance := app.NewApp()

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
