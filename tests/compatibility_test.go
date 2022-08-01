package tests

import (
	"launcher/manager/comp"
	"launcher/memory"
	"testing"
)

func TestCompatibility(t *testing.T) {
	size := memory.GetMemoryTotal()
	if size == -1 {
		t.Fatal("Memory size check failed")
		return
	}

	str := comp.GetLauncherRoot()
	if str == "" {
		t.Fatal("Root path resolution failed")
		return
	}
}
