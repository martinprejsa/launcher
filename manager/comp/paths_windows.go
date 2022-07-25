//go:build windows

package comp

import (
	"os"
	"path/filepath"
)

func GetLauncherRoot() string {
	return filepath.Join(os.Getenv("APPDATA"), ".genecraft")
}
