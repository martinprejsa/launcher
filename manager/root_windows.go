//go:build windows

package manager

import (
	"os"
	"path/filepath"
)

func GetLauncherRoot() string {
	return filepath.Join(os.Getenv("APPDATA"), ".genecraft")
}

func GetLibraryPath() string {
	return filepath.Join(GetLauncherRoot(), "libs")
}

func GetAssetsPath() string {
	return filepath.Join(GetLauncherRoot(), "assets")
}
