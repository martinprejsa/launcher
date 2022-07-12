package compatibility

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func GetMemorySize() int {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return -1
	}
	in := bufio.NewReader(file)
	line, err := in.ReadString('\n')
	if err != nil {
		return -1
	}
	res, err := strconv.Atoi(strings.Fields(line)[2])
	if err != nil {
		return -1
	}
	return res
}
