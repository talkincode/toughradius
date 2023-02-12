// +build windows

package local

import (
	"log"
	"syscall"
	"unsafe"
)

func getFSSize(root string) (uint64, uint64, error) {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	var free uint64
	var total uint64
	var totalFree uint64

	_, _, err := c.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(root))),
		uintptr(unsafe.Pointer(&free)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if err != nil {
		log.Println(err)
	}

	return total - free, free, nil
}
