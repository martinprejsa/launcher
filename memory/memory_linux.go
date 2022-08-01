//go:build linux

package memory

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func GetMemoryTotal() int {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return -1
	}
	defer file.Close()
	in := bufio.NewReader(file)
	line, err := in.ReadString('\n')
	if err != nil {
		return -1
	}
	res, err := strconv.Atoi(strings.Fields(line)[1])
	if err != nil {
		return -1
	}
	return res
}
