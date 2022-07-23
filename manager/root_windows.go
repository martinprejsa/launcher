//go:build windows

package manager

import (
	"os"
	"path/filepath"
)

func GetLauncherRoot() string {
	return filepath.Join(os.Getenv("APPDATA"), ".genecraft", "launcher")
}
