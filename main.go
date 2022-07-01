package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the bridge structure
	bridge := InitBridge()

	// Create application with options

	app := &options.App{
		Title:            "launcher",
		Width:            1024,
		Height:           768,
		Assets:           assets,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Bind: []interface{}{
			bridge,
		},
	}

	err := wails.Run(app)

	if err != nil {
		println("Error:", err)
	}
}
