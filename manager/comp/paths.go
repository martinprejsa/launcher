package comp

import "path/filepath"

func GetLibraryPath() string {
	return filepath.Join(GetLauncherRoot(), "libraries")
}

func GetAssetsPath() string {
	return filepath.Join(GetLauncherRoot(), "assets")
}

func GetLogCfgsPath() string {
	return filepath.Join(GetAssetsPath(), "log_cfgs")
}

func GetIndexesPath() string {
	return filepath.Join(GetAssetsPath(), "indexes")
}
