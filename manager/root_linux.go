package manager

import (
	"os"
	"path/filepath"
)

func GetLauncherRoot() string {
	path, _ := os.UserHomeDir() // By writing it like this, i place my faith in user to not run this on a system without the HOME variable
	return filepath.Join(path, ".genecraft", "launcher")
}
