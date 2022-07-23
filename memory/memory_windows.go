//go:build windows

package memory

func GetMemoryTotal() int {
	/*var mod = syscall.NewLazyDLL("kernel32.dll")
	var proc = mod.NewProc("GetPhysicallyInstalledSystemMemory")
	var mem uint64

	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&mem)))
	return mem*/
	return 160000
}
