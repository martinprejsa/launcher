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

type ProgressBar struct {
	max     int
	current float64
	piece   float64
}

func (p *ProgressBar) SetTaskAmount(amount int) {
	p.max = amount
	p.piece = float64(100) / float64(amount)
}

func (p *ProgressBar) TaskFinished() {
	if p.current+p.piece > float64(p.max) {
		p.current = float64(p.max)
	} else {
		p.current += p.piece
	}
}
