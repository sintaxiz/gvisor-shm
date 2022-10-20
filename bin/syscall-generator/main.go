package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func write_to_shm_mem() {
	addr := uintptr(0x7f45221f7000)
	//*(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))) = 666
	*(*int)(unsafe.Pointer(addr)) = syscall.Getpid()
}

func getpid() int64 {
	write_to_shm_mem()
	return 0
}

func main() {
	for {
		fmt.Println("INFO: start getpid")
		pid := getpid()
		fmt.Printf("INFO: pid = %d\n", pid)
	}
}
