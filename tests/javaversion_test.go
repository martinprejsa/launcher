package tests

import (
	"fmt"
	"launcher/manager"
	"os/exec"
	"strings"
	"testing"
)

func TestJavaVersionCheck(t *testing.T) {
	cmd := exec.Command("java", "--version")
	o, err := cmd.CombinedOutput()
	if err != nil {
		t.Error("execution failed")
	}
	str := string(o)
	words := strings.Split(str, " ")
	version := words[1]
	fmt.Println(version)
	if !manager.CompareVersion("17.0.0", version, manager.PrecisionFull) {
		t.Error("invalid java version")
	}
}
