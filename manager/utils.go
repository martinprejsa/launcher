package manager

import (
	"os"
	"strings"
)

func findJAR(dir string) string {
	dfs, _ := os.ReadDir(dir)
	for _, item := range dfs {
		if strings.HasSuffix(item.Name(), ".jar") {
			return item.Name()
		}
	}
	return ""
}
