package main

import (
	"launcher/manager"
)

/* import (
	"embed"
	"fmt"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	bridge := InitBridge()
	log := logger.NewDefaultLogger()

	opts := &options.App{
		Title:            "Genecraft Launcher",
		Width:            1024,
		Height:           576,
		Assets:           assets,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        bridge.startup,
		Bind: []interface{}{
			bridge,
		},
	}

	err := wails.Run(opts)

	if err != nil {
		log.Error(fmt.Sprintf("failed to initialize application: %s", err))
	}
} */

func main() {
	manager.CreateProfile("/home/martin/.genecraft/launcher/", "none")

}
