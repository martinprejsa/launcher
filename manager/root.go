package manager

import "path/filepath"

func GetLibraryPath() string {
	return filepath.Join(GetLauncherRoot(), "libraries")
}

func GetAssetsPath() string {
	return filepath.Join(GetLauncherRoot(), "assets")
}
