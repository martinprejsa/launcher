//go:build windows

package memory

import (
	"syscall"
	"unsafe"
)

func GetMemoryTotal() int {
	var mod = syscall.NewLazyDLL("kernel32.dll")
	var proc = mod.NewProc("GetPhysicallyInstalledSystemMemory")
	var mem uint64 = 0

	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&mem)))
	if ret != 1 || mem == 0 {
		return -1
	}
	return int(mem)
}
