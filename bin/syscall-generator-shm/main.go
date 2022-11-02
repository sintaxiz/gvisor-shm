package main

import (
	"fmt"
	"unsafe"
)

func write_to_shm_mem() {
	addr := uintptr(0x7f45221f7000)
	//*(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))) = 666
	*(*int)(unsafe.Pointer(addr)) = 666
	for !(*(*int)(unsafe.Pointer(addr)) == 0) {

	}
	fmt.Println("pid = ", *(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))))
}

func getpid() int64 {
	write_to_shm_mem()
	return 0
}

func main() {
	iterations := 3
	for i := 0; i < iterations; i++ {
		getpid()
	}
}
