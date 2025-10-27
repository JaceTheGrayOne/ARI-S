package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts the application.
func main() {
	// Create a new App instance
	app := NewApp()

	// Create a new Wails application by providing the necessary options.
	wailsApp := application.New(application.Options{
		Name:        "ARI.S",
		Description: "Asset Reconfiguration and Integration System",
		Services: []application.Service{
			application.NewService(app),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Register services that need access to the Wails app instance
	retocService := NewRetocService(app)
	uassetService := NewUAssetService(app)
	wailsApp.RegisterService(application.NewService(retocService))
	wailsApp.RegisterService(application.NewService(uassetService))

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
	err := wailsApp.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
