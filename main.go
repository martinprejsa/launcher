package main

import (
	"fmt"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"launcher/bridge"
)

import (
	"embed"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	b := bridge.InitBridge()
	defaultLogger := logger.DefaultLogger{}

	opts := &options.App{
		Title:            "Genecraft Launcher",
		Width:            1024,
		Height:           576,
		MaxWidth:         1024,
		MaxHeight:        576,
		MinWidth:         1024,
		MinHeight:        576,
		Assets:           assets,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        b.Startup,
		Bind: []interface{}{
			b,
		},
	}

	err := wails.Run(opts)

	if err != nil {
		defaultLogger.Error(fmt.Sprintf("failed to initialize application: %s", err))
	}
}
